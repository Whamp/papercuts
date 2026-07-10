package papercuts

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type optionalLabel struct {
	present bool
	value   string
}

type entry struct {
	capturedAt  time.Time
	severity    Severity
	description string
	reporter    optionalLabel
	model       optionalLabel
}

func renderEntry(e entry) ([]byte, error) {
	if !utf8.ValidString(e.description) {
		return nil, fmt.Errorf("description is not valid UTF-8")
	}

	description := strings.ReplaceAll(e.description, "\r\n", "\n")
	description = strings.ReplaceAll(description, "\r", "\n")

	var rendered strings.Builder
	fmt.Fprintf(&rendered, "\n## %s — %s\n\n", e.capturedAt.UTC().Format(time.RFC3339Nano), e.severity)
	if e.reporter.present {
		fmt.Fprintf(&rendered, "- Reporter: %s\n", strconv.QuoteToGraphic(e.reporter.value))
	}
	if e.model.present {
		fmt.Fprintf(&rendered, "- Model: %s\n", strconv.QuoteToGraphic(e.model.value))
	}
	if e.reporter.present || e.model.present {
		rendered.WriteByte('\n')
	}
	for _, line := range strings.Split(description, "\n") {
		if line == "" {
			rendered.WriteString(">\n")
			continue
		}
		rendered.WriteString("> ")
		rendered.WriteString(line)
		rendered.WriteByte('\n')
	}
	return []byte(rendered.String()), nil
}
