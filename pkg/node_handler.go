package pkg

import (
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/git"
	grpcPkg "main/pkg/clients/grpc"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	"main/pkg/metrics"
	cosmovisorQuerierPkg "main/pkg/queriers/cosmovisor"
	nodeConfig "main/pkg/queriers/node_config"
	nodeInfo "main/pkg/queriers/node_info"
	nodeStats "main/pkg/queriers/node_stats"
	"main/pkg/queriers/upgrades"
	"main/pkg/queriers/versions"
	"main/pkg/query_info"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

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
	var grpc *grpcPkg.Client

	if config.TendermintConfig.Enabled.Bool {
		tendermintRPC = tendermint.NewRPC(config, appLogger)
	}

	if config.CosmovisorConfig.Enabled.Bool {
		cosmovisor = cosmovisorPkg.NewCosmovisor(config, appLogger)
	}

	if config.GrpcConfig.Enabled.Bool {
		grpc = grpcPkg.NewClient(config, appLogger)
	}

	gitClient := git.GetClient(config, appLogger)

	queriers := []types.Querier{
		nodeStats.NewQuerier(appLogger, tendermintRPC),
		versions.NewQuerier(appLogger, gitClient, cosmovisor),
		upgrades.NewQuerier(config, appLogger, cosmovisor, tendermintRPC),
		cosmovisorQuerierPkg.NewQuerier(appLogger, cosmovisor),
		nodeConfig.NewQuerier(appLogger, grpc),
		nodeInfo.NewQuerier(appLogger, grpc),
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

func (a *NodeHandler) Process() ([]metrics.MetricInfo, map[string][]query_info.QueryInfo) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	allResults := []metrics.MetricInfo{}
	allQueries := map[string][]query_info.QueryInfo{}

	for _, querier := range a.Queriers {
		allResults = append(allResults, metrics.MetricInfo{
			MetricName: metrics.MetricNameQuerierEnabled,
			Labels: map[string]string{
				"querier": querier.Name(),
				"node":    a.Config.Name,
			},
			Value: utils.BoolToFloat64(querier.Enabled()),
		})

		if !querier.Enabled() {
			continue
		}

		wg.Add(1)
		go func(querier types.Querier) {
			querierResults, queriesInfo := querier.Get()
			mu.Lock()
			allResults = append(allResults, querierResults...)
			allQueries[querier.Name()] = queriesInfo
			mu.Unlock()
			wg.Done()
		}(querier)
	}

	wg.Wait()

	return allResults, allQueries
}
