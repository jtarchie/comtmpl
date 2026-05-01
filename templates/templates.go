package templates

import (
	"fmt"
	"io"
	"reflect"
	textTemplates "text/template"
)

type Template func(*Templates, io.Writer, any) error
type Templates struct {
	templates map[string]Template
	funcs     textTemplates.FuncMap
}

func NewTemplates(templates map[string]Template) *Templates {
	instance := &Templates{
		templates: templates,
		funcs:     make(textTemplates.FuncMap),
	}
	instance.funcs = builtins()

	return instance
}

func (t *Templates) Funcs(funcs textTemplates.FuncMap) *Templates {
	for name, fn := range funcs {
		t.funcs[name] = fn
	}

	return t
}

func (t *Templates) ExecuteTemplate(writer io.Writer, name string, data any) error {
	if template, ok := t.templates[name]; ok {
		return template(t, writer, data)
	}
	return fmt.Errorf("template %q not found", name)
}

// Helper function to evaluate field access like .Name or .User.Name
func EvalField(data any, parts []string) (any, error) {
	if data == nil {
		return "", nil
	}

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

// MustEvalField evaluates a field path and panics on error
// Used for template generation where we need to embed field access in arguments
func MustEvalField(data any, parts []string) any {
	value, err := EvalField(data, parts)
	if err != nil {
		panic(err)
	}
	return value
}

// Call a named function from the FuncMap with provided arguments
func (t *Templates) CallFunc(funcName string, args ...any) (any, error) {
	fn, ok := t.funcs[funcName]
	if !ok {
		return nil, fmt.Errorf("function %q not found", funcName)
	}

	value, err := callFunction(fn, args...)
	if err != nil {
		return nil, fmt.Errorf("error calling function %q: %w", funcName, err)
	}

	return value, nil
}

// GetFunc retrieves a function from the FuncMap by name
func (t *Templates) GetFunc(name string) any {
	if fn, ok := t.funcs[name]; ok {
		return fn
	}
	return nil
}

// Helper function to call a function using reflection
func callFunction(fn any, args ...any) (any, error) {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function: %T", fn)
	}

	// Check if function is variadic
	isVariadic := v.Type().IsVariadic()

	// Prepare arguments
	var in []reflect.Value
	numIn := v.Type().NumIn()

	// Handle regular functions
	if !isVariadic {
		if len(args) != numIn {
			return nil, fmt.Errorf("function expects %d arguments, got %d", numIn, len(args))
		}

		in = make([]reflect.Value, len(args))
		for i, arg := range args {
			if arg == nil {
				// If argument is nil, use zero value of expected type
				in[i] = reflect.Zero(v.Type().In(i))
			} else {
				// Convert the argument to the expected type if necessary
				argValue := reflect.ValueOf(arg)
				paramType := v.Type().In(i)

				if argValue.Type().ConvertibleTo(paramType) {
					in[i] = argValue.Convert(paramType)
				} else {
					return nil, fmt.Errorf("cannot convert argument %d from %s to %s", i, argValue.Type(), paramType)
				}
			}
		}

		// Call function
		out := v.Call(in)

		// Handle return values
		if len(out) == 0 {
			return nil, nil
		} else if len(out) == 1 {
			if out[0].CanInterface() {
				return out[0].Interface(), nil
			}
			return nil, nil
		} else {
			// If multiple return values, return the first one and ignore errors
			// This matches how Go templates handle multiple return values
			if out[0].CanInterface() {
				return out[0].Interface(), nil
			}
			return nil, nil
		}
	} else {
		// Handle variadic functions
		// Number of fixed arguments
		fixedArgs := numIn - 1

		if len(args) < fixedArgs {
			return nil, fmt.Errorf("variadic function expects at least %d arguments, got %d", fixedArgs, len(args))
		}

		in = make([]reflect.Value, len(args))
		for i, arg := range args {
			if arg == nil {
				if i < fixedArgs {
					in[i] = reflect.Zero(v.Type().In(i))
				} else {
					in[i] = reflect.Zero(v.Type().In(fixedArgs).Elem())
				}
			} else {
				in[i] = reflect.ValueOf(arg)
			}
		}

		// Call variadic function
		out := v.CallSlice(in)

		// Handle return values
		if len(out) == 0 {
			return nil, nil
		} else if len(out) == 1 {
			if out[0].CanInterface() {
				return out[0].Interface(), nil
			}
			return nil, nil
		} else {
			if out[0].CanInterface() {
				return out[0].Interface(), nil
			}
			return nil, nil
		}
	}
}

// IsTrue evaluates whether a value is truthy according to Go template rules
func IsTrue(val any) (bool, error) {
	if val == nil {
		return false, nil
	}

	b, ok := val.(bool)
	if ok {
		return b, nil
	}

	// Handle other types using reflection
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() > 0, nil
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() != 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0, nil
	case reflect.Struct:
		return true, nil // Non-nil structs are always true
	case reflect.Pointer, reflect.Interface:
		return !v.IsNil(), nil
	default:
		return false, fmt.Errorf("cannot determine truth value of type %s", v.Type())
	}
}

// GetIterable converts a value into an iterable map or slice for range loops
func GetIterable(val any) (any, error) {
	if val == nil {
		return make(map[string]any), nil // Empty map for nil
	}

	// Handle common types directly
	switch v := val.(type) {
	case map[string]any:
		return v, nil
	case []any:
		return v, nil
	case string:
		// Convert string to a slice of runes
		runes := []rune(v)
		result := make([]any, len(runes))
		for i, r := range runes {
			result[i] = string(r)
		}
		return result, nil
	}

	// Use reflection for other types
	v := reflect.ValueOf(val)

	// Dereference pointers
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return make(map[string]any), nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		// Convert to a map[string]any
		result := make(map[string]any)
		for _, key := range v.MapKeys() {
			if s, ok := key.Interface().(string); ok {
				value := v.MapIndex(key)
				if value.CanInterface() {
					result[s] = value.Interface()
				} else {
					result[s] = nil
				}
			} else {
				// Try to convert key to string
				result[fmt.Sprintf("%v", key.Interface())] = v.MapIndex(key).Interface()
			}
		}
		return result, nil

	case reflect.Slice, reflect.Array:
		// Convert to a []any
		length := v.Len()
		result := make([]any, length)
		for i := 0; i < length; i++ {
			item := v.Index(i)
			if item.CanInterface() {
				result[i] = item.Interface()
			} else {
				result[i] = nil
			}
		}
		return result, nil

	case reflect.Struct:
		// Convert struct to a map of field names to values
		result := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanInterface() {
				result[t.Field(i).Name] = field.Interface()
			}
		}
		return result, nil

	case reflect.String:
		// Split the string into runes
		str := v.String()
		runes := []rune(str)
		result := make([]any, len(runes))
		for i, r := range runes {
			result[i] = string(r)
		}
		return result, nil

	default:
		// Non-iterable type, return a singular item in a slice
		if v.CanInterface() {
			return []any{v.Interface()}, nil
		}
		return []any{}, fmt.Errorf("value of type %s cannot be iterated", v.Type())
	}
}

// ConvertToAnySlice converts any iterable type to []any
func ConvertToAnySlice(val any) ([]any, error) {
	// Already a slice of any
	if slice, ok := val.([]any); ok {
		return slice, nil
	}

	v := reflect.ValueOf(val)

	// Dereference pointers
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return []any{}, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		// Convert to a []any
		length := v.Len()
		result := make([]any, length)
		for i := 0; i < length; i++ {
			item := v.Index(i)
			if item.CanInterface() {
				result[i] = item.Interface()
			} else {
				result[i] = nil
			}
		}
		return result, nil

	case reflect.Map:
		// For maps, return a slice of key-value pairs
		keys := v.MapKeys()
		result := make([]any, len(keys))
		for i, key := range keys {
			pair := map[string]any{
				"key":   key.Interface(),
				"value": v.MapIndex(key).Interface(),
			}
			result[i] = pair
		}
		return result, nil

	case reflect.String:
		// Convert string to slice of runes as strings
		str := v.String()
		runes := []rune(str)
		result := make([]any, len(runes))
		for i, r := range runes {
			result[i] = string(r)
		}
		return result, nil

	default:
		// For non-iterable types, wrap in a single-item slice
		if v.CanInterface() {
			return []any{v.Interface()}, nil
		}
		return []any{}, fmt.Errorf("value of type %s cannot be converted to slice", v.Type())
	}
}

// Dot returns the current dot value. Inside a range scope, the iteration value
// is stored at key "."; outside a range scope, dot is the data itself.
func Dot(data any) any {
	if m, ok := data.(map[string]any); ok {
		if v, exists := m["."]; exists {
			return v
		}
	}
	return data
}

// NewRangeScope creates a new data context for use in a range loop
// It combines the outer context with index and value variables
func NewRangeScope(outerData any, index any, value any) any {
	// Create a map to represent the range scope
	scope := make(map[string]any)

	// Add the special range variables
	scope["."] = value     // Current value becomes the dot
	scope["$"] = outerData // Original data becomes $
	scope["index"] = index // Index is available as index
	scope["value"] = value // Value is available as value

	// If outer data is a map, incorporate its values
	if outerMap, ok := outerData.(map[string]any); ok {
		for k, v := range outerMap {
			if k != "." && k != "$" && k != "index" && k != "value" {
				scope[k] = v
			}
		}
	}

	return scope
}
