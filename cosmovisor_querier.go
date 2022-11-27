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

func (c *CosmovisorQuerier) Get() []prometheus.Collector {
	_, err := c.Cosmovisor.GetUpgrades()
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not get Cosmovisor upgrades")
		return []prometheus.Collector{}
	}

	return []prometheus.Collector{}
}
