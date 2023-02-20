package node_stats

import (
	"main/pkg/constants"
	"main/pkg/query_info"
	"main/pkg/tendermint"
	"main/pkg/utils"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type NodeStatsQuerier struct {
	TendermintRPC *tendermint.TendermintRPC
	Logger        zerolog.Logger
}

func NewNodeStatsQuerier(logger *zerolog.Logger, tendermintRPC *tendermint.TendermintRPC) *NodeStatsQuerier {
	return &NodeStatsQuerier{
		Logger:        logger.With().Str("component", "tendermint_rpc").Logger(),
		TendermintRPC: tendermintRPC,
	}
}

func (n *NodeStatsQuerier) Enabled() bool {
	return n.TendermintRPC != nil
}

func (n *NodeStatsQuerier) Name() string {
	return "node-stats-querier"
}

func (n *NodeStatsQuerier) Get() ([]prometheus.Collector, []query_info.QueryInfo) {
	queryInfo := query_info.QueryInfo{
		Module:  "tendermint",
		Action:  "node_status",
		Success: false,
	}

	status, err := n.TendermintRPC.Status()
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node status")
		return []prometheus.Collector{}, []query_info.QueryInfo{queryInfo}
	}

	catchingUpGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "catching_up",
			Help: "Is node catching up?",
		},
		[]string{},
	)

	timeSinceLatestBlockGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "time_since_latest_block",
			Help: "Time since latest block, in seconds",
		},
		[]string{},
	)

	votingPowerGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "voting_power",
			Help: "Node voting power",
		},
		[]string{},
	)

	catchingUpGauge.
		With(prometheus.Labels{}).
		Set(utils.BoolToFloat64(status.Result.SyncInfo.CatchingUp))

	timeSinceLatestBlockGauge.
		With(prometheus.Labels{}).
		Set(time.Since(status.Result.SyncInfo.LatestBlockTime).Seconds())

	if value, err := utils.StringToFloat64(status.Result.ValidatorInfo.VotingPower); err != nil {
		n.Logger.Error().Err(err).
			Msg("Got error when converting voting power to float64, which should never happen.")
	} else {
		votingPowerGauge.
			With(prometheus.Labels{}).
			Set(value)
	}

	queryInfo.Success = true

	return []prometheus.Collector{
		catchingUpGauge,
		timeSinceLatestBlockGauge,
		votingPowerGauge,
	}, []query_info.QueryInfo{queryInfo}
}
