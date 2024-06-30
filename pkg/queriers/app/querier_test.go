package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppQuerier(t *testing.T) {
	t.Parallel()

	querier := NewQuerier("1.2.3")
	assert.True(t, querier.Enabled())
	assert.Equal(t, "app-querier", querier.Name())

	metrics, queryInfos := querier.Get(context.Background())
	assert.Empty(t, queryInfos)
	assert.NotEmpty(t, metrics)
}
