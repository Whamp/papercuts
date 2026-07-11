package papercuts

import (
	"fmt"
	"strings"
)

const guidancePreamble = `<!-- papercuts:begin -->
## Papercuts

Capture a **papercut**—one concrete friction instance encountered while doing other work—when it occurs:

` + "```text\n" + `papercuts capture --severity <level> "<attempt; friction; impact; workaround or current state>"
` + "```\n\n" + `Severity: `

const guidancePostamble = `. Choose the highest applicable level.

Project captures target ` + "`PAPERCUTS.md`" + ` in the exact working directory; run from the directory containing that log. Use ` + "`--global`" + ` only for shared tooling or environment friction. Run ` + "`papercuts capture --help`" + ` for multiline input and other options. When a log is missing, surface the suggested init command and preserve the intended scope.

A capture is complete when the CLI confirms it and the description records one causal instance. Product bugs and feature requests belong in the product issue tracker. Refer to secrets by role, never value. Resume only safe work.
<!-- papercuts:end -->
`

// ManagedSection returns a copy of the canonical AGENTS.md section.
func ManagedSection() []byte {
	section := managedSection()
	return append([]byte{}, section...)
}

func managedSection() []byte {
	var section strings.Builder
	section.WriteString(guidancePreamble)
	for index, definition := range severityDefinitions {
		if index > 0 {
			section.WriteString("; ")
		}
		fmt.Fprintf(&section, "`%s` = %s", definition.Value, definition.Meaning)
	}
	section.WriteString(guidancePostamble)
	return []byte(section.String())
}
