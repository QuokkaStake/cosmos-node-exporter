package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	metricsPkg "main/pkg/metrics"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

type NodeInfoGenerator struct{}

func NewNodeInfoGenerator() *NodeInfoGenerator {
	return &NodeInfoGenerator{}
}

func (g *NodeInfoGenerator) Get(state fetchers.State) []metricsPkg.MetricInfo {
	nodeInfo, nodeInfoFound := fetchers.StateGet[*cmtservice.GetNodeInfoResponse](state, constants.FetcherNameNodeInfo)
	if !nodeInfoFound {
		return []metricsPkg.MetricInfo{}
	}

	return []metricsPkg.MetricInfo{
		{
			MetricName: metricsPkg.MetricNameCosmosSdkVersion,
			Labels:     map[string]string{"version": nodeInfo.ApplicationVersion.CosmosSdkVersion},
			Value:      1,
		},
		{
			MetricName: metricsPkg.MetricNameRunningAppVersion,
			Labels: map[string]string{
				"version":    nodeInfo.ApplicationVersion.Version,
				"name":       nodeInfo.ApplicationVersion.Name,
				"app_name":   nodeInfo.ApplicationVersion.AppName,
				"git_commit": nodeInfo.ApplicationVersion.GitCommit,
			},
			Value: 1,
		},
		{
			MetricName: metricsPkg.MetricNameGoVersion,
			Labels: map[string]string{
				"version": nodeInfo.ApplicationVersion.GoVersion,
				"tags":    nodeInfo.ApplicationVersion.BuildTags,
			},
			Value: 1,
		},
	}
}
