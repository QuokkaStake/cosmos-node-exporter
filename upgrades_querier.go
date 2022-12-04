package main

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type UpgradesQuerier struct {
	Logger     zerolog.Logger
	Cosmovisor *Cosmovisor
	Tendermint *TendermintRPC
}

func NewUpgradesQuerier(logger *zerolog.Logger, cosmovisor *Cosmovisor, tendermint *TendermintRPC) *UpgradesQuerier {
	return &UpgradesQuerier{
		Logger:     logger.With().Str("component", "upgrades_querier").Logger(),
		Cosmovisor: cosmovisor,
		Tendermint: tendermint,
	}
}

func (u *UpgradesQuerier) Enabled() bool {
	return u.Cosmovisor != nil
}

func (u *UpgradesQuerier) Name() string {
	return "upgrades-querier"
}

func (u *UpgradesQuerier) Get() ([]prometheus.Collector, []QueryInfo) {
	cosmovisorQuery := QueryInfo{
		Action:  "cosmovisor_get_upgrade_plan",
		Success: false,
	}

	upgrade, err := u.Cosmovisor.GetUpgradePlan()
	if err != nil {
		u.Logger.Err(err).Msg("Could not get latest Cosmovisor upgrade plan")
		return []prometheus.Collector{}, []QueryInfo{cosmovisorQuery}
	}

	cosmovisorQuery.Success = true
	isUpgradePresent := upgrade.Name != ""

	upcomingUpgradePresent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "upgrade_coming",
			Help: "Is future upgrade planned?",
		},
		[]string{},
	)

	upcomingUpgradePresent.
		With(prometheus.Labels{}).
		Set(BoolToFloat64(isUpgradePresent))

	queries := []prometheus.Collector{upcomingUpgradePresent}
	queryInfos := []QueryInfo{cosmovisorQuery}

	if isUpgradePresent {
		upgradeInfoGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: MetricsPrefix + "upgrade_info",
				Help: "Future upgrade info",
			},
			[]string{"name", "info"},
		)

		upgradeInfoGauge.
			With(prometheus.Labels{"name": upgrade.Name, "info": upgrade.Info}).
			Set(BoolToFloat64(isUpgradePresent))
		queries = append(queries, upgradeInfoGauge)

		upgradeEstimatedTimeGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: MetricsPrefix + "upgrade_estimated_time",
				Help: "Estimated upgrade time, as Unix timestamp",
			},
			[]string{"name", "info"},
		)

		upgradeTime := upgrade.Time
		if upgradeTime.IsZero() {
			if u.Tendermint == nil {
				u.Logger.Warn().
					Msg("Tendermint RPC not initialized and upgrade time is not specified, not returning upgrade time.")
				return queries, queryInfos
			}

			upgradeHeight, err := strconv.ParseInt(upgrade.Height, 10, 64)
			if err != nil {
				u.Logger.Err(err).Msg("Could not convert expected upgrade height to int64")
				return queries, queryInfos
			}

			tendermintQuery := QueryInfo{
				Action:  "tendermint_get_upgrade_time",
				Success: false,
			}

			upgradeTime, err = u.Tendermint.GetEstimateTimeTillBlock(upgradeHeight)
			if err != nil {
				u.Logger.Err(err).Msg("Could not get estimated upgrade time")
				queryInfos = append(queryInfos, tendermintQuery)
				return queries, queryInfos
			}
			tendermintQuery.Success = true
			queryInfos = append(queryInfos, tendermintQuery)

		}

		upgradeEstimatedTimeGauge.
			With(prometheus.Labels{"name": upgrade.Name, "info": upgrade.Info}).
			Set(float64(upgradeTime.Unix()))
		queries = append(queries, upgradeEstimatedTimeGauge)
	}

	return queries, queryInfos
}
