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

func TestGetGitopiaClientFailToBuildQuery(t *testing.T) {
	t.Parallel()

	config := configPkg.GitConfig{Repository: "gitopia://gitopia/gitopia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGitopia(config, *logger, tracer)
	client.ApiBaseUrl = "://"

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGitopiaClientQueryError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.gitopia.com/gitopia/gitopia/gitopia/gitopia/repository/gitopia/releases/latest",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := configPkg.GitConfig{Repository: "gitopia://gitopia/gitopia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGitopia(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGitopiaClientEndpointError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.gitopia.com/gitopia/gitopia/gitopia/gitopia/repository/gitopia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("gitopia-error.json")),
	)

	config := configPkg.GitConfig{Repository: "gitopia://gitopia/gitopia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGitopia(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "username or address not exists")
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGitopiaClientInvalidJson(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.gitopia.com/gitopia/gitopia/gitopia/gitopia/repository/gitopia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("invalid.toml")),
	)

	config := configPkg.GitConfig{Repository: "gitopia://gitopia/gitopia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGitopia(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character")
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGitopiaClientInvalidResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.gitopia.com/gitopia/gitopia/gitopia/gitopia/repository/gitopia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("empty.json")),
	)

	config := configPkg.GitConfig{Repository: "gitopia://gitopia/gitopia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGitopia(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "malformed response")
	assert.False(t, queryInfo.Success)
	require.Empty(t, release)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetGitopiaClientValid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.gitopia.com/gitopia/gitopia/gitopia/gitopia/repository/gitopia/releases/latest",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("gitopia-valid.json")),
	)

	config := configPkg.GitConfig{Repository: "gitopia://gitopia/gitopia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewGitopia(config, *logger, tracer)

	release, queryInfo, err := client.GetLatestRelease(context.Background())

	require.NoError(t, err)
	assert.True(t, queryInfo.Success)
	require.Equal(t, "v3.3.0", release)
}
