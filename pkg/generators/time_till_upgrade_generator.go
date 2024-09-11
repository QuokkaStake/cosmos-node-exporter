package generators

import (
	"main/pkg/clients/tendermint"
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
	"time"

	"cosmossdk.io/x/upgrade/types"
)

type TimeTillUpgradeGenerator struct{}

func NewTimeTillUpgradeGenerator() *TimeTillUpgradeGenerator {
	return &TimeTillUpgradeGenerator{}
}

func (g *TimeTillUpgradeGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	upgradeRaw, ok := state[constants.FetcherNameUpgrades]
	if !ok || upgradeRaw == nil {
		return []metrics.MetricInfo{}
	}

	blocksInfoRaw, ok := state[constants.FetcherNameBlockTime]
	if !ok || blocksInfoRaw == nil {
		return []metrics.MetricInfo{}
	}

	upgrade, ok := upgradeRaw.(*types.Plan)
	if !ok {
		panic("expected the state entry to be *types.Plan")
	}

	blocksInfo, ok := blocksInfoRaw.(*tendermint.BlocksInfo)
	if !ok {
		panic("expected the state entry to be *tendermint.BlocksInfo")
	}

	blockTime := blocksInfo.BlockTime()

	blocksTillEstimatedBlock := upgrade.Height - blocksInfo.NewerBlock.Result.Block.Header.Height
	secondsTillEstimatedBlock := int64(float64(blocksTillEstimatedBlock) * blockTime)
	durationTillEstimatedBlock := time.Duration(secondsTillEstimatedBlock * int64(time.Second))
	upgradeTime := blocksInfo.NewerBlock.Result.Block.Header.Time.Add(durationTillEstimatedBlock)

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameUpgradeEstimatedTime,
		Labels:     map[string]string{"name": upgrade.Name},
		Value:      float64(upgradeTime.Unix()),
	}}
}
