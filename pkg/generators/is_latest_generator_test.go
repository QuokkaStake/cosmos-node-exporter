package generators

import (
	"context"
	"main/assets"
	cosmovisorPkg "main/pkg/clients/cosmovisor"
	githubPkg "main/pkg/clients/git"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/exec"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"main/pkg/types"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestLocalVersionGeneratorLocalEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	logger := loggerPkg.GetNopLogger()
	generator := NewIsLatestGenerator(*logger)

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestLocalVersionGeneratorLocalInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameLocalVersion: 3,
	}

	logger := loggerPkg.GetNopLogger()
	generator := NewIsLatestGenerator(*logger)
	generator.Get(state)
}

func TestLocalVersionGeneratorRemoteEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{
		constants.FetcherNameLocalVersion: types.VersionInfo{Version: "1.2.3", Name: "gaiad"},
	}

	logger := loggerPkg.GetNopLogger()
	generator := NewIsLatestGenerator(*logger)

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestLocalVersionGeneratorRemoteInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameLocalVersion:  types.VersionInfo{Version: "1.2.3", Name: "gaiad"},
		constants.FetcherNameRemoteVersion: 3,
	}

	logger := loggerPkg.GetNopLogger()
	generator := NewIsLatestGenerator(*logger)
	generator.Get(state)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestIsLatestGeneratorLocalSemverInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.
			NewBytesResponder(200, assets.GetBytesOrPanic("github-valid.json")).
			HeaderAdd(http.Header{"x-ratelimit-reset": []string{"12345"}}),
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

	localFetcher := fetchers.NewLocalVersionFetcher(*logger, cosmovisor, tracer)
	localData, _ := localFetcher.Get(context.Background())

	remoteFetcher := fetchers.NewRemoteVersionFetcher(*logger, githubClient, tracer)
	remoteData, _ := remoteFetcher.Get(context.Background())

	state := fetchers.State{
		constants.FetcherNameLocalVersion:  localData,
		constants.FetcherNameRemoteVersion: remoteData,
	}
	generator := NewIsLatestGenerator(*logger)
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestIsLatestGeneratorRemoteSemverInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.
			NewBytesResponder(200, assets.GetBytesOrPanic("github-invalid.json")).
			HeaderAdd(http.Header{"x-ratelimit-reset": []string{"12345"}}),
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

	localFetcher := fetchers.NewLocalVersionFetcher(*logger, cosmovisor, tracer)
	localData, _ := localFetcher.Get(context.Background())

	remoteFetcher := fetchers.NewRemoteVersionFetcher(*logger, githubClient, tracer)
	remoteData, _ := remoteFetcher.Get(context.Background())

	state := fetchers.State{
		constants.FetcherNameLocalVersion:  localData,
		constants.FetcherNameRemoteVersion: remoteData,
	}
	generator := NewIsLatestGenerator(*logger)
	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestIsLatestGeneratorAllOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.
			NewBytesResponder(200, assets.GetBytesOrPanic("github-valid.json")).
			HeaderAdd(http.Header{"x-ratelimit-reset": []string{"12345"}}),
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

	localFetcher := fetchers.NewLocalVersionFetcher(*logger, cosmovisor, tracer)
	localData, _ := localFetcher.Get(context.Background())

	remoteFetcher := fetchers.NewRemoteVersionFetcher(*logger, githubClient, tracer)
	remoteData, _ := remoteFetcher.Get(context.Background())

	state := fetchers.State{
		constants.FetcherNameLocalVersion:  localData,
		constants.FetcherNameRemoteVersion: remoteData,
	}
	generator := NewIsLatestGenerator(*logger)
	metrics := generator.Get(state)
	assert.Len(t, metrics, 1)

	isLatest := metrics[0]
	assert.Equal(t, map[string]string{
		"local_version":  "1.6.4",
		"remote_version": "17.2.0",
	}, isLatest.Labels)
	assert.Zero(t, isLatest.Value)
}
