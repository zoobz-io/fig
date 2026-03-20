// Package fig provides configuration loading from environment variables and secret providers.
//
// fig uses struct tags to declare configuration sources:
//
//	type Config struct {
//	    Host     string        `env:"DB_HOST" default:"localhost"`
//	    Port     int           `env:"DB_PORT" default:"5432"`
//	    Password string        `secret:"db/password"`
//	    Timeout  time.Duration `env:"TIMEOUT" default:"30s"`
//	    Debug    bool          `env:"DEBUG"`
//	    Tags     []string      `env:"TAGS"`
//	    Name     string        `env:"NAME" required:"true"`
//	}
//
// Resolution order: secret -> env -> default -> zero value.
//
// # Supported Types
//
//   - string
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - bool
//   - time.Duration
//   - []string (comma-separated)
//   - any type implementing encoding.TextUnmarshaler
//
// # Secret Providers
//
// Pass a secret provider when loading:
//
//	fig.Load(&cfg, vault.New(...))
//
// # Validation
//
// If the config struct implements Validator, Validate() is called after loading:
//
//	func (c *Config) Validate() error {
//	    if c.Port <= 0 {
//	        return errors.New("port must be positive")
//	    }
//	    return nil
//	}
package fig

import (
	"context"
	"reflect"

	"github.com/zoobz-io/sentinel"
)

func init() {
	// Register fig's struct tags with sentinel
	sentinel.Tag("env")
	sentinel.Tag("secret")
	sentinel.Tag("default")
	sentinel.Tag("required")
}

// Validator is an optional interface for config validation.
type Validator interface {
	Validate() error
}

// Load populates a struct from environment variables, secrets, and defaults.
// An optional SecretProvider can be passed for secret tag resolution.
func Load[T any](cfg *T, provider ...SecretProvider) error {
	return LoadContext(context.Background(), cfg, provider...)
}

// LoadContext populates a struct with context support for secret provider timeouts.
// An optional SecretProvider can be passed for secret tag resolution.
func LoadContext[T any](ctx context.Context, cfg *T, provider ...SecretProvider) error {
	meta, err := sentinel.TryScan[T]()
	if err != nil {
		return ErrNotStruct
	}

	var p SecretProvider
	if len(provider) > 0 {
		p = provider[0]
	}

	v := reflect.ValueOf(cfg).Elem()

	if err := loadFromMetadata(ctx, v, meta, p); err != nil {
		return err
	}

	if validator, ok := any(cfg).(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
