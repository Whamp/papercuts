package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCommittedIdentifiesPostPublicationFailure(t *testing.T) {
	t.Parallel()

	underlying := errors.New("parent sync failed")
	err := committedError{err: underlying}
	if !Committed(err) {
		t.Error("Committed() = false, want true")
	}
	if !errors.Is(err, underlying) {
		t.Errorf("errors.Is(committed error, underlying) = false")
	}
	if Committed(underlying) {
		t.Error("Committed(uncommitted error) = true, want false")
	}
}

func TestReplaceAtomicallyPublishesCompleteTemporaryFile(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	target := filepath.Join(directory, "AGENTS.md")
	temporary := filepath.Join(directory, ".temporary")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(target) returned error: %v", err)
	}
	if err := os.WriteFile(temporary, []byte("complete replacement"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(temporary) returned error: %v", err)
	}
	if err := Replace(temporary, target); err != nil {
		t.Fatalf("Replace() returned error: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("os.ReadFile(target) returned error: %v", err)
	}
	if string(content) != "complete replacement" {
		t.Errorf("target content = %q", content)
	}
	if _, err := os.Lstat(temporary); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("os.Lstat(temporary) error = %v, want not exist", err)
	}
}

func TestPublishNewDoesNotReplaceExistingFile(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	target := filepath.Join(directory, "PAPERCUTS.md")
	temporary := filepath.Join(directory, ".papercuts-temp")
	if err := os.WriteFile(target, []byte("existing"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(target) returned error: %v", err)
	}
	if err := os.WriteFile(temporary, []byte("new"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(temporary) returned error: %v", err)
	}

	err := PublishNew(temporary, target)
	if !errors.Is(err, os.ErrExist) {
		t.Errorf("PublishNew() error = %v, want errors.Is(os.ErrExist)", err)
	}
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("os.ReadFile(target) returned error: %v", readErr)
	}
	if string(got) != "existing" {
		t.Errorf("target contents = %q, want %q", got, "existing")
	}
}
