package cli

import (
	"fmt"
	"strings"

	"github.com/Whamp/papercuts/internal/papercuts"
)

func rootHelp() string {
	return `Record workflow friction encountered while doing other work.

Usage:
  papercuts <command> [options]

Commands:
  capture  Append one papercut to a project or global log
  init     Initialize a log and optionally add AGENTS.md guidance
  version  Show the version

Options:
  -h, --help  Show help
  --version   Show the version

Run ` + "`papercuts <command> --help`" + ` for command options.
`
}

func initHelp() string {
	return `Initialize a project or global Papercuts log. Project scope is the default.

Usage:
  papercuts init [options]

Options:
  --project             Initialize ./PAPERCUTS.md (default)
  --global              Initialize the global log
  --global-path <path>  Override the global log; requires --global
  --agents              Create or update ./AGENTS.md guidance
  --no-agents           Leave AGENTS.md unchanged and do not prompt
  -h, --help            Show help

Without --agents or --no-agents, a terminal session asks before changing AGENTS.md. Non-terminal execution leaves AGENTS.md unchanged.

Environment:
  PAPERCUTS_GLOBAL_PATH  Global log override used when --global-path is absent
`
}

func captureHelp() string {
	var help strings.Builder
	help.WriteString(`Append one papercut to a project or global log. Project scope is the default.

Usage:
  papercuts capture --severity <low|medium|high> [options] <description>
  papercuts capture --severity <low|medium|high> [options] --stdin

Severity:
`)
	for _, definition := range papercuts.SeverityDefinitions() {
		fmt.Fprintf(&help, "  %-8s%s\n", definition.Value, definition.HelpSummary)
	}
	help.WriteString(`
Options:
  --severity <level>    Required severity: low, medium, or high
  --project             Write to ./PAPERCUTS.md (default)
  --global              Write to the global log
  --global-path <path>  Override the global log; requires --global
  --reporter <label>    Attach a reporter label
  --model <label>       Attach a model label
  --stdin               Read the complete description from standard input
  -h, --help            Show help

Environment:
  PAPERCUTS_GLOBAL_PATH  Global log override used when --global-path is absent
`)
	return help.String()
}
