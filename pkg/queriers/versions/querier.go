package versions

import (
	"main/pkg/constants"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/github"
	"main/pkg/query_info"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/Masterminds/semver"
)

type Querier struct {
	Logger     zerolog.Logger
	Github     *github.Github
	Cosmovisor *cosmovisorPkg.Cosmovisor
}

func NewQuerier(
	logger *zerolog.Logger,
	github *github.Github,
	cosmovisor *cosmovisorPkg.Cosmovisor,
) *Querier {
	return &Querier{
		Logger:     logger.With().Str("component", "versions_querier").Logger(),
		Github:     github,
		Cosmovisor: cosmovisor,
	}
}

func (v *Querier) Enabled() bool {
	return v.Github != nil || v.Cosmovisor != nil
}

func (v *Querier) Name() string {
	return "versions-querier"
}

func (v *Querier) Get() ([]prometheus.Collector, []query_info.QueryInfo) {
	queriesInfo := []query_info.QueryInfo{}
	collectors := []prometheus.Collector{}

	var (
		releaseInfo types.ReleaseInfo
		versionInfo types.VersionInfo
		err         error
	)

	if v.Github != nil {
		queriesInfo = append(queriesInfo, query_info.QueryInfo{
			Module:  "github",
			Action:  "get_latest_release",
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
				Name: constants.MetricsPrefix + "remote_version",
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
		queriesInfo = append(queriesInfo, query_info.QueryInfo{
			Module:  "cosmovisor",
			Action:  "get_version",
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
				Name: constants.MetricsPrefix + "local_version",
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

		semverRemote, err := semver.NewVersion(releaseInfo.TagName)
		if err != nil {
			v.Logger.Err(err).Msg("Could not get remote app version")
			return collectors, queriesInfo
		}

		// 0 is for equal, 1 is when the local version is greater
		isLatestOrSameVersion := semverLocal.Compare(semverRemote) >= 0

		isUsingLatestVersion := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "is_latest",
				Help: "Is the fullnode using the same or latest version?",
			},
			[]string{"local_version", "remote_version"},
		)

		isUsingLatestVersion.
			With(prometheus.Labels{
				"local_version":  versionInfo.Version,
				"remote_version": releaseInfo.TagName,
			}).
			Set(utils.BoolToFloat64(isLatestOrSameVersion))

		collectors = append(collectors, isUsingLatestVersion)
	}

	return collectors, queriesInfo
}
