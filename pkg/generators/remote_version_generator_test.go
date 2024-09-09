package generators

import (
	"context"
	"main/assets"
	"main/pkg/clients/git"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestRemoteVersionGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewRemoteVersionGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestRemoteVersionGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameRemoteVersion: 3,
	}

	generator := NewRemoteVersionGenerator()
	generator.Get(state)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestRemoteVersionGeneratorOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/cosmos/gaia/releases/latest",
		httpmock.
			NewBytesResponder(200, assets.GetBytesOrPanic("github-valid.json")).
			HeaderAdd(http.Header{"x-ratelimit-reset": []string{"12345"}}),
	)

	config := configPkg.GitConfig{
		Repository: "https://github.com/cosmos/gaia",
	}
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := git.NewGithub(config, *logger, tracer)
	fetcher := fetchers.NewRemoteVersionFetcher(*logger, client, tracer)

	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameRemoteVersion: data,
	}

	generator := NewRemoteVersionGenerator()

	metrics := generator.Get(state)
	assert.Len(t, metrics, 1)

	remoteVersion := metrics[0]
	assert.Equal(t, map[string]string{
		"version": "17.2.0",
	}, remoteVersion.Labels)
	assert.InDelta(t, 1, remoteVersion.Value, 0.01)
}
