package papercuts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Whamp/papercuts/internal/atomicfile"
	"github.com/Whamp/papercuts/internal/filelock"
)

// GuidanceState describes the managed AGENTS.md outcome.
type GuidanceState uint8

const (
	// GuidanceNotPerformed indicates that integration did not complete.
	GuidanceNotPerformed GuidanceState = iota
	// GuidanceCreated indicates that a new AGENTS.md was created.
	GuidanceCreated
	// GuidanceUpdated indicates that an existing AGENTS.md was updated.
	GuidanceUpdated
	// GuidanceUnchanged indicates that the managed section was already current.
	GuidanceUnchanged
)

// GuidanceRequest selects the exact AGENTS.md path to integrate.
type GuidanceRequest struct {
	Path string
}

// GuidanceResult reports the managed guidance effect.
type GuidanceResult struct {
	Path   string
	State  GuidanceState
	Effect Effect
}

var errGuidanceTargetChanged = errors.New("AGENTS.md changed during integration")

// IntegrateGuidance creates or reconciles one managed AGENTS.md section.
func (s *Service) IntegrateGuidance(ctx context.Context, request GuidanceRequest) (GuidanceResult, error) {
	result := GuidanceResult{Path: request.Path, State: GuidanceNotPerformed, Effect: EffectUnchanged}
	if !filepath.IsAbs(request.Path) {
		return result, &ValidationError{Field: "AGENTS.md path", Reason: "must be absolute"}
	}
	for attempt := 0; attempt < 3; attempt++ {
		if err := checkContext(ctx); err != nil {
			return result, guidanceOperationError(request.Path, EffectUnchanged, err)
		}
		info, err := os.Lstat(request.Path)
		if errors.Is(err, fs.ErrNotExist) {
			created, createErr := createGuidance(request.Path)
			if errors.Is(createErr, fs.ErrExist) {
				continue
			}
			if createErr != nil {
				return created, createErr
			}
			return created, nil
		}
		if err != nil {
			return result, guidanceOperationError(request.Path, EffectUnchanged, err)
		}
		if err := requireDirectRegular(info); err != nil {
			return result, guidanceOperationError(request.Path, EffectUnchanged, err)
		}
		reconciled, reconcileErr := reconcileGuidance(ctx, request.Path, info)
		if errors.Is(reconcileErr, errGuidanceTargetChanged) {
			continue
		}
		return reconciled, reconcileErr
	}
	return result, guidanceOperationError(request.Path, EffectUnchanged, fmt.Errorf("AGENTS.md changed repeatedly during integration"))
}

func createGuidance(path string) (GuidanceResult, error) {
	result := GuidanceResult{Path: path, State: GuidanceNotPerformed, Effect: EffectUnchanged}
	temporaryPath, err := writeSyncedTemporary(
		path,
		".papercuts-agents-*",
		managedSection(),
		0o644,
		temporaryModeRespectUmask,
	)
	if err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, err)
	}

	if err := atomicfile.PublishNew(temporaryPath, path); err != nil {
		cleanupErr := removeTemporary(temporaryPath)
		if errors.Is(err, fs.ErrExist) {
			if cleanupErr != nil {
				return result, guidanceOperationError(path, EffectUnchanged, cleanupErr)
			}
			return result, err
		}
		effect := EffectUnchanged
		state := GuidanceNotPerformed
		if atomicfile.Committed(err) {
			effect = EffectIndeterminate
			state = GuidanceCreated
		}
		result.State = state
		result.Effect = effect
		return result, guidanceOperationError(path, effect, errors.Join(err, cleanupErr))
	}
	result.State = GuidanceCreated
	result.Effect = EffectDurable
	return result, nil
}

func reconcileGuidance(ctx context.Context, path string, before fs.FileInfo) (GuidanceResult, error) {
	result := GuidanceResult{Path: path, State: GuidanceNotPerformed, Effect: EffectUnchanged}
	lockCtx, cancel := context.WithTimeout(ctx, lockTimeout)
	defer cancel()
	file, err := filelock.Open(lockCtx, path, lockRetry)
	if err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, err)
	}

	pathInfo, err := os.Lstat(path)
	if err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	if err := requireDirectRegular(pathInfo); err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	if !os.SameFile(pathInfo, fileInfo) || !os.SameFile(before, fileInfo) {
		if finishErr := finishLocked(file); finishErr != nil {
			return result, guidanceOperationError(path, EffectUnchanged, errors.Join(fmt.Errorf("AGENTS.md changed before lock acquisition"), finishErr))
		}
		return result, guidanceOperationError(path, EffectUnchanged, errGuidanceTargetChanged)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	existing, err := io.ReadAll(file)
	if err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	merged, change, err := mergeGuidance(existing, true)
	if err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}
	if change == guidanceUnchanged {
		result.State = GuidanceUnchanged
		if err := finishLocked(file); err != nil {
			return result, guidanceOperationError(path, EffectUnchanged, err)
		}
		return result, nil
	}

	preservedMode := fileInfo.Mode() & (os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky)
	temporaryPath, err := writeSyncedTemporary(
		path,
		".papercuts-agents-*",
		merged,
		preservedMode,
		temporaryModeExact,
	)
	if err != nil {
		return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, finishLocked(file)))
	}

	currentInfo, err := os.Lstat(path)
	if err != nil || !os.SameFile(currentInfo, fileInfo) {
		cleanupErr := errors.Join(removeTemporary(temporaryPath), finishLocked(file))
		if cleanupErr != nil {
			return result, guidanceOperationError(path, EffectUnchanged, errors.Join(err, fmt.Errorf("AGENTS.md changed before replacement"), cleanupErr))
		}
		return result, guidanceOperationError(path, EffectUnchanged, errGuidanceTargetChanged)
	}
	if err := atomicfile.Replace(temporaryPath, path); err != nil {
		effect := EffectUnchanged
		if atomicfile.Committed(err) {
			effect = EffectIndeterminate
			result.State = GuidanceUpdated
		}
		result.Effect = effect
		return result, guidanceOperationError(path, effect, errors.Join(err, removeTemporary(temporaryPath), finishLocked(file)))
	}

	result.State = GuidanceUpdated
	result.Effect = EffectDurable
	if err := finishLocked(file); err != nil {
		return result, guidanceOperationError(path, EffectDurable, err)
	}
	return result, nil
}

func guidanceOperationError(path string, effect Effect, err error) error {
	return &OperationError{Operation: "integrate guidance", Path: path, Effect: effect, Err: err}
}
