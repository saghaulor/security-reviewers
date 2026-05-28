package invariants_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInvariantsHasNoHooksImport asserts the architectural boundary rule
// (PATTERNS.md Pattern S4): internal/invariants/*.go must NOT import
// internal/hooks. Predicates are pure functions over typed verdict structs
// and must remain I/O-free.
func TestInvariantsHasNoHooksImport(t *testing.T) {
	const forbidden = `"github.com/saghaulor/claude-security-hooks/internal/hooks"`
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(".", e.Name())
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imp := range f.Imports {
			if imp.Path.Value == forbidden {
				t.Errorf("PATTERNS.md S4 violation: %s imports internal/hooks (predicates must be I/O-free)", path)
			}
		}
	}
}
