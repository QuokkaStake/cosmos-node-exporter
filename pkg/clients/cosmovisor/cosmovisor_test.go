package cosmovisor

import (
	"context"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/exec"
	"main/pkg/fs"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestCosmovisorGetCosmovisorVersionFail(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.CommandExecutor = &exec.TestCommandExecutor{Fail: true}

	version, queryInfo, err := client.GetCosmovisorVersion(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Empty(t, version)
}

func TestCosmovisorGetCosmovisorVersionInvalid(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)

	content := assets.GetBytesOrPanic("empty.json")
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: content}

	version, queryInfo, err := client.GetCosmovisorVersion(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "could not find version")
	assert.False(t, queryInfo.Success)
	assert.Empty(t, version)
}

func TestCosmovisorGetCosmovisorVersionValid(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)

	content := assets.GetBytesOrPanic("cosmovisor-version-ok.txt")
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: content}

	version, queryInfo, err := client.GetCosmovisorVersion(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.Equal(t, "v1.5.0", version)
}

func TestCosmovisorGetAppVersionFail(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.CommandExecutor = &exec.TestCommandExecutor{Fail: true}

	version, queryInfo, err := client.GetVersion(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Empty(t, version)
}

func TestCosmovisorGetAppVersionInvalid(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)

	content := assets.GetBytesOrPanic("invalid.toml")
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: content}

	version, queryInfo, err := client.GetVersion(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character")
	assert.False(t, queryInfo.Success)
	assert.Empty(t, version)
}

func TestCosmovisorGetAppVersionValid(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)

	content := assets.GetBytesOrPanic("cosmovisor-app-version-ok.txt")
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: content}

	version, queryInfo, err := client.GetVersion(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.Equal(t, "1.6.4", version.Version)
}

func TestCosmovisorGetUpgradesFail(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.Filesystem = &fs.TestFS{}

	upgrades, queryInfo, err := client.GetUpgrades(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Empty(t, upgrades)
}

func TestCosmovisorGetUpgradesIsNotDir(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.Filesystem = &fs.TestFS{}
	client.UpgradeSubfolderPath = "cosmovisor/upgrades/v15"

	upgrades, queryInfo, err := client.GetUpgrades(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.Empty(t, upgrades)
}

func TestCosmovisorGetUpgradesGetFileError(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.Filesystem = &fs.TestFS{FileError: true}
	client.UpgradeSubfolderPath = "cosmovisor/upgrades" //nolint:goconst // tests only

	upgrades, queryInfo, err := client.GetUpgrades(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Empty(t, upgrades)
}

func TestCosmovisorGetUpgradesGetFileNotFound(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.Filesystem = &fs.TestFS{FileNotFound: true}
	client.UpgradeSubfolderPath = "cosmovisor/upgrades"

	upgrades, queryInfo, err := client.GetUpgrades(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.False(t, upgrades.HasUpgrade("v15"))
}

func TestCosmovisorGetUpgradesGetFileOk(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.Filesystem = &fs.TestFS{}
	client.UpgradeSubfolderPath = "cosmovisor/upgrades"

	upgrades, queryInfo, err := client.GetUpgrades(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.True(t, upgrades.HasUpgrade("v15"))
}

func TestCosmovisorGetUpgradeInfoFail(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)
	client.CommandExecutor = &exec.TestCommandExecutor{Fail: true}

	upgradeInfo, queryInfo, err := client.GetUpgradeInfo(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Empty(t, upgradeInfo)
}

func TestCosmovisorGetUpgradeInfoNotPresent(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)

	content := assets.GetBytesOrPanic("cosmovisor-upgrade-info-not-found.txt")
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: content}

	upgradeInfo, queryInfo, err := client.GetUpgradeInfo(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.Empty(t, upgradeInfo)
}

func TestCosmovisorGetUpgradeInfoInvalid(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)

	content := assets.GetBytesOrPanic("cosmovisor-version-ok.txt")
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: content}

	upgradeInfo, queryInfo, err := client.GetUpgradeInfo(context.Background())
	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	assert.Empty(t, upgradeInfo)
}

func TestCosmovisorGetUpgradeInfoOk(t *testing.T) {
	t.Parallel()

	config := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewCosmovisor(config, *logger, tracer)

	content := assets.GetBytesOrPanic("cosmovisor-upgrade-info.txt")
	client.CommandExecutor = &exec.TestCommandExecutor{Expected: content}

	upgradeInfo, queryInfo, err := client.GetUpgradeInfo(context.Background())
	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	assert.NotEmpty(t, upgradeInfo)
	assert.Equal(t, int64(999), upgradeInfo.Height)
}
