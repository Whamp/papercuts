// Package filelock opens and exclusively locks an existing regular file.
package filelock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

// File is an open file with a held cross-process exclusive lock.
type File struct {
	*os.File
	locked bool
}

// Open opens path read/write and waits for an exclusive lock until ctx ends.
func Open(ctx context.Context, path string, retryDelay time.Duration) (*File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if retryDelay <= 0 {
		return nil, fmt.Errorf("retry delay must be positive")
	}

	file, err := openFile(path)
	if err != nil {
		return nil, err
	}
	for {
		locked, lockErr := tryLock(file)
		if lockErr != nil {
			return nil, errors.Join(lockErr, file.Close())
		}
		if locked {
			return &File{File: file, locked: true}, nil
		}

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(ctx.Err(), file.Close())
		case <-timer.C:
		}
	}
}

// Unlock releases the held exclusive lock.
func (f *File) Unlock() error {
	if !f.locked {
		return fmt.Errorf("file %q is not locked", f.Name())
	}
	if err := unlock(f.File); err != nil {
		return err
	}
	f.locked = false
	return nil
}
