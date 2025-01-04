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
	upgradePlan, upgradePlanFound := fetchers.StateGet[*upgradeTypes.Plan](state, constants.FetcherNameUpgrades)
	cosmovisorUpgrades, cosmovisorUpgradesFound := fetchers.StateGet[*types.UpgradesPresent](state, constants.FetcherNameCosmovisorUpgrades)

	if !upgradePlanFound || !cosmovisorUpgradesFound {
		return []metrics.MetricInfo{}
	}

	// From cosmovisor docs:
	// The name variable in upgrades/<name> is the lowercase URI-encoded name
	// of the upgrade as specified in the upgrade module plan.
	upgradeName := strings.ToLower(upgradePlan.Name)
	upgradeName = url.QueryEscape(upgradeName)

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameUpgradeBinaryPresent,
		Labels:     map[string]string{"name": upgradePlan.Name},
		Value:      utils.BoolToFloat64(cosmovisorUpgrades.HasUpgrade(upgradeName)),
	}}
}
