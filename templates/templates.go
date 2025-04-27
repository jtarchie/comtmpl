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
