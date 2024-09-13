package tendermint

import (
	"context"
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/http"
	"main/pkg/query_info"
	"net/url"

	"github.com/cosmos/gogoproto/proto"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	upgradeTypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/rs/zerolog"
)

type RPC struct {
	Client       *http.Client
	Logger       zerolog.Logger
	Address      string
	BlocksBehind int64
	Tracer       trace.Tracer
}

func NewRPC(config config.TendermintConfig, logger zerolog.Logger, tracer trace.Tracer) *RPC {
	return &RPC{
		Logger:       logger.With().Str("component", "tendermint_rpc").Logger(),
		Address:      config.Address,
		BlocksBehind: constants.BlocksBehindToCheck,
		Client:       http.NewClient(logger, config.Address, tracer),
		Tracer:       tracer,
	}
}

func (t *RPC) Status(ctx context.Context) (StatusResponse, query_info.QueryInfo, error) {
	childCtx, span := t.Tracer.Start(
		ctx,
		"Fetching node status",
		trace.WithAttributes(attribute.String("address", t.Address)),
	)
	defer span.End()

	queryInfo := query_info.QueryInfo{
		Module:  constants.ModuleTendermint,
		Action:  constants.ActionTendermintGetNodeStatus,
		Success: false,
	}

	res := StatusResponse{}
	err := t.Client.Query(childCtx, "/status", &res)

	if err == nil {
		queryInfo.Success = true
	}

	return res, queryInfo, err
}

func (t *RPC) Block(ctx context.Context, height int64) (BlockResponse, error) {
	childCtx, span := t.Tracer.Start(
		ctx,
		"Fetching block",
		trace.WithAttributes(
			attribute.String("address", t.Address),
			attribute.Int64("height", height),
		),
	)
	defer span.End()

	blockUrl := "/block"
	if height != 0 {
		blockUrl = fmt.Sprintf("/block?height=%d", height)
	}

	res := BlockResponse{}
	err := t.Client.Query(childCtx, blockUrl, &res)
	return res, err
}

func (t *RPC) AbciQuery(
	ctx context.Context,
	method string,
	message proto.Message,
	output codec.ProtoMarshaler,
) error {
	childCtx, span := t.Tracer.Start(
		ctx,
		"Fetching ABCI query",
		trace.WithAttributes(
			attribute.String("address", t.Address),
			attribute.String("query", method),
		),
	)
	defer span.End()

	dataBytes := message.String()
	methodName := fmt.Sprintf("\"%s\"", method)
	queryURL := fmt.Sprintf(
		"/abci_query?path=%s&data=0x%x",
		url.QueryEscape(methodName),
		dataBytes,
	)

	var response AbciQueryResponse
	if err := t.Client.Query(childCtx, queryURL, &response); err != nil {
		return err
	}

	return output.Unmarshal(response.Result.Response.Value)
}

func (t *RPC) GetUpgradePlan(ctx context.Context) (*upgradeTypes.Plan, query_info.QueryInfo, error) {
	childCtx, span := t.Tracer.Start(
		ctx,
		"Fetching upgrade plan",
		trace.WithAttributes(attribute.String("address", t.Address)),
	)
	defer span.End()

	upgradePlanQuery := query_info.QueryInfo{
		Module:  constants.ModuleTendermint,
		Action:  constants.ActionTendermintGetUpgradePlan,
		Success: false,
	}

	query := upgradeTypes.QueryCurrentPlanRequest{}

	var response upgradeTypes.QueryCurrentPlanResponse
	if err := t.AbciQuery(childCtx, "/cosmos.upgrade.v1beta1.Query/CurrentPlan", &query, &response); err != nil {
		return nil, upgradePlanQuery, err
	}

	upgradePlanQuery.Success = true

	return response.Plan, upgradePlanQuery, nil
}

func (t *RPC) GetBlockTime(ctx context.Context) (*BlocksInfo, query_info.QueryInfo, error) {
	childCtx, span := t.Tracer.Start(
		ctx,
		"Fetching block time",
		trace.WithAttributes(attribute.String("address", t.Address)),
	)
	defer span.End()

	blockTimeQuery := query_info.QueryInfo{
		Module:  constants.ModuleTendermint,
		Action:  constants.ActionTendermintGetBlockTime,
		Success: false,
	}

	latestBlock, err := t.Block(childCtx, 0)
	if err != nil {
		t.Logger.Error().Err(err).Msg("Could not fetch current block")
		return nil, blockTimeQuery, err
	}

	latestBlockHeight := latestBlock.Result.Block.Header.Height
	blockToCheck := latestBlockHeight - t.BlocksBehind

	olderBlock, err := t.Block(childCtx, blockToCheck)
	if err != nil {
		t.Logger.Error().Err(err).Msg("Could not fetch older block")
		return nil, blockTimeQuery, err
	}

	blockTimeQuery.Success = true

	return &BlocksInfo{
		NewerBlock: latestBlock,
		OlderBlock: olderBlock,
	}, blockTimeQuery, nil
}
