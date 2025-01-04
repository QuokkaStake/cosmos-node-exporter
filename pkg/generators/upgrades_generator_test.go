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

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestUpgradesGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}
	generator := NewUpgradesGenerator()
	metrics := generator.Get(state)
	assert.Len(t, metrics, 2)

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeComing,
		Labels:     map[string]string{"source": constants.UpgradeSourceGovernance},
		Value:      0,
	}, metrics[0])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeComing,
		Labels:     map[string]string{"source": constants.UpgradeSourceUpgradeInfo},
		Value:      0,
	}, metrics[1])
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
	fetcher := fetchers.NewUpgradesFetcher(*logger, client, true, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.Nil(t, data)

	state := fetchers.State{constants.FetcherNameUpgrades: data}
	generator := NewUpgradesGenerator()

	metrics := generator.Get(state)
	assert.Len(t, metrics, 2)

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeComing,
		Labels:     map[string]string{"source": constants.UpgradeSourceGovernance},
		Value:      0,
	}, metrics[0])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeComing,
		Labels:     map[string]string{"source": constants.UpgradeSourceUpgradeInfo},
		Value:      0,
	}, metrics[1])
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
	fetcher := fetchers.NewUpgradesFetcher(*logger, client, true, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameUpgrades:              data,
		constants.FetcherNameCosmovisorUpgradeInfo: data,
	}

	generator := NewUpgradesGenerator()

	metrics := generator.Get(state)
	assert.Len(t, metrics, 6)

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeComing,
		Labels:     map[string]string{"source": constants.UpgradeSourceGovernance},
		Value:      1,
	}, metrics[0])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeInfo,
		Labels: map[string]string{
			"info":   "{\n  \"binaries\": {\n    \"linux/amd64\": \"https://github.com/NibiruChain/nibiru/releases/download/v1.5.0/nibid_1.5.0_linux_amd64.tar.gz\",\n    \"linux/arm64\": \"https://github.com/NibiruChain/nibiru/releases/download/v1.5.0/nibid_1.5.0_linux_arm64.tar.gz\",\n    \"docker\": \"ghcr.io/nibiruchain/nibiru:1.5.0\"\n  }\n}",
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceGovernance,
		},
		Value: 1,
	}, metrics[1])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeHeight,
		Labels: map[string]string{
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceGovernance,
		},
		Value: 8375044,
	}, metrics[2])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeComing,
		Labels:     map[string]string{"source": constants.UpgradeSourceUpgradeInfo},
		Value:      1,
	}, metrics[3])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeInfo,
		Labels: map[string]string{
			"info":   "{\n  \"binaries\": {\n    \"linux/amd64\": \"https://github.com/NibiruChain/nibiru/releases/download/v1.5.0/nibid_1.5.0_linux_amd64.tar.gz\",\n    \"linux/arm64\": \"https://github.com/NibiruChain/nibiru/releases/download/v1.5.0/nibid_1.5.0_linux_arm64.tar.gz\",\n    \"docker\": \"ghcr.io/nibiruchain/nibiru:1.5.0\"\n  }\n}",
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceUpgradeInfo,
		},
		Value: 1,
	}, metrics[4])

	assert.Equal(t, metricsPkg.MetricInfo{
		MetricName: metricsPkg.MetricNameUpgradeHeight,
		Labels: map[string]string{
			"name":   "v1.5.0",
			"source": constants.UpgradeSourceUpgradeInfo,
		},
		Value: 8375044,
	}, metrics[5])
}
