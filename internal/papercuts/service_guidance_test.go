package papercuts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceIntegrateGuidanceCreatesReviewedSection(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "AGENTS.md")
	service := newService(defaultSystemSources(), time.Now)
	got, err := service.IntegrateGuidance(t.Context(), GuidanceRequest{Path: path})
	if err != nil {
		t.Fatalf("IntegrateGuidance() returned error: %v", err)
	}
	if got.State != GuidanceCreated || got.Effect != EffectDurable || got.Path != path {
		t.Errorf("IntegrateGuidance() = %#v, want durable created result", got)
	}
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("os.ReadFile() returned error: %v", readErr)
	}
	if string(content) != string(managedSection()) {
		t.Errorf("AGENTS.md contents =\n%s\nwant:\n%s", content, managedSection())
	}
}
