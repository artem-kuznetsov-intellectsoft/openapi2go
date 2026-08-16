package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

// fixturesDir is the root of the golden-file fixture tree.
const fixturesDir = "fixtures"

// TestFixtureDirsAreCovered fails on a fixture directory holding a spec that
// no test-table entry references, and on a client.gen.go whose entry leaves
// clientRefFile unset. Both were silent before: two fixture directories sat
// unreferenced for long enough to contain no .go files at all, so they were
// outside ./... and got no type-checking either.
func TestFixtureDirsAreCovered(t *testing.T) {
	referenced, clientRefs := referencedFixtures(t)

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", fixturesDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(fixturesDir, entry.Name())

		specs, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			t.Fatalf("globbing %s: %v", dir, err)
		}

		if len(specs) > 0 && !referenced[entry.Name()] {
			t.Errorf("fixture %s holds a spec but no test case references it; "+
				"add a table entry or delete the directory", entry.Name())
		}

		clientGolden := filepath.Join(dir, "client.gen.go")
		if _, err := os.Stat(clientGolden); err == nil && !clientRefs[filepath.ToSlash(clientGolden)] {
			t.Errorf("%s exists but its test case leaves clientRefFile unset, so it is never compared", clientGolden)
		}
	}
}

// referencedFixtures reports which fixture directories the test table names,
// and which client goldens it compares, by reading the table out of the test
// source rather than duplicating it here.
func referencedFixtures(t *testing.T) (dirs, clientRefs map[string]bool) {
	t.Helper()

	src, err := os.ReadFile("generator_test.go")
	if err != nil {
		t.Fatalf("reading generator_test.go: %v", err)
	}

	dirs, clientRefs = map[string]bool{}, map[string]bool{}

	for line := range strings.Lines(string(src)) {
		for _, path := range fixturePathsIn(line) {
			rest, ok := strings.CutPrefix(path, fixturesDir+"/")
			if !ok {
				continue
			}

			name, _, _ := strings.Cut(rest, "/")
			dirs[name] = true

			if strings.HasSuffix(path, "/client.gen.go") {
				clientRefs[path] = true
			}
		}
	}

	return dirs, clientRefs
}

// fixturePathsIn extracts the quoted fixtures/... paths from one source line.
func fixturePathsIn(line string) []string {
	var out []string

	for rest := line; ; {
		_, after, found := strings.Cut(rest, `"`)
		if !found {
			return out
		}

		value, remainder, closed := strings.Cut(after, `"`)
		if !closed {
			return out
		}

		if strings.HasPrefix(value, fixturesDir+"/") {
			out = append(out, value)
		}

		rest = remainder
	}
}

// plumbingIdents are the identifiers that mark request/response plumbing. The
// point of the runtime extraction is that none of them appear in a generated
// method any more; if one survives, the extraction missed that operation
// shape — which a total rewrite of the goldens would not otherwise reveal.
var plumbingIdents = []string{
	"io.ReadAll",
	"json.Marshal",
	"json.Unmarshal",
	"fmt.Sprint",
	"fmt.Sprintf",
	"httpResp",
	"http.NewRequestWithContext",
}

func TestGeneratedClientsDelegateToRuntime(t *testing.T) {
	goldens, err := filepath.Glob(filepath.Join(fixturesDir, "*", "client.gen.go"))
	if err != nil {
		t.Fatalf("globbing client goldens: %v", err)
	}

	if len(goldens) == 0 {
		t.Fatal("no client goldens found; the glob or the fixture layout changed")
	}

	for _, golden := range goldens {
		t.Run(filepath.Base(filepath.Dir(golden)), func(t *testing.T) {
			src, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("reading %s: %v", golden, err)
			}

			for _, ident := range plumbingIdents {
				if strings.Contains(string(src), ident) {
					t.Errorf("%s still references %s; that plumbing belongs in the client runtime", golden, ident)
				}
			}
		})
	}
}

// TestGeneratedClientMethodsHaveOneRequestPath asserts every generated method
// sends its request through exactly one c.do call and gates the result on
// exactly one 2xx check. A statement-count budget would be the obvious guard
// here, but it false-positives on an operation with many parameters — the
// prelude scales with parameter count, which is not re-inlined plumbing. The
// call structure is the invariant that actually distinguishes the two.
func TestGeneratedClientMethodsHaveOneRequestPath(t *testing.T) {
	goldens, err := filepath.Glob(filepath.Join(fixturesDir, "*", "client.gen.go"))
	if err != nil {
		t.Fatalf("globbing client goldens: %v", err)
	}

	for _, golden := range goldens {
		file, err := parser.ParseFile(token.NewFileSet(), golden, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", golden, err)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil {
				continue
			}

			calls := callCounts(fn.Body)

			if n := calls["c.do"]; n != 1 {
				t.Errorf("%s: %s makes %d c.do calls, want exactly 1", golden, fn.Name.Name, n)
			}
			if n := calls["resp.expectSuccess"] + calls["expectSuccessDefault"]; n != 1 {
				t.Errorf("%s: %s has %d success gates, want exactly 1 — every status must be classified",
					golden, fn.Name.Name, n)
			}
		}
	}
}

// callCounts tallies the calls in body by callee name.
func callCounts(body *ast.BlockStmt) map[string]int {
	counts := map[string]int{}

	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			counts[calleeName(call.Fun)]++
		}

		return true
	})

	return counts
}

// calleeName renders a call's function expression as "recv.name", "name", or
// the same with any generic instantiation stripped.
func calleeName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.IndexExpr: // decodeJSON[T](...)
		return calleeName(f.X)
	case *ast.IndexListExpr:
		return calleeName(f.X)
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		if x, ok := f.X.(*ast.Ident); ok {
			return x.Name + "." + f.Sel.Name
		}

		return f.Sel.Name
	}

	return ""
}

// TestSupportFilesOnlyImportStdlib guards the invariant that makes the
// support-file mechanism work at all: each file is copied into a package that
// cannot resolve this module's imports, so a non-stdlib import there would
// only fail in the consumer's build.
func TestSupportFilesOnlyImportStdlib(t *testing.T) {
	for name, src := range openapi.SupportFiles() {
		t.Run(name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), name, src, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parsing %s: %v", name, err)
			}

			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				// Only non-stdlib import paths contain a dot, which is the
				// domain in a module path.
				if strings.Contains(strings.SplitN(path, "/", 2)[0], ".") {
					t.Errorf("support file %s imports %q, which the generated package cannot resolve", name, path)
				}
			}
		})
	}
}

// TestSupportFilesPackageClauseIsRewritten asserts every support file ends up
// declaring the target package. rewritePackageClause locates the clause
// rather than matching "package openapi" literally, because the client
// runtime lives in its own subpackage — this pins that it still works for
// every file, including the one with a package doc comment.
func TestSupportFilesPackageClauseIsRewritten(t *testing.T) {
	for name, src := range openapi.SupportFiles() {
		t.Run(name, func(t *testing.T) {
			got := rewritePackageClause(src, "generated")

			if !strings.HasPrefix(got, "package generated\n") {
				first, _, _ := strings.Cut(got, "\n")
				t.Errorf("rewritten %s starts with %q, want the package clause", name, first)
			}

			if strings.Contains(got, "package openapi\n") || strings.Contains(got, "package clientruntime\n") {
				t.Errorf("rewritten %s still declares its source package", name)
			}
		})
	}
}
