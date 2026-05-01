package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// TypeResolver loads Go packages on demand and resolves named types and
// FuncMap definitions used by typed codegen. It caches loaded packages so
// repeated lookups across templates only pay the load cost once per run.
type TypeResolver struct {
	pkgs map[string]*packages.Package
}

func NewTypeResolver() *TypeResolver {
	return &TypeResolver{pkgs: map[string]*packages.Package{}}
}

// loadPackage loads (or returns the cached) types information for the
// import path. The configured mode includes Syntax so callers can walk
// the AST when inspecting FuncMap literal values (Phase 3).
func (r *TypeResolver) loadPackage(importPath string) (*packages.Package, error) {
	if pkg, ok := r.pkgs[importPath]; ok {
		return pkg, nil
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo,
	}
	loaded, err := packages.Load(cfg, importPath)
	if err != nil {
		return nil, fmt.Errorf("load package %q: %w", importPath, err)
	}
	if len(loaded) == 0 {
		return nil, fmt.Errorf("no package found at %q", importPath)
	}
	pkg := loaded[0]
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("package %q has errors: %v", importPath, pkg.Errors)
	}
	r.pkgs[importPath] = pkg
	return pkg, nil
}

// ResolveType returns the *types.Type for typeName declared in the
// package at importPath. The returned type is the named type's underlying
// type if you call .Underlying(); the named type itself is suitable for
// referencing in generated source.
func (r *TypeResolver) ResolveType(importPath, typeName string) (types.Type, error) {
	pkg, err := r.loadPackage(importPath)
	if err != nil {
		return nil, err
	}
	if pkg.Types == nil {
		return nil, fmt.Errorf("package %q has no type information", importPath)
	}
	scope := pkg.Types.Scope()
	obj := scope.Lookup(typeName)
	if obj == nil {
		return nil, fmt.Errorf("type %q not found in package %q", typeName, importPath)
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("%s.%s is not a type (got %T)", importPath, typeName, obj)
	}
	return tn.Type(), nil
}
