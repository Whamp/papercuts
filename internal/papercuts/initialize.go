package papercuts

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Whamp/papercuts/internal/atomicfile"
)

const logHeader = "# Papercuts\n"

// InitializeLog creates a complete log or reports an existing direct regular log.
func (s *Service) InitializeLog(ctx context.Context, request InitializeRequest) (InitializeResult, error) {
	target, err := resolveTarget(request.Target, s.sources)
	if err != nil {
		return InitializeResult{}, err
	}
	result := InitializeResult{
		Scope:      target.scope,
		Path:       target.logPath,
		AgentsPath: target.agentsPath,
		State:      InitializeNotPerformed,
		Effect:     EffectUnchanged,
	}
	if err := checkContext(ctx); err != nil {
		return result, operationError("initialize log", target, EffectUnchanged, err)
	}
	if target.globalDirs {
		if err := os.MkdirAll(filepath.Dir(target.logPath), 0o700); err != nil {
			return result, operationError("create global log directory", target, EffectUnchanged, err)
		}
	}

	if info, statErr := os.Lstat(target.logPath); statErr == nil {
		if err := requireDirectRegular(info); err != nil {
			return result, operationError("initialize log", target, EffectUnchanged, err)
		}
		file, _, lockErr := lockExistingTarget(ctx, target)
		if lockErr != nil {
			return result, lockErr
		}
		if err := finishLocked(file); err != nil {
			return result, operationError("finish initialization", target, EffectUnchanged, err)
		}
		result.State = InitializeAlreadyExists
		return result, nil
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return result, operationError("inspect log", target, EffectUnchanged, statErr)
	}

	temporaryPath, err := writeSyncedTemporary(
		target.logPath,
		".papercuts-init-*",
		[]byte(logHeader),
		0o600,
		temporaryModeRespectUmask,
	)
	if err != nil {
		return result, operationError("prepare temporary log", target, EffectUnchanged, err)
	}

	if err := atomicfile.PublishNew(temporaryPath, target.logPath); err != nil {
		cleanupErr := removeTemporary(temporaryPath)
		if errors.Is(err, fs.ErrExist) {
			if cleanupErr != nil {
				return result, operationError("remove temporary log", target, EffectUnchanged, cleanupErr)
			}
			file, _, lockErr := lockExistingTarget(ctx, target)
			if lockErr != nil {
				return result, lockErr
			}
			if finishErr := finishLocked(file); finishErr != nil {
				return result, operationError("finish concurrent initialization", target, EffectUnchanged, finishErr)
			}
			result.State = InitializeAlreadyExists
			return result, nil
		}
		if atomicfile.Committed(err) {
			result.State = InitializeCreated
			result.Effect = EffectIndeterminate
			return result, operationError("finish publishing log", target, EffectIndeterminate, errors.Join(err, cleanupErr))
		}
		return result, operationError("publish log", target, EffectUnchanged, errors.Join(err, cleanupErr))
	}
	result.State = InitializeCreated
	result.Effect = EffectDurable
	return result, nil
}

func requireDirectRegular(info fs.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return &FileKindError{Kind: "symlink"}
	}
	if info.Mode().IsRegular() {
		return nil
	}
	kind := "non-regular file"
	switch {
	case info.IsDir():
		kind = "directory"
	case info.Mode()&os.ModeNamedPipe != 0:
		kind = "named pipe"
	case info.Mode()&os.ModeSocket != 0:
		kind = "socket"
	case info.Mode()&os.ModeDevice != 0:
		kind = "device"
	}
	return &FileKindError{Kind: kind}
}

func writeComplete(file *os.File, content []byte) error {
	written, err := file.Write(content)
	if err != nil {
		return err
	}
	if written != len(content) {
		return fmt.Errorf("short write: wrote %d of %d bytes", written, len(content))
	}
	return nil
}
