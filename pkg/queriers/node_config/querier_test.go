package node_stats

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	nodeTypes "github.com/cosmos/cosmos-sdk/client/grpc/node"

	"github.com/stretchr/testify/require"
	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stretchr/testify/assert"
)

func TestNodeConfigQuerierBase(t *testing.T) {
	t.Parallel()

	config := configPkg.GrpcConfig{
		Address: "localhost:9090",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := grpcPkg.NewClient(config, *logger, tracer)
	querier := NewQuerier(*logger, client, tracer)
	assert.True(t, querier.Enabled())
	assert.Equal(t, "node-config-querier", querier.Name())
}
func TestNodeConfigQuerierFail(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn()(t)

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"invalid",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &grpcPkg.Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

func TestNodeConfigQuerierError(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn()(t)

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"invalid",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &grpcPkg.Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

func TestNodeConfigQuerierNotImplemented(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn()(t)

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &grpcPkg.Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

func TestNodeConfigQuerierInvalidResponse(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn(
		grpcmock.RegisterServiceFromMethods(service.Method{
			ServiceName: "cosmos.base.node.v1beta1.Service",
			MethodName:  "Config",
			MethodType:  service.TypeUnary,
			Input:       &nodeTypes.ConfigRequest{},
			Output:      &nodeTypes.ConfigResponse{},
		}),
		func(s *grpcmock.Server) {
			s.ExpectUnary("cosmos.base.node.v1beta1.Service/Config").
				Return(&nodeTypes.ConfigResponse{MinimumGasPrice: "test"})
		},
	)(t)

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &grpcPkg.Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

func TestNodeConfigQuerierOk(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn(
		grpcmock.RegisterServiceFromMethods(service.Method{
			ServiceName: "cosmos.base.node.v1beta1.Service",
			MethodName:  "Config",
			MethodType:  service.TypeUnary,
			Input:       &nodeTypes.ConfigRequest{},
			Output:      &nodeTypes.ConfigResponse{},
		}),
		func(s *grpcmock.Server) {
			s.ExpectUnary("cosmos.base.node.v1beta1.Service/Config").
				Return(&nodeTypes.ConfigResponse{MinimumGasPrice: "0.1uatom,0.2ustake", HaltHeight: 123})
		},
	)(t)

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &grpcPkg.Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Len(t, metrics, 4)

	pricesCount := metrics[0]
	assert.Empty(t, pricesCount.Labels)
	assert.InDelta(t, 2, pricesCount.Value, 0.01)

	firstPrice := metrics[1]
	assert.Equal(t, map[string]string{
		"denom": "uatom",
	}, firstPrice.Labels)
	assert.InDelta(t, 0.1, firstPrice.Value, 0.01)

	secondPrice := metrics[2]
	assert.Equal(t, map[string]string{
		"denom": "ustake",
	}, secondPrice.Labels)
	assert.InDelta(t, 0.2, secondPrice.Value, 0.01)

	haltHeight := metrics[3]
	assert.Equal(t, map[string]string{}, haltHeight.Labels)
	assert.InDelta(t, 123, haltHeight.Value, 0.01)
}
