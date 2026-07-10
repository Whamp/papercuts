//go:build !windows

package papercuts

import (
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestInitializeLogRejectsNonRegularTargets(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	shortRoot, err := os.MkdirTemp("/tmp", "papercuts-reject-")
	if err != nil {
		t.Fatalf("os.MkdirTemp(/tmp) returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(shortRoot); err != nil {
			t.Errorf("os.RemoveAll(%q) error = %v, want nil", shortRoot, err)
		}
	})
	tests := []struct {
		name  string
		path  string
		kind  string
		setup func(*testing.T, string)
	}{
		{
			name: "directory",
			path: filepath.Join(root, "directory"),
			kind: "directory",
			setup: func(t *testing.T, path string) {
				t.Helper()
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatalf("os.Mkdir(%q) returned error: %v", path, err)
				}
			},
		},
		{
			name: "named pipe",
			path: filepath.Join(root, "pipe"),
			kind: "named pipe",
			setup: func(t *testing.T, path string) {
				t.Helper()
				if err := syscall.Mkfifo(path, 0o600); err != nil {
					t.Fatalf("syscall.Mkfifo(%q) returned error: %v", path, err)
				}
			},
		},
		{
			name: "socket",
			path: filepath.Join(shortRoot, "socket"),
			kind: "socket",
			setup: func(t *testing.T, path string) {
				t.Helper()
				listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
				if err != nil {
					t.Fatalf("ListenConfig.Listen(unix, %q) returned error: %v", path, err)
				}
				t.Cleanup(func() {
					if err := listener.Close(); err != nil {
						t.Errorf("listener.Close(%q) error = %v, want nil", path, err)
					}
				})
			},
		},
		{
			name:  "device",
			path:  "/dev/null",
			kind:  "device",
			setup: func(*testing.T, string) {},
		},
	}

	service := serviceForDirectory(root, time.Now)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup(t, test.path)
			_, err := service.InitializeLog(t.Context(), InitializeRequest{Target: TargetOptions{
				Global:     true,
				GlobalPath: &test.path,
			}})
			var kindError *FileKindError
			if !errors.As(err, &kindError) || kindError.Kind != test.kind {
				t.Errorf("InitializeLog(%q) error = %#v, want FileKindError kind %q", test.path, err, test.kind)
			}
		})
	}
}

func TestCaptureRejectsUnwritableTarget(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permission checks")
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "PAPERCUTS.md")
	if err := os.WriteFile(path, []byte(logHeader), 0o400); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
	}
	_, err := serviceForDirectory(directory, time.Now).Capture(t.Context(), CaptureRequest{
		Severity:    severityLow,
		Description: "friction",
	})
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("Capture(unwritable target) error = %v, want permission error", err)
	}
}
