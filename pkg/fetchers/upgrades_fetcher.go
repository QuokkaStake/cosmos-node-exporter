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

type UpgradesFetcher struct {
	TendermintRPC *tendermint.RPC
	QueryUpgrades bool
	Logger        zerolog.Logger
	Tracer        trace.Tracer
}

func NewUpgradesFetcher(
	logger zerolog.Logger,
	tendermintRPC *tendermint.RPC,
	queryUpgrades bool,
	tracer trace.Tracer,
) *UpgradesFetcher {
	return &UpgradesFetcher{
		Logger:        logger.With().Str("component", "upgrades_fetcher").Logger(),
		TendermintRPC: tendermintRPC,
		QueryUpgrades: queryUpgrades,
		Tracer:        tracer,
	}
}

func (n *UpgradesFetcher) Enabled() bool {
	return n.TendermintRPC != nil && n.QueryUpgrades
}

func (n *UpgradesFetcher) Name() constants.FetcherName {
	return constants.FetcherNameUpgrades
}

func (n *UpgradesFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (n *UpgradesFetcher) Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo) {
	childCtx, span := n.Tracer.Start(
		ctx,
		"Fetcher "+string(n.Name()),
		trace.WithAttributes(attribute.String("node", n.TendermintRPC.Address)),
	)
	defer span.End()

	upgradePlan, queryInfo, err := n.TendermintRPC.GetUpgradePlan(childCtx)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch upgrade plan")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return upgradePlan, []query_info.QueryInfo{queryInfo}
}
