package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	Logger   zerolog.Logger
	Queriers []Querier
}

func NewApp(
	logger *zerolog.Logger,
	config *Config,
) *App {
	appLogger := logger.With().Str("component", "app").Logger()

	var tendermintRPC *TendermintRPC
	var cosmovisor *Cosmovisor
	var github *Github

	if config.TendermintConfig.Address != "" {
		tendermintRPC = NewTendermintRPC(config, logger)
	}

	if config.CosmovisorConfig.IsEnabled() {
		cosmovisor = NewCosmovisor(config, logger)
	}

	if config.GithubConfig.Repository != "" {
		github = NewGithub(config, logger)
	}

	queriers := []Querier{
		NewNodeStatsQuerier(logger, tendermintRPC),
		NewCosmovisorQuerier(logger, cosmovisor),
		NewVersionsQuerier(logger, github, cosmovisor),
		NewUpgradesQuerier(logger, cosmovisor, tendermintRPC),
	}

	for _, querier := range queriers {
		if querier.Enabled() {
			appLogger.Debug().Str("name", querier.Name()).Msg("Querier is enabled")
		} else {
			appLogger.Debug().Str("name", querier.Name()).Msg("Querier is disabled")
		}
	}

	return &App{
		Logger:   appLogger,
		Queriers: queriers,
	}
}

func (a *App) HandleRequest(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	sublogger := a.Logger.With().
		Str("request-id", uuid.New().String()).
		Logger()

	registry := prometheus.NewRegistry()

	querierEnabled := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "querier_enabled",
			Help: "Is querier enabled?",
		},
		[]string{"querier"},
	)
	registry.MustRegister(querierEnabled)

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResults := map[string][]prometheus.Collector{}
	allQueries := map[string][]QueryInfo{}

	for _, querier := range a.Queriers {
		querierEnabled.
			With(prometheus.Labels{
				"querier": querier.Name(),
			}).
			Set(BoolToFloat64(querier.Enabled()))

		if !querier.Enabled() {
			continue
		}

		wg.Add(1)
		go func(querier Querier) {
			querierResults, queriesInfo := querier.Get()
			mu.Lock()
			allResults[querier.Name()] = querierResults
			allQueries[querier.Name()] = queriesInfo
			mu.Unlock()
			wg.Done()
		}(querier)
	}

	wg.Wait()

	for _, querierResults := range allResults {
		for _, result := range querierResults {
			registry.MustRegister(result)
		}
	}

	querySuccessfulGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "query_successful",
			Help: "Was query successful?",
		},
		[]string{"querier", "action"},
	)
	registry.MustRegister(querySuccessfulGauge)

	for name, queryInfos := range allQueries {
		for _, queryInfo := range queryInfos {
			querySuccessfulGauge.
				With(prometheus.Labels{
					"querier": name,
					"action":  queryInfo.Action,
				}).
				Set(BoolToFloat64(queryInfo.Success))
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
