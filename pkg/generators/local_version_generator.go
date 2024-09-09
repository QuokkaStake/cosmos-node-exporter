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
	remoteVersionRaw, ok := state[constants.FetcherNameLocalVersion]
	if !ok || remoteVersionRaw == nil {
		return []metrics.MetricInfo{}
	}

	versionInfo, ok := remoteVersionRaw.(types.VersionInfo)
	if !ok {
		panic("expected the state entry to be string")
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameLocalVersion,
		Labels:     map[string]string{"version": versionInfo.Version},
		Value:      1,
	}}
}
