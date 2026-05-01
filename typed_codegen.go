package main

import (
	"fmt"
	"go/types"
	"io"
	"strings"
	"text/template/parse"
)

// renderFuncName returns the Go identifier for the typed render function
// of templateName. It strips the .html suffix and title-cases the rest:
// "index.html" -> "RenderIndex", "user-profile.html" -> "RenderUserProfile".
func renderFuncName(templateName string) string {
	name := templateName
	if idx := strings.LastIndexByte(name, '.'); idx > 0 {
		name = name[:idx]
	}
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == '.' || r == ' ' || r == '/'
	})
	var b strings.Builder
	b.WriteString("Render")
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

// emitTypedTemplate writes the typed function for a single template into
// the output writer. The function signature is:
//
//	func Render<Name>(w io.Writer, data <DataType>) error
//
// In addition, the caller emits a registry shim that type-asserts `any`
// to the static type and forwards to this function so that
// Parsed.ExecuteTemplate keeps working.
func emitTypedTemplate(out io.Writer, opts GenOptions, templatePath, templateName string,
	tree *parse.Tree, lineIdx *LineIndex, dataType types.Type, dataTypeExpr string) error {

	g := &Generator{
		Writer:       out,
		TemplatePath: templatePath,
		LineIndex:    lineIdx,
		DataType:     dataType,
		DotType:      dataType,
		DataExpr:     "data",
		DotExpr:      "data",
	}

	fnName := renderFuncName(templateName)
	_, _ = fmt.Fprintf(out, "\nfunc %s(writer io.Writer, data %s) error {\n\tvar err error\n", fnName, dataTypeExpr)

	for _, node := range tree.Root.Nodes {
		if err := g.emitNode(node); err != nil {
			return fmt.Errorf("%s: %w", templateName, err)
		}
	}

	_, _ = fmt.Fprintf(out, "\n\treturn nil\n}\n")
	return nil
}

// emitNode dispatches to the right typed-emitter for a parse.Node.
// Unsupported node types return a clear error so users know their
// template can't be compiled in typed mode yet.
func (g *Generator) emitNode(node parse.Node) error {
	g.EmitLine(int64(node.Position()))

	switch n := node.(type) {
	case *parse.TextNode:
		return g.emitTextNode(n)
	case *parse.ActionNode:
		return g.emitActionNode(n)
	case *parse.CommentNode:
		// Comments are no-ops at runtime.
		return nil
	default:
		return fmt.Errorf("typed mode does not yet support %T (line %d)", n,
			lineNumberFor(g.LineIndex, int64(node.Position())))
	}
}

func (g *Generator) emitTextNode(n *parse.TextNode) error {
	g.Writef("\t_, err = io.WriteString(writer, %q)\n", string(n.Text))
	g.Writef("\tif err != nil { return err }\n")
	return nil
}

// emitActionNode handles {{.Field}}, {{.Foo.Bar}}, and {{.}} in typed mode.
// Function calls and pipelines are not yet supported in this MVP.
func (g *Generator) emitActionNode(n *parse.ActionNode) error {
	if n.Pipe == nil || len(n.Pipe.Cmds) == 0 {
		return nil
	}
	if len(n.Pipe.Cmds) > 1 {
		return fmt.Errorf("typed mode does not yet support pipelines (line %d)",
			lineNumberFor(g.LineIndex, int64(n.Position())))
	}
	cmd := n.Pipe.Cmds[0]
	if len(cmd.Args) == 0 {
		return nil
	}

	expr, _, err := g.evalCommandArg(cmd.Args[0])
	if err != nil {
		return err
	}

	g.Writef("\t_, err = fmt.Fprint(writer, %s)\n", expr)
	g.Writef("\tif err != nil { return err }\n")
	return nil
}

// evalCommandArg returns the Go expression and static type for a single
// argument of an action command. Supports FieldNode, DotNode, and
// VariableNode chains. Anything else is rejected.
func (g *Generator) evalCommandArg(arg parse.Node) (expr string, typ types.Type, err error) {
	switch a := arg.(type) {
	case *parse.DotNode:
		return g.DotExpr, g.DotType, nil

	case *parse.FieldNode:
		return g.fieldExpr(g.DotExpr, g.DotType, a.Ident, int64(a.Position()))

	case *parse.VariableNode:
		if len(a.Ident) == 0 {
			return "", nil, fmt.Errorf("empty variable reference")
		}
		bind, ok := g.LookupVar(a.Ident[0])
		if !ok {
			return "", nil, fmt.Errorf("unbound variable %q (line %d)",
				a.Ident[0], lineNumberFor(g.LineIndex, int64(a.Position())))
		}
		if len(a.Ident) == 1 {
			return bind.GoExpr, bind.Type, nil
		}
		return g.fieldExpr(bind.GoExpr, bind.Type, a.Ident[1:], int64(a.Position()))

	case *parse.StringNode:
		return a.Quoted, types.Typ[types.String], nil

	default:
		return "", nil, fmt.Errorf("typed mode does not yet support %T as command arg", a)
	}
}

// fieldExpr resolves a chain of identifiers (e.g. ["User", "Name"])
// against a base expression and its type. Each step navigates either:
//   - a struct field (emits ".Ident")
//   - a map[string]X key (emits "[\"Ident\"]")
//   - a method call on a named type with arity 0 (emits ".Ident()")
//
// It dereferences pointers as needed, like Go selector syntax.
func (g *Generator) fieldExpr(baseExpr string, baseType types.Type, idents []string, pos int64) (string, types.Type, error) {
	expr := baseExpr
	currentType := baseType

	for _, ident := range idents {
		next, nextType, err := g.stepField(expr, currentType, ident)
		if err != nil {
			return "", nil, fmt.Errorf("field path %s: %w (line %d)",
				strings.Join(idents, "."), err, lineNumberFor(g.LineIndex, pos))
		}
		expr = next
		currentType = nextType
	}
	return expr, currentType, nil
}

// stepField navigates a single identifier from a value of the given
// type. Returns the new Go expression and the resulting type.
func (g *Generator) stepField(baseExpr string, baseType types.Type, ident string) (string, types.Type, error) {
	t := baseType
	// Dereference pointers transparently (Go selector handles it; we
	// just need to reason about the underlying type).
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Method lookup happens against the original type (with or without
	// pointer); types.LookupFieldOrMethod handles both addressable
	// receivers and value receivers.
	if obj, _, _ := types.LookupFieldOrMethod(baseType, true, nil, ident); obj != nil {
		switch o := obj.(type) {
		case *types.Var: // struct field
			return baseExpr + "." + ident, o.Type(), nil
		case *types.Func: // method
			sig, ok := o.Type().(*types.Signature)
			if !ok {
				return "", nil, fmt.Errorf("method %s has unexpected type %T", ident, o.Type())
			}
			if sig.Params().Len() != 0 {
				return "", nil, fmt.Errorf("method %s takes %d args; only zero-arg methods are supported in field paths",
					ident, sig.Params().Len())
			}
			if sig.Results().Len() == 0 {
				return "", nil, fmt.Errorf("method %s returns no values", ident)
			}
			return baseExpr + "." + ident + "()", sig.Results().At(0).Type(), nil
		}
	}

	// Map lookup: m["key"] for map[string]V
	if m, ok := t.Underlying().(*types.Map); ok {
		if basic, ok := m.Key().Underlying().(*types.Basic); ok && basic.Kind() == types.String {
			return fmt.Sprintf("%s[%q]", baseExpr, ident), m.Elem(), nil
		}
		return "", nil, fmt.Errorf("map key type must be string, got %s", m.Key())
	}

	return "", nil, fmt.Errorf("type %s has no field, method, or string-key map entry %q", baseType, ident)
}

// lineNumberFor is a small helper that returns 0 if the index is nil so
// error messages can include line info without panicking on tests that
// pass a nil LineIndex.
func lineNumberFor(idx *LineIndex, pos int64) int {
	if idx == nil {
		return 0
	}
	return idx.LineNumberAt(pos)
}
