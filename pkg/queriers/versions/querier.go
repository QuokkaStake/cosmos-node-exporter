package versions

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	cosmovisorPkg "main/pkg/cosmovisor"
	"main/pkg/git"
	"main/pkg/query_info"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/Masterminds/semver"
)

type Querier struct {
	Logger     zerolog.Logger
	GitClient  git.Client
	Config     configPkg.NodeConfig
	Cosmovisor *cosmovisorPkg.Cosmovisor
}

func NewQuerier(
	logger *zerolog.Logger,
	config configPkg.NodeConfig,
	gitClient git.Client,
	cosmovisor *cosmovisorPkg.Cosmovisor,
) *Querier {
	return &Querier{
		Logger:     logger.With().Str("component", "versions_querier").Logger(),
		GitClient:  gitClient,
		Config:     config,
		Cosmovisor: cosmovisor,
	}
}

func (v *Querier) Enabled() bool {
	return v.GitClient != nil || v.Cosmovisor != nil
}

func (v *Querier) Name() string {
	return "versions-querier"
}

func (v *Querier) Get() ([]prometheus.Collector, []query_info.QueryInfo) {
	queriesInfo := []query_info.QueryInfo{}
	collectors := []prometheus.Collector{}

	var (
		latestVersion string
		versionInfo   types.VersionInfo
		err           error
	)

	if v.GitClient != nil {
		queriesInfo = append(queriesInfo, query_info.QueryInfo{
			Module:  "git",
			Action:  "get_latest_release",
			Success: false,
		})

		latestVersion, err = v.GitClient.GetLatestRelease()
		if err != nil {
			v.Logger.Err(err).Msg("Could not get latest Git version")
			return []prometheus.Collector{}, queriesInfo
		}

		// stripping first "v" character: "v1.2.3" => "1.2.3"
		if latestVersion[0] == 'v' {
			latestVersion = latestVersion[1:]
		}

		queriesInfo[len(queriesInfo)-1].Success = true

		remoteVersion := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: constants.MetricsPrefix + "remote_version",
				Help: "Latest version from Github",
			},
			[]string{"node", "version"},
		)

		remoteVersion.
			With(prometheus.Labels{"node": v.Config.Name, "version": latestVersion}).
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
			[]string{"node", "version"},
		)

		localVersion.
			With(prometheus.Labels{"node": v.Config.Name, "version": versionInfo.Version}).
			Set(1)

		collectors = append(collectors, localVersion)
	}

	if v.GitClient != nil && v.Cosmovisor != nil {
		semverLocal, err := semver.NewVersion(versionInfo.Version)
		if err != nil {
			v.Logger.Err(err).Msg("Could not get local app version")
			return collectors, queriesInfo
		}

		semverRemote, err := semver.NewVersion(latestVersion)
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
			[]string{"node", "local_version", "remote_version"},
		)

		isUsingLatestVersion.
			With(prometheus.Labels{
				"node":           v.Config.Name,
				"local_version":  versionInfo.Version,
				"remote_version": latestVersion,
			}).
			Set(utils.BoolToFloat64(isLatestOrSameVersion))

		collectors = append(collectors, isUsingLatestVersion)
	}

	return collectors, queriesInfo
}
