package fig

import (
	"context"
	"testing"
)

func TestResolveField_Secret(t *testing.T) {
	provider := &mockProvider{secrets: map[string]string{"db/password": "secret123"}}
	tags := fieldTags{secret: "db/password"}

	val, found, err := resolveField(context.Background(), tags, provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected value to be found")
	}
	if val != "secret123" {
		t.Errorf("got %q, want %q", val, "secret123")
	}
}

func TestResolveField_Env(t *testing.T) {
	t.Setenv("TEST_FIG_HOST", "localhost")

	tags := fieldTags{env: "TEST_FIG_HOST"}

	val, found, err := resolveField(context.Background(), tags, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected value to be found")
	}
	if val != "localhost" {
		t.Errorf("got %q, want %q", val, "localhost")
	}
}

func TestResolveField_Default(t *testing.T) {
	tags := fieldTags{defValue: "default-value"}

	val, found, err := resolveField(context.Background(), tags, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected value to be found")
	}
	if val != "default-value" {
		t.Errorf("got %q, want %q", val, "default-value")
	}
}

func TestResolveField_Priority(t *testing.T) {
	// Secret takes priority over env
	t.Setenv("TEST_FIG_PRIORITY", "from-env")

	provider := &mockProvider{secrets: map[string]string{"test/priority": "from-secret"}}
	tags := fieldTags{
		secret:   "test/priority",
		env:      "TEST_FIG_PRIORITY",
		defValue: "from-default",
	}

	val, found, err := resolveField(context.Background(), tags, provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected value to be found")
	}
	if val != "from-secret" {
		t.Errorf("got %q, want %q (secret should take priority)", val, "from-secret")
	}
}

func TestResolveField_EnvOverDefault(t *testing.T) {
	t.Setenv("TEST_FIG_ENV_DEF", "from-env")

	tags := fieldTags{
		env:      "TEST_FIG_ENV_DEF",
		defValue: "from-default",
	}

	val, found, err := resolveField(context.Background(), tags, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected value to be found")
	}
	if val != "from-env" {
		t.Errorf("got %q, want %q (env should take priority over default)", val, "from-env")
	}
}

func TestResolveField_NotFound(t *testing.T) {
	tags := fieldTags{env: "TEST_FIG_NONEXISTENT"}

	_, found, err := resolveField(context.Background(), tags, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected value to not be found")
	}
}
