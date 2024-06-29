package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitDisabled(t *testing.T) {
	t.Parallel()

	gitConfig := GitConfig{}
	err := gitConfig.Validate()
	require.NoError(t, err)
}

func TestGitInvalid(t *testing.T) {
	t.Parallel()

	gitConfig := GitConfig{Repository: "invalid"}
	err := gitConfig.Validate()
	require.Error(t, err)
}

func TestGitValidGithub(t *testing.T) {
	t.Parallel()

	gitConfig := GitConfig{Repository: "https://github.com/QuokkaStake/cosmos-node-exporter"}
	err := gitConfig.Validate()
	require.NoError(t, err)
}

func TestGitValidGitopia(t *testing.T) {
	t.Parallel()

	gitConfig := GitConfig{Repository: "gitopia://gitopia/gitopia"}
	err := gitConfig.Validate()
	require.NoError(t, err)
}
