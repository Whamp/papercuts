package papercuts

import (
	"testing"
)

func TestResolveTargetDefaultsToExactWorkingDirectory(t *testing.T) {
	t.Parallel()

	getwdCalls := 0
	got, err := resolveTarget(TargetOptions{}, systemSources{
		getwd: func() (string, error) {
			getwdCalls++
			return "/work/project", nil
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
	if got.scope != ProjectScope || got.logPath != "/work/project/PAPERCUTS.md" || got.agentsPath != "/work/project/AGENTS.md" {
		t.Errorf("resolveTarget() = %#v, want exact project paths", got)
	}
	if getwdCalls != 1 {
		t.Errorf("resolveTarget() called getwd %d times, want 1", getwdCalls)
	}
}
