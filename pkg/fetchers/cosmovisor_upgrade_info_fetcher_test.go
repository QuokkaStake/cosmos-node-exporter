package fetchers

import (
	"context"
	"main/assets"
	"main/pkg/clients/cosmovisor"
	"main/pkg/clients/tendermint"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/exec"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	upgradeTypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestCosmovisorUpgradeInfoFetcherBase(t *testing.T) {
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

	fetcher := NewCosmovisorUpgradeInfoFetcher(*logger, client, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameCosmovisorUpgradeInfo, fetcher.Name())
	assert.Equal(t, []constants.FetcherName{
		constants.FetcherNameNodeStatus,
	}, fetcher.Dependencies())
}

func TestCosmovisorUpgradeInfoFetcherNotEnoughArgs(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

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

	fetcher := NewCosmovisorUpgradeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background())
	assert.Empty(t, queryInfos)
	assert.Nil(t, data)
}

func TestCosmovisorUpgradeInfoFetcherFail(t *testing.T) {
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

	fetcher := NewCosmovisorUpgradeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background(), nil)
	assert.Empty(t, queryInfos)
	assert.Nil(t, data)
}

func TestCosmovisorUpgradeInfoNoUpgradeInfo(t *testing.T) {
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
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-upgrade-info-not-found.txt")}

	nodeStatus := tendermint.StatusResponse{
		Result: tendermint.StatusResult{
			SyncInfo: tendermint.SyncInfo{
				LatestBlockHeight: 99,
			},
		},
	}

	fetcher := NewCosmovisorUpgradeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background(), nodeStatus)
	assert.NotEmpty(t, queryInfos)
	assert.True(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

func TestCosmovisorUpgradeInfoFetcherUpgradeInThePast(t *testing.T) {
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
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-upgrade-info.txt")}

	nodeStatus := tendermint.StatusResponse{
		Result: tendermint.StatusResult{
			SyncInfo: tendermint.SyncInfo{
				LatestBlockHeight: 9999,
			},
		},
	}

	fetcher := NewCosmovisorUpgradeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background(), nodeStatus)
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

func TestCosmovisorUpgradeInfoFetcherUpgradeInTheFuture(t *testing.T) {
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
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-upgrade-info.txt")}

	nodeStatus := tendermint.StatusResponse{
		Result: tendermint.StatusResult{
			SyncInfo: tendermint.SyncInfo{
				LatestBlockHeight: 99,
			},
		},
	}

	fetcher := NewCosmovisorUpgradeInfoFetcher(*logger, client, tracer)
	data, queryInfos := fetcher.Get(context.Background(), nodeStatus)
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.NotNil(t, data)

	converted, ok := data.(*upgradeTypes.Plan)
	assert.True(t, ok)
	assert.Equal(t, int64(999), converted.Height)
}
