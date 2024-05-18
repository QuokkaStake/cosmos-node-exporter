package grpc

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"

	cmtTypes "github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeTypes "github.com/cosmos/cosmos-sdk/client/grpc/node"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Client struct {
	Logger zerolog.Logger
	Client *grpc.ClientConn
	Tracer trace.Tracer
}

func NewClient(config config.NodeConfig, logger zerolog.Logger, tracer trace.Tracer) *Client {
	grpcLogger := logger.With().Str("component", "grpc").Logger()

	grpcConn, err := grpc.Dial(
		config.GrpcConfig.Address,
		grpc.WithInsecure(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		grpcLogger.Panic().Err(err).Msg("Could not connect to gRPC node")
	}

	return &Client{
		Logger: grpcLogger,
		Client: grpcConn,
		Tracer: tracer,
	}
}

func (g *Client) GetNodeConfig(ctx context.Context) (*nodeTypes.ConfigResponse, query_info.QueryInfo, error) {
	childCtx, span := g.Tracer.Start(
		ctx,
		"Fetch gRPC node config",
	)
	defer span.End()

	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleGrpc,
		Action:  constants.ActionGrpcGetNodeConfig,
		Success: false,
	}

	client := nodeTypes.NewServiceClient(g.Client)
	response, err := client.Config(
		childCtx,
		&nodeTypes.ConfigRequest{},
	)

	if err != nil {
		// some chains do not implement this endpoint due to their cosmos-sdk version
		// being too old
		if strings.Contains(err.Error(), "unknown service cosmos.base.node.v1beta1.Service") {
			queryInfo.Success = true
			return nil, queryInfo, nil
		}
		span.RecordError(err)
		return nil, queryInfo, err
	}

	queryInfo.Success = true

	return response, queryInfo, nil
}

func (g *Client) GetNodeInfo(ctx context.Context) (*cmtTypes.GetNodeInfoResponse, query_info.QueryInfo, error) {
	childCtx, span := g.Tracer.Start(
		ctx,
		"Fetch gRPC node info",
	)
	defer span.End()

	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleGrpc,
		Action:  constants.ActionGrpcGetNodeInfo,
		Success: false,
	}

	client := cmtTypes.NewServiceClient(g.Client)
	response, err := client.GetNodeInfo(
		childCtx,
		&cmtTypes.GetNodeInfoRequest{},
	)

	if err != nil {
		span.RecordError(err)
		return nil, queryInfo, err
	}
	queryInfo.Success = true

	return response, queryInfo, nil
}
