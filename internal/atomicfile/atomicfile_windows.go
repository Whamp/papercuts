//go:build windows

package atomicfile

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	replaceFileWProc = kernel32.NewProc("ReplaceFileW")
)

// PublishNew gives temporary its final name without replacing an existing target.
func PublishNew(temporary string, target string) error {
	from, err := windows.UTF16PtrFromString(temporary)
	if err != nil {
		return fmt.Errorf("encode temporary path %q: %w", temporary, err)
	}
	to, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return fmt.Errorf("encode target path %q: %w", target, err)
	}
	if err := windows.MoveFileEx(from, to, windows.MOVEFILE_WRITE_THROUGH); err != nil {
		return &os.LinkError{Op: "publish", Old: temporary, New: target, Err: err}
	}
	return nil
}

// Replace atomically replaces target with temporary.
func Replace(temporary string, target string) error {
	replaced, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return fmt.Errorf("encode target path %q: %w", target, err)
	}
	replacement, err := windows.UTF16PtrFromString(temporary)
	if err != nil {
		return fmt.Errorf("encode temporary path %q: %w", temporary, err)
	}

	result, _, callErr := replaceFileWProc.Call(
		uintptr(unsafe.Pointer(replaced)),
		uintptr(unsafe.Pointer(replacement)),
		0,
		0,
		0,
		0,
	)
	if result == 0 {
		return &os.LinkError{Op: "replace", Old: temporary, New: target, Err: callErr}
	}
	return nil
}
