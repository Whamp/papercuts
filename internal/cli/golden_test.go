package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Whamp/papercuts/internal/buildinfo"
	"github.com/Whamp/papercuts/internal/papercuts"
)

func TestReviewedCLIArtifact(t *testing.T) {
	t.Parallel()

	want, err := os.ReadFile("../../testdata/cli-help.txt")
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	build := buildinfo.Info{Version: "v0.1.0", Commit: "1a2b3c4", BuildDate: "2026-07-09T12:00:00Z"}
	got := "$ papercuts --help\n" + rootHelp() +
		"\n$ papercuts capture --help\n" + captureHelp() +
		"\n$ papercuts init --help\n" + initHelp() +
		"\n$ papercuts version\n" + build.String() + "\n" +
		reviewedMessageTranscript(t)
	if got != string(want) {
		t.Errorf("reviewed CLI artifact changed\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func reviewedMessageTranscript(t *testing.T) string {
	t.Helper()

	severity, err := papercuts.ParseSeverity("low")
	if err != nil {
		t.Fatalf("ParseSeverity() returned error: %v", err)
	}
	var transcript strings.Builder
	transcript.WriteString("\n# Canonical success output\n")
	captureSuccess := cliStreams(t, []string{"capture", "--severity", "low", "friction"}, &recordingOperations{
		captureResult: papercuts.CaptureResult{Scope: papercuts.ProjectScope, Path: "<path>", Severity: severity, Effect: papercuts.EffectDurable},
	}).stdout
	transcript.WriteString(strings.Replace(captureSuccess, "low project", "<severity> <project|global>", 1))

	var output bytes.Buffer
	if err := printInitializeResult(&output, papercuts.InitializeResult{Scope: papercuts.ProjectScope, Path: "<path>", State: papercuts.InitializeCreated}); err != nil {
		t.Fatalf("printInitializeResult(created) returned error: %v", err)
	}
	transcript.WriteString(strings.Replace(output.String(), "project", "<project|global>", 1))
	output.Reset()
	if err := printInitializeResult(&output, papercuts.InitializeResult{Scope: papercuts.ProjectScope, Path: "<path>", State: papercuts.InitializeAlreadyExists}); err != nil {
		t.Fatalf("printInitializeResult(existing) returned error: %v", err)
	}
	transcript.WriteString(strings.Replace(output.String(), "project", "<project|global>", 1))
	for _, state := range []papercuts.GuidanceState{papercuts.GuidanceCreated, papercuts.GuidanceUpdated, papercuts.GuidanceUnchanged} {
		output.Reset()
		if err := printGuidanceResult(&output, papercuts.GuidanceResult{Path: "<path>", State: state}); err != nil {
			t.Fatalf("printGuidanceResult(%d) returned error: %v", state, err)
		}
		transcript.WriteString(output.String())
	}
	transcript.WriteString("Skipped AGENTS.md integration\n")

	transcript.WriteString("\n# Canonical diagnostics\n")
	transcript.WriteString(cliStreams(t, []string{"<command>"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "<value>", "friction"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "low"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "low", "--stdin", "friction"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "low", " "}, &recordingOperations{
		captureErr: &papercuts.ValidationError{Field: "description", Reason: "is empty after trimming"},
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--project", "--global", "--severity", "low", "friction"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--global-path", "/path", "--severity", "low", "friction"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--global", "--severity", "low", "friction"}, &recordingOperations{
		captureErr: &papercuts.ValidationError{Field: "global path", Reason: "must be absolute: \"<path>\""},
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "low", "friction"}, &recordingOperations{
		captureErr: operationFailure(papercuts.ProjectScope, "<path>", papercuts.EffectUnchanged, papercuts.ErrNotInitialized),
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--global", "--severity", "low", "friction"}, &recordingOperations{
		captureErr: operationFailure(papercuts.GlobalScope, "<path>", papercuts.EffectUnchanged, papercuts.ErrNotInitialized),
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "low", "friction"}, &recordingOperations{
		captureErr: operationFailure(papercuts.ProjectScope, "<path>", papercuts.EffectUnchanged, errors.Join(papercuts.ErrLockTimeout, context.DeadlineExceeded)),
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "low", "friction"}, &recordingOperations{
		captureErr: operationFailure(papercuts.ProjectScope, "<path>", papercuts.EffectUnchanged, errors.New("<operating-system error>")),
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"capture", "--severity", "low", "friction"}, &recordingOperations{
		captureErr: operationFailure(papercuts.ProjectScope, "<path>", papercuts.EffectIndeterminate, errors.New("<operating-system error>")),
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"init", "--agents", "--no-agents"}, nil).stderr)
	transcript.WriteString(cliStreams(t, []string{"init", "--no-agents"}, &recordingOperations{
		initializeErr: &papercuts.OperationError{Operation: "initialize", Path: "<path>", Scope: papercuts.ProjectScope, Effect: papercuts.EffectUnchanged, Err: &papercuts.FileKindError{Kind: "<kind>"}},
	}).stderr)
	transcript.WriteString(cliStreams(t, []string{"init", "--agents"}, &recordingOperations{
		initializeRes: papercuts.InitializeResult{Scope: papercuts.ProjectScope, Path: "<log-path>", AgentsPath: "<path>", State: papercuts.InitializeCreated, Effect: papercuts.EffectDurable},
		guidanceErr:   &papercuts.OperationError{Operation: "integrate", Path: "<path>", Effect: papercuts.EffectUnchanged, Err: papercuts.ErrMalformedGuidance},
	}).stderr)
	return transcript.String()
}

type commandStreams struct {
	stdout string
	stderr string
}

func cliStreams(t *testing.T, args []string, operations *recordingOperations) commandStreams {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	Run(t.Context(), args, IO{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr}, operations, buildinfo.Info{})
	return commandStreams{stdout: stdout.String(), stderr: stderr.String()}
}

func operationFailure(scope papercuts.Scope, path string, effect papercuts.Effect, err error) error {
	return &papercuts.OperationError{Operation: fmt.Sprintf("%s operation", scope), Path: path, Scope: scope, Effect: effect, Err: err}
}
