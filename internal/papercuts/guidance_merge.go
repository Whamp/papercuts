package papercuts

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
)

const (
	guidanceBegin = "<!-- papercuts:begin -->"
	guidanceEnd   = "<!-- papercuts:end -->"
)

type guidanceChange uint8

const (
	guidanceUnchanged guidanceChange = iota
	guidanceCreated
	guidanceUpdated
)

type lineSpan struct {
	start int
	end   int
	text  string
}

// ErrMalformedGuidance indicates unsafe or ambiguous managed-region structure.
var ErrMalformedGuidance = errors.New("managed papercuts guidance is malformed")

var (
	papercutsATXHeading    = regexp.MustCompile(`(?i)^[ ]{0,3}#{1,6}[ \t]+papercuts(?:[ \t]+#+)?[ \t]*$`)
	papercutsSetextTitle   = regexp.MustCompile(`(?i)^[ ]{0,3}papercuts[ \t]*$`)
	papercutsSetextDivider = regexp.MustCompile(`^[ ]{0,3}(?:=+|-+)[ \t]*$`)
)

func mergeGuidance(existing []byte, exists bool) ([]byte, guidanceChange, error) {
	section := sectionForNewlineStyle(existing)
	if !exists {
		return section, guidanceCreated, nil
	}

	lines := splitLineSpans(existing)
	var begins []lineSpan
	var ends []lineSpan
	for _, line := range lines {
		switch line.text {
		case guidanceBegin:
			begins = append(begins, line)
		case guidanceEnd:
			ends = append(ends, line)
		}
	}

	if len(begins) == 0 && len(ends) == 0 {
		if containsPapercutsHeading(lines) {
			return nil, guidanceUnchanged, fmt.Errorf("%w: papercuts heading exists without managed markers", ErrMalformedGuidance)
		}
		merged := append([]byte{}, existing...)
		merged = append(merged, appendSeparator(existing)...)
		merged = append(merged, section...)
		return merged, guidanceUpdated, nil
	}
	if len(begins) != 1 || len(ends) != 1 || begins[0].start >= ends[0].start {
		return nil, guidanceUnchanged, fmt.Errorf("%w: papercuts managed markers are malformed", ErrMalformedGuidance)
	}

	merged := make([]byte, 0, len(existing)+len(section))
	merged = append(merged, existing[:begins[0].start]...)
	merged = append(merged, section...)
	merged = append(merged, existing[ends[0].end:]...)
	if bytes.Equal(merged, existing) {
		return existing, guidanceUnchanged, nil
	}
	return merged, guidanceUpdated, nil
}

func containsPapercutsHeading(lines []lineSpan) bool {
	for index, line := range lines {
		if papercutsATXHeading.MatchString(line.text) {
			return true
		}
		if index+1 < len(lines) && papercutsSetextTitle.MatchString(line.text) && papercutsSetextDivider.MatchString(lines[index+1].text) {
			return true
		}
	}
	return false
}

func sectionForNewlineStyle(existing []byte) []byte {
	section := managedSection()
	if usesConsistentCRLF(existing) {
		return bytes.ReplaceAll(section, []byte("\n"), []byte("\r\n"))
	}
	return section
}

func appendSeparator(existing []byte) []byte {
	if len(existing) == 0 {
		return nil
	}
	newline := []byte("\n")
	if usesConsistentCRLF(existing) {
		newline = []byte("\r\n")
	}
	two := append(append([]byte{}, newline...), newline...)
	if bytes.HasSuffix(existing, two) {
		return nil
	}
	if bytes.HasSuffix(existing, newline) {
		return append([]byte{}, newline...)
	}
	return two
}

func usesConsistentCRLF(content []byte) bool {
	if !bytes.Contains(content, []byte("\r\n")) {
		return false
	}
	remainder := bytes.ReplaceAll(content, []byte("\r\n"), nil)
	return !bytes.ContainsAny(remainder, "\r\n")
}

func splitLineSpans(content []byte) []lineSpan {
	var spans []lineSpan
	for start := 0; start < len(content); {
		relativeEnd := bytes.IndexByte(content[start:], '\n')
		end := len(content)
		textEnd := end
		if relativeEnd >= 0 {
			textEnd = start + relativeEnd
			end = textEnd + 1
		}
		if textEnd > start && content[textEnd-1] == '\r' {
			textEnd--
		}
		spans = append(spans, lineSpan{start: start, end: end, text: string(content[start:textEnd])})
		start = end
	}
	return spans
}
