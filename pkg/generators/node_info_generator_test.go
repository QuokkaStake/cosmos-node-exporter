package generators

import (
	"context"
	grpcPkg "main/pkg/clients/grpc"
	"main/pkg/constants"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	cmtTypes "github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"

	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNodeInfoGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewNodeInfoGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestNodeInfoGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameNodeInfo: 3,
	}

	generator := NewNodeInfoGenerator()
	generator.Get(state)
}

func TestNodeInfoGeneratorOk(t *testing.T) {
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

	fetcher := fetchers.NewNodeInfoFetcher(*logger, client, tracer)
	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameNodeInfo: data,
	}

	generator := NewNodeInfoGenerator()

	metrics := generator.Get(state)
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
