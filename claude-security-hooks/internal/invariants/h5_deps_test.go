package invariants_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// TestH5_CoreValidatorPathHasZeroNonStdlibDeps asserts:
//   1. Non-test packages import only stdlib.
//   2. Test packages may additionally import google/go-cmp (subpackages allowed).
//   3. No other non-stdlib module appears anywhere.
func TestH5_CoreValidatorPathHasZeroNonStdlibDeps(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-test", "-json", "./...").Output()
	if err != nil {
		t.Fatalf("go list failed: %v", err)
	}
	dec := json.NewDecoder(strings.NewReader(string(out)))

	const (
		modulePrefix = "github.com/saghaulor/claude-security-hooks/"
		allowedTest  = "github.com/google/go-cmp/"
	)
	type pkg struct {
		ImportPath string
		Standard   bool
		ForTest    string
	}
	for {
		var p pkg
		if err := dec.Decode(&p); err != nil {
			break
		}
		if p.Standard {
			continue
		}
		if strings.HasPrefix(p.ImportPath, modulePrefix) {
			continue
		}
		if strings.HasPrefix(p.ImportPath, allowedTest) {
			if p.ForTest == "" {
				t.Errorf("H5: %s imported by non-test code", p.ImportPath)
			}
			continue
		}
		t.Errorf("H5: disallowed dependency: %s (for-test=%q)", p.ImportPath, p.ForTest)
	}
}
