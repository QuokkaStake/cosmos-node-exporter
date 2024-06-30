package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasUpgrade(t *testing.T) {
	t.Parallel()

	upgrades := UpgradesPresent{"first": true, "second": false}
	assert.True(t, upgrades.HasUpgrade("first"))
	assert.False(t, upgrades.HasUpgrade("second"))
	assert.False(t, upgrades.HasUpgrade("third"))
}
