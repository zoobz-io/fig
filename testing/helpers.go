//go:build testing

// Package testing provides test utilities for fig.
package testing

import (
	"context"
	"testing"

	"github.com/zoobzio/fig"
)

// MockProvider is a test implementation of fig.SecretProvider.
type MockProvider struct {
	Secrets map[string]string
}

// Get retrieves a secret from the mock store.
func (m *MockProvider) Get(_ context.Context, key string) (string, error) {
	if v, ok := m.Secrets[key]; ok {
		return v, nil
	}
	return "", fig.ErrSecretNotFound
}

// NewMockProvider creates a MockProvider with the given secrets.
func NewMockProvider(t *testing.T, secrets map[string]string) *MockProvider {
	t.Helper()
	if secrets == nil {
		secrets = make(map[string]string)
	}
	return &MockProvider{Secrets: secrets}
}

// SetEnv sets an environment variable for the duration of the test.
func SetEnv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}
