package upgrades

import (
	"main/pkg/config"
	"main/pkg/constants"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/query_info"
	"main/pkg/tendermint"
	"main/pkg/utils"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
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
	logger *zerolog.Logger,
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

func (u *Querier) Get() ([]prometheus.Collector, []query_info.QueryInfo) {
	upgradePlanQuery := query_info.QueryInfo{
		Module:  "tendermint",
		Action:  "get_upgrade_plan",
		Success: false,
	}

	upgrade, err := u.Tendermint.GetUpgradePlan()
	if err != nil {
		u.Logger.Err(err).Msg("Could not get latest upgrade plan from Tendermint")
		return []prometheus.Collector{}, []query_info.QueryInfo{upgradePlanQuery}
	}

	upgradePlanQuery.Success = true
	isUpgradePresent := upgrade != nil

	upcomingUpgradePresent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "upgrade_coming",
			Help: "Is future upgrade planned?",
		},
		[]string{"node"},
	)

	upcomingUpgradePresent.
		With(prometheus.Labels{"node": u.Config.Name}).
		Set(utils.BoolToFloat64(isUpgradePresent))

	queries := []prometheus.Collector{upcomingUpgradePresent}
	queryInfos := []query_info.QueryInfo{upgradePlanQuery}

	if !isUpgradePresent {
		return queries, queryInfos
	}

	upgradeInfoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "upgrade_info",
			Help: "Future upgrade info",
		},
		[]string{"node", "name", "info"},
	)

	upgradeHeightGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "upgrade_height",
			Help: "Future upgrade height",
		},
		[]string{"node", "name", "info"},
	)

	upgradeInfoGauge.
		With(prometheus.Labels{"node": u.Config.Name, "name": upgrade.Name, "info": upgrade.Info}).
		Set(utils.BoolToFloat64(isUpgradePresent))
	upgradeHeightGauge.
		With(prometheus.Labels{"node": u.Config.Name, "name": upgrade.Name, "info": upgrade.Info}).
		Set(float64(upgrade.Height))
	queries = append(queries, upgradeInfoGauge, upgradeHeightGauge)

	upgradeEstimatedTimeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "upgrade_estimated_time",
			Help: "Estimated upgrade time, as Unix timestamp",
		},
		[]string{"node", "name", "info"},
	)

	// Calculate upgrade estimated time
	if u.Tendermint == nil {
		u.Logger.Warn().
			Msg("Tendermint RPC not initialized and upgrade time is not specified, not returning upgrade time.")
		return queries, queryInfos
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
		return queries, queryInfos
	}
	upgradeTimeQuery.Success = true
	queryInfos = append(queryInfos, upgradeTimeQuery)

	upgradeEstimatedTimeGauge.
		With(prometheus.Labels{"node": u.Config.Name, "name": upgrade.Name, "info": upgrade.Info}).
		Set(float64(upgradeTime.Unix()))
	queries = append(queries, upgradeEstimatedTimeGauge)

	if u.Cosmovisor == nil {
		u.Logger.Warn().
			Msg("Cosmovisor not initialized, not returning binary presence.")
		return queries, queryInfos
	}

	cosmovisorGetUpgradesQueryInfo := query_info.QueryInfo{
		Action:  "cosmovisor_get_upgrades",
		Success: false,
	}

	upgrades, err := u.Cosmovisor.GetUpgrades()
	if err != nil {
		u.Logger.Error().Err(err).Msg("Could not get Cosmovisor upgrades")
		queryInfos = append(queryInfos, cosmovisorGetUpgradesQueryInfo)
		return []prometheus.Collector{}, queryInfos
	}

	cosmovisorGetUpgradesQueryInfo.Success = true
	queryInfos = append(queryInfos, cosmovisorGetUpgradesQueryInfo)

	upgradeBinaryPresentGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "upgrade_binary_present",
			Help: "Is upgrade binary present?",
		},
		[]string{"node", "name", "info"},
	)

	// From cosmovisor docs:
	// The name variable in upgrades/<name> is the lowercase URI-encoded name
	// of the upgrade as specified in the upgrade module plan.
	upgradeName := strings.ToLower(upgrade.Name)
	upgradeName = url.QueryEscape(upgradeName)

	upgradeBinaryPresentGauge.
		With(prometheus.Labels{"node": u.Config.Name, "name": upgrade.Name, "info": upgrade.Info}).
		Set(utils.BoolToFloat64(upgrades.HasUpgrade(upgradeName)))
	queries = append(queries, upgradeBinaryPresentGauge)

	return queries, queryInfos
}
