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
	upgradePlanRaw, ok := state[constants.FetcherNameUpgrades]
	if !ok || upgradePlanRaw == nil {
		return []metrics.MetricInfo{}
	}

	cosmovisorUpgradesRaw, ok := state[constants.FetcherNameCosmovisorUpgrades]
	if !ok || cosmovisorUpgradesRaw == nil {
		return []metrics.MetricInfo{}
	}

	upgradePlan, ok := upgradePlanRaw.(*upgradeTypes.Plan)
	if !ok {
		panic("expected the state entry to be *types.Plan")
	}

	cosmovisorUpgrades, ok := cosmovisorUpgradesRaw.(types.UpgradesPresent)
	if !ok {
		panic("expected the state entry to be types.UpgradesPresent")
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
