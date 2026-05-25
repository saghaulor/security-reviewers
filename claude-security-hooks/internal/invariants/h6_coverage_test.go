package invariants_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// ids returns the canonical assertion ID list for a prefix-range
// (e.g., ids("T", 1, 11) = ["T1", ..., "T11"]).
func ids(prefix string, lo, hi int) []string {
	out := make([]string, 0, hi-lo+1)
	for i := lo; i <= hi; i++ {
		out = append(out, prefix+strconv.Itoa(i))
	}
	return out
}

// expectedIDs maps each per-agent test file to the assertion IDs that file
// MUST cover with at least one TestXY_* function. This map is the SOURCE
// OF TRUTH for H6 drift detection and is SEALED in Wave 0 — Wave 1 plans
// (02-02 through 02-07) do NOT modify this map; they contribute test
// functions in their per-agent _test.go files which H6 discovers via AST
// walk. If a Wave 1 plan forgets to add a Test func for an ID, H6 fails
// on that per-agent _test.go row.
//
// NOTE: At Wave 0 end, H6 will fail for the per-agent _test.go rows
// because those files do not yet exist. That is intentional — Wave 1
// plans close each row as they land.
func expectedIDs() map[string][]string {
	return map[string][]string{
		"cartographer_test.go":      ids("A", 1, 11),
		"taint_tracer_test.go":      ids("T", 1, 11),
		"authz_tracer_test.go":      ids("AZ", 1, 6),
		"oauth_auditor_test.go":     ids("OA", 1, 7),
		"invariant_checker_test.go": ids("IC", 1, 4),
		"synthesis_test.go":         ids("S", 1, 6),
	}
}

func TestH6_EveryAssertionHasUnitTest(t *testing.T) {
	fset := token.NewFileSet()
	for file, ids := range expectedIDs() {
		path := filepath.Join(".", file)
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Errorf("H6: parse %s: %v (Wave 1 plan owning %s must land first)", path, err, file)
			continue
		}
		funcs := map[string]bool{}
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && strings.HasPrefix(fd.Name.Name, "Test") {
				funcs[fd.Name.Name] = true
			}
		}
		for _, id := range ids {
			prefix := "Test" + id + "_"
			found := false
			for name := range funcs {
				if strings.HasPrefix(name, prefix) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("H6: %s has no test function matching %s*", file, prefix)
			}
		}
	}
}
