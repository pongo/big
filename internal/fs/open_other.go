//go:build !windows

package fs

import (
	"errors"
)

var errUnsupportedPathAction = errors.New("opening paths is only supported on Windows")

func OpenPath(path string) error {
	return errUnsupportedPathAction
}

func RevealPath(path string) error {
	return errUnsupportedPathAction
}
