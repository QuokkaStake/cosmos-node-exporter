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

	if config.CosmovisorConfig.Enabled && config.CosmovisorConfig.ChainFolder != "" && config.CosmovisorConfig.ChainBinaryName != "" {
		cosmovisor = NewCosmovisor(config, logger)
	}

	if config.GithubConfig.Repository != "" {
		github = NewGithub(config, logger)
	}

	queriers := []Querier{
		NewNodeStatsQuerier(logger, tendermintRPC),
		NewCosmovisorQuerier(logger, cosmovisor),
		NewVersionsQuerier(logger, github, cosmovisor),
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

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResults := map[string][]prometheus.Collector{}

	for _, querier := range a.Queriers {
		if !querier.Enabled() {
			continue
		}

		wg.Add(1)
		go func(querier Querier) {
			querierResults := querier.Get()
			mu.Lock()
			allResults[querier.Name()] = querierResults
			mu.Unlock()
			wg.Done()
		}(querier)
	}

	wg.Wait()

	registry := prometheus.NewRegistry()
	for _, querierResults := range allResults {
		for _, result := range querierResults {
			registry.MustRegister(result)
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
