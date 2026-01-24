package gcpsm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zoobzio/fig"
)

func mockServer(secrets map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract secret name from path: /v1/projects/{project}/secrets/{name}/versions/latest:access
		path := r.URL.Path
		// Find the secret name between /secrets/ and /versions/
		const secretsPrefix = "/secrets/"
		const versionsSuffix = "/versions/latest:access"

		start := len("/v1/projects/test-project") + len(secretsPrefix)
		end := len(path) - len(versionsSuffix)

		if start >= end || start < 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		name := path[start:end]
		val, ok := secrets[name]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		resp := accessResponse{
			Name: path[:end],
			Payload: struct {
				Data string `json:"data"`
			}{
				Data: base64.StdEncoding.EncodeToString([]byte(val)),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func newTestProvider(t *testing.T, server *httptest.Server) *Provider {
	t.Helper()
	return &Provider{
		client:  server.Client(),
		project: "test-project",
	}
}

// Override baseURL for testing
func (p *Provider) withBaseURL(url string) *Provider {
	// We'll use a custom client that rewrites URLs
	originalClient := p.client
	p.client = &http.Client{
		Transport: &rewriteTransport{
			base:    originalClient.Transport,
			fromURL: baseURL,
			toURL:   url,
		},
	}
	return p
}

type rewriteTransport struct {
	base    http.RoundTripper
	fromURL string
	toURL   string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL
	newURL := t.toURL + req.URL.Path
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(newReq)
}

func TestProvider_Get(t *testing.T) {
	server := mockServer(map[string]string{
		"my-secret": "secret-value",
	})
	defer server.Close()

	p := newTestProvider(t, server).withBaseURL(server.URL)

	t.Run("existing secret", func(t *testing.T) {
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

func TestProvider_Get_JSON(t *testing.T) {
	server := mockServer(map[string]string{
		"db-creds": `{"username":"admin","password":"secret123"}`,
		"bad-json": `not valid json`,
		"typed":    `{"count":42}`,
	})
	defer server.Close()

	p := newTestProvider(t, server).withBaseURL(server.URL)

	t.Run("extract field from JSON", func(t *testing.T) {
		val, err := p.Get(context.Background(), "db-creds:password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret123" {
			t.Errorf("got %q, want %q", val, "secret123")
		}
	})

	t.Run("missing JSON field", func(t *testing.T) {
		_, err := p.Get(context.Background(), "db-creds:missing")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := p.Get(context.Background(), "bad-json:field")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("non-string JSON field", func(t *testing.T) {
		_, err := p.Get(context.Background(), "typed:count")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestProvider_Close(t *testing.T) {
	p := &Provider{
		client:  http.DefaultClient,
		project: "test-project",
	}

	if err := p.Close(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{}
	p := &Provider{}
	WithHTTPClient(customClient)(p)

	if p.client != customClient {
		t.Error("WithHTTPClient did not set client")
	}
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		key       string
		wantName  string
		wantField string
	}{
		{"my-secret", "my-secret", ""},
		{"my-secret:password", "my-secret", "password"},
		{"my-secret:db:password", "my-secret:db", "password"},
		{"", "", ""},
		{":", "", ""},
		{":field", "", "field"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			name, field := parseKey(tt.key)
			if name != tt.wantName {
				t.Errorf("parseKey(%q) name = %q, want %q", tt.key, name, tt.wantName)
			}
			if field != tt.wantField {
				t.Errorf("parseKey(%q) field = %q, want %q", tt.key, field, tt.wantField)
			}
		})
	}
}
