package fig

import (
	"context"
	"errors"
	"os"
	"reflect"

	"github.com/zoobzio/sentinel"
)

// fieldTags holds parsed struct tags for a field.
type fieldTags struct {
	env      string
	secret   string
	defValue string
	required bool
}

// parseFieldTags extracts fig-related tags from sentinel field metadata.
func parseFieldTags(field sentinel.FieldMetadata) fieldTags {
	return fieldTags{
		env:      field.Tags["env"],
		secret:   field.Tags["secret"],
		defValue: field.Tags["default"],
		required: field.Tags["required"] == "true",
	}
}

// resolveField attempts to resolve a value for a field from configured sources.
// Resolution order: secret -> env -> default -> zero value.
func resolveField(ctx context.Context, tags fieldTags, provider SecretProvider) (string, bool, error) {
	// 1. Try secret provider
	if tags.secret != "" && provider != nil {
		val, err := provider.Get(ctx, tags.secret)
		if err == nil {
			return val, true, nil
		}
		// Only continue if secret not found; other errors should propagate
		if !errors.Is(err, ErrSecretNotFound) {
			return "", false, err
		}
	}

	// 2. Try environment variable
	if tags.env != "" {
		if val := os.Getenv(tags.env); val != "" {
			return val, true, nil
		}
	}

	// 3. Try default value
	if tags.defValue != "" {
		return tags.defValue, true, nil
	}

	// 4. No value found
	return "", false, nil
}

// loadFromMetadata loads configuration using sentinel metadata.
func loadFromMetadata(ctx context.Context, v reflect.Value, meta sentinel.Metadata, provider SecretProvider) error {
	for _, field := range meta.Fields {
		fieldVal := v.FieldByIndex(field.Index)

		// Skip if can't set
		if !fieldVal.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Kind == sentinel.KindStruct {
			tags := parseFieldTags(field)
			// If no fig tags, treat as nested struct and recurse
			if tags.env == "" && tags.secret == "" && tags.defValue == "" && !tags.required {
				// Look up nested struct metadata from sentinel cache
				fqdn := field.ReflectType.PkgPath() + "." + field.ReflectType.Name()
				if nestedMeta, found := sentinel.Lookup(fqdn); found {
					if err := loadFromMetadata(ctx, fieldVal, nestedMeta, provider); err != nil {
						return err
					}
				}
				continue
			}
		}

		tags := parseFieldTags(field)

		// Skip fields with no configuration tags
		if tags.env == "" && tags.secret == "" && tags.defValue == "" && !tags.required {
			continue
		}

		// Resolve the value
		strVal, found, err := resolveField(ctx, tags, provider)
		if err != nil {
			return &FieldError{Field: field.Name, Err: err}
		}

		// Check required
		if !found && tags.required {
			return &FieldError{Field: field.Name, Err: ErrRequired}
		}

		// Convert and set if found
		if found {
			converted, err := convert(strVal, field.ReflectType)
			if err != nil {
				return &FieldError{Field: field.Name, Err: err}
			}
			fieldVal.Set(converted)
		}
	}

	return nil
}
