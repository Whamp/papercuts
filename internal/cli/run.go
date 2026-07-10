// Package cli implements the papercuts command-line interface.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Whamp/papercuts/internal/buildinfo"
	"github.com/Whamp/papercuts/internal/papercuts"
)

// IO contains process streams and terminal state.
type IO struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	StdinIsTTY bool
}

type operations interface {
	Capture(context.Context, papercuts.CaptureRequest) (papercuts.CaptureResult, error)
	InitializeLog(context.Context, papercuts.InitializeRequest) (papercuts.InitializeResult, error)
	IntegrateGuidance(context.Context, papercuts.GuidanceRequest) (papercuts.GuidanceResult, error)
}

// Run executes one command and returns its process exit code.
func Run(
	ctx context.Context,
	args []string,
	streams IO,
	service operations,
	build buildinfo.Info,
) int {
	if len(args) == 0 {
		return writeExit(streams.Stdout, 0, "%s", rootHelp())
	}
	switch args[0] {
	case "-h", "--help":
		return writeExit(streams.Stdout, 0, "%s", rootHelp())
	case "version", "--version":
		return writeExit(streams.Stdout, 0, "%s\n", build.String())
	case "capture":
		return runCapture(ctx, args[1:], streams, service)
	case "init":
		return runInit(ctx, args[1:], streams, service)
	default:
		return writeExit(streams.Stderr, 2, "papercuts: unknown command %q; run `papercuts --help`\n", args[0])
	}
}

func runInit(ctx context.Context, args []string, streams IO, service operations) int {
	command, err := parseInit(args)
	if err != nil {
		return writeExit(streams.Stderr, 2, "papercuts: init: %v\n", err)
	}
	if command.help {
		return writeExit(streams.Stdout, 0, "%s", initHelp())
	}
	result, err := service.InitializeLog(ctx, papercuts.InitializeRequest{Target: command.target})
	if err != nil {
		var validationError *papercuts.ValidationError
		if errors.As(err, &validationError) {
			return writeExit(streams.Stderr, 2, "papercuts: init: %v\n", validationError)
		}
		if writeErr := writeInitError(streams.Stderr, err); writeErr != nil {
			return 1
		}
		return 1
	}
	if err := printInitializeResult(streams.Stdout, result); err != nil {
		return 1
	}

	applyGuidance := command.agents == agentsApply
	if command.agents == agentsAsk && streams.StdinIsTTY {
		applyGuidance, err = promptForGuidance(streams, result.AgentsPath)
		if err != nil {
			return writeExit(streams.Stderr, 1, "papercuts: init: read AGENTS.md consent: %v\n", err)
		}
	}
	if !applyGuidance {
		return writeExit(streams.Stdout, 0, "Skipped AGENTS.md integration\n")
	}

	guidance, err := service.IntegrateGuidance(ctx, papercuts.GuidanceRequest{Path: result.AgentsPath})
	if err != nil {
		if writeErr := writeGuidanceError(streams.Stderr, err, result.Path); writeErr != nil {
			return 1
		}
		return 1
	}
	if err := printGuidanceResult(streams.Stdout, guidance); err != nil {
		return 1
	}
	return 0
}

func printInitializeResult(output io.Writer, result papercuts.InitializeResult) error {
	if result.State == papercuts.InitializeCreated {
		_, err := fmt.Fprintf(output, "Initialized %s papercuts log at %s\n", result.Scope, result.Path)
		return err
	}
	_, err := fmt.Fprintf(output, "%s papercuts log already exists at %s\n", result.Scope, result.Path)
	return err
}

func promptForGuidance(streams IO, path string) (bool, error) {
	if _, err := fmt.Fprintf(streams.Stderr, "Proposed Papercuts guidance for %s:\n\n%s\nAdd Papercuts guidance to %s? [y/N] ", path, papercuts.ManagedSection(), path); err != nil {
		return false, err
	}
	response, err := bufio.NewReader(streams.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

func printGuidanceResult(output io.Writer, result papercuts.GuidanceResult) error {
	var message string
	switch result.State {
	case papercuts.GuidanceNotPerformed:
		return fmt.Errorf("guidance operation did not report an outcome")
	case papercuts.GuidanceCreated:
		message = fmt.Sprintf("Created Papercuts guidance in %s\n", result.Path)
	case papercuts.GuidanceUpdated:
		message = fmt.Sprintf("Updated Papercuts guidance in %s\n", result.Path)
	case papercuts.GuidanceUnchanged:
		message = fmt.Sprintf("Papercuts guidance is already current in %s\n", result.Path)
	default:
		return fmt.Errorf("unknown guidance state %d", result.State)
	}
	_, err := io.WriteString(output, message)
	return err
}

func runCapture(ctx context.Context, args []string, streams IO, service operations) int {
	command, err := parseCapture(args)
	if err != nil {
		return writeExit(streams.Stderr, 2, "papercuts: capture: %v\n", err)
	}
	if command.help {
		return writeExit(streams.Stdout, 0, "%s", captureHelp())
	}
	if command.stdin {
		content, readErr := io.ReadAll(streams.Stdin)
		if readErr != nil {
			return writeExit(streams.Stderr, 1, "papercuts: capture: read description from stdin: %v\n", readErr)
		}
		command.description = string(content)
	}
	result, err := service.Capture(ctx, papercuts.CaptureRequest{
		Target:      command.target,
		Severity:    command.severity,
		Description: command.description,
		Reporter:    command.reporter,
		Model:       command.model,
	})
	if err != nil {
		var validationError *papercuts.ValidationError
		if errors.As(err, &validationError) {
			return writeExit(streams.Stderr, 2, "papercuts: capture: %v\n", validationError)
		}
		if writeErr := writeCaptureError(streams.Stderr, err); writeErr != nil {
			return 1
		}
		return 1
	}
	return writeExit(streams.Stdout, 0, "Captured %s %s papercut in %s\n", result.Severity, result.Scope, result.Path)
}

func writeExit(output io.Writer, exitCode int, format string, values ...any) int {
	if _, err := fmt.Fprintf(output, format, values...); err != nil {
		return 1
	}
	return exitCode
}
