package pkg

import (
	"main/pkg/config"
	"main/pkg/constants"
	cosmovisor2 "main/pkg/cosmovisor"
	github2 "main/pkg/github"
	grpc2 "main/pkg/grpc"
	"main/pkg/queriers/node_stats"
	"main/pkg/queriers/upgrades"
	"main/pkg/queriers/versions"
	"main/pkg/query_info"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
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
	Queriers []types.Querier
}

func NewApp(
	logger *zerolog.Logger,
	config *config.Config,
) *App {
	appLogger := logger.With().Str("component", "app").Logger()

	var tendermintRPC *tendermint.TendermintRPC
	var cosmovisor *cosmovisor2.Cosmovisor
	var grpc *grpc2.Grpc
	var github *github2.Github

	if config.TendermintConfig.Address != "" {
		tendermintRPC = tendermint.NewTendermintRPC(config, logger)
	}

	if config.GrpcConfig.Address != "" {
		grpc = grpc2.NewGrpc(config, logger)
	}

	if config.CosmovisorConfig.IsEnabled() {
		cosmovisor = cosmovisor2.NewCosmovisor(config, logger)
	}

	if config.GithubConfig.Repository != "" {
		github = github2.NewGithub(config, logger)
	}

	queriers := []types.Querier{
		node_stats.NewNodeStatsQuerier(logger, tendermintRPC),
		versions.NewVersionsQuerier(logger, github, cosmovisor),
		upgrades.NewUpgradesQuerier(logger, cosmovisor, grpc, tendermintRPC),
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
			Name: constants.MetricsPrefix + "querier_enabled",
			Help: "Is querier enabled?",
		},
		[]string{"querier"},
	)
	registry.MustRegister(querierEnabled)

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResults := map[string][]prometheus.Collector{}
	allQueries := map[string][]query_info.QueryInfo{}

	for _, querier := range a.Queriers {
		querierEnabled.
			With(prometheus.Labels{"querier": querier.Name()}).
			Set(utils.BoolToFloat64(querier.Enabled()))

		if !querier.Enabled() {
			continue
		}

		wg.Add(1)
		go func(querier types.Querier) {
			querierResults, queriesInfo := querier.Get()
			mu.Lock()
			allResults[querier.Name()] = querierResults
			allQueries[querier.Name()] = queriesInfo
			mu.Unlock()
			wg.Done()
		}(querier)
	}

	wg.Wait()

	allResults["query_infos"] = query_info.GetQueryInfoMetrics(allQueries)

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
