package papercuts

import (
	"os"
	"testing"
)

func TestManagedSectionMatchesReviewedGuidance(t *testing.T) {
	t.Parallel()

	want, err := os.ReadFile("../../testdata/AGENTS.sample.md")
	if err != nil {
		t.Fatalf("os.ReadFile(golden) returned error: %v", err)
	}
	got := managedSection()
	if string(got) != string(want) {
		t.Errorf("managedSection() =\n%s\nwant:\n%s", got, want)
	}
}
