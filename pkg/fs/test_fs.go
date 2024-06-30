package fs

import (
	"errors"
	"main/assets"
	"os"
)

type TestFS struct {
	FileNotFound bool
	FileError    bool
}

func (fs *TestFS) ReadFile(name string) ([]byte, error) {
	return assets.EmbedFS.ReadFile(name)
}

func (fs *TestFS) ReadDir(name string) ([]os.DirEntry, error) {
	return assets.EmbedFS.ReadDir(name)
}

func (fs *TestFS) Stat(name string) (interface{}, error) {
	if fs.FileNotFound {
		return nil, os.ErrNotExist
	}

	if fs.FileError {
		return nil, errors.New("file load error")
	}

	return nil, nil //nolint:nilnil //used for tests only
}
