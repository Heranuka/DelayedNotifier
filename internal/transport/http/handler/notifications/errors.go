package notifications

import "errors"

var (
	// ErrNotFound is returned when a comment does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInternalServerError is returned when the server fails unexpectedly.
	ErrInternalServerError = errors.New("internal server error")

	// ErrBadRequest is returned when the request body cannot be parsed.
	ErrBadRequest = errors.New("invalid request body")
)
