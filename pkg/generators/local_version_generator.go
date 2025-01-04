package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
	"main/pkg/types"
)

type LocalVersionGenerator struct{}

func NewLocalVersionGenerator() *LocalVersionGenerator {
	return &LocalVersionGenerator{}
}

func (g *LocalVersionGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	localVersion, localVersionFound := fetchers.StateGet[types.VersionInfo](state, constants.FetcherNameLocalVersion)
	if !localVersionFound {
		return []metrics.MetricInfo{}
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameLocalVersion,
		Labels:     map[string]string{"version": localVersion.Version},
		Value:      1,
	}}
}
