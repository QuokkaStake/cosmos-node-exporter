package generators

import (
	"context"
	"main/assets"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
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

func TestLocalVersionGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewLocalVersionGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestLocalVersionGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameLocalVersion: 3,
	}

	generator := NewLocalVersionGenerator()
	generator.Get(state)
}

func TestLocalVersionGeneratorOk(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	cosmovisor := cosmovisorPkg.NewCosmovisor(config, *logger, tracer)
	cosmovisor.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-app-version-ok.txt")}

	fetcher := fetchers.NewLocalVersionFetcher(*logger, cosmovisor, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameLocalVersion: data,
	}

	generator := NewLocalVersionGenerator()

	metrics := generator.Get(state)
	assert.Len(t, metrics, 1)

	localVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "1.6.4",
	}, localVersion.Labels)
	assert.InDelta(t, 1, localVersion.Value, 0.01)
}
