package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestImportSetDedupe(t *testing.T) {
	s := NewImportSet()
	a1 := s.Add("github.com/foo/bar", "bar")
	a2 := s.Add("github.com/foo/bar", "")
	if a1 != a2 || a1 != "bar" {
		t.Fatalf("expected same alias 'bar', got %q and %q", a1, a2)
	}
}

func TestImportSetAliasCollision(t *testing.T) {
	s := NewImportSet()
	a1 := s.Add("github.com/foo/bar", "bar")
	a2 := s.Add("github.com/baz/bar", "bar")
	if a1 == a2 {
		t.Fatalf("expected distinct aliases, got %q twice", a1)
	}
	if a1 != "bar" || !strings.HasPrefix(a2, "bar") || a2 == "bar" {
		t.Fatalf("got aliases %q, %q", a1, a2)
	}
}

func TestImportSetWriteImports(t *testing.T) {
	s := NewImportSet()
	s.Add("github.com/foo/bar", "bar")
	s.Add("io", "")
	var buf bytes.Buffer
	s.WriteImports(&buf)
	out := buf.String()
	if !strings.Contains(out, `"github.com/foo/bar"`) {
		t.Errorf("missing bar import:\n%s", out)
	}
	if !strings.Contains(out, `"io"`) {
		t.Errorf("missing io import:\n%s", out)
	}
}

func TestGeneratorScopes(t *testing.T) {
	g := &Generator{}
	g.PushScope()
	g.BindVar("x", ScopeBinding{GoExpr: "x_outer"})
	g.PushScope()
	g.BindVar("x", ScopeBinding{GoExpr: "x_inner"})

	if b, ok := g.LookupVar("x"); !ok || b.GoExpr != "x_inner" {
		t.Errorf("inner lookup got %+v, %v", b, ok)
	}
	g.PopScope()
	if b, ok := g.LookupVar("x"); !ok || b.GoExpr != "x_outer" {
		t.Errorf("after pop got %+v, %v", b, ok)
	}
	g.PopScope()
	if _, ok := g.LookupVar("x"); ok {
		t.Error("expected unbound after popping all scopes")
	}
}
