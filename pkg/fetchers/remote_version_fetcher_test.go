package fetchers

import (
	"context"
	"errors"
	"main/assets"
	"main/pkg/clients/git"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

func TestRemoteVersionFetcherBase(t *testing.T) {
	t.Parallel()

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := git.NewGithub(config, *logger, tracer)
	fetcher := NewRemoteVersionFetcher(*logger, client, tracer)
	assert.True(t, fetcher.Enabled())
	assert.Equal(t, constants.FetcherNameRemoteVersion, fetcher.Name())
	assert.Empty(t, fetcher.Dependencies())
}

//nolint:paralleltest // disabled due to httpmock usage
func TestRemoteVersionFetcherFail(t *testing.T) {
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
	client := git.NewGithub(config, *logger, tracer)
	fetcher := NewRemoteVersionFetcher(*logger, client, tracer)

	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.False(t, queryInfos[0].Success)
	assert.Nil(t, data)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestRemoteVersionFetcherOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.
			NewBytesResponder(200, assets.GetBytesOrPanic("github-valid.json")).
			HeaderAdd(http.Header{"x-ratelimit-reset": []string{"12345"}}),
	)

	config := configPkg.GitConfig{Repository: "https://github.com/cosmos/gaia"}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := git.NewGithub(config, *logger, tracer)
	fetcher := NewRemoteVersionFetcher(*logger, client, tracer)

	data, queryInfos := fetcher.Get(context.Background())
	assert.Len(t, queryInfos, 1)
	assert.True(t, queryInfos[0].Success)
	assert.Equal(t, "17.2.0", data)
}
