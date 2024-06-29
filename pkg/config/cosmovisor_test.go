package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestCosmovisorDisabled(t *testing.T) {
	t.Parallel()

	cosmovisorConfig := CosmovisorConfig{}
	err := cosmovisorConfig.Validate()
	require.NoError(t, err)
}

func TestCosmovisorNoChainBinaryName(t *testing.T) {
	t.Parallel()

	cosmovisorConfig := CosmovisorConfig{Enabled: null.BoolFrom(true)}
	err := cosmovisorConfig.Validate()
	require.Error(t, err)
}

func TestCosmovisorNoChainFolder(t *testing.T) {
	t.Parallel()

	cosmovisorConfig := CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "appd",
	}
	err := cosmovisorConfig.Validate()
	require.Error(t, err)
}

func TestCosmovisorNoCosmovisorPath(t *testing.T) {
	t.Parallel()

	cosmovisorConfig := CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "appd",
		ChainFolder:     "/home/user/.app",
	}
	err := cosmovisorConfig.Validate()
	require.Error(t, err)
}

func TestCosmovisorValid(t *testing.T) {
	t.Parallel()

	cosmovisorConfig := CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "appd",
		ChainFolder:     "/home/user/.app",
		CosmovisorPath:  "/home/user/go/bin/cosmovisor",
	}
	err := cosmovisorConfig.Validate()
	require.NoError(t, err)
}
