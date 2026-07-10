//go:build !windows

package papercuts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInitializeGlobalCreatesPrivateDirectoriesAndLog(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "private", "nested", "PAPERCUTS.md")
	service := serviceForDirectory(root, time.Now)
	result, err := service.InitializeLog(t.Context(), InitializeRequest{Target: TargetOptions{
		Global:     true,
		GlobalPath: &path,
	}})
	if err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}
	for _, createdPath := range []string{filepath.Dir(path), path} {
		info, statErr := os.Stat(createdPath)
		if statErr != nil {
			t.Fatalf("os.Stat(%q) returned error: %v", createdPath, statErr)
		}
		if got := info.Mode().Perm() & 0o077; got != 0 {
			t.Errorf("InitializeLog() group/other mode for %q = %o, want 0", createdPath, got)
		}
	}
	if result.Scope != GlobalScope || result.Effect != EffectDurable {
		t.Errorf("InitializeLog(global) = %#v, want scope %v and effect %v", result, GlobalScope, EffectDurable)
	}
}

func TestCapturePreservesExistingLogMode(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "PAPERCUTS.md")
	if err := os.WriteFile(path, []byte(logHeader), 0o640); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	service := serviceForDirectory(directory, time.Now)
	if _, err := service.Capture(t.Context(), CaptureRequest{Severity: severityLow, Description: "friction"}); err != nil {
		t.Fatalf("Capture() returned error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat() returned error: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Errorf("log mode = %o, want 640", got)
	}
}
