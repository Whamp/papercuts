//go:build linux || darwin

package atomicfile

import (
	"errors"
	"fmt"
	"os"
)

// PublishNew gives temporary its final name without replacing an existing target.
func PublishNew(temporary string, target string) error {
	if err := os.Link(temporary, target); err != nil {
		return fmt.Errorf("publish %q: %w", target, err)
	}
	removeErr := os.Remove(temporary)
	syncErr := syncParent(target)
	if err := errors.Join(removeErr, syncErr); err != nil {
		return afterCommit(fmt.Errorf("finish publishing %q: %w", target, err))
	}
	return nil
}

// Replace atomically replaces target with temporary.
func Replace(temporary string, target string) error {
	if err := os.Rename(temporary, target); err != nil {
		return fmt.Errorf("replace %q: %w", target, err)
	}
	if err := syncParent(target); err != nil {
		return afterCommit(fmt.Errorf("sync parent of %q: %w", target, err))
	}
	return nil
}
