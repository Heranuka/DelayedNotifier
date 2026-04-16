// Package helpers provides utility functions for decoding JSON requests
// and parsing HTTP path parameters.
package helpers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
)

// ErrMissingID indicates that a required URL parameter is missing.
var ErrMissingID = errors.New("id is missing")

// ErrInvalidID indicates that a URL parameter has an invalid format.
var ErrInvalidID = errors.New("id has invalid format")

// DecodeJSON decodes JSON from the request body into dst.
// It disallows unknown fields to catch client errors early.
func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// ParseUUIDParam extracts and parses a UUID from a URL parameter.
// Returns ErrInvalidID if the parameter is missing or malformed.
func ParseUUIDParam(c *ginext.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	if idStr == "" {
		return uuid.Nil, ErrMissingID
	}

	return uuid.Parse(idStr)
}
