package generators

import (
	"context"
	"main/assets"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestUpgradesGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}
	generator := NewUpgradesGenerator()
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestUpgradesGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{constants.FetcherNameUpgrades: 3}
	generator := NewUpgradesGenerator()
	generator.Get(state)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesGeneratorNotPresent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("upgrade-plan-empty.json")),
	)

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := fetchers.NewUpgradesFetcher(*logger, client, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.Nil(t, data)

	state := fetchers.State{constants.FetcherNameUpgrades: data}
	generator := NewUpgradesGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesGeneratorOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("upgrade-plan.json")),
	)

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := fetchers.NewUpgradesFetcher(*logger, client, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{constants.FetcherNameUpgrades: data}

	generator := NewUpgradesGenerator()

	metrics := generator.Get(state)
	assert.Len(t, metrics, 2)

	upgradeInfo := metrics[0]
	assert.NotEmpty(t, upgradeInfo.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeInfo.Labels["name"])
	assert.InDelta(t, 1, upgradeInfo.Value, 0.01)

	upgradeHeight := metrics[1]
	assert.Equal(t, "v1.5.0", upgradeHeight.Labels["name"])
	assert.InDelta(t, 8375044, upgradeHeight.Value, 0.01)
}
