package templates

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Template func(io.Writer, any) error
type Templates map[string]Template

func (t Templates) ExecuteTemplate(writer io.Writer, name string, data any) error {
	if template, ok := t[name]; ok {
		return template(writer, data)
	}
	return fmt.Errorf("template %q not found", name)
}

// Helper function to evaluate field access like .Name or .User.Name
func EvalField(data any, fieldPath string) (any, error) {
	if data == nil {
		return "", nil
	}

	// Trim the leading dot
	fieldPath = strings.TrimPrefix(fieldPath, ".")
	parts := strings.Split(fieldPath, ".")

	var current = data

	// Traverse the parts
	for _, part := range parts {
		// Handle nil values
		if current == nil {
			return "", nil
		}

		// Fast path for common map types without using reflection
		switch m := current.(type) {
		case map[string]interface{}:
			val, ok := m[part]
			if !ok {
				return "", nil
			}
			current = val
			continue

		case map[string]string:
			val, ok := m[part]
			if !ok {
				return "", nil
			}
			current = val
			continue
		}

		// Use reflection only when necessary
		v := reflect.ValueOf(current)

		// Dereference pointers and interfaces
		for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			if v.IsNil() {
				return "", nil
			}
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Map:
			if v.Type().Key().Kind() != reflect.String {
				return "", fmt.Errorf("map key must be string")
			}
			keyValue := reflect.ValueOf(part)
			v = v.MapIndex(keyValue)
			if !v.IsValid() {
				return "", nil
			}
			if !v.CanInterface() {
				return "", fmt.Errorf("cannot access map value")
			}
			current = v.Interface()

		case reflect.Struct:
			v = v.FieldByName(part)
			if !v.IsValid() {
				return "", nil
			}
			if !v.CanInterface() {
				return "", fmt.Errorf("cannot access unexported field %s", part)
			}
			current = v.Interface()

		default:
			return "", fmt.Errorf("cannot access %s in %v type", part, v.Kind())
		}
	}

	return current, nil
}
