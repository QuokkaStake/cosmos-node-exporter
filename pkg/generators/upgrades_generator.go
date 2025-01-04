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
	metricInfos := []metrics.MetricInfo{{
		MetricName: metrics.MetricNameUpgradeComing,
		Labels:     map[string]string{},
		Value:      0,
	}}

	upgradeInfo, upgradeInfoFound := fetchers.StateGet[*upgradeTypes.Plan](state, constants.FetcherNameUpgrades)
	if !upgradeInfoFound {
		return metricInfos
	}

	metricInfos[0].Value = 1

	metricInfos = append(metricInfos, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeInfo,
		Labels:     map[string]string{"name": upgradeInfo.Name, "info": upgradeInfo.Info},
		Value:      1,
	}, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeHeight,
		Labels:     map[string]string{"name": upgradeInfo.Name},
		Value:      float64(upgradeInfo.Height),
	})

	return metricInfos
}
