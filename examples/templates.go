package templates

import (
	"fmt"
	"io"
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
    <title>Example Page</title>
  </head>
  <body>
    <h1>Example Page</h1>
  </body>
</html>`)
		if err != nil {
			return err
		}

		return nil
	},
}
