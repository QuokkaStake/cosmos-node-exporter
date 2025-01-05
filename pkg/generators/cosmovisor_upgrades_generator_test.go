package generators

import (
	"context"
	"main/assets"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	"main/pkg/fs"
	loggerPkg "main/pkg/logger"
	metricsPkg "main/pkg/metrics"
	"main/pkg/tracing"
	"main/pkg/types"
	"testing"

	upgradeTypes "cosmossdk.io/x/upgrade/types"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCosmovisorUpgradesGeneratorEmptyState(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewCosmovisorUpgradesGenerator()
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestCosmovisorUpgradesGeneratorEmptyCosmovisorUpgrades(t *testing.T) {
	t.Parallel()

	state := fetchers.State{
		constants.FetcherNameUpgrades: &upgradeTypes.Plan{},
	}

	generator := NewCosmovisorUpgradesGenerator()
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestCosmovisorUpgradesGeneratorEmptyInvalidUpgrades(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameUpgrades:           3,
		constants.FetcherNameCosmovisorUpgrades: types.UpgradesPresent{},
	}

	generator := NewCosmovisorUpgradesGenerator()
	generator.Get(state)
}

func TestCosmovisorUpgradesGeneratorEmptyInvalidCosmovisorUpgrades(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameUpgrades:           &upgradeTypes.Plan{},
		constants.FetcherNameCosmovisorUpgrades: 3,
	}

	generator := NewCosmovisorUpgradesGenerator()
	generator.Get(state)
}

func TestCosmovisorUpgradesGeneratorEmptyUpgradePlan(t *testing.T) {
	t.Parallel()

	var upgradePlan *upgradeTypes.Plan

	state := fetchers.State{
		constants.FetcherNameUpgrades:           upgradePlan,
		constants.FetcherNameCosmovisorUpgrades: &types.UpgradesPresent{},
	}

	generator := NewCosmovisorUpgradesGenerator()
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestCosmovisorUpgradesGeneratorOk(t *testing.T) {
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
	cosmovisor := cosmovisorPkg.NewCosmovisor(configPkg.CosmovisorConfig{}, *logger, tracer)
	cosmovisor.UpgradeSubfolderPath = "cosmovisor/upgrades"
	cosmovisor.Filesystem = &fs.TestFS{}

	cosmovisorFetcher := fetchers.NewCosmovisorUpgradesFetcher(*logger, cosmovisor, tracer)
	cosmovisorUpgrades, _ := cosmovisorFetcher.Get(context.Background())
	assert.NotNil(t, cosmovisorUpgrades)

	upgradesFetcher := fetchers.NewUpgradesFetcher(*logger, client, true, tracer)
	upgrades, _ := upgradesFetcher.Get(context.Background())
	assert.NotNil(t, upgrades)

	state := fetchers.State{
		constants.FetcherNameUpgrades:              upgrades,
		constants.FetcherNameCosmovisorUpgradeInfo: upgrades,
		constants.FetcherNameCosmovisorUpgrades:    cosmovisorUpgrades,
	}

	generator := NewCosmovisorUpgradesGenerator()
	metrics := generator.Get(state)
	assert.Len(t, metrics, 2)

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeBinaryPresent,
		Labels: map[string]string{
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceGovernance,
		},
		Value: 0,
	}, metrics[0])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeBinaryPresent,
		Labels: map[string]string{
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceUpgradeInfo,
		},
		Value: 0,
	}, metrics[1])
}
