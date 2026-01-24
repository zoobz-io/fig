package fig

import (
	"errors"
	"testing"
)

func TestFieldError_Error(t *testing.T) {
	err := &FieldError{Field: "Host", Err: ErrRequired}
	got := err.Error()
	want := "fig: field Host: fig: required field not set"
	if got != want {
		t.Errorf("FieldError.Error() = %q, want %q", got, want)
	}
}

func TestFieldError_Unwrap(t *testing.T) {
	err := &FieldError{Field: "Host", Err: ErrRequired}
	if !errors.Is(err, ErrRequired) {
		t.Error("FieldError should unwrap to underlying error")
	}
}
