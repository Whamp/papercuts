package cli

import (
	"strings"
	"testing"
)

func FuzzParseCapturePreservesCommandInvariants(f *testing.F) {
	f.Add("--severity\x00low\x00friction")
	f.Add("--global\x00--severity=high\x00--stdin")
	f.Add("--severity\x00medium\x00--\x00--option-shaped-description")
	f.Fuzz(func(t *testing.T, encoded string) {
		args := strings.Split(encoded, "\x00")
		command, err := parseCapture(args)
		if err != nil || command.help {
			return
		}
		if command.severity.String() == "" {
			t.Errorf("parseCapture(%q) severity = empty, want a valid severity after success", args)
		}
		if command.target.Project && command.target.Global {
			t.Errorf("parseCapture(%q) target = %#v, want mutually exclusive scopes", args, command.target)
		}
		if command.target.GlobalPath != nil && !command.target.Global {
			t.Errorf("parseCapture(%q) target = %#v, want global scope with global path", args, command.target)
		}
		if command.stdin && command.description != "" {
			t.Errorf("parseCapture(%q) command = %#v, want empty argument description with stdin", args, command)
		}
	})
}
