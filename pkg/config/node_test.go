package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestNodeNoName(t *testing.T) {
	t.Parallel()

	nodeConfig := NodeConfig{}
	err := nodeConfig.Validate()
	require.Error(t, err)
}

func TestNodeInvalidGitConfig(t *testing.T) {
	t.Parallel()

	nodeConfig := NodeConfig{Name: "node", GitConfig: GitConfig{Repository: "invalid"}}
	err := nodeConfig.Validate()
	require.Error(t, err)
}

func TestNodeInvalidCosmovisorConfig(t *testing.T) {
	t.Parallel()

	nodeConfig := NodeConfig{Name: "node", CosmovisorConfig: CosmovisorConfig{Enabled: null.BoolFrom(true)}}
	err := nodeConfig.Validate()
	require.Error(t, err)
}

func TestNodeValid(t *testing.T) {
	t.Parallel()

	nodeConfig := NodeConfig{Name: "node"}
	err := nodeConfig.Validate()
	require.NoError(t, err)
}
