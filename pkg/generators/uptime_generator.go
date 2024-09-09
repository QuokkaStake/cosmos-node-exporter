package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
	"time"
)

type UptimeGenerator struct{}

func NewUptimeGenerator() *UptimeGenerator {
	return &UptimeGenerator{}
}

func (g *UptimeGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	statusRaw, ok := state[constants.FetcherNameUptime]
	if !ok || statusRaw == nil {
		return []metrics.MetricInfo{}
	}

	startTime, ok := statusRaw.(time.Time)
	if !ok {
		panic("expected the state entry to be time.Time")
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameStartTime,
		Labels:     map[string]string{},
		Value:      float64(startTime.Unix()),
	}}
}
