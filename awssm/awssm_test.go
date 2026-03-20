package awssm

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/zoobz-io/fig"
)

// mockSMClient implements smClient for testing.
type mockSMClient struct {
	secrets map[string]string
	err     error
}

func (m *mockSMClient) GetSecretValue(_ context.Context, params *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if params.SecretId == nil {
		return nil, errors.New("secret id required")
	}
	val, ok := m.secrets[*params.SecretId]
	if !ok {
		return nil, errors.New("secret not found")
	}
	return &secretsmanager.GetSecretValueOutput{
		SecretString: &val,
	}, nil
}

func TestProvider_Get(t *testing.T) {
	t.Run("existing secret", func(t *testing.T) {
		p := &Provider{
			client: &mockSMClient{
				secrets: map[string]string{"my-secret": "secret-value"},
			},
		}

		val, err := p.Get(context.Background(), "my-secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret-value" {
			t.Errorf("got %q, want %q", val, "secret-value")
		}
	})

	t.Run("non-existent secret", func(t *testing.T) {
		p := &Provider{
			client: &mockSMClient{
				secrets: map[string]string{},
			},
		}

		_, err := p.Get(context.Background(), "missing")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("client error", func(t *testing.T) {
		p := &Provider{
			client: &mockSMClient{
				err: errors.New("connection failed"),
			},
		}

		_, err := p.Get(context.Background(), "any")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestProvider_Get_JSON(t *testing.T) {
	t.Run("extract field from JSON", func(t *testing.T) {
		p := &Provider{
			client: &mockSMClient{
				secrets: map[string]string{
					"db-creds": `{"username":"admin","password":"secret123"}`,
				},
			},
		}

		val, err := p.Get(context.Background(), "db-creds:password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret123" {
			t.Errorf("got %q, want %q", val, "secret123")
		}
	})

	t.Run("missing JSON field", func(t *testing.T) {
		p := &Provider{
			client: &mockSMClient{
				secrets: map[string]string{
					"db-creds": `{"username":"admin"}`,
				},
			},
		}

		_, err := p.Get(context.Background(), "db-creds:password")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		p := &Provider{
			client: &mockSMClient{
				secrets: map[string]string{
					"bad-json": `not valid json`,
				},
			},
		}

		_, err := p.Get(context.Background(), "bad-json:field")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("non-string JSON field", func(t *testing.T) {
		p := &Provider{
			client: &mockSMClient{
				secrets: map[string]string{
					"typed": `{"count":42}`,
				},
			},
		}

		_, err := p.Get(context.Background(), "typed:count")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestProvider_Get_NilSecretString(t *testing.T) {
	// Create a mock that returns nil SecretString
	mock := &mockSMClient{
		secrets: map[string]string{"exists": "value"},
	}
	p := &Provider{client: mock}

	// Override to return nil SecretString
	p.client = &nilSecretStringClient{}

	_, err := p.Get(context.Background(), "any")
	if err != fig.ErrSecretNotFound {
		t.Errorf("expected ErrSecretNotFound for nil SecretString, got %v", err)
	}
}

type nilSecretStringClient struct{}

func (n *nilSecretStringClient) GetSecretValue(_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return &secretsmanager.GetSecretValueOutput{SecretString: nil}, nil
}

func TestNewWithClient(t *testing.T) {
	// Verify NewWithClient creates a provider
	p := NewWithClient(nil)
	if p == nil {
		t.Fatal("expected non-nil provider")
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
