//go:build !windows

package papercuts

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

const umaskTestDirectoryEnvironment = "PAPERCUTS_TEST_UMASK_DIRECTORY"

func TestGuidanceCreationHonorsProcessUmask(t *testing.T) {
	directory := t.TempDir()
	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestGuidanceCreationUmaskHelper$")
	command.Env = append(os.Environ(), umaskTestDirectoryEnvironment+"="+directory)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("umask helper returned error: %v\noutput:\n%s", err, output)
	}

	path := filepath.Join(directory, "AGENTS.md")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) returned error: %v", path, err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Errorf("AGENTS.md mode = %o, want %o after umask 0077", got, want)
	}
}

func TestGuidanceCreationUmaskHelper(t *testing.T) {
	directory := os.Getenv(umaskTestDirectoryEnvironment)
	if directory == "" {
		return
	}
	syscall.Umask(0o077)
	service := serviceForDirectory(directory, time.Now)
	initialized, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}
	if _, err := service.IntegrateGuidance(t.Context(), GuidanceRequest{Path: initialized.AgentsPath}); err != nil {
		t.Fatalf("IntegrateGuidance(%q) returned error: %v", initialized.AgentsPath, err)
	}
}
