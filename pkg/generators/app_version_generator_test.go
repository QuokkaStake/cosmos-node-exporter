package generators

import (
	"context"
	"main/pkg/constants"
	"main/pkg/fetchers"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNewAppVersionGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}
	generator := NewAppVersionGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestNewAppVersionGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameAppVersion: 3,
	}

	generator := NewAppVersionGenerator()
	generator.Get(state)
}

func TestNewAppVersionGeneratorOk(t *testing.T) {
	t.Parallel()

	fetcher := fetchers.NewAppVersionFetcher("1.2.3")
	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameAppVersion: data,
	}

	generator := NewAppVersionGenerator()

	metrics := generator.Get(state)
	assert.NotEmpty(t, metrics)
}
