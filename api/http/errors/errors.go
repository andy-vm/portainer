package errors

import "errors"

var (
	// ErrEndpointAccessDenied Access denied to environment(endpoint) error
	ErrEndpointAccessDenied = errors.New("Access denied to environment")
	// ErrUnauthorized Unauthorized error
	ErrUnauthorized = errors.New("Unauthorized")
	// ErrResourceAccessDenied Access denied to resource error
	ErrResourceAccessDenied = errors.New("Access denied to resource")
	// ErrNotAvailableInDemo feature is not allowed in demo
	ErrNotAvailableInDemo = errors.New("This feature is not available in the demo version of Portainer")
)

type ConflictError struct {
	msg string
}

func (e *ConflictError) Error() string {
	return e.msg
}

func NewConflictError(msg string) *ConflictError {
	return &ConflictError{msg: msg}
}
