package fetchers

import (
	"context"
	"main/pkg/clients/cosmovisor"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fs"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestCosmovisorUpgradesFetcherBase(t *testing.T) {
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

	fetcher := NewCosmovisorUpgradesFetcher(*logger, client, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameCosmovisorUpgrades, fetcher.Name())
	assert.Empty(t, fetcher.Dependencies())
}

func TestCosmovisorUpgradesFetcherFail(t *testing.T) {
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
	client.Filesystem = &fs.TestFS{}

	fetcher := NewCosmovisorUpgradesFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

func TestCosmovisorUpgradesFetcherOk(t *testing.T) {
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
	client.UpgradeSubfolderPath = "cosmovisor/upgrades"
	client.Filesystem = &fs.TestFS{}

	fetcher := NewCosmovisorUpgradesFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Equal(t, types.UpgradesPresent{"v15": true}, data)
}
