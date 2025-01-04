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
	metricInfos := []metrics.MetricInfo{}

	governanceUpgrade, governanceUpgradeFound := fetchers.StateGet[*upgradeTypes.Plan](state, constants.FetcherNameUpgrades)
	upgradeInfoJSON, upgradeInfoJSONFound := fetchers.StateGet[*upgradeTypes.Plan](state, constants.FetcherNameCosmovisorUpgradeInfo)

	if governanceUpgradeFound {
		metricInfos = append(metricInfos, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeComing,
			Labels: map[string]string{
				"source": constants.UpgradeSourceGovernance,
			},
			Value: 1,
		}, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeInfo,
			Labels: map[string]string{
				"name":   governanceUpgrade.Name,
				"info":   governanceUpgrade.Info,
				"source": constants.UpgradeSourceGovernance,
			},
			Value: 1,
		}, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeHeight,
			Labels: map[string]string{
				"name":   governanceUpgrade.Name,
				"source": constants.UpgradeSourceGovernance,
			},
			Value: float64(governanceUpgrade.Height),
		})
	} else {
		metricInfos = append(metricInfos, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeComing,
			Labels: map[string]string{
				"source": constants.UpgradeSourceGovernance,
			},
			Value: 0,
		})
	}

	if upgradeInfoJSONFound {
		metricInfos = append(metricInfos, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeComing,
			Labels: map[string]string{
				"source": constants.UpgradeSourceUpgradeInfo,
			},
			Value: 1,
		}, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeInfo,
			Labels: map[string]string{
				"name":   upgradeInfoJSON.Name,
				"info":   upgradeInfoJSON.Info,
				"source": constants.UpgradeSourceUpgradeInfo,
			},
			Value: 1,
		}, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeHeight,
			Labels: map[string]string{
				"name":   upgradeInfoJSON.Name,
				"source": constants.UpgradeSourceUpgradeInfo,
			},
			Value: float64(upgradeInfoJSON.Height),
		})
	} else {
		metricInfos = append(metricInfos, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeComing,
			Labels: map[string]string{
				"source": constants.UpgradeSourceUpgradeInfo,
			},
			Value: 0,
		})
	}

	return metricInfos
}
