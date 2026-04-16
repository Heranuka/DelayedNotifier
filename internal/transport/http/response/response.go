// Package response provides helpers for writing JSON HTTP responses.
package response

import (
	"encoding/json"
	"errors"
	"net/http"
)

// MessageResult represents a success response with a message.
type MessageResult struct {
	Message string `json:"message"`
}

// Error represents the JSON envelope for error responses.
type Error struct {
	Message string `json:"error"`
}

// ErrorWithDetails represents an error response with optional details.
type ErrorWithDetails struct {
	Message string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// OK sends a 200 OK response with the provided result.
func OK(w http.ResponseWriter, result any) error {
	return writeJSON(w, http.StatusOK, result)
}

// Created sends a 201 Created response with the provided result.
func Created(w http.ResponseWriter, result any) error {
	return writeJSON(w, http.StatusCreated, result)
}

// Fail sends an error response with the provided status code and error message.
func Fail(w http.ResponseWriter, status int, err error) error {
	if err == nil {
		err = errors.New("unknown error")
	}
	return writeJSON(w, status, Error{Message: err.Error()})
}

// FailWithDetails sends an error response with optional details.
func FailWithDetails(w http.ResponseWriter, status int, message string, details any) error {
	if message == "" {
		message = "unknown error"
	}

	return writeJSON(w, status, ErrorWithDetails{
		Message: message,
		Details: details,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_, err = w.Write(append(js, '\n'))
	if err != nil {
		return err
	}

	return nil
}
