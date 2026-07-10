package papercuts

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Whamp/papercuts/internal/filelock"
)

func TestServiceCaptureRollsBackPartialWrite(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	service := serviceForDirectory(directory, time.Now)
	initialized, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}
	before, err := os.ReadFile(initialized.Path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", initialized.Path, err)
	}
	service.capturePersistence.write = func(file *filelock.File, content []byte) (int, error) {
		return file.Write(content[:len(content)/2])
	}

	got, err := service.Capture(t.Context(), CaptureRequest{Severity: severityLow, Description: "friction"})
	var operation *OperationError
	if !errors.As(err, &operation) || operation.Operation != "append capture" {
		t.Errorf("Capture() error = %#v, want append-capture OperationError", err)
	}
	if got.Effect != EffectUnchanged {
		t.Errorf("Capture() effect = %v, want %v after successful rollback", got.Effect, EffectUnchanged)
	}
	after, readErr := os.ReadFile(initialized.Path)
	if readErr != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", initialized.Path, readErr)
	}
	if string(after) != string(before) {
		t.Errorf("Capture() log bytes = %q, want original bytes %q after rollback", after, before)
	}
}

func TestServiceCaptureRollsBackAfterSyncFailure(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	service := serviceForDirectory(directory, time.Now)
	initialized, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}
	before, err := os.ReadFile(initialized.Path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", initialized.Path, err)
	}
	syncFailure := errors.New("forced sync failure")
	syncCalls := 0
	service.capturePersistence.sync = func(file *filelock.File) error {
		syncCalls++
		if syncCalls == 1 {
			return syncFailure
		}
		return file.Sync()
	}

	got, err := service.Capture(t.Context(), CaptureRequest{Severity: severityLow, Description: "friction"})
	if !errors.Is(err, syncFailure) {
		t.Errorf("Capture() error = %v, want forced sync failure", err)
	}
	if got.Effect != EffectUnchanged {
		t.Errorf("Capture() effect = %v, want %v after successful rollback", got.Effect, EffectUnchanged)
	}
	after, readErr := os.ReadFile(initialized.Path)
	if readErr != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", initialized.Path, readErr)
	}
	if string(after) != string(before) {
		t.Errorf("Capture() log bytes = %q, want original bytes %q after rollback", after, before)
	}
}

func TestServiceCaptureReportsIndeterminateRollbackFailure(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	service := serviceForDirectory(directory, time.Now)
	initialized, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}
	before, err := os.ReadFile(initialized.Path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", initialized.Path, err)
	}
	service.capturePersistence.write = func(file *filelock.File, content []byte) (int, error) {
		return file.Write(content[:len(content)/2])
	}
	rollbackFailure := errors.New("forced rollback failure")
	service.capturePersistence.truncate = func(*filelock.File, int64) error {
		return rollbackFailure
	}

	got, err := service.Capture(t.Context(), CaptureRequest{Severity: severityLow, Description: "friction"})
	if !errors.Is(err, rollbackFailure) {
		t.Errorf("Capture() error = %v, want forced rollback failure", err)
	}
	if got.Effect != EffectIndeterminate {
		t.Errorf("Capture() effect = %v, want %v after failed rollback", got.Effect, EffectIndeterminate)
	}
	after, readErr := os.ReadFile(initialized.Path)
	if readErr != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", initialized.Path, readErr)
	}
	if len(after) <= len(before) {
		t.Errorf("Capture() log size = %d, want greater than original size %d after failed rollback", len(after), len(before))
	}
}

func TestServiceCaptureAppendsDurableReviewedEntry(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	capturedAt := time.Date(2026, 7, 9, 22, 5, 18, 605_000_000, time.UTC)
	service := newService(systemSources{
		getwd:       func() (string, error) { return directory, nil },
		lookupEnv:   func(string) (string, bool) { return "", false },
		userHomeDir: func() (string, error) { return "", nil },
	}, func() time.Time { return capturedAt })
	if _, err := service.InitializeLog(t.Context(), InitializeRequest{}); err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}

	reporter := "agent"
	model := "gpt-5-codex"
	got, err := service.Capture(t.Context(), CaptureRequest{
		Severity:    severityMedium,
		Description: "The first command failed.\nThe workaround succeeded.",
		Reporter:    &reporter,
		Model:       &model,
	})
	if err != nil {
		t.Fatalf("Capture() returned error: %v", err)
	}
	wantPath := filepath.Join(directory, "PAPERCUTS.md")
	if got.Effect != EffectDurable || got.Path != wantPath || got.Scope != ProjectScope || got.CapturedAt != capturedAt {
		t.Errorf("Capture() = %#v, want durable project result", got)
	}
	content, readErr := os.ReadFile(wantPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile() returned error: %v", readErr)
	}
	want := "# Papercuts\n\n## 2026-07-09T22:05:18.605Z — medium\n\n" +
		"- Reporter: \"agent\"\n- Model: \"gpt-5-codex\"\n\n" +
		"> The first command failed.\n> The workaround succeeded.\n"
	if string(content) != want {
		t.Errorf("log contents =\n%s\nwant:\n%s", content, want)
	}
}
