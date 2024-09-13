package generators

import (
	"context"
	"main/pkg/constants"
	"main/pkg/fetchers"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestUptimeGeneratorEmpty(t *testing.T) {
	t.Parallel()

	state := fetchers.State{}

	generator := NewUptimeGenerator()

	metrics := generator.Get(state)
	assert.Empty(t, metrics)
}

func TestUptimeGeneratorInvalid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	state := fetchers.State{
		constants.FetcherNameUptime: 3,
	}

	generator := NewUptimeGenerator()
	generator.Get(state)
}

func TestUptimeGeneratorOk(t *testing.T) {
	t.Parallel()

	fetcher := fetchers.NewUptimeFetcher()
	data, _ := fetcher.Get(context.Background())
	assert.NotNil(t, data)

	state := fetchers.State{
		constants.FetcherNameUptime: data,
	}

	generator := NewUptimeGenerator()

	metrics := generator.Get(state)
	assert.NotEmpty(t, metrics)
}
