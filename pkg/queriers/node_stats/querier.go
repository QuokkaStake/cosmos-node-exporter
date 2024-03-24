package node_stats

import (
	"main/pkg/clients/tendermint"
	"main/pkg/metrics"
	"main/pkg/query_info"
	"main/pkg/utils"
	"time"

	"github.com/rs/zerolog"
)

type Querier struct {
	TendermintRPC *tendermint.RPC
	Logger        zerolog.Logger
}

func NewQuerier(logger zerolog.Logger, tendermintRPC *tendermint.RPC) *Querier {
	return &Querier{
		Logger:        logger.With().Str("component", "node-stats-querier").Logger(),
		TendermintRPC: tendermintRPC,
	}
}

func (n *Querier) Enabled() bool {
	return n.TendermintRPC != nil
}

func (n *Querier) Name() string {
	return "node-stats-querier"
}

func (n *Querier) Get() ([]metrics.MetricInfo, []query_info.QueryInfo) {
	status, queryInfo, err := n.TendermintRPC.Status()
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node status")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{queryInfo}
	}

	querierMetrics := []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCatchingUp,
			Labels:     map[string]string{},
			Value:      utils.BoolToFloat64(status.Result.SyncInfo.CatchingUp),
		},
		{
			MetricName: metrics.MetricNameTimeSinceLatestBlock,
			Labels:     map[string]string{},
			Value:      time.Since(status.Result.SyncInfo.LatestBlockTime).Seconds(),
		},
		{
			MetricName: metrics.MetricNameNodeInfo,
			Labels: map[string]string{
				"moniker": status.Result.NodeInfo.Moniker,
				"chain":   status.Result.NodeInfo.Network,
			},
			Value: 1,
		},
		{
			MetricName: metrics.MetricNameTendermintVersion,
			Labels:     map[string]string{"version": status.Result.NodeInfo.Version},
			Value:      1,
		},
	}

	if value, err := utils.StringToFloat64(status.Result.ValidatorInfo.VotingPower); err != nil {
		n.Logger.Error().Err(err).
			Msg("Got error when converting voting power to float64, which should never happen.")
	} else {
		querierMetrics = append(querierMetrics, metrics.MetricInfo{
			MetricName: metrics.MetricNameVotingPower,
			Labels:     map[string]string{},
			Value:      value,
		})
	}

	return querierMetrics, []query_info.QueryInfo{queryInfo}
}
