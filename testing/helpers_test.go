//go:build testing

package testing

import (
	"context"
	"testing"

	"github.com/zoobz-io/fig"
)

func TestMockProvider_Get(t *testing.T) {
	p := NewMockProvider(t, map[string]string{"key": "value"})

	val, err := p.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("got %q, want %q", val, "value")
	}
}

func TestMockProvider_NotFound(t *testing.T) {
	p := NewMockProvider(t, nil)

	_, err := p.Get(context.Background(), "nonexistent")
	if err != fig.ErrSecretNotFound {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestNewMockProvider_NilSecrets(t *testing.T) {
	p := NewMockProvider(t, nil)
	if p.Secrets == nil {
		t.Error("Secrets should be initialized to empty map")
	}
}
