package papercuts

import (
	"fmt"
	"strings"
)

const guidancePreamble = `<!-- papercuts:begin -->
## Papercuts

Capture concrete workflow friction encountered while pursuing another task:

` + "```text\n" + `papercuts capture --severity <low|medium|high> "<what you tried, what happened, and the impact>"
papercuts capture --severity <low|medium|high> --stdin
` + "```\n\n"

const guidancePostamble = `- When several definitions apply, use the highest severity.

Project scope is the default. Add ` + "`--global`" + ` only for friction outside this project or in shared tooling. Supply exactly one quoted description, or pipe a multiline description with ` + "`--stdin`" + `. Describe one instance and include a workaround or current state when known. Capture fatal and non-fatal friction; continue safe work, but do not run or continue unsafe work. Never include secrets. Use the product issue tracker for product bugs and feature requests.
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
	for _, definition := range severityDefinitions {
		fmt.Fprintf(&section, "- `%s`: %s\n", definition.Value, definition.Meaning)
	}
	section.WriteString(guidancePostamble)
	return []byte(section.String())
}
