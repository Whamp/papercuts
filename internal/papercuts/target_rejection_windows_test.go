//go:build windows

package papercuts

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestInitializeLogRejectsJunctionReparsePoint(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	junction := filepath.Join(directory, "PAPERCUTS.md")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatalf("os.Mkdir(%q) returned error: %v", target, err)
	}
	output, err := exec.CommandContext(t.Context(), "cmd", "/c", "mklink", "/J", junction, target).CombinedOutput()
	if err != nil {
		t.Fatalf("mklink /J %q %q returned error: %v\noutput:\n%s", junction, target, err, output)
	}

	_, err = serviceForDirectory(directory, time.Now).InitializeLog(t.Context(), InitializeRequest{})
	var kindError *FileKindError
	if !errors.As(err, &kindError) {
		t.Errorf("InitializeLog(junction %q) error = %#v, want FileKindError", junction, err)
	}
	entries, readErr := os.ReadDir(target)
	if readErr != nil {
		t.Fatalf("os.ReadDir(%q) returned error: %v", target, readErr)
	}
	if len(entries) != 0 {
		t.Errorf("InitializeLog(junction) target entries = %v, want none", entries)
	}
}

func TestCaptureRejectsReadOnlyWindowsTarget(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "PAPERCUTS.md")
	if err := os.WriteFile(path, []byte(logHeader), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatalf("os.Chmod(%q, 0400) returned error: %v", path, err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(path, 0o600); err != nil && !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("os.Chmod(%q, 0600) cleanup error = %v, want nil", path, err)
		}
	})

	_, err := serviceForDirectory(directory, time.Now).Capture(t.Context(), CaptureRequest{
		Severity:    severityLow,
		Description: "friction",
	})
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("Capture(read-only Windows target) error = %v, want permission error", err)
	}
}
