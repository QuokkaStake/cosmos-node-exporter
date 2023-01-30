package grpc

import (
	"context"
	"main/pkg/config"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type Grpc struct {
	Logger zerolog.Logger
	Client *grpc.ClientConn
}

func NewGrpc(config *config.Config, logger *zerolog.Logger) *Grpc {
	grpcLogger := logger.With().Str("component", "grpc").Logger()

	grpcConn, err := grpc.Dial(
		config.GrpcConfig.Address,
		grpc.WithInsecure(),
		//nolint:staticcheck
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		grpcLogger.Fatal().Err(err).Msg("Could not connect to gRPC node")
	}

	return &Grpc{
		Logger: grpcLogger,
		Client: grpcConn,
	}
}

func (g *Grpc) GetUpgradePlan() (*upgradeTypes.Plan, error) {
	upgradeClient := upgradeTypes.NewQueryClient(g.Client)
	response, err := upgradeClient.CurrentPlan(
		context.Background(),
		&upgradeTypes.QueryCurrentPlanRequest{},
	)

	if err != nil {
		return nil, err
	}

	return response.Plan, nil
}
