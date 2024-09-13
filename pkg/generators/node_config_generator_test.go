package generators

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	"main/pkg/constants"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	nodeTypes "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNodeConfigGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewNodeConfigGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestNodeConfigGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameNodeConfig: 3,
	}

	generator := NewNodeConfigGenerator()
	generator.Get(state)
}

func TestNodeConfigGeneratorInvalidPrices(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

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

	fetcher := fetchers.NewNodeConfigFetcher(*logger, client, tracer)
	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameNodeConfig: data,
	}

	generator := NewNodeConfigGenerator()
	generator.Get(state)
}

func TestNodeConfigGeneratorOk(t *testing.T) {
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

	fetcher := fetchers.NewNodeConfigFetcher(*logger, client, tracer)
	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameNodeConfig: data,
	}

	generator := NewNodeConfigGenerator()

	metrics := generator.Get(state)
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
