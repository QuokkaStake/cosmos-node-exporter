package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/Masterminds/semver"
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
		Logger:     logger.With().Str("component", "versions_querier").Logger(),
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

func (v *VersionsQuerier) Get() ([]prometheus.Collector, []QueryInfo) {
	githubQuery := QueryInfo{
		Action:  "github_get_latest_release",
		Success: false,
	}

	releaseInfo, err := v.Github.GetLatestRelease()
	if err != nil {
		v.Logger.Err(err).Msg("Could not get latest Github version")
		return []prometheus.Collector{}, []QueryInfo{githubQuery}
	}

	// stripping first "v" character: "v1.2.3" => "1.2.3"
	if releaseInfo.TagName[0] == 'v' {
		releaseInfo.TagName = releaseInfo.TagName[1:]
	}

	githubQuery.Success = true
	localVersionQuery := QueryInfo{
		Action:  "cosmovisor_get_version",
		Success: false,
	}

	versionInfo, err := v.Cosmovisor.GetVersion()
	if err != nil {
		v.Logger.Err(err).Msg("Could not get app version")
		return []prometheus.Collector{}, []QueryInfo{githubQuery, localVersionQuery}
	}

	localVersionQuery.Success = true

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

	collectors := []prometheus.Collector{
		remoteVersion,
		localVersion,
	}

	semverLocal, err := semver.NewVersion(versionInfo.Version)
	if err != nil {
		v.Logger.Err(err).Msg("Could not get local app version")
		return collectors, []QueryInfo{githubQuery, localVersionQuery}
	}

	semverConstraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", releaseInfo.TagName))
	if err != nil {
		v.Logger.Err(err).Msg("Could not get remote app version")
		return collectors, []QueryInfo{githubQuery, localVersionQuery}
	}

	isLatestOrSameVersion := semverConstraint.Check(semverLocal)

	isUsingLatestVersion := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricsPrefix + "is_latest",
			Help: "Is the fullnode using the same or latest version?",
		},
		[]string{"local_version", "remote_version"},
	)

	isUsingLatestVersion.
		With(prometheus.Labels{
			"local_version":  versionInfo.Version,
			"remote_version": releaseInfo.TagName,
		}).
		Set(BoolToFloat64(isLatestOrSameVersion))

	collectors = append(collectors, isUsingLatestVersion)
	return collectors, []QueryInfo{githubQuery, localVersionQuery}
}
