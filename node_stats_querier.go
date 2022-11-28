package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type NodeStatsQuerier struct {
	TendermintRPC *TendermintRPC
	Logger        zerolog.Logger
}

func NewNodeStatsQuerier(logger *zerolog.Logger, tenderminRPC *TendermintRPC) *NodeStatsQuerier {
	return &NodeStatsQuerier{
		Logger:        logger.With().Str("component", "tendermint_rpc").Logger(),
		TendermintRPC: tenderminRPC,
	}
}

func (n *NodeStatsQuerier) Enabled() bool {
	return n.TendermintRPC != nil
}

func (n *NodeStatsQuerier) Name() string {
	return "node-stats-querier"
}

func (n *NodeStatsQuerier) Get() ([]prometheus.Collector, []QueryInfo) {
	queryInfo := QueryInfo{
		Action:  "node_status",
		Success: false,
	}

	status, err := n.TendermintRPC.GetStatus()
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node status")
		return []prometheus.Collector{}, []QueryInfo{queryInfo}
	}

	catchingUpGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "catching_up",
			Help: "Is node catching up?",
		},
		[]string{},
	)

	timeSinceLatestBlockGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "time_since_latest_block",
			Help: "Time since latest block, in seconds",
		},
		[]string{},
	)

	votingPowerGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "voting_power",
			Help: "Node voting power",
		},
		[]string{},
	)

	catchingUpGauge.
		With(prometheus.Labels{}).
		Set(BoolToFloat64(status.SyncInfo.CatchingUp))

	timeSinceLatestBlockGauge.
		With(prometheus.Labels{}).
		Set(time.Since(status.SyncInfo.LatestBlockTime).Seconds())

	votingPowerGauge.
		With(prometheus.Labels{}).
		Set(float64(status.ValidatorInfo.VotingPower))

	queryInfo.Success = true

	return []prometheus.Collector{
		catchingUpGauge,
		timeSinceLatestBlockGauge,
		votingPowerGauge,
	}, []QueryInfo{queryInfo}
}
