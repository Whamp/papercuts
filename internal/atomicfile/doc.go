// Package atomicfile publishes complete files without exposing partial contents.
package atomicfile

import "errors"

type committedError struct {
	err error
}

func (e committedError) Error() string {
	return e.err.Error()
}

func (e committedError) Unwrap() error {
	return e.err
}

func (e committedError) committed() {}

type commitMarker interface {
	committed()
}

// Committed reports whether an error happened after the target changed.
func Committed(err error) bool {
	var marker commitMarker
	return errors.As(err, &marker)
}

func afterCommit(err error) error {
	if err == nil {
		return nil
	}
	return committedError{err: err}
}
