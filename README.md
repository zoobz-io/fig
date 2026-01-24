# fig

Struct tags in, configuration out.

fig loads configuration from environment variables, secret providers, and defaults using Go struct tags. One function call, predictable resolution order.

## Install

```bash
go get github.com/zoobzio/fig
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "log"

    "github.com/zoobzio/fig"
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
go get github.com/zoobzio/fig/awssm

# GCP Secret Manager
go get github.com/zoobzio/fig/gcpsm

# HashiCorp Vault
go get github.com/zoobzio/fig/vault
```

```go
import "github.com/zoobzio/fig/vault"

p, _ := vault.New()
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

- **One function** — `Load` does everything; no builder chains or option structs
- **Predictable** — secret → env → default resolution, every time
- **Minimal** — no external dependencies beyond [sentinel](https://github.com/zoobzio/sentinel)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
