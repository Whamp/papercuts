package papercuts

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/Whamp/papercuts/internal/filelock"
)

const (
	lockTimeout = 5 * time.Second
	lockRetry   = 25 * time.Millisecond
)

func lockExistingTarget(ctx context.Context, target resolvedTarget) (*filelock.File, fs.FileInfo, error) {
	lockCtx, cancel := context.WithTimeout(ctx, lockTimeout)
	defer cancel()
	file, err := filelock.Open(lockCtx, target.logPath, lockRetry)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			err = errors.Join(ErrLockTimeout, err)
		}
		return nil, nil, operationError("lock log", target, EffectUnchanged, err)
	}

	pathInfo, err := os.Lstat(target.logPath)
	if err != nil {
		return nil, nil, operationError("revalidate log", target, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, operationError("inspect locked log", target, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	if err := requireDirectRegular(pathInfo); err != nil {
		return nil, nil, operationError("revalidate log", target, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	if !fileInfo.Mode().IsRegular() || !os.SameFile(pathInfo, fileInfo) {
		identityErr := fmt.Errorf("locked file no longer matches configured path")
		return nil, nil, operationError("revalidate log", target, EffectUnchanged, errors.Join(identityErr, finishLocked(file)))
	}
	return file, fileInfo, nil
}

func finishLocked(file *filelock.File) error {
	unlockErr := file.Unlock()
	closeErr := file.Close()
	return errors.Join(unlockErr, closeErr)
}
