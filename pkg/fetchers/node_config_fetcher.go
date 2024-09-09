package fetchers

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	"main/pkg/constants"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type NodeConfigFetcher struct {
	gRPC   *grpcPkg.Client
	Logger zerolog.Logger
	Tracer trace.Tracer
}

func NewNodeConfigFetcher(logger zerolog.Logger, grpc *grpcPkg.Client, tracer trace.Tracer) *NodeConfigFetcher {
	return &NodeConfigFetcher{
		Logger: logger.With().Str("component", "node_config_fetcher").Logger(),
		gRPC:   grpc,
		Tracer: tracer,
	}
}

func (n *NodeConfigFetcher) Enabled() bool {
	return n.gRPC != nil
}

func (n *NodeConfigFetcher) Name() constants.FetcherName {
	return constants.FetcherNameNodeConfig
}

func (n *NodeConfigFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (n *NodeConfigFetcher) Get(ctx context.Context) (interface{}, []query_info.QueryInfo) {
	childCtx, span := n.Tracer.Start(
		ctx,
		"NodeConfigFetcher "+string(n.Name()),
		trace.WithAttributes(attribute.String("node", string(n.Name()))),
	)
	defer span.End()

	config, queryInfo, err := n.gRPC.GetNodeConfig(childCtx)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node config")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	if config == nil {
		n.Logger.Debug().
			Msg("Node config is nil, probably chain does not implement the node config endpoint.")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return config, []query_info.QueryInfo{queryInfo}
}
