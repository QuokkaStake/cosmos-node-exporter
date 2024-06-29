package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoolToFloat64(t *testing.T) {
	t.Parallel()

	assert.InDelta(t, float64(1), BoolToFloat64(true), 0.001)
	assert.InDelta(t, float64(0), BoolToFloat64(false), 0.001)
}

func TestStringToFloat64(t *testing.T) {
	t.Parallel()

	value, err := StringToFloat64("1.234")
	require.NoError(t, err)
	assert.InDelta(t, 1.234, value, 0.001)

	_, err2 := StringToFloat64("invalid")
	require.Error(t, err2)
}

func TestStringToInt64(t *testing.T) {
	t.Parallel()

	value, err := StringToInt64("1234")
	require.NoError(t, err)
	assert.InDelta(t, int64(1234), value, 0.001)

	_, err2 := StringToInt64("invalid")
	require.Error(t, err2)
}

func TestDecolorityString(t *testing.T) {
	t.Parallel()

	str := "\u001B[90m8:57AM\u001B[0m \u001B[32mINF\u001B[0m committed state \u001B[36mapp_hash=\u001B[0m3DB07B19B0815C3E57D4F716E1434B82CAE48E8E3266183B544901D954F1111A \u001B[36mheight=\u001B[0m14916808 \u001B[36mmodule=\u001B[0mstate \u001B[36mnum_txs=\u001B[0m0"
	expectedStr := "8:57AM INF committed state app_hash=3DB07B19B0815C3E57D4F716E1434B82CAE48E8E3266183B544901D954F1111A height=14916808 module=state num_txs=0"

	assert.Equal(t, expectedStr, DecolorifyString(str))
}
