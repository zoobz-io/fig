package fig

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// convert attempts to convert a string value to the target type.
func convert(value string, target reflect.Type) (reflect.Value, error) {
	// Handle pointer types
	if target.Kind() == reflect.Ptr {
		elemType := target.Elem()
		converted, err := convert(value, elemType)
		if err != nil {
			return reflect.Value{}, err
		}
		ptr := reflect.New(elemType)
		ptr.Elem().Set(converted)
		return ptr, nil
	}

	// Check for TextUnmarshaler
	if target.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
		v := reflect.New(target).Elem()
		unmarshaler, ok := v.Addr().Interface().(encoding.TextUnmarshaler)
		if !ok {
			return reflect.Value{}, ErrInvalidType
		}
		if err := unmarshaler.UnmarshalText([]byte(value)); err != nil {
			return reflect.Value{}, err
		}
		return v, nil
	}

	// Check if pointer to type implements TextUnmarshaler
	ptrType := reflect.PointerTo(target)
	if ptrType.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
		v := reflect.New(target)
		unmarshaler, ok := v.Interface().(encoding.TextUnmarshaler)
		if !ok {
			return reflect.Value{}, ErrInvalidType
		}
		if err := unmarshaler.UnmarshalText([]byte(value)); err != nil {
			return reflect.Value{}, err
		}
		return v.Elem(), nil
	}

	switch target.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special case for time.Duration
		if target == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return reflect.Value{}, ErrInvalidType
			}
			return reflect.ValueOf(d), nil
		}
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, ErrInvalidType
		}
		return reflect.ValueOf(i).Convert(target), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, ErrInvalidType
		}
		return reflect.ValueOf(u).Convert(target), nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, ErrInvalidType
		}
		return reflect.ValueOf(f).Convert(target), nil

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, ErrInvalidType
		}
		return reflect.ValueOf(b), nil

	case reflect.Slice:
		if target.Elem().Kind() == reflect.String {
			parts := splitComma(value)
			return reflect.ValueOf(parts), nil
		}
		return reflect.Value{}, ErrInvalidType

	default:
		return reflect.Value{}, ErrInvalidType
	}
}

// splitComma splits a comma-separated string into a slice, trimming whitespace.
func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
