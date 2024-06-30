package exec

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNativeExec(t *testing.T) {
	t.Parallel()

	exec := NativeCommandExecutor{}
	out, err := exec.RunWithEnv("ls", []string{}, os.Environ())
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}
