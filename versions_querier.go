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
	return v.Github != nil || v.Cosmovisor != nil
}

func (v *VersionsQuerier) Name() string {
	return "versions-querier"
}

func (v *VersionsQuerier) Get() ([]prometheus.Collector, []QueryInfo) {
	queriesInfo := []QueryInfo{}
	collectors := []prometheus.Collector{}

	var (
		releaseInfo ReleaseInfo
		versionInfo VersionInfo
		err         error
	)

	if v.Github != nil {
		queriesInfo = append(queriesInfo, QueryInfo{
			Action:  "github_get_latest_release",
			Success: false,
		})

		releaseInfo, err = v.Github.GetLatestRelease()
		if err != nil {
			v.Logger.Err(err).Msg("Could not get latest Github version")
			return []prometheus.Collector{}, queriesInfo
		}

		if releaseInfo.TagName == "" {
			v.Logger.Err(err).Msg("Malformed Github response when querying version")
			return []prometheus.Collector{}, queriesInfo
		}

		// stripping first "v" character: "v1.2.3" => "1.2.3"
		if releaseInfo.TagName[0] == 'v' {
			releaseInfo.TagName = releaseInfo.TagName[1:]
		}

		queriesInfo[len(queriesInfo)-1].Success = true

		remoteVersion := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: MetricsPrefix + "remote_version",
				Help: "Latest version from Github",
			},
			[]string{"version"},
		)

		remoteVersion.
			With(prometheus.Labels{"version": releaseInfo.TagName}).
			Set(1)

		collectors = append(collectors, remoteVersion)
	}

	if v.Cosmovisor != nil {
		queriesInfo = append(queriesInfo, QueryInfo{
			Action:  "cosmovisor_get_version",
			Success: false,
		})

		versionInfo, err = v.Cosmovisor.GetVersion()
		if err != nil {
			v.Logger.Err(err).Msg("Could not get app version")
			return []prometheus.Collector{}, queriesInfo
		}

		queriesInfo[len(queriesInfo)-1].Success = true

		localVersion := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: MetricsPrefix + "local_version",
				Help: "Fullnode local version",
			},
			[]string{"version"},
		)

		localVersion.
			With(prometheus.Labels{"version": versionInfo.Version}).
			Set(1)

		collectors = append(collectors, localVersion)
	}

	if v.Github != nil && v.Cosmovisor != nil {
		semverLocal, err := semver.NewVersion(versionInfo.Version)
		if err != nil {
			v.Logger.Err(err).Msg("Could not get local app version")
			return collectors, queriesInfo
		}

		semverConstraint, err := semver.NewConstraint(fmt.Sprintf(">= %s", releaseInfo.TagName))
		if err != nil {
			v.Logger.Err(err).Msg("Could not get remote app version")
			return collectors, queriesInfo
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
	}

	return collectors, queriesInfo
}
