package app

import (
	"context"
	"main/pkg/metrics"
	"main/pkg/query_info"
)

type Querier struct {
	Version string
}

func NewQuerier(version string) *Querier {
	return &Querier{
		Version: version,
	}
}

func (u *Querier) Enabled() bool {
	return true
}

func (u *Querier) Name() string {
	return "app-querier"
}

func (u *Querier) Get(ctx context.Context) ([]metrics.MetricInfo, []query_info.QueryInfo) {
	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameAppVersion,
		Labels:     map[string]string{"version": u.Version},
		Value:      1,
	}}, []query_info.QueryInfo{}
}
