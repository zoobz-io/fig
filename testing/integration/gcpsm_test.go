//go:build testing

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/zoobzio/fig"
	"github.com/zoobzio/fig/gcpsm"
)

const (
	gcpTestProject = "test-project"
)

func TestGCPSecretManager_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	project := os.Getenv("GCP_PROJECT")
	if project == "" {
		project = gcpTestProject
	}

	secretName := os.Getenv("GCP_TEST_SECRET")
	if secretName == "" {
		t.Skip("GCP_TEST_SECRET not set, skipping GCP integration test")
	}

	ctx := context.Background()
	p, err := gcpsm.New(ctx, project)
	if err != nil {
		t.Skipf("GCP credentials unavailable: %v", err)
	}
	defer p.Close()

	t.Run("existing secret", func(t *testing.T) {
		val, err := p.Get(ctx, secretName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == "" {
			t.Error("expected non-empty secret value")
		}
		t.Logf("Successfully retrieved secret (length: %d)", len(val))
	})

	t.Run("non-existent secret", func(t *testing.T) {
		_, err := p.Get(ctx, "nonexistent-secret-that-should-not-exist-12345")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestGCPSecretManager_GetJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	project := os.Getenv("GCP_PROJECT")
	if project == "" {
		project = gcpTestProject
	}

	secretName := os.Getenv("GCP_TEST_JSON_SECRET")
	if secretName == "" {
		t.Skip("GCP_TEST_JSON_SECRET not set, skipping GCP JSON integration test")
	}

	fieldName := os.Getenv("GCP_TEST_JSON_FIELD")
	if fieldName == "" {
		fieldName = "password"
	}

	ctx := context.Background()
	p, err := gcpsm.New(ctx, project)
	if err != nil {
		t.Skipf("GCP credentials unavailable: %v", err)
	}
	defer p.Close()

	t.Run("existing JSON field", func(t *testing.T) {
		val, err := p.Get(ctx, secretName+":"+fieldName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == "" {
			t.Error("expected non-empty field value")
		}
		t.Logf("Successfully retrieved JSON field (length: %d)", len(val))
	})

	t.Run("non-existent JSON field", func(t *testing.T) {
		_, err := p.Get(ctx, secretName+":nonexistent_field_12345")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}
