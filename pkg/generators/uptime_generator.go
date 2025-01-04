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
	startTime, startTimeFound := fetchers.StateGet[time.Time](state, constants.FetcherNameUptime)
	if !startTimeFound {
		return []metrics.MetricInfo{}
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameStartTime,
		Labels:     map[string]string{},
		Value:      float64(startTime.Unix()),
	}}
}
