package papercuts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCaptureSerializesConcurrentProcesses(t *testing.T) {
	directory := t.TempDir()
	service := serviceForDirectory(directory, time.Now)
	if _, err := service.InitializeLog(t.Context(), InitializeRequest{}); err != nil {
		t.Fatalf("InitializeLog() returned error: %v", err)
	}

	const processCount = 16
	errorsByProcess := make([]error, processCount)
	var waitGroup sync.WaitGroup
	for index := 0; index < processCount; index++ {
		index := index
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestCaptureHelperProcess$")
			command.Dir = directory
			command.Env = append(os.Environ(),
				"PAPERCUTS_TEST_CAPTURE_HELPER=1",
				fmt.Sprintf("PAPERCUTS_TEST_DESCRIPTION=process-%02d", index),
			)
			output, err := command.CombinedOutput()
			if err != nil {
				errorsByProcess[index] = fmt.Errorf("helper %d: %w: %s", index, err, output)
			}
		}()
	}
	waitGroup.Wait()
	for _, err := range errorsByProcess {
		if err != nil {
			t.Error(err)
		}
	}
	if t.Failed() {
		return
	}

	content, err := os.ReadFile(filepath.Join(directory, "PAPERCUTS.md"))
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}
	if got := strings.Count(string(content), "\n## "); got != processCount {
		t.Errorf("entry count = %d, want %d\n%s", got, processCount, content)
	}
	for index := 0; index < processCount; index++ {
		marker := fmt.Sprintf("> process-%02d\n", index)
		if got := strings.Count(string(content), marker); got != 1 {
			t.Errorf("marker %q count = %d, want 1", marker, got)
		}
	}
	matches, err := filepath.Glob(filepath.Join(directory, ".papercuts-*"))
	if err != nil {
		t.Fatalf("filepath.Glob() returned error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("temporary or lock artifacts remain: %v", matches)
	}
}

func TestCaptureHelperProcess(t *testing.T) {
	if os.Getenv("PAPERCUTS_TEST_CAPTURE_HELPER") != "1" {
		return
	}
	severity, err := ParseSeverity("low")
	if err != nil {
		t.Fatalf("ParseSeverity() returned error: %v", err)
	}
	_, err = NewService().Capture(t.Context(), CaptureRequest{
		Severity:    severity,
		Description: os.Getenv("PAPERCUTS_TEST_DESCRIPTION"),
	})
	if err != nil {
		t.Fatalf("Capture() returned error: %v", err)
	}
}
