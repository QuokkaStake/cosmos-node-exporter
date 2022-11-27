package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type VersionsQuerier struct {
	Logger     zerolog.Logger
	Github     *Github
	Cosmovisor *Cosmovisor
}

func NewVersionsQuerier(
	logger *zerolog.Logger,
	github *Github,
	cosmovisor *Cosmovisor,
) *VersionsQuerier {
	return &VersionsQuerier{
		Logger:     logger.With().Str("component", "tendermint_rpc").Logger(),
		Github:     github,
		Cosmovisor: cosmovisor,
	}
}

func (v *VersionsQuerier) Enabled() bool {
	return v.Github != nil
}

func (v *VersionsQuerier) Name() string {
	return "versions-querier"
}

func (v *VersionsQuerier) Get() []prometheus.Collector {
	releaseInfo, err := v.Github.GetLatestRelease()
	if err != nil {
		v.Logger.Err(err).Msg("Could not get latest Github version")
		return []prometheus.Collector{}
	}

	versionInfo, err := v.Cosmovisor.GetVersion()
	if err != nil {
		v.Logger.Err(err).Msg("Could not get app version")
		return []prometheus.Collector{}
	}

	remoteVersion := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "remote_version",
			Help: "Latest version from Github",
		},
		[]string{"version"},
	)

	localVersion := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "local_version",
			Help: "Fullnode local version",
		},
		[]string{"version"},
	)

	remoteVersion.
		With(prometheus.Labels{"version": releaseInfo.TagName}).
		Set(1)

	localVersion.
		With(prometheus.Labels{"version": versionInfo.Version}).
		Set(1)

	return []prometheus.Collector{
		remoteVersion,
		localVersion,
	}
}
