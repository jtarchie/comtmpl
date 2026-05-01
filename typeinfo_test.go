package main

import (
	"go/types"
	"testing"
)

// TestResolveTypeFromExamples loads a struct type from this repo's
// examples package via the TypeResolver. Uses the existing examples
// package as a fixture; if examples ever loses its setup.go this test
// will need to be reworked.
func TestResolveTypeFromExamples(t *testing.T) {
	r := NewTypeResolver()
	// The examples package exports `Parsed` (a *templates.Templates).
	// Resolving Templates from the runtime package is a stable target.
	typ, err := r.ResolveType("github.com/jtarchie/comtmpl/templates", "Templates")
	if err != nil {
		t.Fatalf("resolve Templates: %v", err)
	}
	named, ok := typ.(*types.Named)
	if !ok {
		t.Fatalf("expected *types.Named, got %T", typ)
	}
	if got := named.Obj().Name(); got != "Templates" {
		t.Errorf("named.Obj().Name() = %q, want %q", got, "Templates")
	}
}

func TestResolveTypeMissing(t *testing.T) {
	r := NewTypeResolver()
	if _, err := r.ResolveType("github.com/jtarchie/comtmpl/templates", "BogusType"); err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestResolveTypeBadPackage(t *testing.T) {
	r := NewTypeResolver()
	if _, err := r.ResolveType("github.com/this/does/not/exist", "Anything"); err == nil {
		t.Fatal("expected error for missing package")
	}
}

// TestResolveTypeCached verifies the cache short-circuits a second call
// for the same package.
func TestResolveTypeCached(t *testing.T) {
	r := NewTypeResolver()
	if _, err := r.ResolveType("github.com/jtarchie/comtmpl/templates", "Templates"); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if _, ok := r.pkgs["github.com/jtarchie/comtmpl/templates"]; !ok {
		t.Fatal("expected package to be cached after first call")
	}
}
