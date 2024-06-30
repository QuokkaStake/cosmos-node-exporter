package grpc

import (
	"context"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.nhat.io/grpcmock/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.nhat.io/grpcmock"

	cmtTypes "github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeTypes "github.com/cosmos/cosmos-sdk/client/grpc/node"
)

func TestGrpcClientInit(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	config := configPkg.GrpcConfig{Address: "localhost:9090"}
	client := NewClient(config, *logger, tracer)
	assert.NotNil(t, client)
}
func TestGrpcClientNodeConfigFail(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn()(t)

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"invalid",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	nodeConfig, queryInfo, err := client.GetNodeConfig(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Nil(t, nodeConfig)
}

func TestGrpcClientNodeConfigNotImplemented(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn()(t)

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	nodeConfig, queryInfo, err := client.GetNodeConfig(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.Nil(t, nodeConfig)
}

func TestGrpcClientNodeConfigOk(t *testing.T) {
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
				Return(&nodeTypes.ConfigResponse{MinimumGasPrice: "0.1uatom"})
		},
	)(t)

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	nodeConfig, queryInfo, err := client.GetNodeConfig(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.NotNil(t, nodeConfig)
	assert.Equal(t, "0.1uatom", nodeConfig.MinimumGasPrice)
}

func TestGrpcClientNodeInfoFail(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn()(t)

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"invalid",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	nodeConfig, queryInfo, err := client.GetNodeInfo(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Nil(t, nodeConfig)
}

func TestGrpcClientNodeInfoOk(t *testing.T) {
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
					DefaultNodeInfo: &p2p.DefaultNodeInfo{Network: "chain"},
				})
		},
	)(t)

	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	grpcConn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithContextDialer(d),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	client := &Client{
		Logger: *logger,
		Client: grpcConn,
		Tracer: tracer,
	}

	nodeInfo, queryInfo, err := client.GetNodeInfo(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.NotNil(t, nodeInfo)
	assert.Equal(t, "chain", nodeInfo.DefaultNodeInfo.Network)
}
