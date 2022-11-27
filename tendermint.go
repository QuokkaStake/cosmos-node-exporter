package main

import (
	"context"

	"github.com/rs/zerolog"

	tmrpc "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/coretypes"
)

type TendermintRPC struct {
	Logger zerolog.Logger
	Client *tmrpc.HTTP
}

func NewTendermintRPC(config *Config, logger *zerolog.Logger) *TendermintRPC {
	client, err := tmrpc.New(config.TendermintConfig.Address)
	if err != nil {
		logger.Fatal().Err(err).Msg("Cannot instantiate Tendermint client")
	}

	return &TendermintRPC{
		Logger: logger.With().Str("component", "tendermint_rpc").Logger(),
		Client: client,
	}
}

func (t *TendermintRPC) GetStatus() (*coretypes.ResultStatus, error) {
	return t.Client.Status(context.Background())
}
