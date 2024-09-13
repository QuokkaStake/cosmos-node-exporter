package generators

import (
	"main/pkg/clients/tendermint"
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
	"main/pkg/utils"
)

type NodeStatusGenerator struct{}

func NewNodeStatusGenerator() *NodeStatusGenerator {
	return &NodeStatusGenerator{}
}

func (g *NodeStatusGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	statusRaw, ok := state[constants.FetcherNameNodeStatus]
	if !ok || statusRaw == nil {
		return []metrics.MetricInfo{}
	}

	status, ok := statusRaw.(tendermint.StatusResponse)
	if !ok {
		panic("expected the state entry to be tendermint.StatusResponse")
	}

	return []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCatchingUp,
			Labels:     map[string]string{},
			Value:      utils.BoolToFloat64(status.Result.SyncInfo.CatchingUp),
		},
		{
			MetricName: metrics.MetricNameLatestBlockTime,
			Labels:     map[string]string{},
			Value:      float64(status.Result.SyncInfo.LatestBlockTime.Unix()),
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
		{
			MetricName: metrics.MetricNameVotingPower,
			Labels:     map[string]string{},
			Value:      float64(status.Result.ValidatorInfo.VotingPower),
		},
	}
}
