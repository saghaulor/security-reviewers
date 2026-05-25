package invariants

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

type CartographerInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.CartographerIndex) []Violation
}

// CartographerGraphPath is the workspace-relative path to graphify-out/graph.json.
// Tests override this to point at a fixture under testdata/. Default per HAND_OFF §3.1 A6.
var CartographerGraphPath = "graphify-out/graph.json"

// CartographerInvariants registers A1..A11 in registration order (D-01).
var CartographerInvariants = []CartographerInvariant{
	{ID: "A1", Description: "output is valid go-index/v1 JSON with required top-level fields", Severity: SeverityCritical, Check: checkA1},
	{ID: "A2", Description: "schema_version equals literal 'go-index/v1'", Severity: SeverityCritical, Check: checkA2},
	{ID: "A3", Description: "every entrypoint has router/method/path/handler.{fqn,file,line} populated", Severity: SeverityCritical, Check: checkA3},
	{ID: "A4", Description: "routers_detected only contains entries from the known router set", Severity: SeverityHigh, Check: checkA4},
	{ID: "A5", Description: "non-empty routers + empty entrypoints requires warning 'router_detected_but_no_routes'", Severity: SeverityMedium, Check: checkA5},
	{ID: "A6", Description: "every cited node id exists in graphify-out/graph.json", Severity: SeverityHigh, Check: checkA6},
	{ID: "A7", Description: "every cited file path exists in the workspace", Severity: SeverityHigh, Check: checkA7},
	{ID: "A8", Description: "every cited line number is within the file's line count", Severity: SeverityHigh, Check: checkA8},
	{ID: "A9", Description: "authz_primitives only contains entries with blocking=true", Severity: SeverityHigh, Check: checkA9},
	{ID: "A10", Description: "agent did not modify graphify-out/ or source tree (no-op stub Phase 2; real check Phase 5)", Severity: SeverityInfo, Check: checkA10},
	{ID: "A11", Description: "agent Bash commands match allowed_commands (no-op stub Phase 2; real check Phase 5)", Severity: SeverityInfo, Check: checkA11},
}

// --- A1: top-level shape ---
func checkA1(idx *schema.CartographerIndex) []Violation {
	var out []Violation
	if idx.SchemaVersion == "" {
		out = append(out, Violation{Path: "schema_version", Expected: "non-empty", Actual: ""})
	}
	if idx.SinksByKind == nil {
		out = append(out, Violation{Path: "sinks_by_kind", Expected: "present", Actual: "nil"})
	}
	return out
}

// --- A2: schema_version literal ---
func checkA2(idx *schema.CartographerIndex) []Violation {
	const want = "go-index/v1"
	if idx.SchemaVersion != want {
		return []Violation{{Path: "schema_version", Expected: want, Actual: idx.SchemaVersion}}
	}
	return nil
}

// --- A3: entrypoint required fields ---
func checkA3(idx *schema.CartographerIndex) []Violation {
	var out []Violation
	for i, ep := range idx.Entrypoints {
		if ep.Router == "" {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].router", i), Expected: "non-empty", Actual: ""})
		}
		if ep.Method == "" {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].method", i), Expected: "non-empty", Actual: ""})
		}
		if ep.Path == "" {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].path", i), Expected: "non-empty", Actual: ""})
		}
		if ep.Handler.FQN == "" {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].handler.fqn", i), Expected: "non-empty", Actual: ""})
		}
		if ep.Handler.File == "" {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].handler.file", i), Expected: "non-empty", Actual: ""})
		}
		if ep.Handler.Line == 0 {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].handler.line", i), Expected: ">0", Actual: "0"})
		}
	}
	return out
}

// --- A4: routers_detected allowed set ---
var cartographerRouterSet = map[string]struct{}{
	"net/http": {}, "chi": {}, "gin": {}, "gorilla/mux": {},
	"echo": {}, "fiber": {}, "httprouter": {}, "custom": {},
}

func checkA4(idx *schema.CartographerIndex) []Violation {
	var out []Violation
	// Build expected description deterministically (sorted).
	names := make([]string, 0, len(cartographerRouterSet))
	for n := range cartographerRouterSet {
		names = append(names, n)
	}
	sort.Strings(names)
	expected := "in {" + strings.Join(names, ",") + "}"
	for i, r := range idx.RoutersDetected {
		if _, ok := cartographerRouterSet[r]; !ok {
			out = append(out, Violation{Path: fmt.Sprintf("routers_detected[%d]", i), Expected: expected, Actual: r})
		}
	}
	return out
}

// --- A5: routers without entrypoints requires warning ---
func checkA5(idx *schema.CartographerIndex) []Violation {
	if len(idx.RoutersDetected) == 0 || len(idx.Entrypoints) > 0 {
		return nil
	}
	const want = "router_detected_but_no_routes"
	for _, w := range idx.Warnings {
		if w == want {
			return nil
		}
	}
	return []Violation{{Path: "warnings", Expected: "contains '" + want + "'", Actual: "missing"}}
}

// --- A6: cited node IDs exist in graphify-out/graph.json ---
func checkA6(idx *schema.CartographerIndex) []Violation {
	if !FileExists(CartographerGraphPath) {
		return []Violation{{Path: CartographerGraphPath, Expected: "exists", Actual: "missing"}}
	}
	data, err := os.ReadFile(CartographerGraphPath)
	if err != nil {
		return []Violation{{Path: CartographerGraphPath, Expected: "readable", Actual: err.Error()}}
	}
	var g struct {
		Nodes []struct {
			ID string `json:"id"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(data, &g); err != nil {
		return []Violation{{Path: CartographerGraphPath, Expected: "valid JSON", Actual: err.Error()}}
	}
	nodeSet := make(map[string]struct{}, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeSet[n.ID] = struct{}{}
	}
	var out []Violation
	for i, ep := range idx.Entrypoints {
		if ep.Handler.FQN == "" {
			continue
		}
		if _, ok := nodeSet[ep.Handler.FQN]; !ok {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].handler.fqn", i), Expected: "present in " + CartographerGraphPath, Actual: ep.Handler.FQN})
		}
	}
	for i, ap := range idx.AuthzPrimitives {
		if ap.FQN == "" {
			continue
		}
		if _, ok := nodeSet[ap.FQN]; !ok {
			out = append(out, Violation{Path: fmt.Sprintf("authz_primitives[%d].fqn", i), Expected: "present in " + CartographerGraphPath, Actual: ap.FQN})
		}
	}
	return out
}

// --- A7: cited file paths exist ---
func checkA7(idx *schema.CartographerIndex) []Violation {
	var out []Violation
	for i, ep := range idx.Entrypoints {
		if ep.Handler.File != "" && !FileExists(ep.Handler.File) {
			out = append(out, Violation{Path: fmt.Sprintf("entrypoints[%d].handler.file", i), Expected: "exists", Actual: ep.Handler.File})
		}
	}
	for kind, sinks := range idx.SinksByKind {
		for i, s := range sinks {
			if s.File != "" && !FileExists(s.File) {
				out = append(out, Violation{Path: fmt.Sprintf("sinks_by_kind[%s][%d].file", kind, i), Expected: "exists", Actual: s.File})
			}
		}
	}
	return out
}

// --- A8: cited line numbers within bounds ---
func checkA8(idx *schema.CartographerIndex) []Violation {
	var out []Violation
	lineCache := make(map[string]int)
	check := func(path string, file string, line int) {
		if file == "" || line <= 0 {
			return
		}
		n, ok := lineCache[file]
		if !ok {
			if !FileExists(file) {
				// A7 reports the missing file; A8 silent on missing-file case.
				return
			}
			lc, err := LineCount(file)
			if err != nil {
				return
			}
			n = lc
			lineCache[file] = n
		}
		if line > n {
			out = append(out, Violation{Path: path, Expected: fmt.Sprintf("<=%d", n), Actual: fmt.Sprintf("%d", line)})
		}
	}
	for i, ep := range idx.Entrypoints {
		check(fmt.Sprintf("entrypoints[%d].handler.line", i), ep.Handler.File, ep.Handler.Line)
	}
	for kind, sinks := range idx.SinksByKind {
		for i, s := range sinks {
			check(fmt.Sprintf("sinks_by_kind[%s][%d].line", kind, i), s.File, s.Line)
		}
	}
	return out
}

// --- A9: authz_primitives only blocking=true ---
func checkA9(idx *schema.CartographerIndex) []Violation {
	var out []Violation
	for i, ap := range idx.AuthzPrimitives {
		if !ap.Blocking {
			out = append(out, Violation{Path: fmt.Sprintf("authz_primitives[%d].blocking", i), Expected: "true", Actual: "false"})
		}
	}
	return out
}

// --- A10: agent did not modify graphify-out/ or source tree ---
// No-op stub for Phase 2; verdict JSON does not carry tool-call telemetry.
// TODO(phase-5): enforce via SubagentStop telemetry or tool-call summary in verdict.
func checkA10(idx *schema.CartographerIndex) []Violation {
	_ = idx
	return nil
}

// --- A11: agent Bash commands match allowed_commands ---
// No-op stub for Phase 2; verdict JSON does not carry tool-call telemetry.
// TODO(phase-5): enforce via SubagentStop telemetry or tool-call summary in verdict.
func checkA11(idx *schema.CartographerIndex) []Violation {
	_ = idx
	return nil
}
