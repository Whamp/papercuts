package papercuts

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestIntegrateGuidanceRetriesAfterConcurrentReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	original := append([]byte("# Existing rules\n\n"), bytes.Repeat([]byte("keep this line\n"), 20_000)...)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}

	const workerCount = 12
	start := make(chan struct{})
	results := make([]GuidanceResult, workerCount)
	errorsByWorker := make([]error, workerCount)
	var waitGroup sync.WaitGroup
	for index := 0; index < workerCount; index++ {
		index := index
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			service := newService(defaultSystemSources(), time.Now)
			results[index], errorsByWorker[index] = service.IntegrateGuidance(t.Context(), GuidanceRequest{Path: path})
		}()
	}
	close(start)
	waitGroup.Wait()

	updated := 0
	for index, err := range errorsByWorker {
		if err != nil {
			t.Errorf("IntegrateGuidance(worker %d) error = %v, want nil", index, err)
			continue
		}
		if results[index].State == GuidanceUpdated {
			updated++
		}
	}
	if updated != 1 {
		t.Errorf("updated result count = %d, want 1", updated)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	if got := bytes.Count(content, []byte(guidanceBegin)); got != 1 {
		t.Errorf("managed section count = %d, want 1", got)
	}
}
