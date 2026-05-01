package main

import (
	"reflect"
	"testing"
)

func TestParseDirectivesNone(t *testing.T) {
	d, err := ParseDirectives([]byte("<h1>{{.Title}}</h1>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Typed() {
		t.Fatalf("expected dynamic mode, got typed: %+v", d)
	}
}

func TestParseDirectivesAll(t *testing.T) {
	src := `{{/* @data examples.IndexData */}}
{{/* @import examples=github.com/jtarchie/comtmpl/examples */}}
{{/* @funcs sprig=github.com/go-task/slim-sprig/v3 */}}
<h1>{{.Title}}</h1>`

	d, err := ParseDirectives([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.DataTypeRef != "examples.IndexData" {
		t.Errorf("DataTypeRef = %q, want %q", d.DataTypeRef, "examples.IndexData")
	}
	wantImports := map[string]string{"examples": "github.com/jtarchie/comtmpl/examples"}
	if !reflect.DeepEqual(d.Imports, wantImports) {
		t.Errorf("Imports = %v, want %v", d.Imports, wantImports)
	}
	wantFuncs := map[string]string{"sprig": "github.com/go-task/slim-sprig/v3"}
	if !reflect.DeepEqual(d.FuncsAlias, wantFuncs) {
		t.Errorf("FuncsAlias = %v, want %v", d.FuncsAlias, wantFuncs)
	}
}

func TestParseDirectivesWhitespaceTrim(t *testing.T) {
	src := `{{- /* @data foo.Bar */ -}}`
	d, err := ParseDirectives([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.DataTypeRef != "foo.Bar" {
		t.Errorf("DataTypeRef = %q, want %q", d.DataTypeRef, "foo.Bar")
	}
}

func TestParseDirectivesErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"duplicate @data", "{{/* @data a.B */}}{{/* @data c.D */}}"},
		{"empty @data", "{{/* @data */}}"},
		{"bad @funcs format", "{{/* @funcs sprig */}}"},
		{"empty @funcs alias", "{{/* @funcs =github.com/foo */}}"},
		{"duplicate @funcs alias", "{{/* @funcs s=a */}}{{/* @funcs s=b */}}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseDirectives([]byte(tc.src)); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestSplitDataTypeRef(t *testing.T) {
	cases := []struct {
		ref      string
		wantPath string
		wantName string
		wantErr  bool
	}{
		{"examples.IndexData", "examples", "IndexData", false},
		{"github.com/foo/bar.T", "github.com/foo/bar", "T", false},
		{"NoDot", "", "", true},
		{"trailing.", "", "", true},
		{".LeadingDot", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.ref, func(t *testing.T) {
			path, name, err := SplitDataTypeRef(tc.ref)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr {
				if path != tc.wantPath || name != tc.wantName {
					t.Errorf("got (%q, %q), want (%q, %q)", path, name, tc.wantPath, tc.wantName)
				}
			}
		})
	}
}
