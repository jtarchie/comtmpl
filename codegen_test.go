package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

type codegenResult struct {
	TmpDir      string
	Generated   string
	BuildErr    error
	BuildStderr string
}

// repoRoot returns the absolute path to the comtmpl repo (the directory
// containing this test file).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine repo root")
	}
	return filepath.Dir(filename)
}

// runCodegen writes templates (name -> content) and supportFiles (relpath
// -> content) to a temp dir, runs Generate in-process, writes a go.mod
// that replaces github.com/jtarchie/comtmpl with the local repo, and runs
// `go build ./...`. The result captures the generator's output and the
// build's stderr so tests can assert on either.
func runCodegen(t *testing.T, srcs map[string]string, supportFiles map[string]string) *codegenResult {
	t.Helper()
	tmp := t.TempDir()
	root := repoRoot(t)

	names := make([]string, 0, len(srcs))
	for name := range srcs {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic

	templatePaths := make([]string, 0, len(names))
	for _, name := range names {
		path := filepath.Join(tmp, name)
		if err := os.WriteFile(path, []byte(srcs[name]), 0o644); err != nil {
			t.Fatalf("write template %s: %v", name, err)
		}
		templatePaths = append(templatePaths, path)
	}

	for relPath, content := range supportFiles {
		path := filepath.Join(tmp, relPath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", relPath, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write support file %s: %v", relPath, err)
		}
	}

	goMod := "module testpkg\n\ngo 1.24\n\nrequire github.com/jtarchie/comtmpl v0.0.0\n\nreplace github.com/jtarchie/comtmpl => " + root + "\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	var buf bytes.Buffer
	if err := Generate(GenOptions{
		Filenames:   templatePaths,
		PackageName: "testpkg",
		Output:      &buf,
	}); err != nil {
		return &codegenResult{TmpDir: tmp, BuildErr: err}
	}
	generated := buf.String()
	if err := os.WriteFile(filepath.Join(tmp, "templates_gen.go"), []byte(generated), 0o644); err != nil {
		t.Fatalf("write generated: %v", err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tmp
	var tidyErr bytes.Buffer
	tidy.Stderr = &tidyErr
	if err := tidy.Run(); err != nil {
		return &codegenResult{
			TmpDir: tmp, Generated: generated,
			BuildErr: err, BuildStderr: tidyErr.String(),
		}
	}

	build := exec.Command("go", "build", "./...")
	build.Dir = tmp
	var buildErr bytes.Buffer
	build.Stderr = &buildErr
	err := build.Run()
	return &codegenResult{
		TmpDir: tmp, Generated: generated,
		BuildErr: err, BuildStderr: buildErr.String(),
	}
}

// TestHarnessSmoke: the existing example templates (re-used as raw strings
// here) must compile through the harness. This is the safety net for
// every subsequent codegen test.
func TestHarnessSmoke(t *testing.T) {
	srcs := map[string]string{
		"index.html": `<html>
  <head>
    <title>{{.Title}}</title>
  </head>
  <body>
    <h1>{{.Title}}</h1>
    <p>Welcome, {{.User.Name}}!</p>
  </body>
</html>`,
		"pipe.html": `<p>{{.Title | upper}}</p>
<p>{{.Title | len}}</p>
<p>{{.Title | title}}</p>`,
	}

	res := runCodegen(t, srcs, nil)
	if res.BuildErr != nil {
		t.Fatalf("smoke test failed to build: %v\nstderr:\n%s\n\ngenerated:\n%s",
			res.BuildErr, res.BuildStderr, res.Generated)
	}
}

// TestLineDirectivesEmitted confirms that every generated template file
// contains at least one //line directive at column 0 referencing the
// original template path. This is the regression guard for Phase 1.2.
func TestLineDirectivesEmitted(t *testing.T) {
	srcs := map[string]string{
		"hello.html": `<h1>{{.Title}}</h1>`,
	}
	res := runCodegen(t, srcs, nil)
	if res.BuildErr != nil {
		t.Fatalf("build failed: %v\nstderr:\n%s", res.BuildErr, res.BuildStderr)
	}
	if !bytes.Contains([]byte(res.Generated), []byte("\n//line ")) {
		t.Fatalf("generated source missing //line directive at column 0:\n%s", res.Generated)
	}
	if !bytes.Contains([]byte(res.Generated), []byte("hello.html:1")) {
		t.Fatalf("generated source missing reference to hello.html:1:\n%s", res.Generated)
	}
}

// TestLineDirectivesAttributeErrors is the linchpin test for the
// //line-directive work: when generated Go fails to compile, the build
// stderr must reference the template file:line, not the generated .go
// file. We do this by injecting a Go syntax error into the generated
// source AT the location where a template node would emit code, then
// asserting that the resulting build error mentions the template path.
func TestLineDirectivesAttributeErrors(t *testing.T) {
	srcs := map[string]string{
		"broken.html": `<h1>{{.Title}}</h1>`,
	}
	res := runCodegen(t, srcs, nil)
	if res.BuildErr != nil {
		t.Fatalf("baseline build failed: %v\nstderr:\n%s", res.BuildErr, res.BuildStderr)
	}

	// Replace the EvalField call with a reference to an undefined symbol
	// so the Go compiler produces a clean error.
	broken := bytes.Replace(
		[]byte(res.Generated),
		[]byte("templates.EvalField(data, []string{\"Title\"})"),
		[]byte("notARealSymbol(data)"),
		1,
	)
	if err := os.WriteFile(filepath.Join(res.TmpDir, "templates_gen.go"), broken, 0o644); err != nil {
		t.Fatalf("rewrite generated: %v", err)
	}

	build := exec.Command("go", "build", "./...")
	build.Dir = res.TmpDir
	var stderr bytes.Buffer
	build.Stderr = &stderr
	if err := build.Run(); err == nil {
		t.Fatal("expected build failure, got success")
	}
	stderrStr := stderr.String()
	// The Go compiler reports errors using the //line directive's path,
	// re-numbered relative to the directive. We don't pin the exact line
	// (a single template node maps to several physical lines of generated
	// Go), only that the template file is referenced.
	if !bytes.Contains([]byte(stderrStr), []byte("broken.html:")) {
		t.Fatalf("expected build error to reference broken.html, got:\n%s", stderrStr)
	}
	if bytes.Contains([]byte(stderrStr), []byte("templates_gen.go")) {
		t.Fatalf("build error leaked generated filename instead of template:\n%s", stderrStr)
	}
}

// TestDynamicFallback confirms that templates without directives still
// emit the legacy reflection-based runtime helpers. This guards the
// Phase 1 backwards-compat contract: existing templates keep working
// unchanged.
func TestDynamicFallback(t *testing.T) {
	srcs := map[string]string{
		"legacy.html": `{{.Field}} {{upper .Other}}`,
	}
	res := runCodegen(t, srcs, nil)
	if res.BuildErr != nil {
		t.Fatalf("build failed: %v\nstderr:\n%s", res.BuildErr, res.BuildStderr)
	}
	if !bytes.Contains([]byte(res.Generated), []byte("templates.EvalField")) {
		t.Fatalf("expected legacy EvalField call in generated source:\n%s", res.Generated)
	}
	if !bytes.Contains([]byte(res.Generated), []byte("t.CallFunc(")) {
		t.Fatalf("expected legacy CallFunc in generated source:\n%s", res.Generated)
	}
}
