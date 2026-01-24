# Testing Infrastructure

Test utilities and shared helpers for fig.

## Structure

```
testing/
├── helpers.go       # Shared test utilities
├── helpers_test.go  # Tests for helpers
├── integration/     # Integration tests
└── benchmarks/      # Performance benchmarks
```

## Usage

Import the testing package in your tests:

```go
import figtesting "github.com/zoobzio/fig/testing"
```

### MockProvider

```go
func TestMyFunction(t *testing.T) {
    p := figtesting.NewMockProvider(t, map[string]string{
        "api-key": "secret-value",
    })
    // Use p as a fig.SecretProvider
}
```

### SetEnv

```go
func TestWithEnv(t *testing.T) {
    figtesting.SetEnv(t, "MY_VAR", "value")
    // Environment variable is automatically cleaned up after test
}
```

## Build Tag

This package uses the `testing` build tag. Include `-tags testing` when running tests:

```bash
go test -tags testing ./...
```
