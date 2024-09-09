package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
)

type AppVersionGenerator struct{}

func NewAppVersionGenerator() *AppVersionGenerator {
	return &AppVersionGenerator{}
}

func (g *AppVersionGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	statusRaw, ok := state[constants.FetcherNameAppVersion]
	if !ok || statusRaw == nil {
		return []metrics.MetricInfo{}
	}

	version, ok := statusRaw.(string)
	if !ok {
		panic("expected the state entry to be string")
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameAppVersion,
		Labels:     map[string]string{"version": version},
		Value:      1,
	}}
}
