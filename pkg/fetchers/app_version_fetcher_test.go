package fetchers

import (
	"context"
	"main/pkg/constants"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppVersionFetcher(t *testing.T) {
	t.Parallel()

	fetchers := NewAppVersionFetcher("1.2.3")
	assert.True(t, fetchers.Enabled())
	assert.Equal(t, constants.FetcherNameAppVersion, fetchers.Name())
	assert.Empty(t, fetchers.Dependencies())

	data, queryInfos := fetchers.Get(context.Background())
	assert.Empty(t, queryInfos)
	assert.Equal(t, "1.2.3", data)
}
