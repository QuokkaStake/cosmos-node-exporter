package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
)

type RemoteVersionGenerator struct{}

func NewRemoteVersionGenerator() *RemoteVersionGenerator {
	return &RemoteVersionGenerator{}
}

func (g *RemoteVersionGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	remoteVersion, remoteVersionFound := fetchers.StateGet[string](state, constants.FetcherNameRemoteVersion)
	if !remoteVersionFound {
		return []metrics.MetricInfo{}
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameRemoteVersion,
		Labels:     map[string]string{"version": remoteVersion},
		Value:      1,
	}}
}
