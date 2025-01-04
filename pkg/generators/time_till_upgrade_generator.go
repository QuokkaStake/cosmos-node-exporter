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
	upgrade, upgradeFound := fetchers.StateGet[*types.Plan](state, constants.FetcherNameUpgrades)
	blocksInfo, blocksInfoFound := fetchers.StateGet[*tendermint.BlocksInfo](state, constants.FetcherNameBlockTime)

	if !upgradeFound || !blocksInfoFound {
		return []metrics.MetricInfo{}
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
