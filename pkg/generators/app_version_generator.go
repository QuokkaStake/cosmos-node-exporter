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
	version, found := fetchers.StateGet[string](state, constants.FetcherNameAppVersion)
	if !found {
		return []metrics.MetricInfo{}
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameAppVersion,
		Labels:     map[string]string{"version": version},
		Value:      1,
	}}
}
