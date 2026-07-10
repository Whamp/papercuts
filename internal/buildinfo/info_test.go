package buildinfo

import "testing"

func TestInfoStringFormatsReleaseMetadata(t *testing.T) {
	t.Parallel()

	got := (Info{
		Version:   "v0.1.0",
		Commit:    "1a2b3c4d5e6f",
		BuildDate: "2026-07-09T12:00:00Z",
	}).String()
	want := "papercuts v0.1.0 (commit 1a2b3c4, built 2026-07-09T12:00:00Z)"
	if got != want {
		t.Errorf("Info.String() = %q, want %q", got, want)
	}
}

func TestInfoStringFormatsLocalBuild(t *testing.T) {
	t.Parallel()

	if got := (Info{Version: "devel"}).String(); got != "papercuts devel" {
		t.Errorf("Info.String() = %q, want %q", got, "papercuts devel")
	}
}
