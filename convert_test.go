package fig

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestConvert_String(t *testing.T) {
	v, err := convert("hello", reflect.TypeOf(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.String() != "hello" {
		t.Errorf("got %q, want %q", v.String(), "hello")
	}
}

func TestConvert_Int(t *testing.T) {
	v, err := convert("42", reflect.TypeOf(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Int() != 42 {
		t.Errorf("got %d, want %d", v.Int(), 42)
	}
}

func TestConvert_Int64(t *testing.T) {
	v, err := convert("9223372036854775807", reflect.TypeOf(int64(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Int() != 9223372036854775807 {
		t.Errorf("got %d, want %d", v.Int(), int64(9223372036854775807))
	}
}

func TestConvert_IntVariants(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		target any
		want   int64
	}{
		{"int8", "127", int8(0), 127},
		{"int8 negative", "-128", int8(0), -128},
		{"int16", "32767", int16(0), 32767},
		{"int16 negative", "-32768", int16(0), -32768},
		{"int32", "2147483647", int32(0), 2147483647},
		{"int32 negative", "-2147483648", int32(0), -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := convert(tt.input, reflect.TypeOf(tt.target))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v.Int() != tt.want {
				t.Errorf("got %d, want %d", v.Int(), tt.want)
			}
		})
	}
}

func TestConvert_Bool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
	}

	for _, tt := range tests {
		v, err := convert(tt.input, reflect.TypeOf(false))
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tt.input, err)
		}
		if v.Bool() != tt.want {
			t.Errorf("convert(%q) = %v, want %v", tt.input, v.Bool(), tt.want)
		}
	}
}

func TestConvert_Float64(t *testing.T) {
	v, err := convert("3.14", reflect.TypeOf(float64(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Float() != 3.14 {
		t.Errorf("got %f, want %f", v.Float(), 3.14)
	}
}

func TestConvert_Duration(t *testing.T) {
	v, err := convert("30s", reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := v.Interface().(time.Duration)
	if !ok {
		t.Fatal("type assertion failed")
	}
	if got != 30*time.Second {
		t.Errorf("got %v, want %v", got, 30*time.Second)
	}
}

func TestConvert_StringSlice(t *testing.T) {
	v, err := convert("a, b, c", reflect.TypeOf([]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := v.Interface().([]string)
	if !ok {
		t.Fatal("type assertion failed")
	}
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestConvert_InvalidInt(t *testing.T) {
	_, err := convert("not-a-number", reflect.TypeOf(0))
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestConvert_InvalidBool(t *testing.T) {
	_, err := convert("maybe", reflect.TypeOf(false))
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestConvert_Pointer(t *testing.T) {
	t.Run("pointer to string", func(t *testing.T) {
		var s *string
		v, err := convert("hello", reflect.TypeOf(s))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Kind() != reflect.Ptr {
			t.Fatalf("expected pointer, got %v", v.Kind())
		}
		if v.Elem().String() != "hello" {
			t.Errorf("got %q, want %q", v.Elem().String(), "hello")
		}
	})

	t.Run("pointer to int", func(t *testing.T) {
		var i *int
		v, err := convert("42", reflect.TypeOf(i))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Elem().Int() != 42 {
			t.Errorf("got %d, want %d", v.Elem().Int(), 42)
		}
	})

	t.Run("invalid pointer conversion", func(t *testing.T) {
		var i *int
		_, err := convert("not-a-number", reflect.TypeOf(i))
		if !errors.Is(err, ErrInvalidType) {
			t.Errorf("expected ErrInvalidType, got %v", err)
		}
	})
}

// textUnmarshalerType implements encoding.TextUnmarshaler with pointer receiver.
type textUnmarshalerType struct {
	value string
}

func (t *textUnmarshalerType) UnmarshalText(text []byte) error {
	if string(text) == "fail" {
		return errors.New("unmarshal failed")
	}
	t.value = "unmarshaled:" + string(text)
	return nil
}

// valueReceiverUnmarshaler implements encoding.TextUnmarshaler with value receiver.
// This exercises the first TextUnmarshaler branch (lines 26-35).
type valueReceiverUnmarshaler string

func (v valueReceiverUnmarshaler) UnmarshalText(_ []byte) error {
	return nil
}

// failingValueUnmarshaler implements encoding.TextUnmarshaler with value receiver
// but returns an error.
type failingValueUnmarshaler string

func (f failingValueUnmarshaler) UnmarshalText(_ []byte) error {
	return errors.New("value receiver unmarshal failed")
}

func TestConvert_TextUnmarshaler(t *testing.T) {
	v, err := convert("test", reflect.TypeOf(textUnmarshalerType{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := v.Interface().(textUnmarshalerType)
	if !ok {
		t.Fatal("type assertion failed")
	}
	if got.value != "unmarshaled:test" {
		t.Errorf("got %q, want %q", got.value, "unmarshaled:test")
	}
}

func TestConvert_TextUnmarshaler_PointerReceiverError(t *testing.T) {
	_, err := convert("fail", reflect.TypeOf(textUnmarshalerType{}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unmarshal failed" {
		t.Errorf("got error %q, want %q", err.Error(), "unmarshal failed")
	}
}

func TestConvert_TextUnmarshaler_ValueReceiver(t *testing.T) {
	_, err := convert("test", reflect.TypeOf(valueReceiverUnmarshaler("")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConvert_TextUnmarshaler_ValueReceiverError(t *testing.T) {
	_, err := convert("test", reflect.TypeOf(failingValueUnmarshaler("")))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "value receiver unmarshal failed" {
		t.Errorf("got error %q, want %q", err.Error(), "value receiver unmarshal failed")
	}
}

func TestConvert_Uint(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		target any
		want   uint64
	}{
		{"uint", "42", uint(0), 42},
		{"uint8", "255", uint8(0), 255},
		{"uint16", "65535", uint16(0), 65535},
		{"uint32", "4294967295", uint32(0), 4294967295},
		{"uint64", "18446744073709551615", uint64(0), 18446744073709551615},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := convert(tt.input, reflect.TypeOf(tt.target))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v.Uint() != tt.want {
				t.Errorf("got %d, want %d", v.Uint(), tt.want)
			}
		})
	}
}

func TestConvert_Float32(t *testing.T) {
	v, err := convert("3.14", reflect.TypeOf(float32(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Use tolerance for float32 comparison
	got := float32(v.Float())
	if got < 3.13 || got > 3.15 {
		t.Errorf("got %f, want ~3.14", got)
	}
}

func TestConvert_InvalidFloat(t *testing.T) {
	_, err := convert("not-a-float", reflect.TypeOf(float64(0)))
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestConvert_InvalidUint(t *testing.T) {
	_, err := convert("not-a-uint", reflect.TypeOf(uint(0)))
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestConvert_InvalidDuration(t *testing.T) {
	_, err := convert("not-a-duration", reflect.TypeOf(time.Duration(0)))
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestConvert_UnsupportedSlice(t *testing.T) {
	_, err := convert("1,2,3", reflect.TypeOf([]int{}))
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType for []int, got %v", err)
	}
}

func TestConvert_UnsupportedType(t *testing.T) {
	type unsupported struct{}
	_, err := convert("value", reflect.TypeOf(unsupported{}))
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType for unsupported type, got %v", err)
	}
}

func TestSplitComma(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"a,b,c", []string{"a", "b", "c"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"a,,b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		got := splitComma(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitComma(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("splitComma(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
