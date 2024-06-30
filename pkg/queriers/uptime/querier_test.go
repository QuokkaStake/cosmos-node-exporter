package uptime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUptimeQuerier(t *testing.T) {
	t.Parallel()

	querier := NewQuerier()
	assert.True(t, querier.Enabled())
	assert.Equal(t, "uptime-querier", querier.Name())

	metrics, queryInfos := querier.Get(context.Background())
	assert.Empty(t, queryInfos)
	assert.NotEmpty(t, metrics)
}
