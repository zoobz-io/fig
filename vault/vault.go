// Package vault provides a fig.SecretProvider backed by HashiCorp Vault.
package vault

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/zoobzio/fig"
)

// Provider retrieves secrets from HashiCorp Vault.
type Provider struct {
	client *http.Client
	addr   string
	token  string
	mount  string
}

// Option configures the Vault provider.
type Option func(*Provider)

// WithMount sets the KV secrets engine mount path. Defaults to "secret".
func WithMount(mount string) Option {
	return func(p *Provider) {
		p.mount = mount
	}
}

// WithHTTPClient sets a custom HTTP client (for testing).
func WithHTTPClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

// WithAddress sets the Vault server address. Defaults to VAULT_ADDR env var.
func WithAddress(addr string) Option {
	return func(p *Provider) {
		p.addr = addr
	}
}

// WithToken sets the Vault token. Defaults to VAULT_TOKEN env var.
func WithToken(token string) Option {
	return func(p *Provider) {
		p.token = token
	}
}

// New creates a Vault provider using environment configuration.
// It reads VAULT_ADDR and VAULT_TOKEN from the environment.
// Optionally reads VAULT_CACERT for custom CA certificates.
func New(opts ...Option) (*Provider, error) {
	p := &Provider{
		addr:  os.Getenv("VAULT_ADDR"),
		token: os.Getenv("VAULT_TOKEN"),
		mount: "secret",
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.addr == "" {
		return nil, fmt.Errorf("vault: VAULT_ADDR not set")
	}

	if p.client == nil {
		client, err := defaultHTTPClient()
		if err != nil {
			return nil, err
		}
		p.client = client
	}

	return p, nil
}

// defaultHTTPClient creates an HTTP client, optionally with custom CA from VAULT_CACERT.
func defaultHTTPClient() (*http.Client, error) {
	caCert := os.Getenv("VAULT_CACERT")
	if caCert == "" {
		return http.DefaultClient, nil
	}

	caCertPEM, err := os.ReadFile(caCert) // #nosec G304 -- path from VAULT_CACERT env var
	if err != nil {
		return nil, fmt.Errorf("vault: failed to read CA cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("vault: failed to parse CA cert")
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    certPool,
				MinVersion: tls.VersionTLS12,
			},
		},
	}, nil
}

// Get retrieves a secret from Vault.
// The key format is "path/to/secret:field" or just "path/to/secret" (uses "value" field).
func (p *Provider) Get(ctx context.Context, key string) (string, error) {
	path, field := parseKey(key)

	// KV v2 API: GET /v1/{mount}/data/{path}
	url := fmt.Sprintf("%s/v1/%s/data/%s", p.addr, p.mount, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fig.ErrSecretNotFound
	}

	if p.token != "" {
		req.Header.Set("X-Vault-Token", p.token)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fig.ErrSecretNotFound
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fig.ErrSecretNotFound
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fig.ErrSecretNotFound
	}

	var result kvResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fig.ErrSecretNotFound
	}

	if result.Data == nil || result.Data.Data == nil {
		return "", fig.ErrSecretNotFound
	}

	val, ok := result.Data.Data[field]
	if !ok {
		return "", fig.ErrSecretNotFound
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("vault: secret %q field %q is not a string", path, field)
	}

	return str, nil
}

// kvResponse is the Vault KV v2 API response structure.
type kvResponse struct {
	Data *kvData `json:"data"`
}

type kvData struct {
	Data map[string]any `json:"data"`
}

// parseKey splits "path/to/secret:field" into path and field.
// If no field is specified, defaults to "value".
func parseKey(key string) (path, field string) {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return key[:i], key[i+1:]
		}
	}
	return key, "value"
}
