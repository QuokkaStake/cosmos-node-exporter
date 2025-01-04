package fetchers

import (
	"context"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/tendermint"
	"main/pkg/constants"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type CosmovisorUpgradeInfoFetcher struct {
	Cosmovisor    *cosmovisorPkg.Cosmovisor
	QueryUpgrades bool
	Logger        zerolog.Logger
	Tracer        trace.Tracer
}

func NewCosmovisorUpgradeInfoFetcher(
	logger zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tracer trace.Tracer,
) *CosmovisorUpgradeInfoFetcher {
	return &CosmovisorUpgradeInfoFetcher{
		Logger:     logger.With().Str("component", "cosmovisor_upgrade_info").Logger(),
		Cosmovisor: cosmovisor,
		Tracer:     tracer,
	}
}

func (n *CosmovisorUpgradeInfoFetcher) Enabled() bool {
	return n.Cosmovisor != nil
}

func (n *CosmovisorUpgradeInfoFetcher) Name() constants.FetcherName {
	return constants.FetcherNameCosmovisorUpgradeInfo
}

func (n *CosmovisorUpgradeInfoFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{constants.FetcherNameNodeStatus}
}

func (n *CosmovisorUpgradeInfoFetcher) Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo) {
	childCtx, span := n.Tracer.Start(
		ctx,
		"Fetcher "+string(n.Name()),
	)
	defer span.End()

	if len(data) < 1 {
		panic("data is empty")
	}

	status, statusConverted := Convert[tendermint.StatusResponse](data[0])
	if !statusConverted {
		n.Logger.Trace().Msg("Node status is empty, cannot check if the upgrade-info plan is in the past.")
		return nil, []query_info.QueryInfo{}
	}

	upgradeInfo, queryInfo, err := n.Cosmovisor.GetUpgradeInfo(childCtx)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch upgrade info")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	if upgradeInfo == nil {
		return nil, []query_info.QueryInfo{queryInfo}
	}

	if upgradeInfo.Height < status.Result.SyncInfo.LatestBlockHeight {
		n.Logger.Trace().
			Int64("node_height", status.Result.SyncInfo.LatestBlockHeight).
			Int64("upgrade_height", upgradeInfo.Height).
			Msg("Cosmovisor upgrade-info is in the past, skipping.")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return upgradeInfo, []query_info.QueryInfo{queryInfo}
}
