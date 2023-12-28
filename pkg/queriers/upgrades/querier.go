package upgrades

import (
	"main/pkg/config"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/metrics"
	"main/pkg/query_info"
	"main/pkg/tendermint"
	"main/pkg/utils"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
)

type Querier struct {
	Config     config.NodeConfig
	Logger     zerolog.Logger
	Cosmovisor *cosmovisorPkg.Cosmovisor
	Tendermint *tendermint.RPC
}

func NewQuerier(
	nodeConfig config.NodeConfig,
	logger zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tendermint *tendermint.RPC,
) *Querier {
	return &Querier{
		Config:     nodeConfig,
		Logger:     logger.With().Str("component", "upgrades_querier").Logger(),
		Cosmovisor: cosmovisor,
		Tendermint: tendermint,
	}
}

func (u *Querier) Enabled() bool {
	return u.Tendermint != nil && u.Config.TendermintConfig.QueryUpgrades.Bool
}

func (u *Querier) Name() string {
	return "upgrades-querier"
}

func (u *Querier) Get() ([]metrics.MetricInfo, []query_info.QueryInfo) {
	upgradePlanQuery := query_info.QueryInfo{
		Module:  "tendermint",
		Action:  "get_upgrade_plan",
		Success: false,
	}

	upgrade, err := u.Tendermint.GetUpgradePlan()
	if err != nil {
		u.Logger.Err(err).Msg("Could not get latest upgrade plan from Tendermint")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{upgradePlanQuery}
	}

	upgradePlanQuery.Success = true
	isUpgradePresent := upgrade != nil

	metricInfos := []metrics.MetricInfo{{
		MetricName: metrics.MetricNameUpgradeComing,
		Labels:     map[string]string{},
		Value:      utils.BoolToFloat64(isUpgradePresent),
	}}
	queryInfos := []query_info.QueryInfo{upgradePlanQuery}

	if !isUpgradePresent {
		return metricInfos, queryInfos
	}

	metricInfos = append(metricInfos, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeInfo,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      utils.BoolToFloat64(isUpgradePresent),
	}, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeHeight,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      float64(upgrade.Height),
	})

	// Calculate upgrade estimated time
	if u.Tendermint == nil {
		u.Logger.Warn().
			Msg("Tendermint RPC not initialized and upgrade time is not specified, not returning upgrade time.")
		return metricInfos, queryInfos
	}

	upgradeTimeQuery := query_info.QueryInfo{
		Module:  "tendermint",
		Action:  "tendermint_get_upgrade_time",
		Success: false,
	}

	upgradeTime, err := u.Tendermint.GetEstimateTimeTillBlock(upgrade.Height)
	if err != nil {
		u.Logger.Err(err).Msg("Could not get estimated upgrade time")
		queryInfos = append(queryInfos, upgradeTimeQuery)
		return metricInfos, queryInfos
	}
	upgradeTimeQuery.Success = true
	queryInfos = append(queryInfos, upgradeTimeQuery)

	metricInfos = append(metricInfos, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeEstimatedTime,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      float64(upgradeTime.Unix()),
	})

	if u.Cosmovisor == nil {
		u.Logger.Warn().
			Msg("Cosmovisor not initialized, not returning binary presence.")
		return metricInfos, queryInfos
	}

	cosmovisorGetUpgradesQueryInfo := query_info.QueryInfo{
		Action:  "cosmovisor_get_upgrades",
		Module:  "cosmovisor",
		Success: false,
	}

	upgrades, err := u.Cosmovisor.GetUpgrades()
	if err != nil {
		u.Logger.Error().Err(err).Msg("Could not get Cosmovisor upgrades")
		queryInfos = append(queryInfos, cosmovisorGetUpgradesQueryInfo)
		return metricInfos, queryInfos
	}

	cosmovisorGetUpgradesQueryInfo.Success = true
	queryInfos = append(queryInfos, cosmovisorGetUpgradesQueryInfo)

	// From cosmovisor docs:
	// The name variable in upgrades/<name> is the lowercase URI-encoded name
	// of the upgrade as specified in the upgrade module plan.
	upgradeName := strings.ToLower(upgrade.Name)
	upgradeName = url.QueryEscape(upgradeName)

	metricInfos = append(metricInfos, metrics.MetricInfo{
		MetricName: metrics.MetricNameUpgradeBinaryPresent,
		Labels:     map[string]string{"name": upgrade.Name, "info": upgrade.Info},
		Value:      utils.BoolToFloat64(upgrades.HasUpgrade(upgradeName)),
	})

	return metricInfos, queryInfos
}
