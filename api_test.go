package fig

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testConfig struct {
	Host     string        `env:"TEST_FIG_API_HOST" default:"localhost"`
	Port     int           `env:"TEST_FIG_API_PORT" default:"8080"`
	Debug    bool          `env:"TEST_FIG_API_DEBUG"`
	Timeout  time.Duration `env:"TEST_FIG_API_TIMEOUT" default:"30s"`
	Tags     []string      `env:"TEST_FIG_API_TAGS"`
	Password string        `secret:"test/password"`
}

func TestLoad_Defaults(t *testing.T) {
	var cfg testConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("TEST_FIG_API_HOST", "example.com")
	t.Setenv("TEST_FIG_API_PORT", "9000")
	t.Setenv("TEST_FIG_API_DEBUG", "true")
	t.Setenv("TEST_FIG_API_TAGS", "api,v2,prod")

	var cfg testConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "example.com" {
		t.Errorf("Host = %q, want %q", cfg.Host, "example.com")
	}
	if cfg.Port != 9000 {
		t.Errorf("Port = %d, want %d", cfg.Port, 9000)
	}
	if !cfg.Debug {
		t.Error("Debug should be true")
	}
	if len(cfg.Tags) != 3 || cfg.Tags[0] != "api" {
		t.Errorf("Tags = %v, want [api v2 prod]", cfg.Tags)
	}
}

func TestLoad_SecretProvider(t *testing.T) {
	provider := &mockProvider{secrets: map[string]string{"test/password": "secret123"}}

	var cfg testConfig
	if err := Load(&cfg, provider); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Password != "secret123" {
		t.Errorf("Password = %q, want %q", cfg.Password, "secret123")
	}
}

type requiredConfig struct {
	Name string `env:"TEST_FIG_REQUIRED_NAME" required:"true"`
}

func TestLoad_Required(t *testing.T) {
	var cfg requiredConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}

	var fieldErr *FieldError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("expected FieldError, got %T", err)
	}
	if fieldErr.Field != "Name" {
		t.Errorf("FieldError.Field = %q, want %q", fieldErr.Field, "Name")
	}
}

type nestedDatabaseConfig struct {
	Host string `env:"TEST_FIG_NESTED_DB_HOST" default:"db.local"`
	Port int    `env:"TEST_FIG_NESTED_DB_PORT" default:"5432"`
}

type nestedAppConfig struct {
	Name     string `env:"TEST_FIG_NESTED_APP_NAME" default:"myapp"`
	Database nestedDatabaseConfig
}

func TestLoad_NestedStruct(t *testing.T) {
	var cfg nestedAppConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "myapp" {
		t.Errorf("Name = %q, want %q", cfg.Name, "myapp")
	}
	if cfg.Database.Host != "db.local" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "db.local")
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want %d", cfg.Database.Port, 5432)
	}
}

type validatingConfig struct {
	Port int `env:"TEST_FIG_VALIDATE_PORT" default:"8080"`
}

func (c *validatingConfig) Validate() error {
	if c.Port <= 0 {
		return errors.New("port must be positive")
	}
	return nil
}

func TestLoad_Validator(t *testing.T) {
	t.Setenv("TEST_FIG_VALIDATE_PORT", "-1")

	var cfg validatingConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "port must be positive" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoadContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cfg testConfig
	if err := LoadContext(ctx, &cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
}

func TestLoadContext_WithProvider(t *testing.T) {
	ctx := context.Background()
	provider := &mockProvider{secrets: map[string]string{"test/password": "ctx-secret"}}

	var cfg testConfig
	if err := LoadContext(ctx, &cfg, provider); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Password != "ctx-secret" {
		t.Errorf("Password = %q, want %q", cfg.Password, "ctx-secret")
	}
}
