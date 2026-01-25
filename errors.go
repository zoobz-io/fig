package fig

import "errors"

// Sentinel errors for configuration loading.
var (
	// ErrRequired indicates a required field was not set.
	ErrRequired = errors.New("fig: required field not set")

	// ErrInvalidType indicates a value could not be converted to the target type.
	ErrInvalidType = errors.New("fig: invalid type conversion")

	// ErrNotStruct indicates Load was called with a non-struct type.
	ErrNotStruct = errors.New("fig: type must be a struct")

	// ErrSecretNotFound indicates a secret was not found in the provider.
	ErrSecretNotFound = errors.New("fig: secret not found")
)

// FieldError wraps an error with field context.
type FieldError struct {
	Err   error
	Field string
}

func (e *FieldError) Error() string {
	return "fig: field " + e.Field + ": " + e.Err.Error()
}

func (e *FieldError) Unwrap() error {
	return e.Err
}
