package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
	"main/pkg/types"
	"main/pkg/utils"
	"net/url"
	"strings"

	upgradeTypes "cosmossdk.io/x/upgrade/types"
)

type CosmovisorUpgradesGenerator struct{}

func NewCosmovisorUpgradesGenerator() *CosmovisorUpgradesGenerator {
	return &CosmovisorUpgradesGenerator{}
}

func (g *CosmovisorUpgradesGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	metricsInfo := []metrics.MetricInfo{}

	governanceUpgrade, governanceUpgradeFound := fetchers.StateGet[*upgradeTypes.Plan](state, constants.FetcherNameUpgrades)
	upgradeJSON, upgradeJSONFound := fetchers.StateGet[*upgradeTypes.Plan](state, constants.FetcherNameCosmovisorUpgradeInfo)
	cosmovisorUpgrades, cosmovisorUpgradesFound := fetchers.StateGet[*types.UpgradesPresent](state, constants.FetcherNameCosmovisorUpgrades)

	if !cosmovisorUpgradesFound {
		return metricsInfo
	}

	if governanceUpgradeFound {
		metricsInfo = append(metricsInfo, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeBinaryPresent,
			Labels: map[string]string{
				"name":   governanceUpgrade.Name,
				"source": constants.UpgradeSourceGovernance,
			},
			Value: utils.BoolToFloat64(cosmovisorUpgrades.HasUpgrade(g.normalizeName(governanceUpgrade.Name))),
		})
	}

	if upgradeJSONFound {
		metricsInfo = append(metricsInfo, metrics.MetricInfo{
			MetricName: metrics.MetricNameUpgradeBinaryPresent,
			Labels: map[string]string{
				"name":   upgradeJSON.Name,
				"source": constants.UpgradeSourceUpgradeInfo,
			},
			Value: utils.BoolToFloat64(cosmovisorUpgrades.HasUpgrade(g.normalizeName(upgradeJSON.Name))),
		})
	}

	return metricsInfo
}

func (g *CosmovisorUpgradesGenerator) normalizeName(originalName string) string {
	// From cosmovisor docs:
	// The name variable in upgrades/<name> is the lowercase URI-encoded name
	// of the upgrade as specified in the upgrade module plan.
	upgradeName := strings.ToLower(originalName)
	upgradeName = url.QueryEscape(upgradeName)

	return upgradeName
}
