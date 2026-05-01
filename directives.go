package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Directives are opt-in metadata declared by a template via Go template
// comments: {{/* @data ... */}}, {{/* @funcs ... */}}, {{/* @import ... */}}.
// They are recognized by a pre-scan over the raw template source (the
// html/template parser may strip some comments before we see them).
type Directives struct {
	// DataTypeRef is the value of {{/* @data <ref> */}} where <ref> is
	// either a fully-qualified path like "github.com/foo/bar.MyType" or a
	// short form like "alias.MyType" (requires a matching @import entry).
	// Empty when the template does not opt into typed mode.
	DataTypeRef string

	// FuncsAlias maps an alias name to a Go import path. Each entry comes
	// from one {{/* @funcs <alias>=<import-path> */}} directive.
	FuncsAlias map[string]string

	// Imports maps an alias name to a Go import path. Each entry comes
	// from one {{/* @import <alias>=<import-path> */}} directive. Used to
	// resolve short @data type refs.
	Imports map[string]string
}

// Typed reports whether the template opts into typed codegen.
func (d Directives) Typed() bool {
	return d.DataTypeRef != ""
}

// directiveRE matches a single template-comment directive, tolerating the
// {{- ... -}} whitespace-trim variants.
var directiveRE = regexp.MustCompile(`\{\{-?\s*/\*\s*@(data|funcs|import)\s+(.*?)\s*\*/\s*-?\}\}`)

// ParseDirectives extracts all comtmpl directives from the raw bytes of a
// template file. Unknown @-directives are reported as errors so typos
// fail loudly rather than silently downgrading to dynamic mode.
func ParseDirectives(raw []byte) (Directives, error) {
	dirs := Directives{
		FuncsAlias: map[string]string{},
		Imports:    map[string]string{},
	}

	matches := directiveRE.FindAllSubmatch(raw, -1)
	for _, m := range matches {
		kind := string(m[1])
		value := strings.TrimSpace(string(m[2]))

		switch kind {
		case "data":
			if dirs.DataTypeRef != "" {
				return dirs, fmt.Errorf("duplicate @data directive: %q (previous: %q)", value, dirs.DataTypeRef)
			}
			if value == "" {
				return dirs, fmt.Errorf("@data directive requires a type reference (e.g. %q)", "github.com/foo.MyType")
			}
			dirs.DataTypeRef = value

		case "funcs":
			alias, path, ok := splitAliasEqPath(value)
			if !ok {
				return dirs, fmt.Errorf("@funcs directive must be alias=<import-path>, got %q", value)
			}
			if existing, dup := dirs.FuncsAlias[alias]; dup {
				return dirs, fmt.Errorf("duplicate @funcs alias %q (previous: %q, new: %q)", alias, existing, path)
			}
			dirs.FuncsAlias[alias] = path

		case "import":
			alias, path, ok := splitAliasEqPath(value)
			if !ok {
				return dirs, fmt.Errorf("@import directive must be alias=<import-path>, got %q", value)
			}
			if existing, dup := dirs.Imports[alias]; dup {
				return dirs, fmt.Errorf("duplicate @import alias %q (previous: %q, new: %q)", alias, existing, path)
			}
			dirs.Imports[alias] = path
		}
	}
	return dirs, nil
}

// splitAliasEqPath parses "alias=path" with whitespace tolerance.
func splitAliasEqPath(s string) (alias, path string, ok bool) {
	idx := strings.IndexByte(s, '=')
	if idx <= 0 {
		return "", "", false
	}
	alias = strings.TrimSpace(s[:idx])
	path = strings.TrimSpace(s[idx+1:])
	if alias == "" || path == "" {
		return "", "", false
	}
	return alias, path, true
}

// SplitDataTypeRef parses a @data ref into (importPath, typeName). For a
// fully-qualified ref like "github.com/foo/bar.MyType" the import path is
// "github.com/foo/bar" and the type name is "MyType". For a short ref
// like "examples.MyType" the path is the alias "examples" (caller must
// resolve via Directives.Imports).
func SplitDataTypeRef(ref string) (pathOrAlias, typeName string, err error) {
	idx := strings.LastIndexByte(ref, '.')
	if idx <= 0 || idx == len(ref)-1 {
		return "", "", fmt.Errorf("@data ref %q is not of the form <import-path-or-alias>.<TypeName>", ref)
	}
	return ref[:idx], ref[idx+1:], nil
}
