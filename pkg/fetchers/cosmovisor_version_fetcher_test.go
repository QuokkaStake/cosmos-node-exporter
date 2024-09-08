package fetchers

import (
	"context"
	"main/assets"
	"main/pkg/clients/cosmovisor"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/exec"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestCosmovisorVersionFetcherBase(t *testing.T) {
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

	fetcher := NewCosmovisorVersionFetcher(*logger, client, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameCosmovisorVersion, fetcher.Name())
	assert.Empty(t, fetcher.Dependencies())
}

func TestCosmovisorVersionFetcherFail(t *testing.T) {
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

	fetcher := NewCosmovisorVersionFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

func TestCosmovisorVersionFetcherOk(t *testing.T) {
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

	fetcher := NewCosmovisorVersionFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Equal(t, "v1.5.0", data)
}
