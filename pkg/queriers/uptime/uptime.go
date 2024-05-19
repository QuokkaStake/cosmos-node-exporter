package uptime

import (
	"context"
	"main/pkg/metrics"
	"main/pkg/query_info"
	"time"
)

type Querier struct {
	StartTime time.Time
}

func NewQuerier() *Querier {
	return &Querier{
		StartTime: time.Now(),
	}
}

func (u *Querier) Enabled() bool {
	return true
}

func (u *Querier) Name() string {
	return "uptime-querier"
}

func (u *Querier) Get(ctx context.Context) ([]metrics.MetricInfo, []query_info.QueryInfo) {
	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameStartTime,
		Labels:     map[string]string{},
		Value:      float64(u.StartTime.Unix()),
	}}, []query_info.QueryInfo{}
}
