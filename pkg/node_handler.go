package pkg

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/git"
	cosmovisorQuerierPkg "main/pkg/queriers/cosmovisor"
	nodeStats "main/pkg/queriers/node_stats"
	"main/pkg/queriers/upgrades"
	"main/pkg/queriers/versions"
	"main/pkg/query_info"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type NodeHandler struct {
	Logger   zerolog.Logger
	Queriers []types.Querier
	Config   configPkg.NodeConfig
}

func NewNodeHandler(
	logger *zerolog.Logger,
	config configPkg.NodeConfig,
) *NodeHandler {
	appLogger := logger.With().
		Str("component", "node_handler").
		Str("node", config.Name).
		Logger()

	var tendermintRPC *tendermint.RPC
	var cosmovisor *cosmovisorPkg.Cosmovisor

	if config.TendermintConfig.Enabled.Bool {
		tendermintRPC = tendermint.NewRPC(config, logger)
	}

	if config.CosmovisorConfig.Enabled.Bool {
		cosmovisor = cosmovisorPkg.NewCosmovisor(config, logger)
	}

	gitClient := git.GetClient(config, logger)

	queriers := []types.Querier{
		nodeStats.NewQuerier(logger, config, tendermintRPC),
		versions.NewQuerier(logger, config, gitClient, cosmovisor),
		upgrades.NewQuerier(config, logger, cosmovisor, tendermintRPC),
		cosmovisorQuerierPkg.NewQuerier(config, logger, cosmovisor),
	}

	for _, querier := range queriers {
		if querier.Enabled() {
			appLogger.Debug().Str("name", querier.Name()).Msg("Querier is enabled")
		} else {
			appLogger.Debug().Str("name", querier.Name()).Msg("Querier is disabled")
		}
	}

	return &NodeHandler{
		Logger:   appLogger,
		Queriers: queriers,
		Config:   config,
	}
}

func (a *NodeHandler) Process() (map[string][]prometheus.Collector, map[string][]query_info.QueryInfo) {
	querierEnabled := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "querier_enabled",
			Help: "Is querier enabled?",
		},
		[]string{"node", "querier"},
	)

	var wg sync.WaitGroup
	var mu sync.Mutex
	allResults := map[string][]prometheus.Collector{"querier_enabled": {querierEnabled}}
	allQueries := map[string][]query_info.QueryInfo{}

	for _, querier := range a.Queriers {
		querierEnabled.
			With(prometheus.Labels{
				"querier": querier.Name(),
				"node":    a.Config.Name,
			}).
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

	return allResults, allQueries
}
