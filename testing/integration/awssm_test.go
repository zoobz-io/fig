//go:build testing

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/zoobz-io/fig"
	"github.com/zoobz-io/fig/awssm"
)

const (
	localstackEndpoint = "http://localhost:4566"
	testSecretName     = "fig-integration-test"
	testSecretValue    = "integration-test-secret-value"
	testJSONSecretName = "fig-integration-test-json"
)

func localstackClient(t *testing.T) *secretsmanager.Client {
	t.Helper()

	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	if endpoint == "" {
		endpoint = localstackEndpoint
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = &endpoint
	})
}

func setupTestSecret(t *testing.T, client *secretsmanager.Client) {
	t.Helper()
	ctx := context.Background()

	// Clean up any existing secret
	_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   ptr(testSecretName),
		ForceDeleteWithoutRecovery: ptr(true),
	})

	// Create test secret
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         ptr(testSecretName),
		SecretString: ptr(testSecretValue),
	})
	if err != nil {
		t.Fatalf("failed to create test secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   ptr(testSecretName),
			ForceDeleteWithoutRecovery: ptr(true),
		})
	})
}

func setupJSONTestSecret(t *testing.T, client *secretsmanager.Client) {
	t.Helper()
	ctx := context.Background()

	// Clean up any existing secret
	_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   ptr(testJSONSecretName),
		ForceDeleteWithoutRecovery: ptr(true),
	})

	// Create JSON test secret
	jsonValue := `{"username":"admin","password":"secret123"}`
	_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         ptr(testJSONSecretName),
		SecretString: ptr(jsonValue),
	})
	if err != nil {
		t.Fatalf("failed to create JSON test secret: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   ptr(testJSONSecretName),
			ForceDeleteWithoutRecovery: ptr(true),
		})
	})
}

func TestAWSSecretsManager_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := localstackClient(t)

	// Check if localstack is available
	ctx := context.Background()
	_, err := client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})
	if err != nil {
		t.Skipf("localstack unavailable: %v", err)
	}

	setupTestSecret(t, client)

	p := awssm.NewWithClient(client)

	t.Run("existing secret", func(t *testing.T) {
		val, err := p.Get(ctx, testSecretName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != testSecretValue {
			t.Errorf("got %q, want %q", val, testSecretValue)
		}
	})

	t.Run("non-existent secret", func(t *testing.T) {
		_, err := p.Get(ctx, "nonexistent-secret")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func TestAWSSecretsManager_GetJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := localstackClient(t)

	// Check if localstack is available
	ctx := context.Background()
	_, err := client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})
	if err != nil {
		t.Skipf("localstack unavailable: %v", err)
	}

	setupJSONTestSecret(t, client)

	p := awssm.NewWithClient(client)

	t.Run("existing JSON field", func(t *testing.T) {
		val, err := p.Get(ctx, testJSONSecretName+":password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret123" {
			t.Errorf("got %q, want %q", val, "secret123")
		}
	})

	t.Run("non-existent JSON field", func(t *testing.T) {
		_, err := p.Get(ctx, testJSONSecretName+":missing")
		if err != fig.ErrSecretNotFound {
			t.Errorf("expected ErrSecretNotFound, got %v", err)
		}
	})
}

func ptr[T any](v T) *T {
	return &v
}
