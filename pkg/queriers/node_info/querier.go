package node_stats

import (
	grpcPkg "main/pkg/clients/grpc"
	"main/pkg/metrics"
	"main/pkg/query_info"

	"github.com/rs/zerolog"
)

type Querier struct {
	gRPC   *grpcPkg.Client
	Logger zerolog.Logger
}

func NewQuerier(logger zerolog.Logger, grpc *grpcPkg.Client) *Querier {
	return &Querier{
		Logger: logger.With().Str("component", "node-info-querier").Logger(),
		gRPC:   grpc,
	}
}

func (n *Querier) Enabled() bool {
	return n.gRPC != nil
}

func (n *Querier) Name() string {
	return "node-info-querier"
}

func (n *Querier) Get() ([]metrics.MetricInfo, []query_info.QueryInfo) {
	config, queryInfo, err := n.gRPC.GetNodeInfo()
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node info")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{queryInfo}
	}

	querierMetrics := []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCosmosSdkVersion,
			Labels:     map[string]string{"version": config.ApplicationVersion.CosmosSdkVersion},
			Value:      1,
		},
		{
			MetricName: metrics.MetricNameRunningAppVersion,
			Labels: map[string]string{
				"version":    config.ApplicationVersion.Version,
				"name":       config.ApplicationVersion.Name,
				"app_name":   config.ApplicationVersion.AppName,
				"git_commit": config.ApplicationVersion.GitCommit,
			},
			Value: 1,
		},
		{
			MetricName: metrics.MetricNameGoVersion,
			Labels: map[string]string{
				"version": config.ApplicationVersion.GoVersion,
				"tags":    config.ApplicationVersion.BuildTags,
			},
			Value: 1,
		},
	}

	return querierMetrics, []query_info.QueryInfo{queryInfo}
}
