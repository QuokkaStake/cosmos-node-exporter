package fetchers

import (
	"context"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type CosmovisorVersionFetcher struct {
	Config     configPkg.NodeConfig
	Logger     zerolog.Logger
	Cosmovisor *cosmovisorPkg.Cosmovisor
	Tracer     trace.Tracer
}

func NewCosmovisorVersionFetcher(
	logger zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tracer trace.Tracer,
) *CosmovisorVersionFetcher {
	return &CosmovisorVersionFetcher{
		Logger:     logger.With().Str("component", "cosmovisor_fetcher").Logger(),
		Cosmovisor: cosmovisor,
		Tracer:     tracer,
	}
}

func (v *CosmovisorVersionFetcher) Enabled() bool {
	return v.Cosmovisor != nil
}

func (v *CosmovisorVersionFetcher) Name() constants.FetcherName {
	return constants.FetcherNameCosmovisorVersion
}

func (v *CosmovisorVersionFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (v *CosmovisorVersionFetcher) Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo) {
	childCtx, span := v.Tracer.Start(
		ctx,
		"CosmovisorVersionFetcher "+string(v.Name()),
		trace.WithAttributes(attribute.String("node", string(v.Name()))),
	)
	defer span.End()

	cosmovisorVersion, queryInfo, err := v.Cosmovisor.GetCosmovisorVersion(childCtx)
	if err != nil {
		v.Logger.Err(err).Msg("Could not get Cosmovisor version")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return cosmovisorVersion, []query_info.QueryInfo{queryInfo}
}
