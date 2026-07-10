package papercuts

import (
	"context"
	"fmt"
	"time"
)

// Effect describes what persistence can prove after an operation.
type Effect uint8

const (
	// EffectUnchanged proves that the target was not changed.
	EffectUnchanged Effect = iota
	// EffectDurable proves that the intended bytes were synced.
	EffectDurable
	// EffectIndeterminate means bytes may have changed and callers must inspect before retrying.
	EffectIndeterminate
)

// InitializeState describes the initialized log state.
type InitializeState uint8

const (
	// InitializeNotPerformed indicates that initialization did not complete.
	InitializeNotPerformed InitializeState = iota
	// InitializeCreated indicates that this invocation created the log.
	InitializeCreated
	// InitializeAlreadyExists indicates that a direct regular log already existed.
	InitializeAlreadyExists
)

// InitializeRequest requests initialization for one selected target.
type InitializeRequest struct {
	Target TargetOptions
}

// InitializeResult reports the independently durable log outcome.
type InitializeResult struct {
	Scope      Scope
	Path       string
	AgentsPath string
	State      InitializeState
	Effect     Effect
}

// FileKindError indicates that a configured path is not a direct regular file.
type FileKindError struct {
	Kind string
}

// Error describes the unexpected filesystem object kind.
func (e *FileKindError) Error() string {
	return fmt.Sprintf("expected a direct regular file, found %s", e.Kind)
}

// OperationError describes a failed filesystem operation without choosing CLI prose.
type OperationError struct {
	Operation        string
	Path             string
	Scope            Scope
	CustomGlobalPath bool
	Effect           Effect
	Err              error
}

// Error returns operation and path context.
func (e *OperationError) Error() string {
	return fmt.Sprintf("%s %q: %v", e.Operation, e.Path, e.Err)
}

// Unwrap preserves the underlying semantic error.
func (e *OperationError) Unwrap() error {
	return e.Err
}

// Service performs complete papercuts filesystem transactions.
type Service struct {
	sources            systemSources
	now                func() time.Time
	capturePersistence capturePersistence
}

// NewService returns a service wired to the operating system.
func NewService() *Service {
	return newService(defaultSystemSources(), time.Now)
}

func newService(sources systemSources, now func() time.Time) *Service {
	return &Service{
		sources:            sources,
		now:                now,
		capturePersistence: defaultCapturePersistence(),
	}
}

func operationError(operation string, target resolvedTarget, effect Effect, err error) error {
	return &OperationError{
		Operation:        operation,
		Path:             target.logPath,
		Scope:            target.scope,
		CustomGlobalPath: target.customGlobalPath,
		Effect:           effect,
		Err:              err,
	}
}

func checkContext(ctx context.Context) error {
	return ctx.Err()
}
