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
)

type CLI struct {
	Filenames []string `arg:"" help:"Files to process"`
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

	writeString(writer, `
	package templates

	import (
		"io"
		"fmt"
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
	`)

	for _, filename := range c.Filenames {
		baseFilename := filepath.Base(filename)
		template := templates.Lookup(baseFilename)
		offset, err := NewLineIndex(filename)
		if err != nil {
			return fmt.Errorf("failed to create line index: %w", err)
		}

		writeString(writer, fmt.Sprintf("\t%q: func(writer io.Writer, data any) error {\nvar err error\n", template.Name()))
		for _, node := range template.Tree.Root.Nodes {
			writeString(writer, fmt.Sprintf("//%s:%d\n", template.Name(), offset.LineNumberAt(int64(node.Position()))))
			switch typed := node.(type) {
			case *parse.TextNode:
				writeString(writer, "_, err = ")
				writeString(writer, "io.WriteString(writer, `")
				writeString(writer, string(typed.Text))
				writeString(writer, "`)\n")
				writeString(writer, "if err != nil {\nreturn err\n}\n")
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

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cli := &CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
