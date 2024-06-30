package node_stats

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	configPkg "main/pkg/config"
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

func TestNodeInfoQuerierBase(t *testing.T) {
	t.Parallel()

	config := configPkg.GrpcConfig{
		Address: "localhost:9090",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := grpcPkg.NewClient(config, *logger, tracer)
	querier := NewQuerier(*logger, client, tracer)
	assert.True(t, querier.Enabled())
	assert.Equal(t, "node-info-querier", querier.Name())
}
func TestNodeInfoQuerierFail(t *testing.T) {
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

func TestNodeInfoQuerierError(t *testing.T) {
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

func TestNodeInfoQuerierOk(t *testing.T) {
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

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Len(t, metrics, 3)

	cosmosSdkVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "0.50.7",
	}, cosmosSdkVersion.Labels)
	assert.InDelta(t, 1, cosmosSdkVersion.Value, 0.01)

	appInfo := metrics[1]
	assert.Equal(t, map[string]string{
		"name":       "appd",
		"app_name":   "app",
		"git_commit": "123abc",
		"version":    "1.2.3",
	}, appInfo.Labels)
	assert.InDelta(t, 1, appInfo.Value, 0.01)

	goVersion := metrics[2]
	assert.Equal(t, map[string]string{
		"tags":    "netgo,ledger",
		"version": "1.22.3",
	}, goVersion.Labels)
	assert.InDelta(t, 1, goVersion.Value, 0.01)
}
