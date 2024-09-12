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

type CosmovisorUpgradesFetcher struct {
	Config     configPkg.NodeConfig
	Logger     zerolog.Logger
	Cosmovisor *cosmovisorPkg.Cosmovisor
	Tracer     trace.Tracer
}

func NewCosmovisorUpgradesFetcher(
	logger zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tracer trace.Tracer,
) *CosmovisorUpgradesFetcher {
	return &CosmovisorUpgradesFetcher{
		Logger:     logger.With().Str("component", "cosmovisor_upgrades").Logger(),
		Cosmovisor: cosmovisor,
		Tracer:     tracer,
	}
}

func (v *CosmovisorUpgradesFetcher) Enabled() bool {
	return v.Cosmovisor != nil
}

func (v *CosmovisorUpgradesFetcher) Name() constants.FetcherName {
	return constants.FetcherNameCosmovisorUpgrades
}

func (v *CosmovisorUpgradesFetcher) Dependencies() []constants.FetcherName {
	return []constants.FetcherName{}
}

func (v *CosmovisorUpgradesFetcher) Get(ctx context.Context, data ...interface{}) (interface{}, []query_info.QueryInfo) {
	childCtx, span := v.Tracer.Start(
		ctx,
		"CosmovisorUpgradesFetcher "+string(v.Name()),
		trace.WithAttributes(attribute.String("node", string(v.Name()))),
	)
	defer span.End()

	cosmovisorUpgrades, queryInfo, err := v.Cosmovisor.GetUpgrades(childCtx)
	if err != nil {
		v.Logger.Err(err).Msg("Could not get Cosmovisor upgrades")
		return nil, []query_info.QueryInfo{queryInfo}
	}

	return cosmovisorUpgrades, []query_info.QueryInfo{queryInfo}
}
