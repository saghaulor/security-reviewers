package invariants_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// TestH5_CoreValidatorPathHasZeroNonStdlibDeps asserts:
//  1. Non-test packages import only stdlib (checked via go list WITHOUT -test).
//  2. Test code may additionally import google/go-cmp (subpackages allowed).
//  3. No other non-stdlib module appears anywhere (checked via go list WITH -test).
//
// NOTE: The old implementation used ForTest == "" as a proxy for "non-test code",
// but that is incorrect: external packages always have ForTest=="" regardless of
// whether they appear only in test binaries. We now use two separate go list
// invocations — one without -test (pure non-test graph) and one with -test
// (full graph including test deps) — which accurately separates the two checks.
func TestH5_CoreValidatorPathHasZeroNonStdlibDeps(t *testing.T) {
	const (
		modulePrefix = "github.com/saghaulor/claude-security-hooks/"
		allowedTest  = "github.com/google/go-cmp/"
	)
	type pkg struct {
		ImportPath string
		Standard   bool
	}

	// --- Phase 1: non-test code must only use stdlib ---
	// go list without -test produces the non-test dependency graph only.
	out1, err := exec.Command("go", "list", "-deps", "-json", "./...").Output()
	if err != nil {
		t.Fatalf("go list (non-test) failed: %v", err)
	}
	dec1 := json.NewDecoder(strings.NewReader(string(out1)))
	for {
		var p pkg
		if err := dec1.Decode(&p); err != nil {
			break
		}
		if p.Standard {
			continue
		}
		if strings.HasPrefix(p.ImportPath, modulePrefix) {
			continue
		}
		t.Errorf("H5: non-test code depends on non-stdlib package: %s", p.ImportPath)
	}

	// --- Phase 2: test code may only additionally use go-cmp; no other external packages ---
	// go list with -test includes both the normal and test binary dependency graphs.
	out2, err := exec.Command("go", "list", "-deps", "-test", "-json", "./...").Output()
	if err != nil {
		t.Fatalf("go list (test) failed: %v", err)
	}
	dec2 := json.NewDecoder(strings.NewReader(string(out2)))
	seen := map[string]bool{}
	for {
		var p pkg
		if err := dec2.Decode(&p); err != nil {
			break
		}
		if p.Standard {
			continue
		}
		if strings.HasPrefix(p.ImportPath, modulePrefix) {
			continue
		}
		if strings.HasPrefix(p.ImportPath, allowedTest) {
			continue // go-cmp is the only allowed external test dependency
		}
		if !seen[p.ImportPath] {
			seen[p.ImportPath] = true
			t.Errorf("H5: disallowed external dependency (even in tests): %s", p.ImportPath)
		}
	}
}
