package tendermint

import (
	"encoding/json"
	"fmt"
	"main/pkg/config"
	"main/pkg/utils"
	"net/http"
	"net/url"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/rs/zerolog"
)

type RPC struct {
	Logger       zerolog.Logger
	Address      string
	BlocksBehind int64
}

func NewRPC(config *config.Config, logger *zerolog.Logger) *RPC {
	return &RPC{
		Logger:       logger.With().Str("component", "tendermint_rpc").Logger(),
		Address:      config.TendermintConfig.Address,
		BlocksBehind: 1000,
	}
}

func (t *RPC) Query(url string, output interface{}) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "cosmos-node-exporter")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(&output)
}

func (t *RPC) Status() (StatusResponse, error) {
	url := fmt.Sprintf("%s/status", t.Address)
	res := StatusResponse{}
	err := t.Query(url, &res)
	return res, err
}

func (t *RPC) Block(height int64) (BlockResponse, error) {
	url := fmt.Sprintf("%s/block", t.Address)
	if height != 0 {
		url = fmt.Sprintf("%s/block?height=%d", t.Address, height)
	}

	res := BlockResponse{}
	err := t.Query(url, &res)
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
		"%s/abci_query?path=%s&data=0x%x",
		t.Address,
		url.QueryEscape(methodName),
		dataBytes,
	)

	var response AbciQueryResponse
	if err := t.Query(queryURL, &response); err != nil {
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
	secondsTillEstimatedBlock := blocksTillEstimatedBlock * int64(blockTime)
	durationTillEstimatedBlock := time.Duration(secondsTillEstimatedBlock * int64(time.Second))

	return time.Now().Add(durationTillEstimatedBlock), nil
}
