package tendermint

import (
	"context"
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintStatusFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/status",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	status, queryInfo, err := rpc.Status(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	assert.False(t, queryInfo.Success)
	assert.Empty(t, status)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintStatusSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/status",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("status.json")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	status, queryInfo, err := rpc.Status(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.NotEmpty(t, status)
	assert.Equal(t, int64(0), status.Result.ValidatorInfo.VotingPower)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintBlockFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	status, err := rpc.Block(context.Background(), 0)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	assert.Empty(t, status)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintBlockSuccessLatest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block.json")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	status, err := rpc.Block(context.Background(), 0)
	require.NoError(t, err)
	assert.Equal(t, int64(21077108), status.Result.Block.Header.Height)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintBlockSpecific(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block?height=21077108",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block.json")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	status, err := rpc.Block(context.Background(), 21077108)
	require.NoError(t, err)
	assert.Equal(t, int64(21077108), status.Result.Block.Header.Height)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintGetBlockTimeFailCurrent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	_, queryInfo, err := rpc.GetBlockTime(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintGetBlockTimeFailOlder(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block?height=21076108",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	_, queryInfo, err := rpc.GetBlockTime(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintGetBlockTimeSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/block?height=21076108",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block2.json")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	blockTime, queryInfo, err := rpc.GetBlockTime(context.Background())
	require.NoError(t, err)
	require.True(t, queryInfo.Success)
	assert.NotNil(t, blockTime)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintGetUpgradePlanError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	_, queryInfo, err := rpc.GetUpgradePlan(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.False(t, queryInfo.Success)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTendermintGetUpgradePlanSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("upgrade-plan.json")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	plan, queryInfo, err := rpc.GetUpgradePlan(context.Background())
	require.NoError(t, err)
	require.True(t, queryInfo.Success)
	require.Equal(t, "v1.5.0", plan.Name)
}

func TestTendermintNotSuccessCode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com:443/status",
		httpmock.NewBytesResponder(501, assets.GetBytesOrPanic("status.json")),
	)

	config := configPkg.TendermintConfig{
		Enabled: null.BoolFrom(true),
		Address: "https://example.com:443",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	rpc := NewRPC(config, *logger, tracer)

	_, queryInfo, err := rpc.Status(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "http request failed with status code 501")
	assert.False(t, queryInfo.Success)
}
