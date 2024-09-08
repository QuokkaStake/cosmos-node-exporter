package generators

import (
	"context"
	"main/assets"
	"main/pkg/clients/cosmovisor"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/exec"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"gopkg.in/guregu/null.v4"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestCosmovisorVersionGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewCosmovisorVersionGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestCosmovisorVersionGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameCosmovisorVersion: 3,
	}

	generator := NewCosmovisorVersionGenerator()
	generator.Get(state)
}

func TestCosmovisorVersionGeneratorOk(t *testing.T) {
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
	fetcher := fetchers.NewCosmovisorVersionFetcher(*logger, client, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameCosmovisorVersion: data,
	}

	generator := NewCosmovisorVersionGenerator()

	metrics := generator.Get(state)
	assert.Len(t, metrics, 1)

	metric := metrics[0]
	assert.Equal(t, map[string]string{"version": "v1.5.0"}, metric.Labels)
	assert.InDelta(t, 1, metric.Value, 0.01)
}
