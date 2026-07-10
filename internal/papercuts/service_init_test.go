package papercuts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceInitializeLogCreatesCompleteProjectLog(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	service := newService(systemSources{
		getwd:       func() (string, error) { return directory, nil },
		lookupEnv:   func(string) (string, bool) { return "", false },
		userHomeDir: func() (string, error) { return "", nil },
	}, time.Now)

	got, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}
	wantPath := filepath.Join(directory, "PAPERCUTS.md")
	if got.State != InitializeCreated || got.Effect != EffectDurable || got.Path != wantPath || got.Scope != ProjectScope {
		t.Errorf("InitializeLog() = %#v, want durable created project result", got)
	}
	content, readErr := os.ReadFile(wantPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(log) returned error: %v", readErr)
	}
	if string(content) != "# Papercuts\n" {
		t.Errorf("log contents = %q, want %q", content, "# Papercuts\\n")
	}
	entries, readDirErr := os.ReadDir(directory)
	if readDirErr != nil {
		t.Fatalf("os.ReadDir() returned error: %v", readDirErr)
	}
	if len(entries) != 1 || entries[0].Name() != "PAPERCUTS.md" {
		t.Errorf("project directory entries = %v, want only PAPERCUTS.md", entries)
	}
}
