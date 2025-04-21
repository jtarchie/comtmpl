package main

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template/parse"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Filenames   []string `arg:"" help:"Files to process"`
	PackageName string   `help:"Package name" default:"templates"`
}

func writeString(writer io.Writer, str string) {
	_, err := writer.Write([]byte(str))
	if err != nil {
		panic(err)
	}
}

func (c *CLI) Run() error {
	templates, err := template.ParseFiles(c.Filenames...)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	writer := os.Stdout

	writeString(writer, fmt.Sprintf(`
	package %s

	import (
		"io"
		"fmt"
		"github.com/jtarchie/comtmpl/templates"
	)

	var Parsed = templates.Templates{
	`, c.PackageName))

	for _, filename := range c.Filenames {
		baseFilename := filepath.Base(filename)
		template := templates.Lookup(baseFilename)
		offset, err := NewLineIndex(filename)
		if err != nil {
			return fmt.Errorf("failed to create line index: %w", err)
		}

		writeString(writer, fmt.Sprintf("\t%q: func(writer io.Writer, data any) error {\nvar err error\n", template.Name()))

		// Counter for unique variable names
		varCounter := 0

		for _, node := range template.Tree.Root.Nodes {
			writeString(writer, fmt.Sprintf("//%s:%d\n", template.Name(), offset.LineNumberAt(int64(node.Position()))))
			switch typed := node.(type) {
			case *parse.TextNode:
				writeString(writer, "_, err = ")
				writeString(writer, "io.WriteString(writer, `")
				writeString(writer, string(typed.Text))
				writeString(writer, "`)\n")
				writeString(writer, "if err != nil {\nreturn err\n}\n")
			case *parse.ActionNode:
				pipe := typed.Pipe
				if len(pipe.Cmds) == 1 {
					cmd := pipe.Cmds[0]
					if len(cmd.Args) == 1 {
						// Handle field access like {{.Name}} or {{.User.Name}}
						if fieldNode, ok := cmd.Args[0].(*parse.FieldNode); ok {
							// Generate the field access path (e.g., ".Name" or ".User.Name")
							fieldPath := fieldPathToString(fieldNode.Ident)

							// Use a unique variable name for each field access
							valueVar := fmt.Sprintf("value%d", varCounter)
							varCounter++

							// Write the field accessor code
							writeString(writer, `
// Handle {{`+fieldPath+`}}
var `+valueVar+` any
`+valueVar+`, err = templates.EvalField(data, "`+fieldPath+`")
if err != nil {
	return err
}
_, err = fmt.Fprint(writer, `+valueVar+`)
if err != nil {
	return err
}
`)
						}
					}
				}
			}
		}

		writeString(writer, `
		return nil
		},
		`)
	}

	writeString(writer, `
	}
	`)

	return nil
}

// Helper function to convert field path to string
func fieldPathToString(ident []string) string {
	return "." + strings.Join(ident, ".")
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cli := &CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
