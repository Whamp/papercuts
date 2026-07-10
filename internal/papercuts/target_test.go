package papercuts

import (
	"path/filepath"
	"testing"
)

func TestResolveTargetDefaultsToExactWorkingDirectory(t *testing.T) {
	t.Parallel()

	getwdCalls := 0
	workingDirectory := filepath.Join(string(filepath.Separator), "work", "project")
	got, err := resolveTarget(TargetOptions{}, systemSources{
		getwd: func() (string, error) {
			getwdCalls++
			return workingDirectory, nil
		},
		lookupEnv: func(string) (string, bool) {
			t.Fatal("resolveTarget() unexpectedly read environment for project scope")
			return "", false
		},
		userHomeDir: func() (string, error) {
			t.Fatal("resolveTarget() unexpectedly read home for project scope")
			return "", nil
		},
	})
	if err != nil {
		t.Fatalf("resolveTarget() returned error: %v", err)
	}
	if got.scope != ProjectScope || got.logPath != filepath.Join(workingDirectory, "PAPERCUTS.md") || got.agentsPath != filepath.Join(workingDirectory, "AGENTS.md") {
		t.Errorf("resolveTarget() = %#v, want exact project paths", got)
	}
	if getwdCalls != 1 {
		t.Errorf("resolveTarget() called getwd %d times, want 1", getwdCalls)
	}
}
