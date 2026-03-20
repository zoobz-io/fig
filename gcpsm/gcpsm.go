// Package gcpsm provides a fig.SecretProvider backed by Google Cloud Secret Manager.
package gcpsm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zoobz-io/fig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	baseURL = "https://secretmanager.googleapis.com/v1"
	scope   = "https://www.googleapis.com/auth/cloud-platform"
)

// Provider retrieves secrets from Google Cloud Secret Manager.
type Provider struct {
	client  *http.Client
	project string
}

// Option configures the GCP Secret Manager provider.
type Option func(*Provider)

// WithHTTPClient sets a custom HTTP client (for testing).
func WithHTTPClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

// New creates a GCP Secret Manager provider using Application Default Credentials.
// The project ID is required.
func New(ctx context.Context, project string, opts ...Option) (*Provider, error) {
	p := &Provider{
		project: project,
	}
	for _, opt := range opts {
		opt(p)
	}

	// Only set up ADC if no custom client provided
	if p.client == nil {
		creds, err := google.FindDefaultCredentials(ctx, scope)
		if err != nil {
			return nil, fmt.Errorf("gcpsm: %w", err)
		}
		p.client = oauth2.NewClient(ctx, creds.TokenSource)
	}

	return p, nil
}

// Get retrieves a secret from GCP Secret Manager.
// The key format is "secret-name" or "secret-name:json-field" for JSON secrets.
// Version defaults to "latest".
func (p *Provider) Get(ctx context.Context, key string) (string, error) {
	name, field := parseKey(key)

	url := fmt.Sprintf("%s/projects/%s/secrets/%s/versions/latest:access", baseURL, p.project, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fig.ErrSecretNotFound
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

	var result accessResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fig.ErrSecretNotFound
	}

	// Payload data is base64 encoded
	decoded, err := base64.StdEncoding.DecodeString(result.Payload.Data)
	if err != nil {
		return "", fig.ErrSecretNotFound
	}

	value := string(decoded)

	// If a field is specified, parse as JSON and extract
	if field != "" {
		var data map[string]any
		if err := json.Unmarshal([]byte(value), &data); err != nil {
			return "", fig.ErrSecretNotFound
		}
		v, ok := data[field]
		if !ok {
			return "", fig.ErrSecretNotFound
		}
		str, ok := v.(string)
		if !ok {
			return "", fig.ErrSecretNotFound
		}
		return str, nil
	}

	return value, nil
}

// Close is a no-op for the REST client (kept for interface compatibility).
func (p *Provider) Close() error {
	return nil
}

// accessResponse is the REST API response for accessing a secret version.
type accessResponse struct {
	Name    string `json:"name"`
	Payload struct {
		Data string `json:"data"`
	} `json:"payload"`
}

// parseKey splits "secret-name:field" into name and field.
func parseKey(key string) (name, field string) {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}
