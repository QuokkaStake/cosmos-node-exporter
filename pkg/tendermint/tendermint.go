package tendermint

import (
	"context"
	"main/pkg/config"
	"time"

	"github.com/rs/zerolog"

	tmrpc "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type TendermintRPC struct {
	Logger       zerolog.Logger
	Client       *tmrpc.HTTP
	BlocksBehind int64
}

func NewTendermintRPC(config *config.Config, logger *zerolog.Logger) *TendermintRPC {
	client, err := tmrpc.New(config.TendermintConfig.Address, "/websocket")
	if err != nil {
		logger.Fatal().Err(err).Msg("Cannot instantiate Tendermint client")
	}

	return &TendermintRPC{
		Logger:       logger.With().Str("component", "tendermint_rpc").Logger(),
		Client:       client,
		BlocksBehind: 1000,
	}
}

func (t *TendermintRPC) GetStatus() (*coretypes.ResultStatus, error) {
	return t.Client.Status(context.Background())
}

func (t *TendermintRPC) GetEstimateTimeTillBlock(height int64) (time.Time, error) {
	latestBlock, err := t.Client.Block(context.Background(), nil)
	if err != nil {
		t.Logger.Error().Err(err).Msg("Could not fetch current block")
		return time.Now(), err
	}

	blockToCheck := latestBlock.Block.Height - t.BlocksBehind

	olderBlock, err := t.Client.Block(context.Background(), &blockToCheck)
	if err != nil {
		t.Logger.Error().Err(err).Msg("Could not fetch older block")
		return time.Now(), err
	}

	blocksDiffTime := latestBlock.Block.Time.Sub(olderBlock.Block.Time)
	blockTime := blocksDiffTime.Seconds() / float64(t.BlocksBehind)
	blocksTillEstimatedBlock := height - latestBlock.Block.Height
	secondsTillEstimatedBlock := blocksTillEstimatedBlock * int64(blockTime)
	durationTillEstimatedBlock := time.Duration(secondsTillEstimatedBlock * int64(time.Second))

	return time.Now().Add(durationTillEstimatedBlock), nil
}
