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
		handled, nextIndex, err := parseTargetOption(name, inlineValue, hasInlineValue, args, index, &command.target)
		if err != nil {
			return command, err
		}
		if handled {
			index = nextIndex
			continue
		}
		switch name {
		case "--help":
			if hasInlineValue {
				return command, &usageError{message: "--help does not take a value"}
			}
			command.help = true
		case "--stdin":
			if hasInlineValue {
				return command, &usageError{message: "--stdin does not take a value"}
			}
			command.stdin = true
		case "--severity", "--reporter", "--model":
			value, valueIndex, err := optionValue(name, inlineValue, hasInlineValue, args, index)
			if err != nil {
				return command, err
			}
			index = valueIndex
			switch name {
			case "--severity":
				severityValue = &value
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
	if err := validateTargetOptions(command.target); err != nil {
		return command, err
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
		handled, nextIndex, err := parseTargetOption(name, inlineValue, hasInlineValue, args, index, &command.target)
		if err != nil {
			return command, err
		}
		if handled {
			index = nextIndex
			continue
		}
		switch name {
		case "--help":
			if hasInlineValue {
				return command, &usageError{message: "--help does not take a value"}
			}
			command.help = true
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
	if err := validateTargetOptions(command.target); err != nil {
		return command, err
	}
	return command, nil
}

func parseTargetOption(
	name string,
	inlineValue string,
	hasInlineValue bool,
	args []string,
	index int,
	target *papercuts.TargetOptions,
) (bool, int, error) {
	switch name {
	case "--project":
		if hasInlineValue {
			return true, index, &usageError{message: "--project does not take a value"}
		}
		target.Project = true
		return true, index, nil
	case "--global":
		if hasInlineValue {
			return true, index, &usageError{message: "--global does not take a value"}
		}
		target.Global = true
		return true, index, nil
	case "--global-path":
		value, nextIndex, err := optionValue(name, inlineValue, hasInlineValue, args, index)
		if err != nil {
			return true, index, err
		}
		target.GlobalPath = &value
		return true, nextIndex, nil
	default:
		return false, index, nil
	}
}

func validateTargetOptions(target papercuts.TargetOptions) error {
	if target.Project && target.Global {
		return &usageError{code: usageScopeConflict, message: "--project and --global cannot be used together"}
	}
	if target.GlobalPath != nil && !target.Global {
		return &usageError{code: usageGlobalPathWithoutGlobal, message: "--global-path requires --global"}
	}
	return nil
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
