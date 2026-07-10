package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Whamp/papercuts/internal/papercuts"
)

func TestCaptureDurableFinalizationErrorSaysDoNotRetry(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	err := writeCaptureError(&output, &papercuts.OperationError{
		Operation: "finish capture",
		Path:      "/work/PAPERCUTS.md",
		Scope:     papercuts.ProjectScope,
		Effect:    papercuts.EffectDurable,
		Err:       errors.New("unlock failed"),
	})
	if err != nil {
		t.Fatalf("writeCaptureError() returned error: %v", err)
	}
	if got := output.String(); !strings.Contains(got, "was durably appended") || !strings.Contains(got, "do not retry") {
		t.Errorf("writeCaptureError() = %q, want durable no-retry guidance", got)
	}
}

func TestGuidanceIndeterminateErrorDoesNotClaimUnchanged(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	err := writeGuidanceError(&output, &papercuts.OperationError{
		Operation: "integrate guidance",
		Path:      "/work/AGENTS.md",
		Effect:    papercuts.EffectIndeterminate,
		Err:       errors.New("parent sync failed"),
	}, "/work/PAPERCUTS.md")
	if err != nil {
		t.Fatalf("writeGuidanceError() returned error: %v", err)
	}
	got := output.String()
	if !strings.Contains(got, "may have changed") || !strings.Contains(got, "inspect it before retrying") || strings.Contains(got, "AGENTS.md unchanged") {
		t.Errorf("writeGuidanceError() = %q, want indeterminate inspection guidance", got)
	}
}
