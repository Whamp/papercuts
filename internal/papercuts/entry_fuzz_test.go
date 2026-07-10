package papercuts

import (
	"bytes"
	"testing"
	"time"
)

func FuzzRenderedDescriptionCannotCreateEntryBoundary(f *testing.F) {
	f.Add("ordinary friction")
	f.Add("## forged entry\n\nbody")
	f.Add("line one\r\n\r\nline two")
	f.Fuzz(func(t *testing.T, description string) {
		prepared, err := prepareEntry(severityLow, description, nil, nil, func() time.Time {
			return time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC)
		})
		if err != nil {
			return
		}
		rendered, err := renderEntry(prepared)
		if err != nil {
			t.Fatalf("renderEntry() returned error after prepareEntry() succeeded: %v", err)
		}
		if !bytes.HasPrefix(rendered, []byte("\n## ")) || !bytes.HasSuffix(rendered, []byte("\n")) {
			t.Errorf("renderEntry() framing = %q", rendered)
		}
		if got := bytes.Count(rendered, []byte("\n## ")); got != 1 {
			t.Errorf("renderEntry() emitted %d entry boundaries for one description\n%q", got, rendered)
		}
	})
}
