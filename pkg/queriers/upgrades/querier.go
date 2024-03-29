package upgrades

import (
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/tendermint"
	"main/pkg/config"
	"main/pkg/metrics"
	"main/pkg/query_info"
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
	upgrade, upgradePlanQuery, err := u.Tendermint.GetUpgradePlan()
	if err != nil {
		u.Logger.Err(err).Msg("Could not get latest upgrade plan from Tendermint")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{upgradePlanQuery}
	}

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

	upgradeTime, upgradeTimeQuery, err := u.Tendermint.GetEstimateTimeTillBlock(upgrade.Height)
	queryInfos = append(queryInfos, upgradeTimeQuery)

	if err != nil {
		u.Logger.Err(err).Msg("Could not get estimated upgrade time")
		return metricInfos, queryInfos
	}

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

	upgrades, cosmovisorGetUpgradesQueryInfo, err := u.Cosmovisor.GetUpgrades()
	if err != nil {
		u.Logger.Error().Err(err).Msg("Could not get Cosmovisor upgrades")
		queryInfos = append(queryInfos, cosmovisorGetUpgradesQueryInfo)
		return metricInfos, queryInfos
	}

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
