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
	statusRaw, ok := state[constants.FetcherNameCosmovisorVersion]
	if !ok || statusRaw == nil {
		return []metrics.MetricInfo{}
	}

	version, ok := statusRaw.(string)
	if !ok {
		panic("expected the state entry to be string")
	}

	return []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCosmovisorVersion,
			Labels:     map[string]string{"version": version},
			Value:      1,
		},
	}
}
