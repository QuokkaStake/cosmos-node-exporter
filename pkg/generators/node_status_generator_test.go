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

func TestNodeStatusGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewNodeStatusGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestNodeStatusGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameNodeStatus: 3,
	}

	generator := NewNodeStatusGenerator()
	generator.Get(state)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNodeStatusGeneratorOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/status",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("status.json")),
	)

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	fetcher := fetchers.NewNodeStatusFetcher(*logger, client, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameNodeStatus: data,
	}

	generator := NewNodeStatusGenerator()

	metrics := generator.Get(state)
	assert.Len(t, metrics, 5)

	catchingUp := metrics[0]
	assert.Empty(t, catchingUp.Labels)
	assert.Zero(t, catchingUp.Value)

	timeSinceLatest := metrics[1]
	assert.Empty(t, timeSinceLatest.Labels)

	nodeInfo := metrics[2]
	assert.Equal(t, map[string]string{
		"chain":   "cosmoshub-4",
		"moniker": "freak12techno",
	}, nodeInfo.Labels)
	assert.InDelta(t, 1, nodeInfo.Value, 0.01)

	tendermintVersion := metrics[3]
	assert.Equal(t, map[string]string{
		"version": "0.37.6",
	}, tendermintVersion.Labels)
	assert.InDelta(t, 1, tendermintVersion.Value, 0.01)

	votingPower := metrics[4]
	assert.Empty(t, votingPower.Labels)
	assert.Zero(t, votingPower.Value)
}
