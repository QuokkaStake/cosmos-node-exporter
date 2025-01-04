package fetchers

import (
	"context"
	"errors"
	"main/assets"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestBlockTimeFetcherBase(t *testing.T) {
	t.Parallel()

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, []constants.FetcherName{
		constants.FetcherNameUpgrades,
		constants.FetcherNameCosmovisorUpgradeInfo,
	}, fetcher.Dependencies())
	assert.Equal(t, constants.FetcherNameBlockTime, fetcher.Name())
}

func TestBlockTimeFetcherDataEmpty(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)
	fetcher.Get(context.Background())
}

func TestBlockTimeFetcherDataNil(t *testing.T) {
	t.Parallel()

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)

	var upgradePlan *types.Plan

	data, queryInfos := fetcher.Get(context.Background(), upgradePlan, upgradePlan)
	assert.Empty(t, queryInfos)
	assert.Nil(t, data)
}

func TestBlockTimeFetcherDataWrong(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)
	fetcher.Get(context.Background(), 3)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBlockTimeFetcherTendermintFailBlock(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)

	data, queryInfos := fetcher.Get(context.Background(), &types.Plan{}, &types.Plan{})
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBlockTimeFetcherTendermintFailOlderBlock(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block?height=21076108",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)

	data, queryInfos := fetcher.Get(context.Background(), &types.Plan{}, &types.Plan{})
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBlockTimeFetcherNoUpgrade(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)

	var upgradePlan *types.Plan

	data, queryInfos := fetcher.Get(context.Background(), upgradePlan, upgradePlan)
	assert.Empty(t, queryInfos)
	assert.Nil(t, data)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBlockTimeFetcherTendermintOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block?height=21076108",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("block2.json")),
	)

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewBlockTimeFetcher(*logger, client, tracer)

	data, queryInfos := fetcher.Get(context.Background(), &types.Plan{}, &types.Plan{})
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.NotNil(t, data)
}
