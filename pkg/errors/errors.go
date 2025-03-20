package errors

import (
	"fmt"
)

type ErrorType string

const (
	ValidationError    ErrorType = "VALIDATION_ERROR"
	NotFoundError      ErrorType = "NOT_FOUND"
	AuthorizationError ErrorType = "AUTHORIZATION_ERROR"
	DatabaseError      ErrorType = "DATABASE_ERROR"
	NetworkError       ErrorType = "NETWORK_ERROR"
	ConflictError      ErrorType = "CONFLICT_ERROR"
	InternalError      ErrorType = "INTERNAL_ERROR"
)

// AppError represents a custom application error
type AppError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Err     error                  `json:"-"` // Internal error, not exposed in JSON
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Helper functions to create specific error types
func NewValidationError(msg string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    ValidationError,
		Message: msg,
		Details: details,
	}
}

func NewNotFoundError(msg string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    NotFoundError,
		Message: msg,
		Details: details,
	}
}

func NewAuthorizationError(msg string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    AuthorizationError,
		Message: msg,
		Details: details,
	}
}

func NewDatabaseError(msg string, err error, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    DatabaseError,
		Message: msg,
		Details: details,
		Err:     err,
	}
}

func NewNetworkError(msg string, err error, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    NetworkError,
		Message: msg,
		Details: details,
		Err:     err,
	}
}

func NewConflictError(msg string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    ConflictError,
		Message: msg,
		Details: details,
	}
}

func NewInternalError(msg string, err error, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    InternalError,
		Message: msg,
		Details: details,
		Err:     err,
	}
}

func IsType(err error, target ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == target
	}
	return false
}
