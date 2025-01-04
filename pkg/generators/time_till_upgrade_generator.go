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
	governanceUpgrade, governanceUpgradeFound := fetchers.StateGet[*types.Plan](state, constants.FetcherNameUpgrades)
	upgradeInfoJson, upgradeInfoJsonFound := fetchers.StateGet[*types.Plan](state, constants.FetcherNameCosmovisorUpgradeInfo)
	blocksInfo, blocksInfoFound := fetchers.StateGet[*tendermint.BlocksInfo](state, constants.FetcherNameBlockTime)

	if !blocksInfoFound {
		return []metrics.MetricInfo{}
	}

	metricsInfo := []metrics.MetricInfo{}

	if governanceUpgradeFound {
		metricsInfo = append(metricsInfo, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeEstimatedTime,
			Labels: map[string]string{
				"name":   governanceUpgrade.Name,
				"source": constants.UpgradeSourceGovernance,
			},
			Value: float64(g.calculateBlockTime(governanceUpgrade.Height, blocksInfo).Unix()),
		})
	}

	if upgradeInfoJsonFound {
		metricsInfo = append(metricsInfo, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeEstimatedTime,
			Labels: map[string]string{
				"name":   upgradeInfoJson.Name,
				"source": constants.UpgradeSourceUpgradeInfo,
			},
			Value: float64(g.calculateBlockTime(upgradeInfoJson.Height, blocksInfo).Unix()),
		})
	}

	return metricsInfo
}

func (g *TimeTillUpgradeGenerator) calculateBlockTime(height int64, blocksInfo *tendermint.BlocksInfo) time.Time {
	blockTime := blocksInfo.BlockTime()

	blocksTillEstimatedBlock := height - blocksInfo.NewerBlock.Result.Block.Header.Height
	secondsTillEstimatedBlock := int64(float64(blocksTillEstimatedBlock) * blockTime)
	durationTillEstimatedBlock := time.Duration(secondsTillEstimatedBlock * int64(time.Second))
	upgradeTime := blocksInfo.NewerBlock.Result.Block.Header.Time.Add(durationTillEstimatedBlock)

	return upgradeTime
}
