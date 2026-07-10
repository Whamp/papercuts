package cli

import (
	"fmt"
	"strings"

	"github.com/Whamp/papercuts/internal/papercuts"
)

type usageErrorCode uint8

const (
	usageOther usageErrorCode = iota
	usageDuplicateOption
	usageScopeConflict
	usageGlobalPathWithoutGlobal
	usageContentConflict
	usageInvalidSeverity
)

type usageError struct {
	code    usageErrorCode
	message string
}

func (e *usageError) Error() string {
	return e.message
}

type agentsChoice uint8

const (
	agentsAsk agentsChoice = iota
	agentsApply
	agentsSkip
)

type initCommand struct {
	target papercuts.TargetOptions
	agents agentsChoice
	help   bool
}

type captureCommand struct {
	target      papercuts.TargetOptions
	severity    papercuts.Severity
	stdin       bool
	description string
	reporter    *string
	model       *string
	help        bool
}

func parseCapture(args []string) (captureCommand, error) {
	var command captureCommand
	var severityValue *string
	var positional []string
	stopOptions := false
	seen := make(map[string]struct{})

	for index := 0; index < len(args); index++ {
		argument := args[index]
		if stopOptions || !strings.HasPrefix(argument, "-") || argument == "-" {
			positional = append(positional, argument)
			continue
		}
		if argument == "--" {
			stopOptions = true
			continue
		}
		name, inlineValue, hasInlineValue := splitOption(argument)
		if name == "-h" {
			name = "--help"
		}
		if err := markSeen(seen, name); err != nil {
			return command, err
		}
		switch name {
		case "--help":
			if hasInlineValue {
				return command, &usageError{message: "--help does not take a value"}
			}
			command.help = true
		case "--project":
			if hasInlineValue {
				return command, &usageError{message: "--project does not take a value"}
			}
			command.target.Project = true
		case "--global":
			if hasInlineValue {
				return command, &usageError{message: "--global does not take a value"}
			}
			command.target.Global = true
		case "--stdin":
			if hasInlineValue {
				return command, &usageError{message: "--stdin does not take a value"}
			}
			command.stdin = true
		case "--severity", "--global-path", "--reporter", "--model":
			value, nextIndex, err := optionValue(name, inlineValue, hasInlineValue, args, index)
			if err != nil {
				return command, err
			}
			index = nextIndex
			switch name {
			case "--severity":
				severityValue = &value
			case "--global-path":
				command.target.GlobalPath = &value
			case "--reporter":
				command.reporter = &value
			case "--model":
				command.model = &value
			}
		default:
			return command, &usageError{message: fmt.Sprintf("unknown option %q", name)}
		}
	}

	if command.help {
		return command, nil
	}
	if severityValue == nil {
		return command, &usageError{message: "--severity is required; choose low, medium, or high"}
	}
	severity, err := papercuts.ParseSeverity(*severityValue)
	if err != nil {
		return command, &usageError{code: usageInvalidSeverity, message: fmt.Sprintf("invalid severity %q; choose low, medium, or high", *severityValue)}
	}
	command.severity = severity
	if command.target.Project && command.target.Global {
		return command, &usageError{code: usageScopeConflict, message: "--project and --global cannot be used together"}
	}
	if command.target.GlobalPath != nil && !command.target.Global {
		return command, &usageError{code: usageGlobalPathWithoutGlobal, message: "--global-path requires --global"}
	}
	if command.stdin && len(positional) > 0 {
		return command, &usageError{code: usageContentConflict, message: "provide either one description argument or --stdin, not both"}
	}
	if !command.stdin && len(positional) != 1 {
		return command, &usageError{message: "provide one description argument or --stdin"}
	}
	if !command.stdin {
		command.description = positional[0]
	}
	return command, nil
}

func parseInit(args []string) (initCommand, error) {
	command := initCommand{agents: agentsAsk}
	seen := make(map[string]struct{})
	for index := 0; index < len(args); index++ {
		argument := args[index]
		name, inlineValue, hasInlineValue := splitOption(argument)
		if name == "-h" {
			name = "--help"
		}
		if err := markSeen(seen, name); err != nil {
			return command, err
		}
		switch name {
		case "--help":
			if hasInlineValue {
				return command, &usageError{message: "--help does not take a value"}
			}
			command.help = true
		case "--project":
			if hasInlineValue {
				return command, &usageError{message: "--project does not take a value"}
			}
			command.target.Project = true
		case "--global":
			if hasInlineValue {
				return command, &usageError{message: "--global does not take a value"}
			}
			command.target.Global = true
		case "--agents":
			if hasInlineValue {
				return command, &usageError{message: "--agents does not take a value"}
			}
			if command.agents == agentsSkip {
				return command, &usageError{message: "--agents and --no-agents cannot be used together"}
			}
			command.agents = agentsApply
		case "--no-agents":
			if hasInlineValue {
				return command, &usageError{message: "--no-agents does not take a value"}
			}
			if command.agents == agentsApply {
				return command, &usageError{message: "--agents and --no-agents cannot be used together"}
			}
			command.agents = agentsSkip
		case "--global-path":
			value, nextIndex, err := optionValue(name, inlineValue, hasInlineValue, args, index)
			if err != nil {
				return command, err
			}
			index = nextIndex
			command.target.GlobalPath = &value
		default:
			if !strings.HasPrefix(argument, "-") {
				return command, &usageError{message: fmt.Sprintf("unexpected argument %q", argument)}
			}
			return command, &usageError{message: fmt.Sprintf("unknown option %q", name)}
		}
	}
	if command.help {
		return command, nil
	}
	if command.target.Project && command.target.Global {
		return command, &usageError{code: usageScopeConflict, message: "--project and --global cannot be used together"}
	}
	if command.target.GlobalPath != nil && !command.target.Global {
		return command, &usageError{code: usageGlobalPathWithoutGlobal, message: "--global-path requires --global"}
	}
	return command, nil
}

func markSeen(seen map[string]struct{}, name string) error {
	if _, ok := seen[name]; ok {
		return &usageError{code: usageDuplicateOption, message: fmt.Sprintf("%s may be specified only once", name)}
	}
	seen[name] = struct{}{}
	return nil
}

func splitOption(argument string) (string, string, bool) {
	name, value, found := strings.Cut(argument, "=")
	return name, value, found
}

func optionValue(
	name string,
	inlineValue string,
	hasInlineValue bool,
	args []string,
	index int,
) (string, int, error) {
	if hasInlineValue {
		return inlineValue, index, nil
	}
	if index+1 >= len(args) {
		return "", index, &usageError{message: fmt.Sprintf("%s requires a value", name)}
	}
	return args[index+1], index + 1, nil
}
