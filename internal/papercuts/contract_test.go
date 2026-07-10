package papercuts

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveTargetUsesGlobalPrecedence(t *testing.T) {
	t.Parallel()

	flagPath := filepath.Join(t.TempDir(), "flag.md")
	environmentPath := filepath.Join(t.TempDir(), "environment.md")
	home := t.TempDir()
	for _, test := range []struct {
		name        string
		options     TargetOptions
		environment string
		want        string
	}{
		{name: "flag", options: TargetOptions{Global: true, GlobalPath: &flagPath}, environment: environmentPath, want: flagPath},
		{name: "environment", options: TargetOptions{Global: true}, environment: environmentPath, want: environmentPath},
		{name: "home default", options: TargetOptions{Global: true}, want: filepath.Join(home, ".papercuts", "PAPERCUTS.md")},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveTarget(test.options, systemSources{
				getwd: func() (string, error) { return "/work", nil },
				lookupEnv: func(string) (string, bool) {
					return test.environment, test.environment != ""
				},
				userHomeDir: func() (string, error) { return home, nil },
			})
			if err != nil {
				t.Fatalf("resolveTarget() returned error: %v", err)
			}
			if got.scope != GlobalScope || got.logPath != test.want {
				t.Errorf("resolveTarget() = %#v, want global path %q", got, test.want)
			}
		})
	}
}

func TestInitializeLogIsIdempotentAndPreservesExistingBytes(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	service := serviceForDirectory(directory, time.Now)
	first, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog(first) returned error: %v", err)
	}
	if err := os.WriteFile(first.Path, []byte("custom bytes"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	second, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog(second) returned error: %v", err)
	}
	if second.State != InitializeAlreadyExists || second.Effect != EffectUnchanged {
		t.Errorf("InitializeLog(second) = %#v", second)
	}
	content, err := os.ReadFile(first.Path)
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	if string(content) != "custom bytes" {
		t.Errorf("existing bytes = %q, want preserved", content)
	}
}

func TestCaptureRejectsSymlinkTarget(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	realPath := filepath.Join(directory, "real.md")
	if err := os.WriteFile(realPath, []byte(logHeader), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	if err := os.Symlink(realPath, filepath.Join(directory, "PAPERCUTS.md")); err != nil {
		t.Skipf("os.Symlink() unavailable: %v", err)
	}
	service := serviceForDirectory(directory, time.Now)
	_, err := service.Capture(t.Context(), CaptureRequest{Severity: severityLow, Description: "friction"})
	var kindError *FileKindError
	if !errors.As(err, &kindError) || kindError.Kind != "symlink" {
		t.Errorf("Capture() error = %#v, want symlink FileKindError", err)
	}
	content, readErr := os.ReadFile(realPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(real) returned error: %v", readErr)
	}
	if string(content) != logHeader {
		t.Errorf("real target changed to %q", content)
	}
}

func TestCaptureDoesNotInitializeEmptyExistingLog(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "PAPERCUTS.md")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	service := serviceForDirectory(directory, func() time.Time {
		return time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC)
	})
	if _, err := service.Capture(t.Context(), CaptureRequest{Severity: severityLow, Description: "friction"}); err != nil {
		t.Fatalf("Capture() returned error: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	if strings.HasPrefix(string(content), logHeader) || !strings.HasPrefix(string(content), "\n## 2026-07-09T00:00:00Z") {
		t.Errorf("Capture() empty-log append = %q", content)
	}
}

func TestCaptureRepairsMissingFinalNewline(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "PAPERCUTS.md")
	if err := os.WriteFile(path, []byte("# Existing without newline"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	service := serviceForDirectory(directory, func() time.Time {
		return time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC)
	})
	_, err := service.Capture(t.Context(), CaptureRequest{Severity: severityLow, Description: "friction"})
	if err != nil {
		t.Fatalf("Capture() returned error: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	if !strings.Contains(string(content), "# Existing without newline\n\n## 2026-07-09T00:00:00Z") {
		t.Errorf("Capture() framing = %q", content)
	}
}

func FuzzMergeGuidanceIsIdempotent(f *testing.F) {
	f.Add([]byte("# Project\n\nRules.\n"), true)
	f.Add(managedSection(), true)
	f.Add([]byte{}, false)
	f.Fuzz(func(t *testing.T, existing []byte, exists bool) {
		first, _, err := mergeGuidance(existing, exists)
		if err != nil {
			return
		}
		second, change, err := mergeGuidance(first, true)
		if err != nil {
			t.Fatalf("mergeGuidance(second) returned error after successful first merge: %v", err)
		}
		if string(first) != string(second) || change != guidanceUnchanged {
			t.Errorf("mergeGuidance() is not idempotent")
		}
	})
}

func serviceForDirectory(directory string, now func() time.Time) *Service {
	return newService(systemSources{
		getwd:       func() (string, error) { return directory, nil },
		lookupEnv:   func(string) (string, bool) { return "", false },
		userHomeDir: func() (string, error) { return "", errors.New("unexpected home lookup") },
	}, now)
}
