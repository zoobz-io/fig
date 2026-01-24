package fig

import "context"

// SecretProvider defines the interface for secret backends.
type SecretProvider interface {
	// Get retrieves a secret by key.
	// Returns ErrSecretNotFound if the secret does not exist.
	Get(ctx context.Context, key string) (string, error)
}
