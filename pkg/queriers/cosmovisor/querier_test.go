package cosmovisor

import (
	"context"
	"main/assets"
	"main/pkg/clients/cosmovisor"
	configPkg "main/pkg/config"
	"main/pkg/exec"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestCosmovisorQuerierBase(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := cosmovisor.NewCosmovisor(config, *logger, tracer)

	querier := NewQuerier(*logger, client, tracer)
	assert.True(t, querier.Enabled())
	assert.Equal(t, "cosmovisor-querier", querier.Name())
}

func TestCosmovisorQuerierFail(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := cosmovisor.NewCosmovisor(config, *logger, tracer)
	client.CommandExecutor = &exec.TestCommandExecutor{Fail: true}

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

func TestCosmovisorQuerierOk(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := cosmovisor.NewCosmovisor(config, *logger, tracer)
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-version-ok.txt")}

	querier := NewQuerier(*logger, client, tracer)
	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Len(t, metrics, 1)

	metric := metrics[0]
	assert.Equal(t, map[string]string{"version": "v1.5.0"}, metric.Labels)
	assert.InDelta(t, 1, metric.Value, 0.01)
}
