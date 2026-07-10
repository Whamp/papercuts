package papercuts

import (
	"errors"
	"testing"
)

func TestMergeGuidanceAppendsToPlainFileWithoutChangingExistingBytes(t *testing.T) {
	t.Parallel()

	existing := []byte("# Project rules\n\nKeep this exact.\n")
	got, change, err := mergeGuidance(existing, true)
	if err != nil {
		t.Fatalf("mergeGuidance() returned error: %v", err)
	}
	if change != guidanceUpdated {
		t.Errorf("mergeGuidance() change = %v, want guidanceUpdated", change)
	}
	want := append(append([]byte{}, existing...), '\n')
	want = append(want, managedSection()...)
	if string(got) != string(want) {
		t.Errorf("mergeGuidance() =\n%s\nwant:\n%s", got, want)
	}
}

func TestMergeGuidanceRejectsEveryPapercutsMarkdownHeading(t *testing.T) {
	t.Parallel()

	for _, content := range []string{
		"## Papercuts ##\n",
		"Papercuts\n==========\n",
		"Papercuts\n----------\n",
	} {
		content := content
		t.Run(content, func(t *testing.T) {
			original := []byte(content)
			got, _, err := mergeGuidance(original, true)
			if !errors.Is(err, ErrMalformedGuidance) {
				t.Errorf("mergeGuidance(%q) error = %v, want ErrMalformedGuidance", content, err)
			}
			if got != nil {
				t.Errorf("mergeGuidance(%q) = %q, want nil", content, got)
			}
		})
	}
}
