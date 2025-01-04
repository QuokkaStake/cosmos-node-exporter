package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
)

type CosmovisorVersionGenerator struct{}

func NewCosmovisorVersionGenerator() *CosmovisorVersionGenerator {
	return &CosmovisorVersionGenerator{}
}

func (g *CosmovisorVersionGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	version, versionFound := fetchers.StateGet[string](state, constants.FetcherNameCosmovisorVersion)
	if !versionFound {
		return []metrics.MetricInfo{}
	}

	return []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCosmovisorVersion,
			Labels:     map[string]string{"version": version},
			Value:      1,
		},
	}
}
