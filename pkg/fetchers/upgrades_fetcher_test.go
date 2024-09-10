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

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestUpgradesFetcherBase(t *testing.T) {
	t.Parallel()

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewUpgradesFetcher(*logger, client, true, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameUpgrades, fetcher.Name())
	assert.Empty(t, fetcher.Dependencies())
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesFetcherFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewUpgradesFetcher(*logger, client, true, tracer)

	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesFetcherOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("upgrade-plan.json")),
	)

	config := configPkg.TendermintConfig{Address: "https://example.com"}
	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := NewUpgradesFetcher(*logger, client, true, tracer)

	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.NotNil(t, data)
}
