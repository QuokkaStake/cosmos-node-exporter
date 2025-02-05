package generators

import (
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/metrics"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/Masterminds/semver"
	"github.com/rs/zerolog"
)

type IsLatestGenerator struct {
	Logger zerolog.Logger
}

func NewIsLatestGenerator(logger zerolog.Logger) *IsLatestGenerator {
	return &IsLatestGenerator{
		Logger: logger.With().Str("component", "is_latest_generator").Logger(),
	}
}

func (g *IsLatestGenerator) Get(state fetchers.State) []metrics.MetricInfo {
	localVersion, localVersionFound := fetchers.StateGet[types.VersionInfo](state, constants.FetcherNameLocalVersion)
	if !localVersionFound {
		return []metrics.MetricInfo{}
	}

	remoteVersion, remoteVersionFound := fetchers.StateGet[string](state, constants.FetcherNameRemoteVersion)
	if !remoteVersionFound {
		return []metrics.MetricInfo{}
	}

	semverLocal, err := semver.NewVersion(localVersion.Version)
	if err != nil {
		g.Logger.Err(err).Msg("Could not get local app version")
		return []metrics.MetricInfo{}
	}

	semverRemote, err := semver.NewVersion(remoteVersion)
	if err != nil {
		g.Logger.Err(err).Msg("Could not get remote app version")
		return []metrics.MetricInfo{}
	}

	// 0 is for equal, 1 is when the local version is greater
	isLatestOrSameVersion := semverLocal.Compare(semverRemote) >= 0

	return []metrics.MetricInfo{{
		MetricName: metrics.MetricNameIsLatest,
		Labels: map[string]string{
			"local_version":  localVersion.Version,
			"remote_version": remoteVersion,
		},
		Value: utils.BoolToFloat64(isLatestOrSameVersion),
	}}
}
