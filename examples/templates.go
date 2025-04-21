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

var Parsed = Templates{
	"index.html": func(writer io.Writer, data any) error {
		var err error
		// index.html:1
		_, err = io.WriteString(writer, `<html>
  <head>
    <title>`)
		if err != nil {
			return err
		}
		//index.html:3

		// Handle {{.Title}}
		var value0 any
		value0, err = evalField(data, ".Title")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, value0)
		if err != nil {
			return err
		}
		// index.html:3
		_, err = io.WriteString(writer, `</title>
  </head>
  <body>
    <h1>`)
		if err != nil {
			return err
		}
		//index.html:6

		// Handle {{.Title}}
		var value1 any
		value1, err = evalField(data, ".Title")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, value1)
		if err != nil {
			return err
		}
		// index.html:6
		_, err = io.WriteString(writer, `</h1>
    <p>Welcome, `)
		if err != nil {
			return err
		}
		//index.html:7

		// Handle {{.User.Name}}
		var value2 any
		value2, err = evalField(data, ".User.Name")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, value2)
		if err != nil {
			return err
		}
		// index.html:7
		_, err = io.WriteString(writer, `!</p>
  </body>
</html>`)
		if err != nil {
			return err
		}

		return nil
	},
}

// Helper function to evaluate field access like .Name or .User.Name
func evalField(data any, fieldPath string) (any, error) {
	if data == nil {
		return "", nil
	}

	fieldPath = strings.TrimPrefix(fieldPath, ".")

	// Handle map[string]interface{} case
	if m, ok := data.(map[string]interface{}); ok {
		parts := strings.Split(fieldPath, ".")
		current := m

		// Handle nested fields except the last one
		for i := 0; i < len(parts)-1; i++ {
			if nextMap, ok := current[parts[i]].(map[string]interface{}); ok {
				current = nextMap
			} else if nextMap, ok := current[parts[i]]; ok {
				// Try to continue with whatever we got
				if m, ok := nextMap.(map[string]interface{}); ok {
					current = m
				} else {
					return "", fmt.Errorf("cannot access %s in %s", parts[i+1], parts[i])
				}
			} else {
				return "", nil
			}
		}

		// Access the final field
		if val, ok := current[parts[len(parts)-1]]; ok {
			return val, nil
		}
		return "", nil
	}

	// Handle struct case with reflection
	v := reflect.ValueOf(data)
	parts := strings.Split(fieldPath, ".")

	for _, part := range parts {
		// Dereference pointer if needed
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return "", nil
			}
			v = v.Elem()
		}

		// Handle structs
		if v.Kind() == reflect.Struct {
			v = v.FieldByName(part)
			if !v.IsValid() {
				return "", nil
			}
		} else {
			return "", fmt.Errorf("cannot access %s in non-struct value", part)
		}
	}

	return v.Interface(), nil
}
