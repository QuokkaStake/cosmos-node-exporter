package pkg

import (
	"context"
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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type NodeHandler struct {
	Logger   zerolog.Logger
	Queriers []types.Querier
	Config   configPkg.NodeConfig
	Tracer   trace.Tracer
}

func NewNodeHandler(
	logger *zerolog.Logger,
	config configPkg.NodeConfig,
	tracer trace.Tracer,
) *NodeHandler {
	appLogger := logger.With().
		Str("component", "node_handler").
		Str("node", config.Name).
		Logger()

	var tendermintRPC *tendermint.RPC
	var cosmovisor *cosmovisorPkg.Cosmovisor
	var grpc *grpcPkg.Client

	if config.TendermintConfig.Enabled.Bool {
		tendermintRPC = tendermint.NewRPC(config.TendermintConfig, appLogger, tracer)
	}

	if config.CosmovisorConfig.Enabled.Bool {
		cosmovisor = cosmovisorPkg.NewCosmovisor(config.CosmovisorConfig, appLogger, tracer)
	}

	if config.GrpcConfig.Enabled.Bool {
		grpc = grpcPkg.NewClient(config.GrpcConfig, appLogger, tracer)
	}

	gitClient := git.GetClient(config.GitConfig, appLogger, tracer)

	queriers := []types.Querier{
		nodeStats.NewQuerier(appLogger, tendermintRPC, tracer),
		versions.NewQuerier(appLogger, gitClient, cosmovisor, tracer),
		upgrades.NewQuerier(config.TendermintConfig.QueryUpgrades.Bool, appLogger, cosmovisor, tendermintRPC, tracer),
		cosmovisorQuerierPkg.NewQuerier(appLogger, cosmovisor, tracer),
		nodeConfig.NewQuerier(appLogger, grpc, tracer),
		nodeInfo.NewQuerier(appLogger, grpc, tracer),
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
		Tracer:   tracer,
	}
}

func (a *NodeHandler) Process(ctx context.Context) ([]metrics.MetricInfo, map[string][]query_info.QueryInfo) {
	childCtx, span := a.Tracer.Start(
		ctx,
		"Node "+a.Config.Name,
		trace.WithAttributes(attribute.String("node", a.Config.Name)),
	)
	defer span.End()

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
			querierResults, queriesInfo := querier.Get(childCtx)
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
