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

type NodeInfoFetcher struct {
	gRPC   *grpcPkg.Client
	Logger zerolog.Logger
	Tracer trace.Tracer
}

func NewNodeInfoFetcher(logger zerolog.Logger, grpc *grpcPkg.Client, tracer trace.Tracer) *NodeInfoFetcher {
	return &NodeInfoFetcher{
		Logger: logger.With().Str("component", "node_info_fetcher").Logger(),
		gRPC:   grpc,
		Tracer: tracer,
	}
}

func (n *NodeInfoFetcher) Enabled() bool {
	return n.gRPC != nil
}

func (n *NodeInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameNodeInfo
}

func (n *NodeInfoFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (n *NodeInfoFetcher) Get(ctx context.Context) (interface{}, []query_info.QueryInfo) {
	childCtx, span := n.Tracer.Start(
		ctx,
		"NodeInfoFetcher "+string(n.Name()),
		trace.WithAttributes(attribute.String("node", string(n.Name()))),
	)
	defer span.End()

	config, queryInfo, err := n.gRPC.GetNodeInfo(childCtx)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node info")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return config, []query_info.QueryInfo{queryInfo}
}
