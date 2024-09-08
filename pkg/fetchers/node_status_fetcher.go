package fetchers

import (
	"context"
	"main/pkg/clients/tendermint"
	"main/pkg/constants"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type NodeStatusFetcher struct {
	TendermintRPC *tendermint.RPC
	Logger        zerolog.Logger
	Tracer        trace.Tracer
}

func NewNodeStatusFetcher(logger zerolog.Logger, tendermintRPC *tendermint.RPC, tracer trace.Tracer) *NodeStatusFetcher {
	return &NodeStatusFetcher{
		Logger:        logger.With().Str("component", "node-stats-querier").Logger(),
		TendermintRPC: tendermintRPC,
		Tracer:        tracer,
	}
}

func (n *NodeStatusFetcher) Enabled() bool {
	return n.TendermintRPC != nil
}

func (n *NodeStatusFetcher) Name() constants.FetcherName {
	return constants.FetcherNameNodeStatus
}

func (n *NodeStatusFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (n *NodeStatusFetcher) Get(ctx context.Context) (interface{}, []query_info.QueryInfo) {
	childCtx, span := n.Tracer.Start(
		ctx,
		"Fetcher "+string(n.Name()),
		trace.WithAttributes(attribute.String("node", n.TendermintRPC.Address)),
	)
	defer span.End()

	status, queryInfo, err := n.TendermintRPC.Status(childCtx)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node status")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return status, []query_info.QueryInfo{queryInfo}
}
