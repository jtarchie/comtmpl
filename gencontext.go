package main

import (
	"fmt"
	"go/types"
	"io"
	"sort"
	"strings"
)

// Generator drives a single template's typed codegen. It holds the
// running state (VarCounter, current dot type, lexical scope of typed
// $variables) plus the writer and supporting helpers (line directives,
// import collection). The dynamic codepath in main.go does not use this
// struct — it predates the typed work and stays parameter-driven.
type Generator struct {
	Writer       io.Writer
	TemplatePath string
	LineIndex    *LineIndex
	VarCounter   int

	Resolver *TypeResolver
	Imports  *ImportSet
	DataType types.Type
	DotType  types.Type
	DataExpr string // expression that refers to the root data value
	DotExpr  string // expression that refers to the current dot value
	Scopes   []SymbolScope
}

// SymbolScope records the typed bindings for $variables introduced by
// {{range}} or {{with}}. The outermost scope is the function arguments
// (data, $).
type SymbolScope struct {
	Vars map[string]ScopeBinding
}

// ScopeBinding describes one $variable in scope: the Go expression that
// refers to it and the static type of that expression.
type ScopeBinding struct {
	GoExpr string
	Type   types.Type
}

// PushScope opens a new lexical scope. Call Pop when leaving.
func (g *Generator) PushScope() {
	g.Scopes = append(g.Scopes, SymbolScope{Vars: map[string]ScopeBinding{}})
}

// PopScope closes the most-recently-opened scope.
func (g *Generator) PopScope() {
	if len(g.Scopes) == 0 {
		return
	}
	g.Scopes = g.Scopes[:len(g.Scopes)-1]
}

// BindVar records a $variable binding in the innermost scope.
func (g *Generator) BindVar(name string, b ScopeBinding) {
	if len(g.Scopes) == 0 {
		g.PushScope()
	}
	g.Scopes[len(g.Scopes)-1].Vars[name] = b
}

// LookupVar returns the binding for a $variable, walking outward through
// nested scopes. The bool is false if the variable is unbound.
func (g *Generator) LookupVar(name string) (ScopeBinding, bool) {
	for i := len(g.Scopes) - 1; i >= 0; i-- {
		if b, ok := g.Scopes[i].Vars[name]; ok {
			return b, true
		}
	}
	return ScopeBinding{}, false
}

// NextVar returns a new unique-name suffix and increments the counter.
func (g *Generator) NextVar() int {
	v := g.VarCounter
	g.VarCounter++
	return v
}

// EmitLine writes a //line directive for the given byte offset (a
// parse.Pos value) in the source template. It is the typed-mode
// equivalent of emitLineDirective.
func (g *Generator) EmitLine(pos int64) {
	if g.LineIndex == nil || g.TemplatePath == "" {
		return
	}
	_, _ = fmt.Fprintf(g.Writer, "\n//line %s:%d\n", g.TemplatePath, g.LineIndex.LineNumberAt(pos))
}

// Writef writes formatted text to the generator's writer. Errors panic
// (matching the dynamic-path writeString convention).
func (g *Generator) Writef(format string, args ...any) {
	if _, err := fmt.Fprintf(g.Writer, format, args...); err != nil {
		panic(err)
	}
}

// ImportSet tracks the set of Go imports that the generated file needs.
// Aliases are deduplicated so two unrelated packages that suggest the
// same alias get distinct names.
type ImportSet struct {
	byAlias map[string]string // alias -> path
	byPath  map[string]string // path -> alias
}

func NewImportSet() *ImportSet {
	return &ImportSet{
		byAlias: map[string]string{},
		byPath:  map[string]string{},
	}
}

// Add records an import for path, suggesting preferredAlias. Returns the
// alias that was actually used (may differ if preferredAlias is taken).
// If path was already added the existing alias is returned unchanged.
func (s *ImportSet) Add(path, preferredAlias string) string {
	if alias, ok := s.byPath[path]; ok {
		return alias
	}
	alias := preferredAlias
	if alias == "" {
		alias = lastPathSegment(path)
	}
	for i := 1; ; i++ {
		if _, taken := s.byAlias[alias]; !taken {
			break
		}
		alias = fmt.Sprintf("%s%d", preferredAlias, i)
	}
	s.byAlias[alias] = path
	s.byPath[path] = alias
	return alias
}

// WriteImports writes a Go import block listing all collected imports in
// alias-sorted order. If the set is empty, nothing is written.
func (s *ImportSet) WriteImports(w io.Writer) {
	if len(s.byAlias) == 0 {
		return
	}
	aliases := make([]string, 0, len(s.byAlias))
	for a := range s.byAlias {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)
	_, _ = fmt.Fprintln(w, "import (")
	for _, a := range aliases {
		path := s.byAlias[a]
		if a == lastPathSegment(path) {
			_, _ = fmt.Fprintf(w, "\t%q\n", path)
		} else {
			_, _ = fmt.Fprintf(w, "\t%s %q\n", a, path)
		}
	}
	_, _ = fmt.Fprintln(w, ")")
}

func lastPathSegment(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}
