package papercuts

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestIntegrateGuidancePreservesPlainFileBytesModeAndNewlines(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "AGENTS.md")
	original := []byte("# Project rules\r\n\r\nKeep these bytes.\r\n")
	if err := os.WriteFile(path, original, 0o640); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	var originalMode os.FileMode
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o640|os.ModeSticky); err != nil {
			t.Fatalf("os.Chmod() returned error: %v", err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("os.Stat(original) returned error: %v", err)
		}
		originalMode = info.Mode()
	}
	service := newService(defaultSystemSources(), time.Now)
	first, err := service.IntegrateGuidance(t.Context(), GuidanceRequest{Path: path})
	if err != nil {
		t.Fatalf("IntegrateGuidance(first) returned error: %v", err)
	}
	if first.State != GuidanceUpdated || first.Effect != EffectDurable {
		t.Errorf("IntegrateGuidance(first) = %#v, want state %v and effect %v", first, GuidanceUpdated, EffectDurable)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	if !bytes.HasPrefix(content, original) {
		t.Errorf("existing prefix changed\ngot: %q\nwant prefix: %q", content, original)
	}
	withoutCRLF := bytes.ReplaceAll(content, []byte("\r\n"), nil)
	if bytes.ContainsAny(withoutCRLF, "\r\n") {
		t.Errorf("IntegrateGuidance(CRLF file) bytes = %q, want CRLF-only line endings", content)
	}
	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("os.Stat() returned error: %v", statErr)
		}
		if got := info.Mode(); got != originalMode {
			t.Errorf("AGENTS.md mode = %v, want %v", got, originalMode)
		}
	}

	second, err := service.IntegrateGuidance(t.Context(), GuidanceRequest{Path: path})
	if err != nil {
		t.Fatalf("IntegrateGuidance(second) returned error: %v", err)
	}
	if second.State != GuidanceUnchanged || second.Effect != EffectUnchanged {
		t.Errorf("IntegrateGuidance(second) = %#v, want state %v and effect %v", second, GuidanceUnchanged, EffectUnchanged)
	}
	secondContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(second) returned error: %v", err)
	}
	if !bytes.Equal(secondContent, content) {
		t.Errorf("IntegrateGuidance(second) bytes = %q, want first result bytes %q", secondContent, content)
	}
}

func TestIntegrateGuidanceRejectsMalformedMarkersWithoutChangingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "AGENTS.md")
	original := []byte("before\n<!-- papercuts:begin -->\nmissing end\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	service := newService(defaultSystemSources(), time.Now)
	result, err := service.IntegrateGuidance(t.Context(), GuidanceRequest{Path: path})
	if !errors.Is(err, ErrMalformedGuidance) {
		t.Errorf("IntegrateGuidance() error = %v, want ErrMalformedGuidance", err)
	}
	if result.Effect != EffectUnchanged {
		t.Errorf("IntegrateGuidance() effect = %v, want unchanged", result.Effect)
	}
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("os.ReadFile() returned error: %v", readErr)
	}
	if !bytes.Equal(content, original) {
		t.Errorf("IntegrateGuidance(malformed markers) bytes = %q, want original bytes %q", content, original)
	}
}
