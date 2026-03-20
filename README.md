# fig

[![CI](https://github.com/zoobz-io/fig/actions/workflows/ci.yml/badge.svg)](https://github.com/zoobz-io/fig/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobz-io/fig/branch/main/graph/badge.svg)](https://codecov.io/gh/zoobz-io/fig)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobz-io/fig)](https://goreportcard.com/report/github.com/zoobz-io/fig)
[![CodeQL](https://github.com/zoobz-io/fig/actions/workflows/codeql.yml/badge.svg)](https://github.com/zoobz-io/fig/actions/workflows/codeql.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobz-io/fig.svg)](https://pkg.go.dev/github.com/zoobz-io/fig)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobz-io/fig)](go.mod)
[![Release](https://img.shields.io/github/v/release/zoobz-io/fig)](https://github.com/zoobz-io/fig/releases/latest)

Struct tags in, configuration out.

fig loads configuration from environment variables, secret providers, and defaults using Go struct tags. One function call, predictable resolution order.

## Install

```bash
go get github.com/zoobz-io/fig
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "log"

    "github.com/zoobz-io/fig"
)

type Config struct {
    Host     string   `env:"APP_HOST" default:"localhost"`
    Port     int      `env:"APP_PORT" default:"8080"`
    Password string   `secret:"db/password"`
    Tags     []string `env:"APP_TAGS"`
    APIKey   string   `env:"API_KEY" required:"true"`
}

func main() {
    var cfg Config
    if err := fig.Load(&cfg); err != nil {
        log.Fatal(err)
    }
    // cfg is now populated
}
```

Resolution order: `secret` → `env` → `default` → zero value.

## Capabilities

| Feature | Description |
|---------|-------------|
| Environment variables | `env:"VAR_NAME"` tag |
| Secret providers | `secret:"path/to/secret"` tag with pluggable backends |
| Default values | `default:"value"` tag |
| Required fields | `required:"true"` tag |
| Nested structs | Automatic recursion into embedded structs |
| Validation | Implement `Validator` interface for custom checks |
| Context support | `LoadContext` for secret provider timeouts |

### Supported Types

`string`, `int`, `int8-64`, `uint`, `uint8-64`, `float32`, `float64`, `bool`, `time.Duration`, `[]string` (comma-separated), and any type implementing `encoding.TextUnmarshaler`.

## Secret Providers

Pass a provider implementing `SecretProvider` to load secrets:

```go
type SecretProvider interface {
    Get(ctx context.Context, key string) (string, error)
}
```

### Available Providers

Each provider is a separate module — import only what you need:

```bash
# AWS Secrets Manager
go get github.com/zoobz-io/fig/awssm

# GCP Secret Manager
go get github.com/zoobz-io/fig/gcpsm

# HashiCorp Vault
go get github.com/zoobz-io/fig/vault
```

```go
import "github.com/zoobz-io/fig/vault"

p, err := vault.New()
if err != nil {
    log.Fatal(err)
}
fig.Load(&cfg, p)
```

Secrets take precedence over environment variables, allowing secure overrides.

## Validation

Implement the `Validator` interface for custom validation after loading:

```go
func (c *Config) Validate() error {
    if c.Port <= 0 || c.Port > 65535 {
        return errors.New("port must be between 1 and 65535")
    }
    return nil
}
```

## Why fig?

No config files. No YAML. No JSON. No builder chains. One function, one resolution order, done.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
