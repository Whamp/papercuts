package papercuts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/Whamp/papercuts/internal/filelock"
)

type capturePersistence struct {
	write    func(*filelock.File, []byte) (int, error)
	sync     func(*filelock.File) error
	truncate func(*filelock.File, int64) error
}

func defaultCapturePersistence() capturePersistence {
	return capturePersistence{
		write:    func(file *filelock.File, content []byte) (int, error) { return file.Write(content) },
		sync:     func(file *filelock.File) error { return file.Sync() },
		truncate: func(file *filelock.File, size int64) error { return file.Truncate(size) },
	}
}

// ErrNotInitialized indicates that the selected log does not exist.
var ErrNotInitialized = errors.New("papercuts log is not initialized")

// ErrLockTimeout indicates that the service's five-second lock cap elapsed.
var ErrLockTimeout = errors.New("papercuts lock wait timed out")

// CaptureRequest contains one validated capture intent.
type CaptureRequest struct {
	Target      TargetOptions
	Severity    Severity
	Description string
	Reporter    *string
	Model       *string
}

// CaptureResult reports the selected target and persistence effect.
type CaptureResult struct {
	Scope      Scope
	Path       string
	Severity   Severity
	CapturedAt time.Time
	Effect     Effect
}

// Capture appends one complete entry to the selected initialized log.
func (s *Service) Capture(ctx context.Context, request CaptureRequest) (CaptureResult, error) {
	target, err := resolveTarget(request.Target, s.sources)
	if err != nil {
		return CaptureResult{}, err
	}
	result := CaptureResult{Scope: target.scope, Path: target.logPath, Severity: request.Severity, Effect: EffectUnchanged}

	prepared, err := prepareEntry(request.Severity, request.Description, request.Reporter, request.Model, s.now)
	if err != nil {
		return result, err
	}
	result.CapturedAt = prepared.capturedAt
	rendered, err := renderEntry(prepared)
	if err != nil {
		return result, err
	}

	pathInfo, err := os.Lstat(target.logPath)
	if errors.Is(err, fs.ErrNotExist) {
		return result, operationError("capture", target, EffectUnchanged, ErrNotInitialized)
	}
	if err != nil {
		return result, operationError("inspect log", target, EffectUnchanged, err)
	}
	if err := requireDirectRegular(pathInfo); err != nil {
		return result, operationError("capture", target, EffectUnchanged, err)
	}

	file, fileInfo, err := lockExistingTarget(ctx, target)
	if err != nil {
		return result, err
	}

	originalSize := fileInfo.Size()
	payload, err := appendPayload(file, originalSize, rendered)
	if err != nil {
		return result, operationError("prepare append", target, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return result, operationError("seek log", target, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}

	written, writeErr := s.capturePersistence.write(file, payload)
	if writeErr != nil || written != len(payload) {
		if writeErr == nil {
			writeErr = fmt.Errorf("short write: wrote %d of %d bytes", written, len(payload))
		}
		effect, rollbackErr := s.rollbackAppend(file, originalSize)
		return resultWithEffect(result, effect), operationError("append capture", target, effect, errors.Join(writeErr, rollbackErr, finishLocked(file)))
	}
	if err := s.capturePersistence.sync(file); err != nil {
		effect, rollbackErr := s.rollbackAppend(file, originalSize)
		return resultWithEffect(result, effect), operationError("sync capture", target, effect, errors.Join(err, rollbackErr, finishLocked(file)))
	}

	result.Effect = EffectDurable
	if err := finishLocked(file); err != nil {
		return result, operationError("finish capture", target, EffectDurable, err)
	}
	return result, nil
}

func appendPayload(file *filelock.File, size int64, rendered []byte) ([]byte, error) {
	if size == 0 {
		return rendered, nil
	}
	last := []byte{0}
	if _, err := file.ReadAt(last, size-1); err != nil {
		return nil, err
	}
	if last[0] == '\n' {
		return rendered, nil
	}
	payload := make([]byte, 0, 1+len(rendered))
	payload = append(payload, '\n')
	payload = append(payload, rendered...)
	return payload, nil
}

func (s *Service) rollbackAppend(file *filelock.File, originalSize int64) (Effect, error) {
	if err := s.capturePersistence.truncate(file, originalSize); err != nil {
		return EffectIndeterminate, fmt.Errorf("rollback append: %w", err)
	}
	if err := s.capturePersistence.sync(file); err != nil {
		return EffectIndeterminate, fmt.Errorf("sync rollback: %w", err)
	}
	return EffectUnchanged, nil
}

func resultWithEffect(result CaptureResult, effect Effect) CaptureResult {
	result.Effect = effect
	return result
}
