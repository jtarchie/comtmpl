package main

import (
	"bytes"
	"fmt"
	"go/types"
	"html/template"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template/parse"

	"github.com/alecthomas/kong"
	"github.com/go-task/slim-sprig/v3"
)

type CLI struct {
	Filenames   []string `arg:"" help:"Files to process"`
	PackageName string   `help:"Package name" default:"templates"`
}

// GenOptions controls a codegen run. It is the in-process equivalent of CLI flags.
type GenOptions struct {
	Filenames   []string
	PackageName string
	Output      io.Writer
}

func writeString(writer io.Writer, str string) {
	_, err := writer.Write([]byte(str))
	if err != nil {
		panic(err)
	}
}

func (c *CLI) Run() error {
	return Generate(GenOptions{
		Filenames:   c.Filenames,
		PackageName: c.PackageName,
		Output:      os.Stdout,
	})
}

// resolvedTemplate holds the per-template state derived from CLI input
// and template directives. It is populated in the first pass of
// Generate and consumed in the second.
type resolvedTemplate struct {
	Filename     string
	BaseName     string
	TemplatePath string // absolute path; used for //line directives
	Tree         *parse.Tree
	LineIndex    *LineIndex
	Directives   Directives
	DataType     types.Type // nil for dynamic templates
	DataTypeExpr string     // Go expression to refer to DataType
}

// Generate runs codegen for the given templates and writes the result to opts.Output.
func Generate(opts GenOptions) error {
	tmpl, err := template.New("").Funcs(sprig.FuncMap()).ParseFiles(opts.Filenames...)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	imports := NewImportSet()
	imports.Add("io", "")
	imports.Add("fmt", "")
	imports.Add("github.com/jtarchie/comtmpl/templates", "templates")

	resolver := NewTypeResolver()

	resolved := make([]*resolvedTemplate, 0, len(opts.Filenames))
	for _, filename := range opts.Filenames {
		raw, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read %s: %w", filename, err)
		}
		dirs, err := ParseDirectives(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}

		baseFilename := filepath.Base(filename)
		t := tmpl.Lookup(baseFilename)
		if t == nil {
			return fmt.Errorf("template %q not found after parse", baseFilename)
		}
		idx, err := NewLineIndex(filename)
		if err != nil {
			return fmt.Errorf("line index for %s: %w", filename, err)
		}
		absPath, err := filepath.Abs(filename)
		if err != nil {
			absPath = filename
		}

		rt := &resolvedTemplate{
			Filename:     filename,
			BaseName:     t.Name(),
			TemplatePath: absPath,
			Tree:         t.Tree,
			LineIndex:    idx,
			Directives:   dirs,
		}

		if dirs.Typed() {
			pathOrAlias, typeName, err := SplitDataTypeRef(dirs.DataTypeRef)
			if err != nil {
				return fmt.Errorf("%s: %w", filename, err)
			}
			importPath := pathOrAlias
			if alias, ok := dirs.Imports[pathOrAlias]; ok {
				importPath = alias
			}
			typ, err := resolver.ResolveType(importPath, typeName)
			if err != nil {
				return fmt.Errorf("%s: @data %q: %w", filename, dirs.DataTypeRef, err)
			}
			rt.DataType = typ

			// Decide how to refer to the type. Same-package refs use the
			// bare name; cross-package refs add an alias to the import set.
			pkgName := lastPathSegment(importPath)
			if pkgName == opts.PackageName {
				rt.DataTypeExpr = typeName
			} else {
				alias := imports.Add(importPath, pkgName)
				rt.DataTypeExpr = alias + "." + typeName
			}
		}

		resolved = append(resolved, rt)
	}

	writer := opts.Output

	// Emit typed render functions to a side buffer; they are appended
	// after the registry so the file stays readable.
	typedBody := &bytes.Buffer{}
	for _, rt := range resolved {
		if rt.DataType == nil {
			continue
		}
		if err := emitTypedTemplate(typedBody, opts, rt.TemplatePath, rt.BaseName,
			rt.Tree, rt.LineIndex, rt.DataType, rt.DataTypeExpr); err != nil {
			return err
		}
	}

	writeString(writer, fmt.Sprintf("package %s\n\n", opts.PackageName))
	imports.WriteImports(writer)
	writeString(writer, "\nvar Parsed = templates.NewTemplates(map[string]templates.Template{\n")

	for _, rt := range resolved {
		if rt.DataType != nil {
			// Typed template: emit a registry shim that delegates to the
			// typed render function.
			fnName := renderFuncName(rt.BaseName)
			writeString(writer, fmt.Sprintf("\t%q: func(t *templates.Templates, writer io.Writer, data any) error {\n", rt.BaseName))
			writeString(writer, fmt.Sprintf("\t\ttyped, ok := data.(%s)\n", rt.DataTypeExpr))
			writeString(writer, "\t\tif !ok {\n")
			writeString(writer, fmt.Sprintf("\t\t\treturn fmt.Errorf(\"%s: expected %s, got %%T\", data)\n", rt.BaseName, rt.DataTypeExpr))
			writeString(writer, "\t\t}\n")
			writeString(writer, fmt.Sprintf("\t\treturn %s(writer, typed)\n", fnName))
			writeString(writer, "\t},\n")
			continue
		}

		// Dynamic template: existing reflection-based emit.
		writeString(writer, fmt.Sprintf("\t%q: func(t *templates.Templates, writer io.Writer, data any) error {\n\t\tvar err error\n", rt.BaseName))
		varCounter := 0
		processTreeNodes(writer, rt.Tree.Root.Nodes, rt.TemplatePath, rt.LineIndex, &varCounter)
		writeString(writer, "\n\t\treturn nil\n\t},\n")
	}

	writeString(writer, "})\n")

	if typedBody.Len() > 0 {
		writeString(writer, typedBody.String())
	}

	return nil
}

// emitLineDirective writes a Go //line directive pointing at the given
// template position. The directive must start at column 0 (no leading
// whitespace) for the Go compiler to honor it. Subsequent lines of the
// generated file are reported as originating from templatePath until the
// next //line directive.
func emitLineDirective(writer io.Writer, templatePath string, offset *LineIndex, pos int64) {
	if offset == nil || templatePath == "" {
		return
	}
	writeString(writer, fmt.Sprintf("\n//line %s:%d\n", templatePath, offset.LineNumberAt(pos)))
}

// processTreeNodes processes all nodes at the current level
func processTreeNodes(writer io.Writer, nodes []parse.Node, templatePath string, offset *LineIndex, varCounter *int) {
	for _, node := range nodes {
		emitLineDirective(writer, templatePath, offset, int64(node.Position()))

		switch n := node.(type) {
		case *parse.TextNode:
			// Simple text node
			writeString(writer, fmt.Sprintf("\t\t_, err = io.WriteString(writer, %q)\n", string(n.Text)))
			writeString(writer, "\t\tif err != nil { return err }\n")

		case *parse.ActionNode:
			// Simple action node like {{ .Field }} or {{ functionCall }}
			generateActionCode(writer, n, varCounter)

		case *parse.IfNode:
			// If node
			generateIfCode(writer, n, templatePath, offset, varCounter)

		case *parse.RangeNode:
			// Range node
			generateRangeCode(writer, n, templatePath, offset, varCounter)

		case *parse.WithNode:
			// With node
			generateWithCode(writer, n, templatePath, offset, varCounter)

		case *parse.TemplateNode:
			// Template inclusion
			generateTemplateCode(writer, n, varCounter)

		case *parse.CommentNode:
			// Skip comments in templates
			writeString(writer, fmt.Sprintf("\t\t// Template comment: %s\n", strings.ReplaceAll(n.String(), "\n", " ")))
		}
	}
}

// generateActionCode handles {{ .Field }} or {{ functionCall }} expressions
func generateActionCode(writer io.Writer, action *parse.ActionNode, varCounter *int) {
	resultVar := fmt.Sprintf("result%d", *varCounter)
	(*varCounter)++

	writeString(writer, fmt.Sprintf("\t\tvar %s any\n", resultVar))

	// Process the action's pipeline
	if len(action.Pipe.Cmds) > 0 {
		cmd := action.Pipe.Cmds[0]

		if len(cmd.Args) > 0 {
			switch arg := cmd.Args[0].(type) {
			case *parse.FieldNode:
				// Field access like {{ .Field }}
				fields := "[]string{"
				for i, ident := range arg.Ident {
					if i > 0 {
						fields += ", "
					}
					fields += fmt.Sprintf("%q", ident)
				}
				fields += "}"

				writeString(writer, fmt.Sprintf("\t\t%s, err = templates.EvalField(data, %s)\n", resultVar, fields))
				writeString(writer, "\t\tif err != nil { return err }\n")

			case *parse.IdentifierNode:
				// Function call like {{ funcName }}
				funcName := arg.Ident
				writeString(writer, fmt.Sprintf("\t\t%s, err = t.CallFunc(%q", resultVar, funcName))

				// Add arguments if any
				for i := 1; i < len(cmd.Args); i++ {
					// Handle different argument types
					switch argItem := cmd.Args[i].(type) {
					case *parse.FieldNode:
						// Field argument like {{ funcName .Field }}
						argFields := "[]string{"
						for j, ident := range argItem.Ident {
							if j > 0 {
								argFields += ", "
							}
							argFields += fmt.Sprintf("%q", ident)
						}
						argFields += "}"

						argVar := fmt.Sprintf("arg%d_%d", *varCounter, i)
						(*varCounter)++

						writeString(writer, fmt.Sprintf(")\n\t\tvar %s any\n\t\t%s, err = templates.EvalField(data, %s)", argVar, argVar, argFields))
						writeString(writer, fmt.Sprintf("\n\t\tif err != nil { return err }\n\t\t%s, err = t.CallFunc(%q, %s", resultVar, funcName, argVar))

					case *parse.DotNode:
						// Argument is dot itself like {{ funcName . }}
						writeString(writer, ", templates.Dot(data)")

					case *parse.VariableNode:
						// Variable reference like {{ funcName $var }}
						varName := sanitizeVarName(argItem.Ident[0])
						writeString(writer, fmt.Sprintf(", %s", varName))

					default:
						// Fallback for unsupported argument types
						writeString(writer, ", nil")
					}
				}

				writeString(writer, ")\n")
				writeString(writer, "\t\tif err != nil { return err }\n")

			case *parse.DotNode:
				// {{ . }} itself
				writeString(writer, fmt.Sprintf("\t\t%s = templates.Dot(data)\n", resultVar))

			case *parse.VariableNode:
				// {{ $var }} or {{ $var.Field }} variable reference
				emitVariableRef(writer, resultVar, arg.Ident)

			default:
				writeString(writer, fmt.Sprintf("\t\t%s = nil // Unsupported node type: %T\n", resultVar, arg))
			}
		}

		// Handle pipes (simplified for now)
		for i := 1; i < len(action.Pipe.Cmds); i++ {
			if len(action.Pipe.Cmds[i].Args) > 0 {
				if ident, ok := action.Pipe.Cmds[i].Args[0].(*parse.IdentifierNode); ok {
					funcName := ident.Ident
					writeString(writer, fmt.Sprintf("\t\t%s, err = t.CallFunc(%q, %s", resultVar, funcName, resultVar))

					// Add arguments if any
					for j := 1; j < len(action.Pipe.Cmds[i].Args); j++ {
						switch argItem := action.Pipe.Cmds[i].Args[j].(type) {
						case *parse.FieldNode:
							// Field argument like {{ .Field | funcName .OtherField }}
							argFields := "[]string{"
							for k, ident := range argItem.Ident {
								if k > 0 {
									argFields += ", "
								}
								argFields += fmt.Sprintf("%q", ident)
							}
							argFields += "}"

							argVar := fmt.Sprintf("arg%d_%d", *varCounter, j)
							(*varCounter)++

							writeString(writer, fmt.Sprintf(")\n\t\tvar %s any\n\t\t%s, err = templates.EvalField(data, %s)", argVar, argVar, argFields))
							writeString(writer, fmt.Sprintf("\n\t\tif err != nil { return err }\n\t\t%s, err = t.CallFunc(%q, %s", resultVar, funcName, resultVar))
							writeString(writer, fmt.Sprintf(", %s", argVar))

						case *parse.DotNode:
							// Dot argument like {{ .Field | funcName . }}
							writeString(writer, ", templates.Dot(data)")

						case *parse.VariableNode:
							// Variable reference like {{ .Field | funcName $var }}
							varName := sanitizeVarName(argItem.Ident[0])
							writeString(writer, fmt.Sprintf(", %s", varName))

						default:
							// Fallback for unsupported argument types
							writeString(writer, ", nil")
						}
					}

					writeString(writer, ")\n")
					writeString(writer, "\t\tif err != nil { return err }\n")
				}
			}
		}
	}

	// Output the result
	writeString(writer, fmt.Sprintf("\t\t_, err = fmt.Fprint(writer, %s)\n", resultVar))
	writeString(writer, "\t\tif err != nil { return err }\n")
}

// generateIfCode handles if/else statements
func generateIfCode(writer io.Writer, ifNode *parse.IfNode, templatePath string, offset *LineIndex, varCounter *int) {
	condVar := fmt.Sprintf("cond%d", *varCounter)
	(*varCounter)++

	writeString(writer, "\t\t// If statement\n")
	writeString(writer, fmt.Sprintf("\t\tvar %s bool\n", condVar))

	// Evaluate the condition (simplified)
	if len(ifNode.Pipe.Cmds) > 0 {
		resultVar := fmt.Sprintf("ifResult%d", *varCounter)
		(*varCounter)++

		writeString(writer, fmt.Sprintf("\t\tvar %s any\n", resultVar))

		cmd := ifNode.Pipe.Cmds[0]
		if len(cmd.Args) > 0 {
			switch arg := cmd.Args[0].(type) {
			case *parse.FieldNode:
				// Field access like {{ if .Field }}
				fields := "[]string{"
				for i := 0; i < len(arg.Ident); i++ {
					if i > 0 {
						fields += ", "
					}
					fields += fmt.Sprintf("%q", arg.Ident[i])
				}
				fields += "}"

				writeString(writer, fmt.Sprintf("\t\t%s, err = templates.EvalField(data, %s)\n", resultVar, fields))
				writeString(writer, "\t\tif err != nil { return err }\n")

			case *parse.IdentifierNode:
				// Function call like {{ if funcName }}
				funcName := arg.Ident
				writeString(writer, fmt.Sprintf("\t\t%s, err = t.CallFunc(%q)\n", resultVar, funcName))
				writeString(writer, "\t\tif err != nil { return err }\n")

			case *parse.DotNode:
				// {{ if . }}
				writeString(writer, fmt.Sprintf("\t\t%s = templates.Dot(data)\n", resultVar))

			case *parse.VariableNode:
				// Variable reference like {{ if $x }} or {{ if $x.Field }}
				emitVariableRef(writer, resultVar, arg.Ident)

			default:
				writeString(writer, fmt.Sprintf("\t\t%s = nil // Unsupported node type: %T\n", resultVar, arg))
			}
		}

		// Convert to boolean
		writeString(writer, fmt.Sprintf("\t\t%s, err = templates.IsTrue(%s)\n", condVar, resultVar))
		writeString(writer, "\t\tif err != nil { return err }\n")
	}

	// Generate if block
	writeString(writer, fmt.Sprintf("\t\tif %s {\n", condVar))

	// Process the if body with one more level of indentation
	if ifNode.List != nil {
		// Process all nodes in the if body recursively
		for _, node := range ifNode.List.Nodes {
			processNodeWithIndent(writer, node, templatePath, offset, varCounter, 1)
		}
	}

	// Process the else block if it exists
	if ifNode.ElseList != nil {
		writeString(writer, "\t\t} else {\n")

		// Process all nodes in the else body recursively
		for _, node := range ifNode.ElseList.Nodes {
			processNodeWithIndent(writer, node, templatePath, offset, varCounter, 1)
		}
	}

	writeString(writer, "\t\t}\n")
}

// processNodeWithIndent processes a single node with additional indentation
func processNodeWithIndent(writer io.Writer, node parse.Node, templatePath string, offset *LineIndex, varCounter *int, indentLevel int) {
	indent := strings.Repeat("\t", indentLevel+2) // Base indent (2) + additional levels

	emitLineDirective(writer, templatePath, offset, int64(node.Position()))

	switch n := node.(type) {
	case *parse.TextNode:
		// Simple text node
		writeString(writer, fmt.Sprintf("%s_, err = io.WriteString(writer, %q)\n", indent, string(n.Text)))
		writeString(writer, fmt.Sprintf("%sif err != nil { return err }\n", indent))

	case *parse.ActionNode:
		// Action node - adjust indentation of generated code
		origOutput := strings.Builder{}
		generateActionCode(&origOutput, n, varCounter)
		indented := strings.ReplaceAll(origOutput.String(), "\t\t", indent)
		writeString(writer, indented)

	case *parse.IfNode:
		// Nested if node
		origOutput := strings.Builder{}
		generateIfCode(&origOutput, n, templatePath, offset, varCounter)
		indented := strings.ReplaceAll(origOutput.String(), "\t\t", indent)
		writeString(writer, indented)

	case *parse.RangeNode:
		// Nested range node
		origOutput := strings.Builder{}
		generateRangeCode(&origOutput, n, templatePath, offset, varCounter)
		indented := strings.ReplaceAll(origOutput.String(), "\t\t", indent)
		writeString(writer, indented)

	case *parse.WithNode:
		// Nested with node
		origOutput := strings.Builder{}
		generateWithCode(&origOutput, n, templatePath, offset, varCounter)
		indented := strings.ReplaceAll(origOutput.String(), "\t\t", indent)
		writeString(writer, indented)

	case *parse.TemplateNode:
		// Template inclusion
		origOutput := strings.Builder{}
		generateTemplateCode(&origOutput, n, varCounter)
		indented := strings.ReplaceAll(origOutput.String(), "\t\t", indent)
		writeString(writer, indented)

	case *parse.CommentNode:
		// Skip comments in templates
		writeString(writer, fmt.Sprintf("%s// Template comment: %s\n", indent,
			strings.ReplaceAll(n.String(), "\n", " ")))

	default:
		// Handle any other node types
		writeString(writer, fmt.Sprintf("%s// Unsupported nested node type: %T\n", indent, n))
	}
}

// generateRangeCode handles range loops
func generateRangeCode(writer io.Writer, rangeNode *parse.RangeNode, templatePath string, offset *LineIndex, varCounter *int) {
	rangeVar := fmt.Sprintf("rangeData%d", *varCounter)
	(*varCounter)++

	writeString(writer, "\t\t// Range statement\n")
	writeString(writer, fmt.Sprintf("\t\tvar %s any\n", rangeVar))

	// Get the range data
	if len(rangeNode.Pipe.Cmds) > 0 {
		cmd := rangeNode.Pipe.Cmds[0]

		if len(cmd.Args) > 0 {
			switch arg := cmd.Args[0].(type) {
			case *parse.FieldNode:
				// Field access like {{ range .Items }}
				fields := "[]string{"
				for i := 0; i < len(arg.Ident); i++ {
					if i > 0 {
						fields += ", "
					}
					fields += fmt.Sprintf("%q", arg.Ident[i])
				}
				fields += "}"

				writeString(writer, fmt.Sprintf("\t\t%s, err = templates.EvalField(data, %s)\n", rangeVar, fields))
				writeString(writer, "\t\tif err != nil { return err }\n")

			case *parse.DotNode:
				// {{ range . }}
				writeString(writer, fmt.Sprintf("\t\t%s = templates.Dot(data)\n", rangeVar))

			case *parse.VariableNode:
				// {{ range $var }} or {{ range $var.Field }}
				emitVariableRef(writer, rangeVar, arg.Ident)

			default:
				writeString(writer, fmt.Sprintf("\t\t%s = nil // Unsupported node type: %T\n", rangeVar, arg))
			}
		}
	}

	// Create iterable and range loop
	iterVar := fmt.Sprintf("iter%d", *varCounter)
	(*varCounter)++

	writeString(writer, fmt.Sprintf("\t\t%s, err := templates.GetIterable(%s)\n", iterVar, rangeVar))
	writeString(writer, "\t\tif err != nil { return err }\n")
	hasElse := rangeNode.ElseList != nil
	if hasElse {
		writeString(writer, "\t\thasItems := false\n")
	}

	// Handle variable declarations in range
	var indexVarName, valueVarName string
	if len(rangeNode.Pipe.Decl) >= 1 {
		// Create a safe variable name from the template variable
		origIndexVar := rangeNode.Pipe.Decl[0].Ident[0]
		indexVarName = sanitizeVarName(origIndexVar)
		writeString(writer, fmt.Sprintf("\t\tvar %s any // Template variable for index: %s\n", indexVarName, origIndexVar))
	}
	if len(rangeNode.Pipe.Decl) >= 2 {
		// Create a safe variable name from the template variable
		origValueVar := rangeNode.Pipe.Decl[1].Ident[0]
		valueVarName = sanitizeVarName(origValueVar)
		writeString(writer, fmt.Sprintf("\t\tvar %s any // Template variable for value: %s\n", valueVarName, origValueVar))
	}
	(*varCounter)++

	// Generate code for map iteration
	writeString(writer, fmt.Sprintf(`
		if mapData, isMap := %s.(map[string]any); isMap {
			for k, v := range mapData {`, iterVar))
	if hasElse {
		writeString(writer, "\n\t\t\t\thasItems = true")
	}

	// Assign to template variables if they exist
	if indexVarName != "" {
		writeString(writer, fmt.Sprintf("\n\t\t\t\t%s = k", indexVarName))
	}
	if valueVarName != "" {
		writeString(writer, fmt.Sprintf("\n\t\t\t\t%s = v", valueVarName))
	}

	// Setup range context
	writeString(writer, `

				// Create range scope
				rangeContext := templates.NewRangeScope(data, k, v)
				oldData := data
				data = rangeContext
`)

	// Process range body with proper node handling
	writeString(writer, "\t\t\t\t// Range body\n")
	if rangeNode.List != nil && rangeNode.List.Nodes != nil {
		for _, node := range rangeNode.List.Nodes {
			processNodeWithIndent(writer, node, templatePath, offset, varCounter, 2)
		}
	}

	// Restore original context
	writeString(writer, "\t\t\t\tdata = oldData\n")
	writeString(writer, "\t\t\t}\n")

	// Generate code for slice iteration
	writeString(writer, fmt.Sprintf(`
		} else if sliceData, isSlice := %s.([]any); isSlice {
			for i, v := range sliceData {`, iterVar))
	if hasElse {
		writeString(writer, "\n\t\t\t\thasItems = true")
	}

	// Assign to template variables if they exist
	if indexVarName != "" {
		writeString(writer, fmt.Sprintf("\n\t\t\t\t%s = i", indexVarName))
	}
	if valueVarName != "" {
		writeString(writer, fmt.Sprintf("\n\t\t\t\t%s = v", valueVarName))
	}

	// Setup range context
	writeString(writer, `

				// Create range scope
				rangeContext := templates.NewRangeScope(data, i, v)
				oldData := data
				data = rangeContext
`)

	// Process range body again for slice iteration with proper node handling
	writeString(writer, "\t\t\t\t// Range body\n")
	if rangeNode.List != nil && rangeNode.List.Nodes != nil {
		for _, node := range rangeNode.List.Nodes {
			processNodeWithIndent(writer, node, templatePath, offset, varCounter, 2)
		}
	}

	// Restore original context
	writeString(writer, "\t\t\t\tdata = oldData\n")
	writeString(writer, "\t\t\t}\n")
	writeString(writer, "\t\t}\n")

	// Handle range else clause if present
	if hasElse {
		writeString(writer, "\t\tif !hasItems {\n")

		for _, node := range rangeNode.ElseList.Nodes {
			processNodeWithIndent(writer, node, templatePath, offset, varCounter, 1)
		}

		writeString(writer, "\t\t}\n")
	}
}

// generateWithCode handles with blocks
func generateWithCode(writer io.Writer, withNode *parse.WithNode, templatePath string, offset *LineIndex, varCounter *int) {
	withVar := fmt.Sprintf("withData%d", *varCounter)
	(*varCounter)++

	writeString(writer, "\t\t// With statement\n")
	writeString(writer, fmt.Sprintf("\t\tvar %s any\n", withVar))

	// Get the with value
	if len(withNode.Pipe.Cmds) > 0 {
		cmd := withNode.Pipe.Cmds[0]

		if len(cmd.Args) > 0 {
			switch arg := cmd.Args[0].(type) {
			case *parse.FieldNode:
				// Field access like {{ with .Field }}
				fields := "[]string{"
				for i := 0; i < len(arg.Ident); i++ {
					if i > 0 {
						fields += ", "
					}
					fields += fmt.Sprintf("%q", arg.Ident[i])
				}
				fields += "}"

				writeString(writer, fmt.Sprintf("\t\t%s, err = templates.EvalField(data, %s)\n", withVar, fields))
				writeString(writer, "\t\tif err != nil { return err }\n")

			case *parse.DotNode:
				// {{ with . }}
				writeString(writer, fmt.Sprintf("\t\t%s = templates.Dot(data)\n", withVar))

			case *parse.VariableNode:
				// {{ with $var }} or {{ with $var.Field }}
				emitVariableRef(writer, withVar, arg.Ident)

			default:
				writeString(writer, fmt.Sprintf("\t\t%s = nil // Unsupported node type: %T\n", withVar, arg))
			}
		}
	}

	// Check if with value is truthy
	condVar := fmt.Sprintf("withCond%d", *varCounter)
	(*varCounter)++

	writeString(writer, fmt.Sprintf("\t\t%s, err := templates.IsTrue(%s)\n", condVar, withVar))
	writeString(writer, "\t\tif err != nil { return err }\n")

	// Generate with block
	writeString(writer, fmt.Sprintf("\t\tif %s {\n", condVar))
	writeString(writer, "\t\t\t// Save old data context and set new one\n")
	writeString(writer, "\t\t\toldData := data\n")
	writeString(writer, fmt.Sprintf("\t\t\tdata = %s\n", withVar))

	// Process the with body with proper node handling
	if withNode.List != nil {
		for _, node := range withNode.List.Nodes {
			processNodeWithIndent(writer, node, templatePath, offset, varCounter, 1)
		}
	}

	// Restore data context
	writeString(writer, "\t\t\tdata = oldData\n")

	// Process the else block if it exists
	if withNode.ElseList != nil {
		writeString(writer, "\t\t} else {\n")

		for _, node := range withNode.ElseList.Nodes {
			processNodeWithIndent(writer, node, templatePath, offset, varCounter, 1)
		}
	}

	writeString(writer, "\t\t}\n")
}

// generateTemplateCode handles template inclusion
func generateTemplateCode(writer io.Writer, tmplNode *parse.TemplateNode, varCounter *int) {
	dataVar := fmt.Sprintf("tmplData%d", *varCounter)
	(*varCounter)++

	writeString(writer, fmt.Sprintf("\t\t// Include template: %s\n", tmplNode.Name))

	// Set up data for the template
	if tmplNode.Pipe != nil && len(tmplNode.Pipe.Cmds) > 0 {
		writeString(writer, fmt.Sprintf("\t\tvar %s any\n", dataVar))

		cmd := tmplNode.Pipe.Cmds[0]
		if len(cmd.Args) > 0 {
			switch arg := cmd.Args[0].(type) {
			case *parse.FieldNode:
				// Field access like {{ template "name" .Field }}
				fields := "[]string{"
				for i, ident := range arg.Ident {
					if i > 0 {
						fields += ", "
					}
					fields += fmt.Sprintf("%q", ident)
				}
				fields += "}"

				writeString(writer, fmt.Sprintf("\t\t%s, err = templates.EvalField(data, %s)\n", dataVar, fields))
				writeString(writer, "\t\tif err != nil { return err }\n")

			case *parse.DotNode:
				// {{ template "name" . }}
				writeString(writer, fmt.Sprintf("\t\t%s = templates.Dot(data)\n", dataVar))

			default:
				writeString(writer, fmt.Sprintf("\t\t%s = nil // Unsupported node type: %T\n", dataVar, arg))
			}
		}
	} else {
		// No data specified, use nil
		writeString(writer, fmt.Sprintf("\t\t%s := data\n", dataVar))
	}

	// Execute the template
	writeString(writer, fmt.Sprintf("\t\terr = t.ExecuteTemplate(writer, %q, %s)\n", tmplNode.Name, dataVar))
	writeString(writer, "\t\tif err != nil { return err }\n")
}

// Helper function to sanitize variable names (convert $var to valid Go identifier)
func sanitizeVarName(varName string) string {
	// Remove $ prefix for template variables and make a valid Go identifier
	if strings.HasPrefix(varName, "$") {
		return "var_" + varName[1:]
	}
	return varName
}

// emitVariableRef writes code that assigns a template variable (with optional
// field path, e.g. $item.Name.First) into the destination variable.
func emitVariableRef(writer io.Writer, dest string, idents []string) {
	varName := sanitizeVarName(idents[0])
	if len(idents) == 1 {
		writeString(writer, fmt.Sprintf("\t\t%s = %s // Variable reference\n", dest, varName))
		return
	}

	fields := "[]string{"
	for i, ident := range idents[1:] {
		if i > 0 {
			fields += ", "
		}
		fields += fmt.Sprintf("%q", ident)
	}
	fields += "}"

	writeString(writer, fmt.Sprintf("\t\t%s, err = templates.EvalField(%s, %s)\n", dest, varName, fields))
	writeString(writer, "\t\tif err != nil { return err }\n")
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cli := &CLI{}
	ctx := kong.Parse(cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
