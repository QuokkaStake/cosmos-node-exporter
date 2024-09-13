package fetchers

import (
	"context"
	"main/pkg/constants"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUptimeFetcher(t *testing.T) {
	t.Parallel()

	fetchers := NewUptimeFetcher()
	assert.True(t, fetchers.Enabled())
	assert.Equal(t, constants.FetcherNameUptime, fetchers.Name())
	assert.Empty(t, fetchers.Dependencies())

	data, queryInfos := fetchers.Get(context.Background())
	assert.Empty(t, queryInfos)
	assert.NotEmpty(t, data)
}
