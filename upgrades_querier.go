package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type UpgradesQuerier struct {
	Logger     zerolog.Logger
	Cosmovisor *Cosmovisor
}

func NewUpgradesQuerier(logger *zerolog.Logger, cosmovisor *Cosmovisor) *UpgradesQuerier {
	return &UpgradesQuerier{
		Logger:     logger.With().Str("component", "upgrades_querier").Logger(),
		Cosmovisor: cosmovisor,
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

	_, err := u.Cosmovisor.GetUpgradePlan()
	if err != nil {
		u.Logger.Err(err).Msg("Could not get latest Cosmovisor upgrade plan")
		return []prometheus.Collector{}, []QueryInfo{cosmovisorQuery}
	}

	cosmovisorQuery.Success = true

	upcomingUpgradePresent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "upgrade_coming",
			Help: "Is future upgrade planned?",
		},
		[]string{},
	)

	upcomingUpgradePresent.
		With(prometheus.Labels{}).
		Set(0)

	return []prometheus.Collector{
		upcomingUpgradePresent,
	}, []QueryInfo{cosmovisorQuery}
}
