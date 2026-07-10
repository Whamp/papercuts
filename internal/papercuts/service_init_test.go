package papercuts

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestServiceInitializeLogSerializesConcurrentCreation(t *testing.T) {
	t.Parallel()

	const workers = 16
	directory := t.TempDir()
	service := serviceForDirectory(directory, time.Now)
	results := make([]InitializeResult, workers)
	errorsByWorker := make([]error, workers)
	start := make(chan struct{})
	var group sync.WaitGroup
	for index := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			results[index], errorsByWorker[index] = service.InitializeLog(t.Context(), InitializeRequest{})
		}()
	}
	close(start)
	group.Wait()

	created := 0
	existing := 0
	for index, err := range errorsByWorker {
		if err != nil {
			t.Errorf("InitializeLog(worker %d) error = %v, want nil", index, err)
			continue
		}
		switch results[index].State {
		case InitializeCreated:
			created++
		case InitializeAlreadyExists:
			existing++
		case InitializeNotPerformed:
			t.Errorf("InitializeLog(worker %d) state = %v, want created or already exists", index, results[index].State)
		default:
			t.Errorf("InitializeLog(worker %d) state = %v, want a known completed state", index, results[index].State)
		}
	}
	if created != 1 || existing != workers-1 {
		t.Errorf("InitializeLog() outcomes = created %d, existing %d; want created 1, existing %d", created, existing, workers-1)
	}
	path := filepath.Join(directory, "PAPERCUTS.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}
	if string(content) != logHeader {
		t.Errorf("InitializeLog() bytes = %q, want %q", content, logHeader)
	}
	artifacts, err := filepath.Glob(filepath.Join(directory, ".papercuts-init-*"))
	if err != nil {
		t.Fatalf("filepath.Glob(init artifacts) returned error: %v", err)
	}
	if len(artifacts) != 0 {
		t.Errorf("InitializeLog() temporary artifacts = %v, want none", artifacts)
	}
}

func TestServiceInitializeLogCreatesCompleteProjectLog(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	service := newService(systemSources{
		getwd:       func() (string, error) { return directory, nil },
		lookupEnv:   func(string) (string, bool) { return "", false },
		userHomeDir: func() (string, error) { return "", nil },
	}, time.Now)

	got, err := service.InitializeLog(t.Context(), InitializeRequest{})
	if err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}
	wantPath := filepath.Join(directory, "PAPERCUTS.md")
	if got.State != InitializeCreated || got.Effect != EffectDurable || got.Path != wantPath || got.Scope != ProjectScope {
		t.Errorf("InitializeLog() = %#v, want durable created project result", got)
	}
	content, readErr := os.ReadFile(wantPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(log) returned error: %v", readErr)
	}
	if string(content) != "# Papercuts\n" {
		t.Errorf("log contents = %q, want %q", content, "# Papercuts\\n")
	}
	entries, readDirErr := os.ReadDir(directory)
	if readDirErr != nil {
		t.Fatalf("os.ReadDir() returned error: %v", readDirErr)
	}
	if len(entries) != 1 || entries[0].Name() != "PAPERCUTS.md" {
		t.Errorf("project directory entries = %v, want only PAPERCUTS.md", entries)
	}
}
