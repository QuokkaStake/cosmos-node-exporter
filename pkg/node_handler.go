package pkg

import (
	"context"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/git"
	grpcPkg "main/pkg/clients/grpc"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	fetchersPkg "main/pkg/fetchers"
	generatorsPkg "main/pkg/generators"
	metricsPkg "main/pkg/metrics"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type NodeHandler struct {
	Logger     zerolog.Logger
	Config     configPkg.NodeConfig
	Tracer     trace.Tracer
	Generators []generatorsPkg.Generator
	Controller *fetchersPkg.Controller
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

	fetchers := fetchersPkg.Fetchers{
		fetchersPkg.NewNodeStatusFetcher(appLogger, tendermintRPC, tracer),
		fetchersPkg.NewCosmovisorVersionFetcher(appLogger, cosmovisor, tracer),
		fetchersPkg.NewNodeConfigFetcher(appLogger, grpc, tracer),
		fetchersPkg.NewNodeInfoFetcher(appLogger, grpc, tracer),
		fetchersPkg.NewRemoteVersionFetcher(appLogger, gitClient, tracer),
		fetchersPkg.NewLocalVersionFetcher(appLogger, cosmovisor, tracer),
		fetchersPkg.NewUpgradesFetcher(appLogger, tendermintRPC, config.TendermintConfig.QueryUpgrades.Bool, tracer),
		fetchersPkg.NewBlockTimeFetcher(appLogger, tendermintRPC, tracer),
		fetchersPkg.NewCosmovisorUpgradesFetcher(appLogger, cosmovisor, tracer),
		fetchersPkg.NewCosmovisorUpgradeInfoFetcher(appLogger, cosmovisor, tracer),
	}

	generators := []generatorsPkg.Generator{
		generatorsPkg.NewNodeStatusGenerator(),
		generatorsPkg.NewCosmovisorVersionGenerator(),
		generatorsPkg.NewNodeConfigGenerator(),
		generatorsPkg.NewNodeInfoGenerator(),
		generatorsPkg.NewRemoteVersionGenerator(),
		generatorsPkg.NewLocalVersionGenerator(),
		generatorsPkg.NewIsLatestGenerator(appLogger),
		generatorsPkg.NewUpgradesGenerator(),
		generatorsPkg.NewTimeTillUpgradeGenerator(),
		generatorsPkg.NewCosmovisorUpgradesGenerator(),
	}

	controller := fetchersPkg.NewController(fetchers, appLogger, config.Name)

	for _, fetcher := range fetchers {
		if fetcher.Enabled() {
			appLogger.Debug().Str("name", string(fetcher.Name())).Msg("Fetcher is enabled")
		} else {
			appLogger.Debug().Str("name", string(fetcher.Name())).Msg("Fetcher is disabled")
		}
	}

	return &NodeHandler{
		Logger:     appLogger,
		Config:     config,
		Tracer:     tracer,
		Generators: generators,
		Controller: controller,
	}
}

func (a *NodeHandler) Process(ctx context.Context) ([]metricsPkg.MetricInfo, map[string][]query_info.QueryInfo) {
	childCtx, span := a.Tracer.Start(
		ctx,
		"Node "+a.Config.Name,
		trace.WithAttributes(attribute.String("node", a.Config.Name)),
	)
	defer span.End()

	allResults := []metricsPkg.MetricInfo{}
	allQueries := map[string][]query_info.QueryInfo{}

	state, queries := a.Controller.Fetch(childCtx)

	for _, generator := range a.Generators {
		metrics := generator.Get(state)
		allResults = append(allResults, metrics...)
	}

	for key, fetcherQueries := range queries {
		allQueries[string(key)] = fetcherQueries
	}

	return allResults, allQueries
}
