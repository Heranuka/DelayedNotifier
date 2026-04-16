package notifications

import "errors"

var (
	// ErrNotFound is returned when a comment does not exist.
	ErrNotFound = errors.New("not found")

	// ErrBadRequest is returned when the request body cannot be parsed.
	ErrBadRequest = errors.New("invalid request body")

	ErrMessageRequired = errors.New("message is required")

	ErrDataSent = errors.New("data sent must be in the future")
)
