package papercuts

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func writeSyncedTemporary(
	target string,
	pattern string,
	content []byte,
	mode fs.FileMode,
) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(target), pattern)
	if err != nil {
		return "", err
	}
	path := file.Name()
	fail := func(cause error) (string, error) {
		return "", errors.Join(cause, file.Close(), removeTemporary(path))
	}
	if err := writeComplete(file, content); err != nil {
		return fail(err)
	}
	if err := file.Chmod(mode); err != nil {
		return fail(err)
	}
	if err := file.Sync(); err != nil {
		return fail(err)
	}
	if err := file.Close(); err != nil {
		return "", errors.Join(err, removeTemporary(path))
	}
	return path, nil
}

func removeTemporary(path string) error {
	err := os.Remove(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
