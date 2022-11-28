package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type CosmovisorQuerier struct {
	Logger     zerolog.Logger
	Cosmovisor *Cosmovisor
}

func NewCosmovisorQuerier(logger *zerolog.Logger, cosmovisor *Cosmovisor) *CosmovisorQuerier {
	return &CosmovisorQuerier{
		Logger:     logger.With().Str("component", "cosmovisor_querier").Logger(),
		Cosmovisor: cosmovisor,
	}
}

func (c *CosmovisorQuerier) Enabled() bool {
	return c.Cosmovisor != nil
}

func (c *CosmovisorQuerier) Name() string {
	return "cosmovisor-querier"
}

func (c *CosmovisorQuerier) Get() ([]prometheus.Collector, []QueryInfo) {
	queryInfo := QueryInfo{
		Action:  "cosmovisor_get_upgrades",
		Success: false,
	}

	upgrades, err := c.Cosmovisor.GetUpgrades()
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not get Cosmovisor upgrades")
		return []prometheus.Collector{}, []QueryInfo{queryInfo}
	}

	upgradeBinaryPresent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "upgrade_binary_present",
			Help: "Is upgrade binary present?",
		},
		[]string{"name"},
	)

	for _, upgrade := range upgrades {
		upgradeBinaryPresent.
			With(prometheus.Labels{"name": upgrade.Name}).
			Set(BoolToFloat64(upgrade.BinaryPresent))
	}

	queryInfo.Success = true

	return []prometheus.Collector{
		upgradeBinaryPresent,
	}, []QueryInfo{queryInfo}
}
