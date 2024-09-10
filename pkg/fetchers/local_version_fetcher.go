package fetchers

import (
	"context"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/constants"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type LocalVersionFetcher struct {
	Cosmovisor *cosmovisorPkg.Cosmovisor
	Logger     zerolog.Logger
	Tracer     trace.Tracer
}

func NewLocalVersionFetcher(
	logger zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tracer trace.Tracer,
) *LocalVersionFetcher {
	return &LocalVersionFetcher{
		Logger:     logger.With().Str("component", "local_version_fetcher").Logger(),
		Cosmovisor: cosmovisor,
		Tracer:     tracer,
	}
}

func (n *LocalVersionFetcher) Enabled() bool {
	return n.Cosmovisor != nil
}

func (n *LocalVersionFetcher) Name() constants.FetcherName {
	return constants.FetcherNameLocalVersion
}

func (n *LocalVersionFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (f *LocalVersionFetcher) Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo) {
	childCtx, span := f.Tracer.Start(
		ctx,
		"Fetcher "+string(f.Name()),
	)
	defer span.End()

	versionInfo, queryInfo, cosmovisorErr := f.Cosmovisor.GetVersion(childCtx)
	if cosmovisorErr != nil {
		f.Logger.Err(cosmovisorErr).Msg("Could not get app version")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return versionInfo, []query_info.QueryInfo{queryInfo}
}
