package config

import (
	"main/pkg/fs"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestConfigZeroNodes(t *testing.T) {
	t.Parallel()

	appConfig := Config{}
	err := appConfig.Validate()
	require.Error(t, err)
}

func TestConfigInvalidNode(t *testing.T) {
	t.Parallel()

	appConfig := Config{NodeConfigs: []NodeConfig{{}}}
	err := appConfig.Validate()
	require.Error(t, err)
}

func TestConfigInvalidTracing(t *testing.T) {
	t.Parallel()

	appConfig := Config{
		NodeConfigs:   []NodeConfig{{Name: "node"}},
		TracingConfig: TracingConfig{Enabled: null.BoolFrom(true)},
	}
	err := appConfig.Validate()
	require.Error(t, err)
}

func TestConfigValid(t *testing.T) {
	t.Parallel()

	appConfig := Config{
		NodeConfigs: []NodeConfig{{Name: "node"}},
	}
	err := appConfig.Validate()
	require.NoError(t, err)
}

func TestGetConfigErrorLoadingFile(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}

	config, err := GetConfig(filesystem, "config-not-found.toml")
	require.Error(t, err)
	require.Nil(t, config)
}

func TestGetConfigErrorParsing(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}

	config, err := GetConfig(filesystem, "invalid.toml")
	require.Error(t, err)
	require.Nil(t, config)
}

func TestGetConfigValid(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}

	config, err := GetConfig(filesystem, "config-valid.toml")
	require.NoError(t, err)
	require.NotNil(t, config)
}
