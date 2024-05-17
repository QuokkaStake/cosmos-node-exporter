package versions

import (
	"context"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/git"
	"main/pkg/metrics"
	"main/pkg/query_info"
	"main/pkg/types"
	"main/pkg/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"

	"github.com/Masterminds/semver"
)

type Querier struct {
	Logger     zerolog.Logger
	GitClient  git.Client
	Cosmovisor *cosmovisorPkg.Cosmovisor
	Tracer     trace.Tracer
}

func NewQuerier(
	logger zerolog.Logger,
	gitClient git.Client,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tracer trace.Tracer,
) *Querier {
	return &Querier{
		Logger:     logger.With().Str("component", "versions_querier").Logger(),
		GitClient:  gitClient,
		Cosmovisor: cosmovisor,
		Tracer:     tracer,
	}
}

func (v *Querier) Enabled() bool {
	return v.GitClient != nil || v.Cosmovisor != nil
}

func (v *Querier) Name() string {
	return "versions-querier"
}

func (v *Querier) Get(ctx context.Context) ([]metrics.MetricInfo, []query_info.QueryInfo) {
	childCtx, span := v.Tracer.Start(
		ctx,
		"Querier "+v.Name(),
		trace.WithAttributes(attribute.String("node", v.Name())),
	)
	defer span.End()

	queriesInfo := []query_info.QueryInfo{}
	metricsInfos := []metrics.MetricInfo{}

	var (
		latestVersion              string
		versionInfo                types.VersionInfo
		gitQueryInfo               query_info.QueryInfo
		cosmovisorVersionQueryInfo query_info.QueryInfo
		err                        error
	)

	if v.GitClient != nil {
		latestVersion, gitQueryInfo, err = v.GitClient.GetLatestRelease(childCtx)
		queriesInfo = append(queriesInfo, gitQueryInfo)
		if err != nil {
			v.Logger.Err(err).Msg("Could not get latest Git version")
			return []metrics.MetricInfo{}, queriesInfo
		} else {
			// stripping first "v" character: "v1.2.3" => "1.2.3"
			if latestVersion[0] == 'v' {
				latestVersion = latestVersion[1:]
			}

			metricsInfos = append(metricsInfos, metrics.MetricInfo{
				MetricName: metrics.MetricNameRemoteVersion,
				Labels:     map[string]string{"version": latestVersion},
				Value:      1,
			})
		}
	}

	if v.Cosmovisor != nil {
		versionInfo, cosmovisorVersionQueryInfo, err = v.Cosmovisor.GetVersion(childCtx)
		queriesInfo = append(queriesInfo, cosmovisorVersionQueryInfo)
		if err != nil {
			v.Logger.Err(err).Msg("Could not get app version")
		} else {
			metricsInfos = append(metricsInfos, metrics.MetricInfo{
				MetricName: metrics.MetricNameLocalVersion,
				Labels:     map[string]string{"version": versionInfo.Version},
				Value:      1,
			})
		}
	}

	if v.GitClient != nil && v.Cosmovisor != nil {
		semverLocal, err := semver.NewVersion(versionInfo.Version)
		if err != nil {
			v.Logger.Err(err).Msg("Could not get local app version")
			return metricsInfos, queriesInfo
		}

		semverRemote, err := semver.NewVersion(latestVersion)
		if err != nil {
			v.Logger.Err(err).Msg("Could not get remote app version")
			return metricsInfos, queriesInfo
		}

		// 0 is for equal, 1 is when the local version is greater
		isLatestOrSameVersion := semverLocal.Compare(semverRemote) >= 0

		metricsInfos = append(metricsInfos, metrics.MetricInfo{
			MetricName: metrics.MetricNameIsLatest,
			Labels: map[string]string{
				"local_version":  versionInfo.Version,
				"remote_version": latestVersion,
			},
			Value: utils.BoolToFloat64(isLatestOrSameVersion),
		})
	}

	return metricsInfos, queriesInfo
}
