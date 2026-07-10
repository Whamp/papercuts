package papercuts

import (
	"errors"
	"testing"
	"time"
)

func TestPrepareEntryValidatesBeforeReadingClock(t *testing.T) {
	t.Parallel()

	clockCalls := 0
	now := func() time.Time {
		clockCalls++
		return time.Date(2026, 7, 9, 22, 5, 18, 0, time.UTC)
	}

	for _, test := range []struct {
		name        string
		severity    Severity
		description string
		reporter    *string
	}{
		{name: "invalid severity", description: "friction"},
		{name: "empty description", severity: severityLow, description: " \n\t"},
		{name: "invalid UTF-8", severity: severityLow, description: string([]byte{0xff})},
		{name: "empty reporter", severity: severityLow, description: "friction", reporter: stringPointer("  ")},
		{name: "multiline reporter", severity: severityLow, description: "friction", reporter: stringPointer("agent\nother")},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := prepareEntry(test.severity, test.description, test.reporter, nil, now); err == nil {
				t.Errorf("prepareEntry() returned nil error, want rejection")
			}
		})
	}

	if clockCalls != 0 {
		t.Errorf("prepareEntry() called clock %d times for invalid input, want 0", clockCalls)
	}
}

func TestPrepareEntryReportsTrimmedEmptyDescription(t *testing.T) {
	t.Parallel()

	_, err := prepareEntry(severityLow, " \n\t", nil, nil, time.Now)
	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("prepareEntry() error = %v, want ValidationError", err)
	}
	if validationError.Field != "description" || validationError.Reason != "is empty after trimming" {
		t.Errorf("prepareEntry(trimmed-empty description) ValidationError = %#v, want field %q and reason %q", validationError, "description", "is empty after trimming")
	}
}

func TestPrepareEntryAllowsEscapableLabelControls(t *testing.T) {
	t.Parallel()

	reporter := "agent\tworker"
	got, err := prepareEntry(severityLow, "friction", &reporter, nil, time.Now)
	if err != nil {
		t.Fatalf("prepareEntry() returned error: %v", err)
	}
	if got.reporter.value != reporter {
		t.Errorf("prepareEntry() reporter = %q, want %q", got.reporter.value, reporter)
	}
}

func stringPointer(value string) *string {
	return &value
}
