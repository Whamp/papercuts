package papercuts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
