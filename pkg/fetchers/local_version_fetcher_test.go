package fetchers

import (
	"context"
	"main/assets"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/exec"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestLocalVersionFetcherBase(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	fetcher := NewLocalVersionFetcher(*logger, nil, tracer)
	assert.False(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameLocalVersion, fetcher.Name())
	assert.Empty(t, fetcher.Dependencies())
}

func TestLocalVersionFetcherCosmovisorFail(t *testing.T) {
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
	cosmovisor.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("invalid.toml")}

	fetcher := NewLocalVersionFetcher(*logger, cosmovisor, tracer)

	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, data)
}

func TestLocalVersionFetcherCosmovisorOk(t *testing.T) {
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

	fetcher := NewLocalVersionFetcher(*logger, cosmovisor, tracer)

	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Equal(t, types.VersionInfo{Name: "decentr", Version: "1.6.4"}, data)
}
