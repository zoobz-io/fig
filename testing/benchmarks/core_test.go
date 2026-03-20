//go:build testing

package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/zoobz-io/fig"
	figtesting "github.com/zoobz-io/fig/testing"
)

// Small config with 3 fields
type SmallConfig struct {
	Host string `env:"BENCH_HOST" default:"localhost"`
	Port int    `env:"BENCH_PORT" default:"8080"`
	Name string `env:"BENCH_NAME" default:"app"`
}

// Medium config with 10 fields
type MediumConfig struct {
	Host        string        `env:"BENCH_HOST" default:"localhost"`
	Port        int           `env:"BENCH_PORT" default:"8080"`
	Name        string        `env:"BENCH_NAME" default:"app"`
	Debug       bool          `env:"BENCH_DEBUG" default:"false"`
	Timeout     time.Duration `env:"BENCH_TIMEOUT" default:"30s"`
	MaxRetries  int           `env:"BENCH_RETRIES" default:"3"`
	LogLevel    string        `env:"BENCH_LOG" default:"info"`
	Environment string        `env:"BENCH_ENV" default:"development"`
	Version     string        `env:"BENCH_VERSION" default:"1.0.0"`
	Workers     int           `env:"BENCH_WORKERS" default:"4"`
}

// Large config with 20 fields
type LargeConfig struct {
	Host           string        `env:"BENCH_HOST" default:"localhost"`
	Port           int           `env:"BENCH_PORT" default:"8080"`
	Name           string        `env:"BENCH_NAME" default:"app"`
	Debug          bool          `env:"BENCH_DEBUG" default:"false"`
	Timeout        time.Duration `env:"BENCH_TIMEOUT" default:"30s"`
	MaxRetries     int           `env:"BENCH_RETRIES" default:"3"`
	LogLevel       string        `env:"BENCH_LOG" default:"info"`
	Environment    string        `env:"BENCH_ENV" default:"development"`
	Version        string        `env:"BENCH_VERSION" default:"1.0.0"`
	Workers        int           `env:"BENCH_WORKERS" default:"4"`
	CacheSize      int           `env:"BENCH_CACHE" default:"1000"`
	BatchSize      int           `env:"BENCH_BATCH" default:"100"`
	ReadTimeout    time.Duration `env:"BENCH_READ_TO" default:"10s"`
	WriteTimeout   time.Duration `env:"BENCH_WRITE_TO" default:"10s"`
	IdleTimeout    time.Duration `env:"BENCH_IDLE_TO" default:"60s"`
	MaxConnections int           `env:"BENCH_MAX_CONN" default:"100"`
	BufferSize     int           `env:"BENCH_BUFFER" default:"4096"`
	CompressLevel  int           `env:"BENCH_COMPRESS" default:"6"`
	RateLimit      float64       `env:"BENCH_RATE" default:"1000.0"`
	EnableMetrics  bool          `env:"BENCH_METRICS" default:"true"`
}

// Config with secrets
type SecretConfig struct {
	Host     string `env:"BENCH_HOST" default:"localhost"`
	Password string `secret:"db/password"`
	APIKey   string `secret:"api/key"`
}

func BenchmarkResolve_Small(b *testing.B) {
	var cfg SmallConfig
	for b.Loop() {
		_ = fig.Load(&cfg)
	}
}

func BenchmarkResolve_Medium(b *testing.B) {
	var cfg MediumConfig
	for b.Loop() {
		_ = fig.Load(&cfg)
	}
}

func BenchmarkResolve_Large(b *testing.B) {
	var cfg LargeConfig
	for b.Loop() {
		_ = fig.Load(&cfg)
	}
}

func BenchmarkResolve_WithProvider(b *testing.B) {
	p := &figtesting.MockProvider{
		Secrets: map[string]string{
			"db/password": "secret-password",
			"api/key":     "secret-api-key",
		},
	}
	var cfg SecretConfig
	for b.Loop() {
		_ = fig.Load(&cfg, p)
	}
}

func BenchmarkProviderGet_Hit(b *testing.B) {
	p := &figtesting.MockProvider{
		Secrets: map[string]string{
			"key": "value",
		},
	}
	ctx := context.Background()
	for b.Loop() {
		_, _ = p.Get(ctx, "key")
	}
}

func BenchmarkProviderGet_Miss(b *testing.B) {
	p := &figtesting.MockProvider{
		Secrets: map[string]string{},
	}
	ctx := context.Background()
	for b.Loop() {
		_, _ = p.Get(ctx, "nonexistent")
	}
}
