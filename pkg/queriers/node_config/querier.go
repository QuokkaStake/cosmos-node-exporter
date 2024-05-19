package node_stats

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	"main/pkg/metrics"
	"main/pkg/query_info"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	cosmosTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
)

type Querier struct {
	gRPC   *grpcPkg.Client
	Logger zerolog.Logger
	Tracer trace.Tracer
}

func NewQuerier(logger zerolog.Logger, grpc *grpcPkg.Client, tracer trace.Tracer) *Querier {
	return &Querier{
		Logger: logger.With().Str("component", "node-config-querier").Logger(),
		gRPC:   grpc,
		Tracer: tracer,
	}
}

func (n *Querier) Enabled() bool {
	return n.gRPC != nil
}

func (n *Querier) Name() string {
	return "node-config-querier"
}

func (n *Querier) Get(ctx context.Context) ([]metrics.MetricInfo, []query_info.QueryInfo) {
	childCtx, span := n.Tracer.Start(
		ctx,
		"Querier "+n.Name(),
		trace.WithAttributes(attribute.String("node", n.Name())),
	)
	defer span.End()

	config, queryInfo, err := n.gRPC.GetNodeConfig(childCtx)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Could not fetch node config")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{queryInfo}
	}

	if config == nil {
		n.Logger.Debug().
			Msg("Node config is nil, probably chain does not implement the node config endpoint.")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{queryInfo}
	}

	coinsParsed, err := cosmosTypes.ParseDecCoins(config.MinimumGasPrice)
	if err != nil {
		n.Logger.Error().Err(err).Msg("Error decoding minimum gas prices")
		return []metrics.MetricInfo{}, []query_info.QueryInfo{queryInfo}
	}

	querierMetrics := make([]metrics.MetricInfo, 1+len(coinsParsed))
	querierMetrics[0] = metrics.MetricInfo{
		MetricName: metrics.MetricNameMinimumGasPricesCount,
		Labels:     map[string]string{},
		Value:      float64(len(coinsParsed)),
	}

	for index, amount := range coinsParsed {
		querierMetrics[index+1] = metrics.MetricInfo{
			MetricName: metrics.MetricNameMinimumGasPrice,
			Labels:     map[string]string{"denom": amount.Denom},
			Value:      amount.Amount.MustFloat64(),
		}
	}

	return querierMetrics, []query_info.QueryInfo{queryInfo}
}
