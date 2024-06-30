package git

import (
	"context"
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGithubClientFailToBuildQuery(t *testing.T) {
	t.Parallel()

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGithub(config, *logger, tracer)
	client.ApiBaseUrl = "://"

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGithubClientQueryError(t *testing.T) {
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
	client := NewGithub(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGithubClientInvalidResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("invalid.toml")),
	)

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGithub(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character")
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGithubClientRateLimit(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("github-error.json")),
	)

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGithub(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "got error from Github: API rate limit exceeded")
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGithubClientValidQueried(t *testing.T) {
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
	client := NewGithub(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	require.Equal(t, "v17.2.0", release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGithubClientCachedNoPreviousResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(304, assets.GetBytesOrPanic("github-valid.json")),
	)

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGithub(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	require.Equal(t, "v17.2.0", release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGithubClientCachedWithPreviousResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.NewBytesResponder(304, assets.GetBytesOrPanic("github-valid.json")),
	)

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia", Token: "aaa"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGithub(config, *logger, tracer)
	client.LastResult = "v1.2.3"

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	require.Equal(t, "v1.2.3", release)
}
