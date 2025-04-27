package templates

import (
	"fmt"
	"reflect"
	textTemplates "text/template"
)

func builtins() textTemplates.FuncMap {
	return textTemplates.FuncMap{
		"and": func(booleans ...bool) bool {
			for _, b := range booleans {
				if !b {
					return false
				}
			}
			return true
		},
		"not": func(b bool) bool {
			return !b
		},
		"or": func(booleans ...bool) bool {
			for _, b := range booleans {
				if b {
					return true
				}
			}
			return false
		},
		// "call":  emptyCall,
		// "index": index,
		// "slice": slice,
		"len": func(value any) (int, error) {
			item := reflect.ValueOf(value)

			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}
			if item.Kind() == reflect.Interface {
				item = item.Elem()
			}
			switch item.Kind() {
			case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
				return item.Len(), nil
			}
			return 0, fmt.Errorf("len of type %s", item.Type())
		},
		"print":   fmt.Sprint,
		"printf":  fmt.Sprintf,
		"println": fmt.Sprintln,
		// "urlquery": URLQueryEscaper,
		// "js":       JSEscaper,
		// "html":     HTMLEscaper,

		// Comparisons
		// "eq": eq, // ==
		// "ge": ge, // >=
		// "gt": gt, // >
		// "le": le, // <=
		// "lt": lt, // <
		// "ne": ne, // !=
	}
}
