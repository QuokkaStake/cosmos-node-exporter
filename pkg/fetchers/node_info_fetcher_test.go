package fetchers

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	cmtTypes "github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/stretchr/testify/require"
	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stretchr/testify/assert"
)

func TestNodeInfoFetcherBase(t *testing.T) {
	t.Parallel()

	config := configPkg.GrpcConfig{
		Address: "localhost:9090",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := grpcPkg.NewClient(config, *logger, tracer)
	fetcher := NewNodeInfoFetcher(*logger, client, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameNodeInfo, fetcher.Name())
	assert.Empty(t, fetcher.Dependencies())
}
func TestNodeInfoFetcherFail(t *testing.T) {
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

	fetcher := NewNodeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, data)
}

func TestNodeInfoFetcherError(t *testing.T) {
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

	fetcher := NewNodeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, data)
}

func TestNodeInfoFetcherOk(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn(
		grpcmock.RegisterServiceFromMethods(service.Method{
			ServiceName: "cosmos.base.tendermint.v1beta1.Service",
			MethodName:  "GetNodeInfo",
			MethodType:  service.TypeUnary,
			Input:       &cmtTypes.GetNodeInfoRequest{},
			Output:      &cmtTypes.GetNodeInfoResponse{},
		}),
		func(s *grpcmock.Server) {
			s.ExpectUnary("cosmos.base.tendermint.v1beta1.Service/GetNodeInfo").
				Return(&cmtTypes.GetNodeInfoResponse{
					ApplicationVersion: &cmtTypes.VersionInfo{
						Name:             "appd",
						AppName:          "app",
						Version:          "1.2.3",
						CosmosSdkVersion: "0.50.7",
						GoVersion:        "1.22.3",
						BuildTags:        "netgo,ledger",
						GitCommit:        "123abc",
					},
				})
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

	fetcher := NewNodeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.NotEmpty(t, data)
}
