//go:build testing

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/zoobz-io/fig"
	"github.com/zoobz-io/fig/vault"
)

func TestVault_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		t.Skip("VAULT_ADDR not set, skipping Vault integration test")
	}

	secretPath := os.Getenv("VAULT_TEST_SECRET")
	if secretPath == "" {
		t.Skip("VAULT_TEST_SECRET not set, skipping Vault integration test")
	}

	ctx := context.Background()
	p, err := vault.New()
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	t.Run("existing secret default field", func(t *testing.T) {
		val, err := p.Get(ctx, secretPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == "" {
			t.Error("expected non-empty secret value")
		}
		t.Logf("Successfully retrieved secret (length: %d)", len(val))
	})

	t.Run("non-existent secret", func(t *testing.T) {
		_, err := p.Get(ctx, "nonexistent/secret/path/12345")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestVault_GetField(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		t.Skip("VAULT_ADDR not set, skipping Vault integration test")
	}

	secretPath := os.Getenv("VAULT_TEST_SECRET")
	if secretPath == "" {
		t.Skip("VAULT_TEST_SECRET not set, skipping Vault integration test")
	}

	fieldName := os.Getenv("VAULT_TEST_FIELD")
	if fieldName == "" {
		fieldName = "password"
	}

	ctx := context.Background()
	p, err := vault.New()
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	t.Run("existing secret named field", func(t *testing.T) {
		val, err := p.Get(ctx, secretPath+":"+fieldName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == "" {
			t.Error("expected non-empty field value")
		}
		t.Logf("Successfully retrieved field (length: %d)", len(val))
	})

	t.Run("non-existent field", func(t *testing.T) {
		_, err := p.Get(ctx, secretPath+":nonexistent_field_12345")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestVault_WithMount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		t.Skip("VAULT_ADDR not set, skipping Vault integration test")
	}

	mount := os.Getenv("VAULT_TEST_MOUNT")
	if mount == "" {
		t.Skip("VAULT_TEST_MOUNT not set, skipping mount test")
	}

	secretPath := os.Getenv("VAULT_TEST_SECRET")
	if secretPath == "" {
		t.Skip("VAULT_TEST_SECRET not set, skipping Vault integration test")
	}

	ctx := context.Background()
	p, err := vault.New(vault.WithMount(mount))
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	t.Run("with explicit mount", func(t *testing.T) {
		val, err := p.Get(ctx, secretPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == "" {
			t.Error("expected non-empty secret value")
		}
		t.Logf("Successfully retrieved secret from mount %q (length: %d)", mount, len(val))
	})
}
