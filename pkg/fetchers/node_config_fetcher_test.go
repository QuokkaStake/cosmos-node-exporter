package fetchers

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	configPkg "main/pkg/config"
	"main/pkg/constants"
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

func TestNodeConfigFetcherBase(t *testing.T) {
	t.Parallel()

	config := configPkg.GrpcConfig{
		Address: "localhost:9090",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := grpcPkg.NewClient(config, *logger, tracer)
	fetcher := NewNodeConfigFetcher(*logger, client, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameNodeConfig, fetcher.Name())
	assert.Empty(t, fetcher.Dependencies())
}
func TestNodeConfigFetcherFail(t *testing.T) {
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

	fetcher := NewNodeConfigFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

func TestNodeConfigFetcherError(t *testing.T) {
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

	fetcher := NewNodeConfigFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

func TestNodeConfigFetcherNotImplemented(t *testing.T) {
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

	fetcher := NewNodeConfigFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

func TestNodeConfigFetcherOk(t *testing.T) {
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

	fetcher := NewNodeConfigFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.NotNil(t, data)
}
