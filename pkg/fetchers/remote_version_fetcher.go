package fetchers

import (
	"context"
	"main/pkg/clients/git"
	"main/pkg/constants"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type RemoteVersionFetcher struct {
	GitClient git.Client
	Logger    zerolog.Logger
	Tracer    trace.Tracer
}

func NewRemoteVersionFetcher(logger zerolog.Logger, gitClient git.Client, tracer trace.Tracer) *RemoteVersionFetcher {
	return &RemoteVersionFetcher{
		Logger:    logger.With().Str("component", "remote_version_fetcher").Logger(),
		GitClient: gitClient,
		Tracer:    tracer,
	}
}

func (n *RemoteVersionFetcher) Enabled() bool {
	return n.GitClient != nil
}

func (n *RemoteVersionFetcher) Name() constants.FetcherName {
	return constants.FetcherNameRemoteVersion
}

func (n *RemoteVersionFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (f *RemoteVersionFetcher) Get(ctx context.Context) (interface{}, []query_info.QueryInfo) {
	childCtx, span := f.Tracer.Start(
		ctx,
		"Fetcher "+string(f.Name()),
	)
	defer span.End()

	latestVersion, queryInfo, latestVersionError := f.GitClient.GetLatestRelease(childCtx)
	if latestVersionError != nil {
		f.Logger.Err(latestVersionError).Msg("Could not get latest Git version")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	if latestVersion[0] == 'v' {
		latestVersion = latestVersion[1:]
	}

	return latestVersion, []query_info.QueryInfo{queryInfo}
}
