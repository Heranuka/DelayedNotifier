package comments

import (
	"errors"
)

var (
	// ErrNotFound is returned when a comment does not exist.
	ErrNotFound = errors.New("not found")
)
