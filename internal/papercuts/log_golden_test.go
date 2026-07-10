package papercuts

import (
	"os"
	"testing"
	"time"
)

func TestReviewedSampleLogMatchesRenderer(t *testing.T) {
	t.Parallel()

	reporterAgent := optionalLabel{present: true, value: "agent"}
	model := optionalLabel{present: true, value: "gpt-5-codex"}
	reporterWill := optionalLabel{present: true, value: "Will Hampson"}
	entries := []entry{
		{
			capturedAt:  time.Date(2026, 7, 9, 21, 13, 30, 864_000_000, time.UTC),
			severity:    severityLow,
			description: "The documented search example used an unquoted glob, so zsh rejected the command before `rg` ran. Quoting the glob worked.",
		},
		{
			capturedAt:  time.Date(2026, 7, 9, 22, 5, 18, 605_000_000, time.UTC),
			severity:    severityMedium,
			reporter:    reporterAgent,
			model:       model,
			description: "While verifying the project, `go test ./...` failed because the fixture path assumed a repository-root working directory.\n\nThe workaround was:\n\n```sh\ncd internal/capture && go test ./...\n```",
		},
		{
			capturedAt:  time.Date(2026, 7, 9, 22, 9, 10, 915_000_000, time.UTC),
			severity:    severityHigh,
			reporter:    reporterWill,
			description: "The deployment command selected the production account despite the documented staging flag, so I stopped before applying changes.\n\n## This heading belongs to the description\n\nIt remains inside the blockquote and cannot look like the next papercut.",
		},
	}
	got := []byte(logHeader)
	for _, value := range entries {
		rendered, err := renderEntry(value)
		if err != nil {
			t.Fatalf("renderEntry() returned error: %v", err)
		}
		got = append(got, rendered...)
	}
	want, err := os.ReadFile("../../testdata/PAPERCUTS.sample.md")
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("rendered sample changed\ngot:\n%s\nwant:\n%s", got, want)
	}
}
