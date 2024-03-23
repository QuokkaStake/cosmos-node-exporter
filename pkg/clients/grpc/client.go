package grpc

import (
	"context"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/query_info"
	"strings"
	"time"

	nodeTypes "github.com/cosmos/cosmos-sdk/client/grpc/node"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Client struct {
	Logger zerolog.Logger
	Client *grpc.ClientConn
}

func NewClient(config config.NodeConfig, logger zerolog.Logger) *Client {
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
	}
}

func (g *Client) GetNodeConfig() (*nodeTypes.ConfigResponse, query_info.QueryInfo, error) {
	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleGrpc,
		Action:  constants.ActionGrpcGetNodeConfig,
		Success: false,
	}

	client := nodeTypes.NewServiceClient(g.Client)
	response, err := client.Config(
		context.Background(),
		&nodeTypes.ConfigRequest{},
	)

	if err != nil {
		// some chains do not implement this endpoint due to their cosmos-sdk version
		// being too old
		if strings.Contains(err.Error(), "unknown service cosmos.base.node.v1beta1.Service") {
			queryInfo.Success = true
			return nil, queryInfo, nil
		}
		return nil, queryInfo, err
	}

	queryInfo.Success = true

	return response, queryInfo, nil
}
