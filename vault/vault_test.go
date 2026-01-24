package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zoobzio/fig"
)

func mockVaultServer(secrets map[string]map[string]any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract path from: /v1/{mount}/data/{path}
		// We expect mount to be "secret" in tests
		path := r.URL.Path
		const prefix = "/v1/secret/data/"

		if len(path) <= len(prefix) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		secretPath := path[len(prefix):]
		data, ok := secrets[secretPath]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		resp := kvResponse{
			Data: &kvData{
				Data: data,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func newTestProvider(t *testing.T, server *httptest.Server) *Provider {
	t.Helper()
	return &Provider{
		client: server.Client(),
		addr:   server.URL,
		token:  "test-token",
		mount:  "secret",
	}
}

func TestProvider_Get(t *testing.T) {
	server := mockVaultServer(map[string]map[string]any{
		"my-secret": {"value": "secret-value"},
	})
	defer server.Close()

	p := newTestProvider(t, server)

	t.Run("existing secret default field", func(t *testing.T) {
		val, err := p.Get(context.Background(), "my-secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret-value" {
			t.Errorf("got %q, want %q", val, "secret-value")
		}
	})

	t.Run("non-existent secret", func(t *testing.T) {
		_, err := p.Get(context.Background(), "missing")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestProvider_Get_NamedField(t *testing.T) {
	server := mockVaultServer(map[string]map[string]any{
		"db-creds": {"username": "admin", "password": "secret123"},
	})
	defer server.Close()

	p := newTestProvider(t, server)

	t.Run("existing secret named field", func(t *testing.T) {
		val, err := p.Get(context.Background(), "db-creds:password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret123" {
			t.Errorf("got %q, want %q", val, "secret123")
		}
	})

	t.Run("missing field", func(t *testing.T) {
		_, err := p.Get(context.Background(), "db-creds:missing")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("non-string field", func(t *testing.T) {
		server := mockVaultServer(map[string]map[string]any{
			"typed": {"count": 42},
		})
		defer server.Close()

		p := newTestProvider(t, server)
		_, err := p.Get(context.Background(), "typed:count")
		if err == nil {
			t.Fatal("expected error for non-string field")
		}
		if err == fig.ErrSecretNotFound {
			t.Error("expected non-string error, not ErrSecretNotFound")
		}
	})
}

func TestProvider_Get_NilData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return response with nil data
		resp := kvResponse{Data: nil}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := newTestProvider(t, server)
	_, err := p.Get(context.Background(), "any")
	if err != fig.ErrSecretNotFound {
		t.Errorf("expected ErrSecretNotFound for nil data, got %v", err)
	}
}

func TestNew(t *testing.T) {
	t.Run("missing VAULT_ADDR", func(t *testing.T) {
		t.Setenv("VAULT_ADDR", "")
		_, err := New()
		if err == nil {
			t.Fatal("expected error for missing VAULT_ADDR")
		}
	})

	t.Run("with VAULT_ADDR", func(t *testing.T) {
		t.Setenv("VAULT_ADDR", "http://localhost:8200")
		t.Setenv("VAULT_TOKEN", "test-token")
		p, err := New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.addr != "http://localhost:8200" {
			t.Errorf("addr = %q, want %q", p.addr, "http://localhost:8200")
		}
		if p.token != "test-token" {
			t.Errorf("token = %q, want %q", p.token, "test-token")
		}
		if p.mount != "secret" {
			t.Errorf("mount = %q, want %q", p.mount, "secret")
		}
	})
}

func TestWithOptions(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://localhost:8200")

	t.Run("WithMount", func(t *testing.T) {
		p, err := New(WithMount("custom"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.mount != "custom" {
			t.Errorf("mount = %q, want %q", p.mount, "custom")
		}
	})

	t.Run("WithAddress", func(t *testing.T) {
		p, err := New(WithAddress("http://other:8200"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.addr != "http://other:8200" {
			t.Errorf("addr = %q, want %q", p.addr, "http://other:8200")
		}
	})

	t.Run("WithToken", func(t *testing.T) {
		p, err := New(WithToken("my-token"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.token != "my-token" {
			t.Errorf("token = %q, want %q", p.token, "my-token")
		}
	})

	t.Run("WithHTTPClient", func(t *testing.T) {
		client := &http.Client{}
		p, err := New(WithHTTPClient(client))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.client != client {
			t.Error("client not set")
		}
	})
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		key       string
		wantPath  string
		wantField string
	}{
		{"my-secret", "my-secret", "value"},
		{"my-secret:password", "my-secret", "password"},
		{"path/to/secret", "path/to/secret", "value"},
		{"path/to/secret:api_key", "path/to/secret", "api_key"},
		{"", "", "value"},
		{":", "", ""},
		{":field", "", "field"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			path, field := parseKey(tt.key)
			if path != tt.wantPath {
				t.Errorf("parseKey(%q) path = %q, want %q", tt.key, path, tt.wantPath)
			}
			if field != tt.wantField {
				t.Errorf("parseKey(%q) field = %q, want %q", tt.key, field, tt.wantField)
			}
		})
	}
}
