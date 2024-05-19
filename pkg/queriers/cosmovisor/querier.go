package cosmovisor

import (
	"context"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	configPkg "main/pkg/config"
	"main/pkg/metrics"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type Querier struct {
	Config     configPkg.NodeConfig
	Logger     zerolog.Logger
	Cosmovisor *cosmovisorPkg.Cosmovisor
	Tracer     trace.Tracer
}

func NewQuerier(
	logger zerolog.Logger,
	cosmovisor *cosmovisorPkg.Cosmovisor,
	tracer trace.Tracer,
) *Querier {
	return &Querier{
		Logger:     logger.With().Str("component", "cosmovisor_querier").Logger(),
		Cosmovisor: cosmovisor,
		Tracer:     tracer,
	}
}

func (v *Querier) Enabled() bool {
	return v.Cosmovisor != nil
}

func (v *Querier) Name() string {
	return "cosmovisor-querier"
}

func (v *Querier) Get(ctx context.Context) ([]metrics.MetricInfo, []query_info.QueryInfo) {
	childCtx, span := v.Tracer.Start(
		ctx,
		"Querier "+v.Name(),
		trace.WithAttributes(attribute.String("node", v.Name())),
	)
	defer span.End()

	cosmovisorVersion, queryInfo, err := v.Cosmovisor.GetCosmovisorVersion(childCtx)
	if err != nil {
		v.Logger.Err(err).Msg("Could not get Cosmovisor version")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{queryInfo}
	}

	return []metrics.MetricInfo{
		{
			MetricName: metrics.MetricNameCosmovisorVersion,
			Labels:     map[string]string{"version": cosmovisorVersion},
			Value:      1,
		},
	}, []query_info.QueryInfo{queryInfo}
}
