package filelock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenTimesOutThroughHardLinkAlias(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "PAPERCUTS.md")
	alias := filepath.Join(directory, "alias.md")
	if err := os.WriteFile(path, []byte("# Papercuts\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	if err := os.Link(path, alias); err != nil {
		t.Skipf("os.Link() unavailable: %v", err)
	}
	first, err := Open(t.Context(), path, time.Millisecond)
	if err != nil {
		t.Fatalf("Open(first) returned error: %v", err)
	}
	defer func() {
		if err := first.Unlock(); err != nil {
			t.Errorf("first.Unlock() returned error: %v", err)
		}
		if err := first.Close(); err != nil {
			t.Errorf("first.Close() returned error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Millisecond)
	defer cancel()
	_, err = Open(ctx, alias, time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Open(alias) error = %v, want context deadline exceeded", err)
	}
}

func TestOpenStopsWhenCallerCancels(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "PAPERCUTS.md")
	if err := os.WriteFile(path, []byte("# Papercuts\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	first, err := Open(t.Context(), path, time.Millisecond)
	if err != nil {
		t.Fatalf("Open(first) returned error: %v", err)
	}
	defer func() {
		if err := first.Unlock(); err != nil {
			t.Errorf("first.Unlock() returned error: %v", err)
		}
		if err := first.Close(); err != nil {
			t.Errorf("first.Close() returned error: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(t.Context())
	result := make(chan error, 1)
	go func() {
		_, openErr := Open(ctx, path, time.Millisecond)
		result <- openErr
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Errorf("Open(canceled) error = %v, want context.Canceled", err)
	}
}

func TestOpenTimesOutWhenTargetIsLocked(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "PAPERCUTS.md")
	if err := os.WriteFile(path, []byte("# Papercuts\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	first, err := Open(context.Background(), path, time.Millisecond)
	if err != nil {
		t.Fatalf("Open(first) returned error: %v", err)
	}
	defer func() {
		if err := first.Unlock(); err != nil {
			t.Errorf("first.Unlock() returned error: %v", err)
		}
		if err := first.Close(); err != nil {
			t.Errorf("first.Close() returned error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	_, err = Open(ctx, path, time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Open(second) error = %v, want context deadline exceeded", err)
	}
}
