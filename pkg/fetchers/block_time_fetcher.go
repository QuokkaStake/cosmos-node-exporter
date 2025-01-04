package fetchers

import (
	"context"
	"main/pkg/clients/tendermint"
	"main/pkg/constants"
	"main/pkg/query_info"

	"cosmossdk.io/x/upgrade/types"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type BlockTimeFetcher struct {
	TendermintRPC *tendermint.RPC
	Logger        zerolog.Logger
	Tracer        trace.Tracer
}

func NewBlockTimeFetcher(logger zerolog.Logger, tendermintRPC *tendermint.RPC, tracer trace.Tracer) *BlockTimeFetcher {
	return &BlockTimeFetcher{
		Logger:        logger.With().Str("component", "block_time_fetcher").Logger(),
		TendermintRPC: tendermintRPC,
		Tracer:        tracer,
	}
}

func (n *BlockTimeFetcher) Enabled() bool {
	return n.TendermintRPC != nil
}

func (n *BlockTimeFetcher) Name() constants.FetcherName {
	return constants.FetcherNameBlockTime
}

func (n *BlockTimeFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{
		constants.FetcherNameUpgrades,
		constants.FetcherNameCosmovisorUpgradeInfo,
	}
}

func (n *BlockTimeFetcher) Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo) {
	if len(data) < 2 {
		panic("data is empty")
	}

	_, governanceUpgradePlanConverted := Convert[*types.Plan](data[0])
	_, upgradeInfoJSONConverted := Convert[*types.Plan](data[1])

	if !governanceUpgradePlanConverted && !upgradeInfoJSONConverted {
		n.Logger.Trace().Msg("Upgrade plan is empty, not fetching block time.")
		return nil, []query_info.QueryInfo{}
	}

	childCtx, span := n.Tracer.Start(
		ctx,
		"Fetcher "+string(n.Name()),
		trace.WithAttributes(attribute.String("node", n.TendermintRPC.Address)),
	)
	defer span.End()

	blockTime, queryInfo, err := n.TendermintRPC.GetBlockTime(childCtx)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch block time info")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return blockTime, []query_info.QueryInfo{queryInfo}
}
