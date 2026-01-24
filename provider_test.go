package fig

import (
	"context"
	"testing"
)

// mockProvider is a test implementation of SecretProvider.
type mockProvider struct {
	secrets map[string]string
}

func (m *mockProvider) Get(_ context.Context, key string) (string, error) {
	if v, ok := m.secrets[key]; ok {
		return v, nil
	}
	return "", ErrSecretNotFound
}

func TestSecretProvider_Interface(_ *testing.T) {
	// Verify mockProvider satisfies SecretProvider
	var _ SecretProvider = (*mockProvider)(nil)
}
