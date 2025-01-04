package generators

import (
	"context"
	"main/assets"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	metricsPkg "main/pkg/metrics"
	"main/pkg/tracing"
	"testing"

	"cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestTimeTillUpgradeGeneratorStateEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewTimeTillUpgradeGenerator()
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestTimeTillUpgradeGeneratorBlockTimeEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{
		constants.FetcherNameUpgrades: &types.Plan{},
	}

	generator := NewTimeTillUpgradeGenerator()
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestTimeTillUpgradeGeneratorWrongUpgrade(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameUpgrades:  3,
		constants.FetcherNameBlockTime: &tendermint.BlocksInfo{},
	}

	generator := NewTimeTillUpgradeGenerator()
	generator.Get(state)
}

func TestTimeTillUpgradeGeneratorWrongBlockTime(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameUpgrades:  &types.Plan{},
		constants.FetcherNameBlockTime: 3,
	}

	generator := NewTimeTillUpgradeGenerator()
	generator.Get(state)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestTimeTillUpgradeGeneratorOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("upgrade-plan.json")),
	)

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

	upgradesFetcher := fetchers.NewUpgradesFetcher(*logger, client, true, tracer)
	upgradesInfo, _ := upgradesFetcher.Get(context.Background())
	assert.NotNil(t, upgradesInfo)

	blockTimeFetcher := fetchers.NewBlockTimeFetcher(*logger, client, tracer)
	blockTimeData, _ := blockTimeFetcher.Get(context.Background(), upgradesInfo, upgradesInfo)
	assert.NotNil(t, blockTimeData)

	state := fetchers.State{
		constants.FetcherNameUpgrades:              upgradesInfo,
		constants.FetcherNameCosmovisorUpgradeInfo: upgradesInfo,
		constants.FetcherNameBlockTime:             blockTimeData,
	}

	generator := NewTimeTillUpgradeGenerator()
	metrics := generator.Get(state)
	assert.Len(t, metrics, 2)

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeEstimatedTime,
		Labels: map[string]string{
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceGovernance,
		},
		Value: 1642330177,
	}, metrics[0])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeEstimatedTime,
		Labels: map[string]string{
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceUpgradeInfo,
		},
		Value: 1642330177,
	}, metrics[1])
}
