package fs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativeFsReadFile(t *testing.T) {
	t.Parallel()

	fs := OsFS{}
	file, err := fs.ReadFile("not-found.txt")
	require.Error(t, err)
	require.Nil(t, file)
}

func TestNativeFsReadDir(t *testing.T) {
	t.Parallel()

	fs := OsFS{}
	file, err := fs.ReadDir("not-found.txt")
	require.Error(t, err)
	require.Nil(t, file)
}

func TestNativeFsStat(t *testing.T) {
	t.Parallel()

	fs := OsFS{}
	stat, err := fs.Stat("not-found.txt")
	require.Error(t, err)
	require.Nil(t, stat)
}
