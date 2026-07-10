package papercuts

import (
	"testing"
	"time"
)

func TestRenderEntryProducesReviewedMarkdown(t *testing.T) {
	t.Parallel()

	capturedAt := time.Date(2026, 7, 9, 22, 5, 18, 605_000_000, time.UTC)
	got, err := renderEntry(entry{
		capturedAt: capturedAt,
		severity:   severityMedium,
		description: "While verifying, the command failed.\r\n\r\n" +
			"## This heading belongs to the description",
		reporter: optionalLabel{present: true, value: "agent"},
		model:    optionalLabel{present: true, value: "gpt-5-codex"},
	})
	if err != nil {
		t.Fatalf("renderEntry() returned error: %v", err)
	}

	want := "\n## 2026-07-09T22:05:18.605Z — medium\n\n" +
		"- Reporter: \"agent\"\n" +
		"- Model: \"gpt-5-codex\"\n\n" +
		"> While verifying, the command failed.\n" +
		">\n" +
		"> ## This heading belongs to the description\n"
	if string(got) != want {
		t.Errorf("renderEntry() =\n%s\nwant:\n%s", got, want)
	}
}
