package versions

import (
	"context"
	"errors"
	"main/assets"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	githubPkg "main/pkg/clients/git"
	configPkg "main/pkg/config"
	"main/pkg/exec"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestVersionsQuerierBase(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	querier := NewQuerier(*logger, nil, nil, tracer)
	assert.False(t, querier.Enabled())
	assert.Equal(t, "versions-querier", querier.Name())
}

//nolint:paralleltest // disabled due to httpmock usage
func TestVersionsQuerierGitFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	githubClient := githubPkg.NewGithub(config, *logger, tracer)
	querier := NewQuerier(*logger, githubClient, nil, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestVersionsQuerierGitOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("github-valid.json")),
	)

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	githubClient := githubPkg.NewGithub(config, *logger, tracer)
	querier := NewQuerier(*logger, githubClient, nil, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Len(t, metrics, 1)

	remoteVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "17.2.0",
	}, remoteVersion.Labels)
	assert.InDelta(t, 1, remoteVersion.Value, 0.01)
}

func TestVersionsQuerierCosmovisorFail(t *testing.T) {
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

	querier := NewQuerier(*logger, nil, cosmovisor, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Empty(t, metrics)
}

func TestVersionsQuerierCosmovisorOk(t *testing.T) {
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

	querier := NewQuerier(*logger, nil, cosmovisor, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Len(t, metrics, 1)

	localVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "1.6.4",
	}, localVersion.Labels)
	assert.InDelta(t, 1, localVersion.Value, 0.01)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestVersionsQuerierLocalSemverInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("github-valid.json")),
	)

	gitConfig := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	githubClient := githubPkg.NewGithub(gitConfig, *logger, tracer)

	cosmovisorConfig := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	cosmovisor := cosmovisorPkg.NewCosmovisor(cosmovisorConfig, *logger, tracer)
	cosmovisor.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-app-version-invalid.txt")}

	querier := NewQuerier(*logger, githubClient, cosmovisor, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 2)
	assert.True(t, queryInfos[0].Success)
	assert.True(t, queryInfos[1].Success)
	assert.Len(t, metrics, 2)

	remoteVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "17.2.0",
	}, remoteVersion.Labels)
	assert.InDelta(t, 1, remoteVersion.Value, 0.01)

	localVersion := metrics[1]
	assert.Equal(t, map[string]string{
		"version": "test",
	}, localVersion.Labels)
	assert.InDelta(t, 1, localVersion.Value, 0.01)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestVersionsQuerierRemoteSemverInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("github-invalid.json")),
	)

	gitConfig := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	githubClient := githubPkg.NewGithub(gitConfig, *logger, tracer)

	cosmovisorConfig := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	cosmovisor := cosmovisorPkg.NewCosmovisor(cosmovisorConfig, *logger, tracer)
	cosmovisor.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-app-version-ok.txt")}

	querier := NewQuerier(*logger, githubClient, cosmovisor, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 2)
	assert.True(t, queryInfos[0].Success)
	assert.True(t, queryInfos[1].Success)
	assert.Len(t, metrics, 2)

	remoteVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "test",
	}, remoteVersion.Labels)
	assert.InDelta(t, 1, remoteVersion.Value, 0.01)

	localVersion := metrics[1]
	assert.Equal(t, map[string]string{
		"version": "1.6.4",
	}, localVersion.Labels)
	assert.InDelta(t, 1, localVersion.Value, 0.01)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestVersionsQuerierAllOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("github-valid.json")),
	)

	gitConfig := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	githubClient := githubPkg.NewGithub(gitConfig, *logger, tracer)

	cosmovisorConfig := configPkg.CosmovisorConfig{
		Enabled:         null.BoolFrom(true),
		ChainBinaryName: "gaiad",
		ChainFolder:     "/home/validator/.gaia",
		CosmovisorPath:  "/home/validator/go/bin/cosmovisor",
	}
	cosmovisor := cosmovisorPkg.NewCosmovisor(cosmovisorConfig, *logger, tracer)
	cosmovisor.CommandExecutor = &exec.TestCommandExecutor{Expected: assets.GetBytesOrPanic("cosmovisor-app-version-ok.txt")}

	querier := NewQuerier(*logger, githubClient, cosmovisor, tracer)

	metrics, queryInfos := querier.Get(context.Background())
	assert.Len(t, queryInfos, 2)
	assert.True(t, queryInfos[0].Success)
	assert.True(t, queryInfos[1].Success)
	assert.Len(t, metrics, 3)

	remoteVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "17.2.0",
	}, remoteVersion.Labels)
	assert.InDelta(t, 1, remoteVersion.Value, 0.01)

	localVersion := metrics[1]
	assert.Equal(t, map[string]string{
		"version": "1.6.4",
	}, localVersion.Labels)
	assert.InDelta(t, 1, localVersion.Value, 0.01)

	isLatest := metrics[2]
	assert.Equal(t, map[string]string{
		"local_version":  "1.6.4",
		"remote_version": "17.2.0",
	}, isLatest.Labels)
	assert.Zero(t, isLatest.Value)
}
