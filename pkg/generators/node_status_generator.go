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
	status, statusFound := fetchers.StateGet[tendermint.StatusResponse](state, constants.FetcherNameNodeStatus)
	if !statusFound {
		return []metrics.MetricInfo{}
	}

	return []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCatchingUp,
			Labels:     map[string]string{},
			Value:      utils.BoolToFloat64(status.Result.SyncInfo.CatchingUp),
		},
		{
			MetricName: metrics.MetricNameLatestBlockHeight,
			Labels:     map[string]string{},
			Value:      float64(status.Result.SyncInfo.LatestBlockHeight),
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
