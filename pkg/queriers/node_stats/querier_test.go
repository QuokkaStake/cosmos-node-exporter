package node_stats

import (
	"context"
	"errors"
	"main/assets"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestNodeStatsQuerierBase(t *testing.T) {
	t.Parallel()

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	querier := NewQuerier(*logger, client, tracer)
	assert.True(t, querier.Enabled())
	assert.Equal(t, "node-stats-querier", querier.Name())
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNodeStatsQuerierFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/status",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.TendermintConfig{
		Address: "https://example.com",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	querier := NewQuerier(*logger, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNodeStatsQuerierOk(t *testing.T) {
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
	logger := loggerPkg.GetDefaultLogger()
	tracer := tracing.InitNoopTracer()
	client := tendermint.NewRPC(config, *logger, tracer)
	querier := NewQuerier(*logger, client, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
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
