package upgrades

import (
	"context"
	"errors"
	"main/assets"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	"main/pkg/fs"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestUpgradesQuerierBase(t *testing.T) {
	t.Parallel()

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	querier := NewQuerier(true, *logger, nil, client, tracer)
	assert.True(t, querier.Enabled())
	assert.Equal(t, "upgrades-querier", querier.Name())
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesQuerierTendermintError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.upgrade.v1beta1.Query%2FCurrentPlan%22&data=0x",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	querier := NewQuerier(true, *logger, nil, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesQuerierTendermintFailBlock(t *testing.T) {
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
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	querier := NewQuerier(true, *logger, nil, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 2)
	assert.True(t, queryInfos[0].Success)
	assert.False(t, queryInfos[1].Success)
	assert.Len(t, metrics, 3)

	upgradeComing := metrics[0]
	assert.Empty(t, upgradeComing.Labels)
	assert.InDelta(t, 1, upgradeComing.Value, 0.01)

	upgradeInfo := metrics[1]
	assert.NotEmpty(t, upgradeInfo.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeInfo.Labels["name"])
	assert.InDelta(t, 1, upgradeInfo.Value, 0.01)

	upgradeHeight := metrics[2]
	assert.NotEmpty(t, upgradeHeight.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeHeight.Labels["name"])
	assert.InDelta(t, 8375044, upgradeHeight.Value, 0.01)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesQuerierNoUpgrade(t *testing.T) {
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
	querier := NewQuerier(true, *logger, nil, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Len(t, metrics, 1)

	upgradeComing := metrics[0]
	assert.Empty(t, upgradeComing.Labels)
	assert.Zero(t, upgradeComing.Value)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesQuerierTendermintOk(t *testing.T) {
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

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	querier := NewQuerier(true, *logger, nil, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 2)
	assert.True(t, queryInfos[0].Success)
	assert.True(t, queryInfos[1].Success)
	assert.Len(t, metrics, 4)

	upgradeComing := metrics[0]
	assert.Empty(t, upgradeComing.Labels)
	assert.InDelta(t, 1, upgradeComing.Value, 0.01)

	upgradeInfo := metrics[1]
	assert.NotEmpty(t, upgradeInfo.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeInfo.Labels["name"])
	assert.InDelta(t, 1, upgradeInfo.Value, 0.01)

	upgradeHeight := metrics[2]
	assert.NotEmpty(t, upgradeHeight.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeHeight.Labels["name"])
	assert.InDelta(t, 8375044, upgradeHeight.Value, 0.01)

	upgradeTime := metrics[3]
	assert.NotEmpty(t, upgradeTime.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeTime.Labels["name"])
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesQuerierTendermintCosmovisorFail(t *testing.T) {
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

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	cosmovisor := cosmovisorPkg.NewCosmovisor(configPkg.CosmovisorConfig{}, *logger, tracer)
	cosmovisor.Filesystem = &fs.TestFS{}
	querier := NewQuerier(true, *logger, cosmovisor, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 3)
	assert.True(t, queryInfos[0].Success)
	assert.True(t, queryInfos[1].Success)
	assert.False(t, queryInfos[2].Success)
	assert.Len(t, metrics, 4)

	upgradeComing := metrics[0]
	assert.Empty(t, upgradeComing.Labels)
	assert.InDelta(t, 1, upgradeComing.Value, 0.01)

	upgradeInfo := metrics[1]
	assert.NotEmpty(t, upgradeInfo.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeInfo.Labels["name"])
	assert.InDelta(t, 1, upgradeInfo.Value, 0.01)

	upgradeHeight := metrics[2]
	assert.NotEmpty(t, upgradeHeight.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeHeight.Labels["name"])
	assert.InDelta(t, 8375044, upgradeHeight.Value, 0.01)

	upgradeTime := metrics[3]
	assert.NotEmpty(t, upgradeTime.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeTime.Labels["name"])
}

//nolint:paralleltest // disabled due to httpmock usage
func TestUpgradesQuerierTendermintCosmovisorOk(t *testing.T) {
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

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	cosmovisor := cosmovisorPkg.NewCosmovisor(configPkg.CosmovisorConfig{}, *logger, tracer)
	cosmovisor.UpgradeSubfolderPath = "cosmovisor/upgrades"
	cosmovisor.Filesystem = &fs.TestFS{}
	querier := NewQuerier(true, *logger, cosmovisor, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 3)
	assert.True(t, queryInfos[0].Success)
	assert.True(t, queryInfos[1].Success)
	assert.True(t, queryInfos[2].Success)
	assert.Len(t, metrics, 5)

	upgradeComing := metrics[0]
	assert.Empty(t, upgradeComing.Labels)
	assert.InDelta(t, 1, upgradeComing.Value, 0.01)

	upgradeInfo := metrics[1]
	assert.NotEmpty(t, upgradeInfo.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeInfo.Labels["name"])
	assert.InDelta(t, 1, upgradeInfo.Value, 0.01)

	upgradeHeight := metrics[2]
	assert.NotEmpty(t, upgradeHeight.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeHeight.Labels["name"])
	assert.InDelta(t, 8375044, upgradeHeight.Value, 0.01)

	upgradeTime := metrics[3]
	assert.NotEmpty(t, upgradeTime.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradeTime.Labels["name"])
	assert.InDelta(t, 1642330177, upgradeTime.Value, 0.01)

	upgradePresent := metrics[4]
	assert.NotEmpty(t, upgradePresent.Labels["info"])
	assert.Equal(t, "v1.5.0", upgradePresent.Labels["name"])
	assert.Zero(t, upgradePresent.Value)
}
