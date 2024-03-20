package tendermint

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/utils"
	"net/url"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/rs/zerolog"
)

type RPC struct {
	Client       *http.Client
	Logger       zerolog.Logger
	Address      string
	BlocksBehind int64
}

func NewRPC(config config.NodeConfig, logger zerolog.Logger) *RPC {
	return &RPC{
		Logger:       logger.With().Str("component", "tendermint_rpc").Logger(),
		Address:      config.TendermintConfig.Address,
		BlocksBehind: 1000,
		Client:       http.NewClient(logger, config.TendermintConfig.Address),
	}
}

func (t *RPC) Status() (StatusResponse, error) {
	res := StatusResponse{}
	err := t.Client.Query("/status", &res)
	return res, err
}

func (t *RPC) Block(height int64) (BlockResponse, error) {
	blockUrl := "/block"
	if height != 0 {
		blockUrl = fmt.Sprintf("/block?height=%d", height)
	}

	res := BlockResponse{}
	err := t.Client.Query(blockUrl, &res)
	return res, err
}

func (t *RPC) AbciQuery(
	method string,
	message codec.ProtoMarshaler,
	output codec.ProtoMarshaler,
) error {
	dataBytes, err := message.Marshal()
	if err != nil {
		return err
	}

	methodName := fmt.Sprintf("\"%s\"", method)
	queryURL := fmt.Sprintf(
		"/abci_query?path=%s&data=0x%x",
		url.QueryEscape(methodName),
		dataBytes,
	)

	var response AbciQueryResponse
	if err := t.Client.Query(queryURL, &response); err != nil {
		return err
	}

	return output.Unmarshal(response.Result.Response.Value)
}

func (t *RPC) GetUpgradePlan() (*upgradeTypes.Plan, error) {
	query := upgradeTypes.QueryCurrentPlanRequest{}

	var response upgradeTypes.QueryCurrentPlanResponse
	if err := t.AbciQuery("/cosmos.upgrade.v1beta1.Query/CurrentPlan", &query, &response); err != nil {
		return nil, err
	}

	return response.Plan, nil
}

func (t *RPC) GetEstimateTimeTillBlock(height int64) (time.Time, error) {
	latestBlock, err := t.Block(0)
	if err != nil {
		t.Logger.Error().Err(err).Msg("Could not fetch current block")
		return time.Now(), err
	}

	latestBlockHeight, err := utils.StringToInt64(latestBlock.Result.Block.Header.Height)
	if err != nil {
		t.Logger.Error().
			Err(err).
			Msg("Error converting latest block height to int64, which should never happen.")
		return time.Now(), err
	}
	blockToCheck := latestBlockHeight - t.BlocksBehind

	olderBlock, err := t.Block(blockToCheck)
	if err != nil {
		t.Logger.Error().Err(err).Msg("Could not fetch older block")
		return time.Now(), err
	}

	blocksDiffTime := latestBlock.Result.Block.Header.Time.Sub(olderBlock.Result.Block.Header.Time)
	blockTime := blocksDiffTime.Seconds() / float64(t.BlocksBehind)
	blocksTillEstimatedBlock := height - latestBlockHeight
	secondsTillEstimatedBlock := int64(float64(blocksTillEstimatedBlock) * blockTime)
	durationTillEstimatedBlock := time.Duration(secondsTillEstimatedBlock * int64(time.Second))

	return time.Now().Add(durationTillEstimatedBlock), nil
}
