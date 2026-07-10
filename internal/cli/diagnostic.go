package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Whamp/papercuts/internal/papercuts"
)

func writeInitError(output io.Writer, err error) error {
	var operationError *papercuts.OperationError
	var kindError *papercuts.FileKindError
	if errors.As(err, &operationError) && operationError.Effect == papercuts.EffectIndeterminate {
		_, writeErr := fmt.Fprintf(output, "papercuts: init: initialization of %q may have completed; inspect it before retrying: %v\n", operationError.Path, operationError.Err)
		return writeErr
	}
	if errors.As(err, &operationError) && operationError.Effect == papercuts.EffectDurable {
		_, writeErr := fmt.Fprintf(output, "papercuts: init: %q was durably initialized, but finalization failed; do not retry: %v\n", operationError.Path, operationError.Err)
		return writeErr
	}
	if errors.As(err, &operationError) && errors.As(err, &kindError) {
		_, writeErr := fmt.Fprintf(output, "papercuts: init: cannot use %q as a log: expected a regular file, not %s\n", operationError.Path, kindError.Kind)
		return writeErr
	}
	_, writeErr := fmt.Fprintf(output, "papercuts: init: %v\n", err)
	return writeErr
}

func writeGuidanceError(output io.Writer, err error, logPath string) error {
	var operationError *papercuts.OperationError
	if errors.As(err, &operationError) && operationError.Effect == papercuts.EffectIndeterminate {
		_, writeErr := fmt.Fprintf(output, "papercuts: init: %q may have changed; log initialized at %s; inspect it before retrying: %v\n", operationError.Path, logPath, operationError.Err)
		return writeErr
	}
	if errors.As(err, &operationError) && operationError.Effect == papercuts.EffectDurable {
		_, writeErr := fmt.Fprintf(output, "papercuts: init: guidance in %q was durably updated, but finalization failed; log initialized at %s; do not retry: %v\n", operationError.Path, logPath, operationError.Err)
		return writeErr
	}
	if errors.As(err, &operationError) && errors.Is(err, papercuts.ErrMalformedGuidance) {
		_, writeErr := fmt.Fprintf(output, "papercuts: init: Papercuts markers in %q are malformed; log initialized at %q; AGENTS.md unchanged\n", operationError.Path, logPath)
		return writeErr
	}
	_, writeErr := fmt.Fprintf(output, "papercuts: init: %v; log initialized at %s; AGENTS.md unchanged\n", err, logPath)
	return writeErr
}

func writeCaptureError(output io.Writer, err error) error {
	var operationError *papercuts.OperationError
	if !errors.As(err, &operationError) {
		_, writeErr := fmt.Fprintf(output, "papercuts: capture: %v\n", err)
		return writeErr
	}
	if errors.Is(err, papercuts.ErrNotInitialized) {
		if operationError.Scope == papercuts.GlobalScope {
			if operationError.CustomGlobalPath {
				_, writeErr := fmt.Fprintf(output, "papercuts: capture: global log not found at %q; run `papercuts init --global --global-path %q`\n", operationError.Path, operationError.Path)
				return writeErr
			}
			_, writeErr := fmt.Fprintf(output, "papercuts: capture: global log not found at %q; run `papercuts init --global`\n", operationError.Path)
			return writeErr
		}
		_, writeErr := fmt.Fprintf(output, "papercuts: capture: project log not found at %q; run `papercuts init --project` in that directory\n", operationError.Path)
		return writeErr
	}
	if errors.Is(err, papercuts.ErrLockTimeout) {
		_, writeErr := fmt.Fprintf(output, "papercuts: capture: timed out after 5s waiting to write %q\n", operationError.Path)
		return writeErr
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		_, writeErr := fmt.Fprintf(output, "papercuts: capture: canceled while waiting to write %q: %v\n", operationError.Path, operationError.Err)
		return writeErr
	}
	if operationError.Effect == papercuts.EffectIndeterminate {
		_, writeErr := fmt.Fprintf(output, "papercuts: capture: append to %q may be incomplete; inspect the log before retrying: %v\n", operationError.Path, operationError.Err)
		return writeErr
	}
	if operationError.Effect == papercuts.EffectDurable {
		_, writeErr := fmt.Fprintf(output, "papercuts: capture: papercut was durably appended to %q, but finalization failed; do not retry: %v\n", operationError.Path, operationError.Err)
		return writeErr
	}
	_, writeErr := fmt.Fprintf(output, "papercuts: capture: cannot append to %q: %v\n", operationError.Path, operationError.Err)
	return writeErr
}
