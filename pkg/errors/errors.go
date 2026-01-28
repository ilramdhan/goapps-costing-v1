package errors

import (
	"errors"
	"fmt"
)

// Standard error types
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternal      = errors.New("internal server error")
	ErrUnavailable   = errors.New("service unavailable")
	ErrTimeout       = errors.New("request timeout")
	ErrRateLimited   = errors.New("rate limit exceeded")
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

// NewValidationErrors creates a new ValidationErrors
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}

// Add adds a validation error
func (v *ValidationErrors) Add(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "no validation errors"
	}
	return fmt.Sprintf("validation failed: %d errors", len(v.Errors))
}

// AppError represents an application error with context
type AppError struct {
	Code       string
	Message    string
	Err        error
	Validation *ValidationErrors
}

// NewAppError creates a new application error
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a validation error
func NewValidationError(validation *ValidationErrors) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		Validation: validation,
	}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches target
func (e *AppError) Is(target error) bool {
	if e.Err == nil {
		return false
	}
	return errors.Is(e.Err, target)
}

// IsNotFound checks if error is not found
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if error is already exists
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsValidation checks if error is validation error
func IsValidation(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Validation != nil
	}
	return false
}

// Wrap wraps an error with context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapWithCode wraps an error with code and message
func WrapWithCode(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
