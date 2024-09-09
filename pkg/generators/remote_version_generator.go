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
	remoteVersionRaw, ok := state[constants.FetcherNameRemoteVersion]
	if !ok || remoteVersionRaw == nil {
		return []metrics.MetricInfo{}
	}

	version, ok := remoteVersionRaw.(string)
	if !ok {
		panic("expected the state entry to be string")
	}

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameRemoteVersion,
		Labels:     map[string]string{"version": version},
		Value:      1,
	}}
}
