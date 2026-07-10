package cli

import (
	"errors"
	"testing"
)

func TestParseCaptureRejectsAmbiguousInputs(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		args []string
		want usageErrorCode
	}{
		{name: "duplicate severity", args: []string{"--severity", "low", "--severity", "high", "friction"}, want: usageDuplicateOption},
		{name: "duplicate global", args: []string{"--global", "--global", "--severity", "low", "friction"}, want: usageDuplicateOption},
		{name: "conflicting scope", args: []string{"--project", "--global", "--severity", "low", "friction"}, want: usageScopeConflict},
		{name: "path without global", args: []string{"--global-path", "/tmp/log", "--severity", "low", "friction"}, want: usageGlobalPathWithoutGlobal},
		{name: "stdin and argument", args: []string{"--severity", "low", "--stdin", "friction"}, want: usageContentConflict},
		{name: "case sensitive severity", args: []string{"--severity", "LOW", "friction"}, want: usageInvalidSeverity},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseCapture(test.args)
			var usageErr *usageError
			if !errors.As(err, &usageErr) || usageErr.code != test.want {
				t.Errorf("parseCapture(%v) error = %#v, want usage code %d", test.args, err, test.want)
			}
		})
	}
}

func TestParseCaptureDoubleDashAllowsOptionShapedDescription(t *testing.T) {
	t.Parallel()

	got, err := parseCapture([]string{"--severity", "low", "--", "--unexpected-tool-flag"})
	if err != nil {
		t.Fatalf("parseCapture() returned error: %v", err)
	}
	if got.description != "--unexpected-tool-flag" {
		t.Errorf("parseCapture() description = %q", got.description)
	}
}

func TestParseInitRejectsRepeatedConsent(t *testing.T) {
	t.Parallel()

	_, err := parseInit([]string{"--agents", "--agents"})
	var usageErr *usageError
	if !errors.As(err, &usageErr) || usageErr.code != usageDuplicateOption {
		t.Errorf("parseInit() error = %#v, want duplicate option code", err)
	}
}
