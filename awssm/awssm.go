// Package awssm provides a fig.SecretProvider backed by AWS Secrets Manager.
package awssm

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/zoobz-io/fig"
)

// smClient defines the Secrets Manager operations used by Provider.
type smClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// Provider retrieves secrets from AWS Secrets Manager.
type Provider struct {
	client smClient
}

// Option configures the AWS Secrets Manager provider.
type Option func(*Provider)

// New creates an AWS Secrets Manager provider using default configuration.
// It uses the standard AWS credential chain (env vars, shared config, IAM role).
func New(ctx context.Context, opts ...Option) (*Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	p := &Provider{
		client: secretsmanager.NewFromConfig(cfg),
	}
	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

// NewWithClient creates a provider with a pre-configured client.
func NewWithClient(client *secretsmanager.Client, opts ...Option) *Provider {
	p := &Provider{client: client}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Get retrieves a secret from AWS Secrets Manager.
// The key format is "secret-name" or "secret-name:json-field" for JSON secrets.
func (p *Provider) Get(ctx context.Context, key string) (string, error) {
	name, field := parseKey(key)

	out, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &name,
	})
	if err != nil {
		return "", fig.ErrSecretNotFound
	}

	if out.SecretString == nil {
		return "", fig.ErrSecretNotFound
	}

	value := *out.SecretString

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

// parseKey splits "secret-name:field" into name and field.
func parseKey(key string) (name, field string) {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}
