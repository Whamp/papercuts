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
			t.Error("parseCapture() accepted an invalid severity")
		}
		if command.target.Project && command.target.Global {
			t.Error("parseCapture() accepted conflicting scopes")
		}
		if command.target.GlobalPath != nil && !command.target.Global {
			t.Error("parseCapture() accepted global path without global scope")
		}
		if command.stdin && command.description != "" {
			t.Errorf("parseCapture() accepted stdin with an argument description: %#v", command)
		}
	})
}
