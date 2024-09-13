package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"

	upgradeTypes "cosmossdk.io/x/upgrade/types"
)

type UpgradesGenerator struct{}

func NewUpgradesGenerator() *UpgradesGenerator {
	return &UpgradesGenerator{}
}

func (g *UpgradesGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	upgradesRaw, ok := state[constants.FetcherNameUpgrades]
	if !ok || upgradesRaw == nil {
		return []metrics.MetricInfo{}
	}

	upgradeInfo, ok := upgradesRaw.(*upgradeTypes.Plan)
	if !ok {
		panic("expected the state entry to be string")
	}

	if upgradeInfo == nil {
		return []metrics.MetricInfo{}
	}

	return []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameUpgradeInfo,
			Labels:     map[string]string{"name": upgradeInfo.Name, "info": upgradeInfo.Info},
			Value:      1,
		},
		{
			MetricName: metrics.MetricNameUpgradeHeight,
			Labels:     map[string]string{"name": upgradeInfo.Name},
			Value:      float64(upgradeInfo.Height),
		},
	}
}
