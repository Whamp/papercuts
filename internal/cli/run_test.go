package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/Whamp/papercuts/internal/buildinfo"
	"github.com/Whamp/papercuts/internal/papercuts"
)

type recordingOperations struct {
	captureRequest  papercuts.CaptureRequest
	captureResult   papercuts.CaptureResult
	captureErr      error
	initializeCalls int
	initializeReq   papercuts.InitializeRequest
	initializeRes   papercuts.InitializeResult
	initializeErr   error
	guidanceCalls   int
	guidanceReq     papercuts.GuidanceRequest
	guidanceRes     papercuts.GuidanceResult
	guidanceErr     error
}

func (o *recordingOperations) Capture(_ context.Context, request papercuts.CaptureRequest) (papercuts.CaptureResult, error) {
	o.captureRequest = request
	return o.captureResult, o.captureErr
}

func (o *recordingOperations) InitializeLog(_ context.Context, request papercuts.InitializeRequest) (papercuts.InitializeResult, error) {
	o.initializeCalls++
	o.initializeReq = request
	return o.initializeRes, o.initializeErr
}

func (o *recordingOperations) IntegrateGuidance(_ context.Context, request papercuts.GuidanceRequest) (papercuts.GuidanceResult, error) {
	o.guidanceCalls++
	o.guidanceReq = request
	return o.guidanceRes, o.guidanceErr
}

func TestRunCapturesOneProjectPapercut(t *testing.T) {
	t.Parallel()

	severity, err := papercuts.ParseSeverity("medium")
	if err != nil {
		t.Fatalf("ParseSeverity() returned error: %v", err)
	}
	operations := &recordingOperations{captureResult: papercuts.CaptureResult{
		Scope:    papercuts.ProjectScope,
		Path:     "/work/PAPERCUTS.md",
		Severity: severity,
		Effect:   papercuts.EffectDurable,
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{
		"capture", "--severity", "medium", "The test command required an undocumented working directory.",
	}, IO{
		Stdin:  bytes.NewBufferString("ignored"),
		Stdout: &stdout,
		Stderr: &stderr,
	}, operations, buildinfo.Info{})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0", exitCode)
	}
	if stdout.String() != "Captured medium project papercut in /work/PAPERCUTS.md\n" {
		t.Errorf("Run() stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("Run() stderr = %q, want empty", stderr.String())
	}
	if operations.captureRequest.Description != "The test command required an undocumented working directory." || operations.captureRequest.Severity.String() != "medium" {
		t.Errorf("Capture() request = %#v", operations.captureRequest)
	}
}

func TestRunReadsMultilineInputOnlyWithExplicitStdin(t *testing.T) {
	t.Parallel()

	severity, err := papercuts.ParseSeverity("high")
	if err != nil {
		t.Fatalf("ParseSeverity() returned error: %v", err)
	}
	operations := &recordingOperations{captureResult: papercuts.CaptureResult{
		Scope:    papercuts.GlobalScope,
		Path:     "/var/papercuts.md",
		Severity: severity,
		Effect:   papercuts.EffectDurable,
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{
		"capture", "--global", "--global-path", "/var/papercuts.md", "--severity", "high",
		"--reporter", "agent", "--model", "gpt-5-codex", "--stdin",
	}, IO{
		Stdin:  bytes.NewBufferString("The first command failed.\nThe task is blocked.\n"),
		Stdout: &stdout,
		Stderr: &stderr,
	}, operations, buildinfo.Info{})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if operations.captureRequest.Description != "The first command failed.\nThe task is blocked.\n" {
		t.Errorf("Capture() description = %q", operations.captureRequest.Description)
	}
	if !operations.captureRequest.Target.Global || operations.captureRequest.Target.GlobalPath == nil || *operations.captureRequest.Target.GlobalPath != "/var/papercuts.md" {
		t.Errorf("Capture() target = %#v", operations.captureRequest.Target)
	}
	if operations.captureRequest.Reporter == nil || *operations.captureRequest.Reporter != "agent" || operations.captureRequest.Model == nil || *operations.captureRequest.Model != "gpt-5-codex" {
		t.Errorf("Capture() attribution = reporter %v, model %v", operations.captureRequest.Reporter, operations.captureRequest.Model)
	}
}

func TestRunInitNonInteractiveSkipsGuidanceWithoutReadingStdin(t *testing.T) {
	t.Parallel()

	operations := &recordingOperations{initializeRes: papercuts.InitializeResult{
		Scope:      papercuts.ProjectScope,
		Path:       "/work/PAPERCUTS.md",
		AgentsPath: "/work/AGENTS.md",
		State:      papercuts.InitializeCreated,
		Effect:     papercuts.EffectDurable,
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{"init"}, IO{
		Stdin:  failingReader{},
		Stdout: &stdout,
		Stderr: &stderr,
	}, operations, buildinfo.Info{})

	if exitCode != 0 {
		t.Errorf("Run() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	want := "Initialized project papercuts log at /work/PAPERCUTS.md\nSkipped AGENTS.md integration\n"
	if stdout.String() != want {
		t.Errorf("Run() stdout = %q, want %q", stdout.String(), want)
	}
	if operations.initializeCalls != 1 || operations.guidanceCalls != 0 {
		t.Errorf("operation calls = initialize %d, guidance %d", operations.initializeCalls, operations.guidanceCalls)
	}
}

func TestRunReportsTrimmedEmptyDescription(t *testing.T) {
	t.Parallel()

	operations := &recordingOperations{captureErr: &papercuts.ValidationError{
		Field:  "description",
		Reason: "is empty after trimming",
	}}
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{"capture", "--severity", "low", "   "}, IO{
		Stdin:  failingReader{},
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	}, operations, buildinfo.Info{})
	want := "papercuts: capture: description is empty after trimming\n"
	if exitCode != 2 || stderr.String() != want {
		t.Errorf("Run() = exit %d, stderr %q; want exit 2, stderr %q", exitCode, stderr.String(), want)
	}
}

func TestRunReportsWrongInitTargetKind(t *testing.T) {
	t.Parallel()

	operations := &recordingOperations{initializeErr: &papercuts.OperationError{
		Operation: "initialize log",
		Path:      "/work/PAPERCUTS.md",
		Scope:     papercuts.ProjectScope,
		Effect:    papercuts.EffectUnchanged,
		Err:       &papercuts.FileKindError{Kind: "directory"},
	}}
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{"init", "--no-agents"}, IO{
		Stdin:  failingReader{},
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	}, operations, buildinfo.Info{})

	want := "papercuts: init: cannot use \"/work/PAPERCUTS.md\" as a log: expected a regular file, not directory\n"
	if exitCode != 1 || stderr.String() != want {
		t.Errorf("Run() = exit %d, stderr %q; want exit 1, stderr %q", exitCode, stderr.String(), want)
	}
}

func TestRunReportsMalformedGuidanceAfterDurableLogInit(t *testing.T) {
	t.Parallel()

	operations := &recordingOperations{
		initializeRes: papercuts.InitializeResult{
			Scope:      papercuts.ProjectScope,
			Path:       "/work/PAPERCUTS.md",
			AgentsPath: "/work/AGENTS.md",
			State:      papercuts.InitializeCreated,
			Effect:     papercuts.EffectDurable,
		},
		guidanceErr: &papercuts.OperationError{
			Operation: "integrate guidance",
			Path:      "/work/AGENTS.md",
			Effect:    papercuts.EffectUnchanged,
			Err:       papercuts.ErrMalformedGuidance,
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{"init", "--agents"}, IO{
		Stdin:  failingReader{},
		Stdout: &stdout,
		Stderr: &stderr,
	}, operations, buildinfo.Info{})

	wantOut := "Initialized project papercuts log at /work/PAPERCUTS.md\n"
	wantErr := "papercuts: init: Papercuts markers in \"/work/AGENTS.md\" are malformed; log initialized at \"/work/PAPERCUTS.md\"; AGENTS.md unchanged\n"
	if exitCode != 1 || stdout.String() != wantOut || stderr.String() != wantErr {
		t.Errorf("Run() = exit %d, stdout %q, stderr %q", exitCode, stdout.String(), stderr.String())
	}
}

func TestRunReportsMissingCustomGlobalLogWithExactInitCommand(t *testing.T) {
	t.Parallel()

	operations := &recordingOperations{captureErr: &papercuts.OperationError{
		Operation:        "capture",
		Path:             "/custom/PAPERCUTS.md",
		Scope:            papercuts.GlobalScope,
		CustomGlobalPath: true,
		Effect:           papercuts.EffectUnchanged,
		Err:              papercuts.ErrNotInitialized,
	}}
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{"capture", "--global", "--global-path", "/custom/PAPERCUTS.md", "--severity", "low", "friction"}, IO{
		Stdin:  failingReader{},
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	}, operations, buildinfo.Info{})
	want := "papercuts: capture: global log not found at \"/custom/PAPERCUTS.md\"; run `papercuts init --global --global-path \"/custom/PAPERCUTS.md\"`\n"
	if exitCode != 1 || stderr.String() != want {
		t.Errorf("Run() = exit %d, stderr %q; want exit 1, stderr %q", exitCode, stderr.String(), want)
	}
}

func TestRunReportsMissingProjectLogWithInitCommand(t *testing.T) {
	t.Parallel()

	operations := &recordingOperations{captureErr: &papercuts.OperationError{
		Operation: "capture",
		Path:      "/work/PAPERCUTS.md",
		Scope:     papercuts.ProjectScope,
		Effect:    papercuts.EffectUnchanged,
		Err:       papercuts.ErrNotInitialized,
	}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(t.Context(), []string{"capture", "--severity", "low", "friction"}, IO{
		Stdin:  failingReader{},
		Stdout: &stdout,
		Stderr: &stderr,
	}, operations, buildinfo.Info{})

	if exitCode != 1 {
		t.Errorf("Run() exit code = %d, want 1", exitCode)
	}
	want := "papercuts: capture: project log not found at \"/work/PAPERCUTS.md\"; run `papercuts init --project` in that directory\n"
	if stderr.String() != want {
		t.Errorf("Run() stderr = %q, want %q", stderr.String(), want)
	}
	if stdout.Len() != 0 {
		t.Errorf("Run() stdout = %q, want empty", stdout.String())
	}
}

func TestRunShowsVersionWithoutOperations(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"version"}, {"--version"}} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		exitCode := Run(t.Context(), args, IO{
			Stdin:  failingReader{},
			Stdout: &stdout,
			Stderr: &stderr,
		}, nil, buildinfo.Info{Version: "v0.1.0", Commit: "1a2b3c4", BuildDate: "2026-07-09T12:00:00Z"})
		if exitCode != 0 || stderr.Len() != 0 {
			t.Errorf("Run(%v) = exit %d, stderr %q", args, exitCode, stderr.String())
		}
		want := "papercuts v0.1.0 (commit 1a2b3c4, built 2026-07-09T12:00:00Z)\n"
		if stdout.String() != want {
			t.Errorf("Run(%v) stdout = %q, want %q", args, stdout.String(), want)
		}
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("stdin must not be read")
}
