package main

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"text/template/parse"

	"github.com/alecthomas/kong"
	"github.com/go-task/slim-sprig/v3"
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
	tmpl, err := template.New("").Funcs(sprig.FuncMap()).ParseFiles(c.Filenames...)
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

	var Parsed = templates.NewTemplates(map[string]templates.Template{
	`, c.PackageName))

	for _, filename := range c.Filenames {
		baseFilename := filepath.Base(filename)
		template := tmpl.Lookup(baseFilename)
		offset, err := NewLineIndex(filename)
		if err != nil {
			return fmt.Errorf("failed to create line index: %w", err)
		}

		writeString(writer, fmt.Sprintf("\t%q: func(t *templates.Templates, writer io.Writer, data any) error {\nvar err error\n", template.Name()))

		// Counter for unique variable names
		varCounter := 0

		for _, node := range template.Tree.Root.Nodes {
			writeString(writer, fmt.Sprintf("\n\n//%s:%d\n\n", template.Name(), offset.LineNumberAt(int64(node.Position()))))
			switch typed := node.(type) {
			case *parse.TextNode:
				writeString(writer, "_, err = ")
				writeString(writer, "io.WriteString(writer, `")
				writeString(writer, string(typed.Text))
				writeString(writer, "`)\n")
				writeString(writer, "if err != nil {\nreturn err\n}\n")
			case *parse.ActionNode:
				pipe := typed.Pipe

				// Get a string representation of the pipe for comments
				pipeStr := typed.String()

				// Variable to store value from previous pipe command
				pipeValueVar := fmt.Sprintf("pipeValue%d", varCounter)
				varCounter++

				writeString(writer, fmt.Sprintf("// Handle %s\n", pipeStr))
				writeString(writer, fmt.Sprintf("var %s any\n", pipeValueVar))

				// Process each command in the pipe
				for i, cmd := range pipe.Cmds {
					if i == 0 {
						// First command - could be a field access or identifier
						if len(cmd.Args) == 1 {
							// Field access like {{.Name}} or {{.User.Name}}
							if fieldNode, ok := cmd.Args[0].(*parse.FieldNode); ok {
								// Generate the field access path
								fieldPathAsSlice := "[]string{"
								for j, ident := range fieldNode.Ident {
									if j > 0 {
										fieldPathAsSlice += ", "
									}
									fieldPathAsSlice += fmt.Sprintf("%q", ident)
								}
								fieldPathAsSlice += "}"

								writeString(writer, fmt.Sprintf(`%s, err = templates.EvalField(data, %s)
if err != nil {
	return err
}
`, pipeValueVar, fieldPathAsSlice))
							} else if _, ok := cmd.Args[0].(*parse.IdentifierNode); ok {
								// This is just a function with no args like {{len}}
								// It will be handled in the next pipe section if it exists
								writeString(writer, fmt.Sprintf(`%s = nil
`, pipeValueVar))
							}
						} else {
							// Handle more complex first command
							writeString(writer, fmt.Sprintf(`// TODO: Handle more complex first command
%s = nil
`, pipeValueVar))
						}
					} else {
						// Subsequent commands - function calls that take the previous value
						if len(cmd.Args) > 0 {
							// First argument should be function name
							if identNode, ok := cmd.Args[0].(*parse.IdentifierNode); ok {
								funcName := identNode.Ident

								// Check if there are additional arguments for the function
								var additionalArgs string
								if len(cmd.Args) > 1 {
									// Process additional arguments
									for argIdx := 1; argIdx < len(cmd.Args); argIdx++ {
										arg := cmd.Args[argIdx]

										// Handle different argument types
										if literalNode, ok := arg.(*parse.NumberNode); ok {
											// Number argument like truncate 10
											additionalArgs += fmt.Sprintf(", %s", literalNode.Text)
										} else if stringNode, ok := arg.(*parse.StringNode); ok {
											// String argument
											additionalArgs += fmt.Sprintf(", %q", stringNode.Text)
										} else if boolNode, ok := arg.(*parse.BoolNode); ok {
											// Boolean argument
											additionalArgs += fmt.Sprintf(", %t", boolNode.True)
										} else {
											// Other argument types - add placeholder
											additionalArgs += ", nil /*unsupported arg type*/"
										}
									}
								}

								writeString(writer, fmt.Sprintf(`// Pipe to function %s
%s, err = t.CallFunc(%q, %s%s)
if err != nil {
	return err
}
`, funcName, pipeValueVar, funcName, pipeValueVar, additionalArgs))
							}
						}
					}
				}

				// Final output of pipe result
				writeString(writer, fmt.Sprintf(`_, err = fmt.Fprint(writer, %s)
if err != nil {
	return err
}
`, pipeValueVar))
			}
		}

		writeString(writer, `
		return nil
		},
		`)
	}

	writeString(writer, `
})
	`)

	return nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cli := &CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
