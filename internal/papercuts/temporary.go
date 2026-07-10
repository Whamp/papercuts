package papercuts

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type temporaryModePolicy uint8

const (
	temporaryModeRespectUmask temporaryModePolicy = iota
	temporaryModeExact
)

func writeSyncedTemporary(
	target string,
	pattern string,
	content []byte,
	mode fs.FileMode,
	modePolicy temporaryModePolicy,
) (string, error) {
	creationMode := mode
	if modePolicy == temporaryModeExact {
		creationMode = 0o600
	}
	file, err := openExclusiveTemporary(filepath.Dir(target), pattern, creationMode)
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
	if modePolicy == temporaryModeExact {
		if err := file.Chmod(mode); err != nil {
			return fail(err)
		}
	}
	if err := file.Sync(); err != nil {
		return fail(err)
	}
	if err := file.Close(); err != nil {
		return "", errors.Join(err, removeTemporary(path))
	}
	return path, nil
}

func openExclusiveTemporary(directory string, pattern string, mode fs.FileMode) (*os.File, error) {
	if filepath.Base(pattern) != pattern || strings.Count(pattern, "*") != 1 {
		return nil, fmt.Errorf("invalid temporary pattern %q", pattern)
	}
	for range 100 {
		var random [16]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, fmt.Errorf("generate temporary name: %w", err)
		}
		name := strings.Replace(pattern, "*", hex.EncodeToString(random[:]), 1)
		path := filepath.Join(directory, name)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, mode)
		if err == nil {
			return file, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("create unique temporary file in %q", directory)
}

func removeTemporary(path string) error {
	err := os.Remove(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
