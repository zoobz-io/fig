# Integration Tests

Integration tests that require external services.

## Running

```bash
make test-integration
```

## Requirements

Tests in this directory may require:

- AWS credentials (for AWS Secrets Manager tests)
- GCP credentials (for GCP Secret Manager tests)
- Vault server (for HashiCorp Vault tests)

## Skipping

Integration tests are excluded from short mode:

```bash
go test -short ./...  # Skips integration tests
```
