# Phase 2: claude-security-hooks Go binary - Pattern Map

**Mapped:** 2026-05-19
**Files analyzed:** 38 new files + 1 modified (`go.mod`/`go.sum`)
**Analogs found in-repo:** 0 / 38 (Phase 1 scaffold is empty directories + a single `go.mod`)
**Canonical external analogs cited:** 38 / 38 (verified excerpts pulled from RESEARCH.md § Code Examples and § Recommended Project Structure)

## Pattern Shortage Notice

The Phase 1 scaffold (`scaffold: initial project structure`, commit `adfbdbf`) created **directory shells only**: `claude-security-hooks/cmd/claude-security-hooks/`, `claude-security-hooks/internal/{hooks,invariants,schema}/`, plus `claude-security-hooks/go.mod` (3 lines: module + go 1.22). There are **zero Go source files anywhere in the repository** as of this mapping — verified by `find /home/saghaulor/code/security_reviewer -type f -name "*.go"` returning empty. The sibling `opengrep-mcp/` is likewise scaffold-only.

**Consequence:** the planner cannot pull excerpts from "the closest existing file in the repo." Every concrete pattern below is therefore sourced from one of:
1. **RESEARCH.md § Code Examples** (verified, mentally-type-checked Go snippets — the primary planner reference).
2. **Canonical external libraries / docs** named in RESEARCH.md (kubernetes apimachinery `field.ErrorList`, gobyexample.com subcommand pattern, stdlib `encoding/json` / `go/parser` / `flag` packages, `github.com/google/go-cmp`).

Treat RESEARCH.md as the de-facto analog source for this phase. Each `## Pattern Assignments` entry below cites a verbatim line range in RESEARCH.md so the planner can reference excerpts directly without re-reading.

## File Classification

Files are grouped by wave per RESEARCH.md "Recommendations to Planner" (Wave 0 = infra; Wave 1 = per-agent registries; Wave 2 = subcommands + meta-tests + Makefile).

| Wave | New File | Role | Data Flow | Closest Analog | Match Quality |
|------|----------|------|-----------|----------------|---------------|
| 0 | `claude-security-hooks/internal/invariants/registry.go` | type-defs (shared) | pure-data | kubernetes apimachinery `pkg/util/validation/field/errors.go` (`field.ErrorList`) [external] | role-match (errors-as-values) |
| 0 | `claude-security-hooks/internal/invariants/fsutil.go` | utility | file-I/O | stdlib `os.Stat` + `bufio.Scanner` over `os.Open` [stdlib] | exact-pattern (stdlib idiom) |
| 0 | `claude-security-hooks/internal/invariants/fsutil_test.go` | test | request-response | RESEARCH.md Pattern 4 (table-driven `cmp.Diff`) lines 411-482 | exact-pattern |
| 0 | `claude-security-hooks/internal/invariants/h5_deps_test.go` | meta-test | transform (go-list parse) | RESEARCH.md § H5 meta-test lines 655-708 (verbatim) | exact-pattern (use as-is) |
| 0 | `claude-security-hooks/internal/invariants/h6_coverage_test.go` | meta-test | AST walk | RESEARCH.md § H6 meta-test lines 712-775 (verbatim) | exact-pattern (use as-is) |
| 0 | `claude-security-hooks/internal/invariants/testdata/workspace/` (fixture dir + a few `.go` files at known line counts) | fixture | static | RESEARCH.md § Wave 0 Gaps line 913; Validation Architecture A7/A8/T9 rows | no analog — invent minimal fixture |
| 0 | `claude-security-hooks/internal/invariants/testdata/graphify-out/graph.json` | fixture | static JSON | RESEARCH.md A6 row line 820 | no analog — invent minimal valid graph |
| 0 | `claude-security-hooks/internal/invariants/testdata/synthesis_out/{review-report.json, review-report.md}` | fixture | static | RESEARCH.md S1 row line 879 | no analog — invent minimal pair |
| 0 | `claude-security-hooks/internal/hooks/doc.go` | package-doc | none (godoc) | stdlib `doc.go` convention (any stdlib `doc.go`) | exact-pattern (idiomatic) |
| 0 | `claude-security-hooks/internal/hooks/events.go` | type-defs | request-response | RESEARCH.md § events.go example lines 597-651 (verbatim) | exact-pattern (use as-is) |
| 0 | `claude-security-hooks/internal/hooks/events_test.go` | test | request-response | RESEARCH.md Pattern 4 + § Pitfall 3 (separate event-struct tests) | role-match |
| 0 | `claude-security-hooks/internal/hooks/decision.go` | utility | transform | D-03 format spec (CONTEXT.md lines 54, 168); no external analog needed | derive from spec |
| 0 | `claude-security-hooks/internal/hooks/decision_test.go` | test | transform | RESEARCH.md Pattern 4 | exact-pattern |
| 0 | `claude-security-hooks/internal/hooks/agents.go` | type-defs (const) | none | D-07 (CONTEXT.md lines 62-63); no external analog needed | derive from spec |
| 0 | `claude-security-hooks/internal/hooks/agents_test.go` | drift-test | none | RESEARCH.md § Specific Ideas line 169 (`TestSecurityAgentSet_MatchesSpec`) | exact-pattern |
| 0 | `claude-security-hooks/internal/schema/doc.go` | package-doc | none | Pitfall 2 (RESEARCH.md lines 532-540) — D-10 contract loud documentation | derive from D-10 |
| 0 | `claude-security-hooks/internal/schema/cartographer.go` | type-defs | static | CON-schema-cartographer (HAND_OFF §3.1) — planner reads HAND_OFF for field names | role-match (DTO) |
| 0 | `claude-security-hooks/internal/schema/taint_tracer.go` | type-defs | static | RESEARCH.md Pattern 2 line 433 (`schema.TaintVerdict` / `TaintPathStep` shape) + CON-schema-taint | role-match (DTO) |
| 0 | `claude-security-hooks/internal/schema/authz_tracer.go` | type-defs | static | CON-schema-authz (HAND_OFF §3.3) | role-match (DTO) |
| 0 | `claude-security-hooks/internal/schema/oauth_auditor.go` | type-defs | static | CON-schema-oauth (HAND_OFF §3.4) | role-match (DTO) |
| 0 | `claude-security-hooks/internal/schema/invariant_checker.go` | type-defs | static | CON-schema-invariant (HAND_OFF §3.5) | role-match (DTO) |
| 0 | `claude-security-hooks/internal/schema/synthesis.go` | type-defs | static | CON-schema-synthesis (HAND_OFF §3.6, review-report/v1) | role-match (DTO) |
| 0 | `claude-security-hooks/Makefile` | build-config | none | D-13 (CONTEXT.md line 83); H1 verification step (RESEARCH.md line 902) | derive from spec |
| 1 | `claude-security-hooks/internal/invariants/cartographer.go` | registry + predicates | pure-function | RESEARCH.md Pattern 2 lines 327-388 (`TaintInvariant` pattern, rebadged) | exact-pattern (template) |
| 1 | `claude-security-hooks/internal/invariants/cartographer_test.go` | test (11 funcs A1..A11) | request-response | RESEARCH.md Pattern 4 lines 411-482 | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/taint_tracer.go` | registry + predicates | pure-function | RESEARCH.md Pattern 2 lines 350-388 (verbatim) | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/taint_tracer_test.go` | test (11 funcs T1..T11) | request-response | RESEARCH.md Pattern 4 lines 411-482 (verbatim) | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/authz_tracer.go` | registry + predicates | pure-function | RESEARCH.md Pattern 2 (template) | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/authz_tracer_test.go` | test (6 funcs AZ1..AZ6) | request-response | RESEARCH.md Pattern 4 | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/oauth_auditor.go` | registry + predicates | pure-function | RESEARCH.md Pattern 2 (template) | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/oauth_auditor_test.go` | test (7 funcs OA1..OA7) | request-response | RESEARCH.md Pattern 4 | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/invariant_checker.go` | registry + predicates | pure-function | RESEARCH.md Pattern 2 (template) | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/invariant_checker_test.go` | test (4 funcs IC1..IC4) | request-response | RESEARCH.md Pattern 4 | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/synthesis.go` | registry + predicates | pure-function + file-I/O (S1) | RESEARCH.md Pattern 2 + fsutil for S1 dir check | exact-pattern |
| 1 | `claude-security-hooks/internal/invariants/synthesis_test.go` | test (6 funcs S1..S6) | request-response | RESEARCH.md Pattern 4 + RESEARCH.md S1 row line 879 | exact-pattern |
| 2 | `claude-security-hooks/cmd/claude-security-hooks/main.go` | entrypoint / dispatch | request-response | RESEARCH.md Pattern 1 lines 276-325 (verbatim) | exact-pattern (use as-is) |
| 2 | `claude-security-hooks/internal/hooks/preflight.go` | subcommand body | request-response | D-09 (CONTEXT.md lines 70-72) + RESEARCH.md Pattern 3 lines 396-410 | role-match (compose Pattern 3) |
| 2 | `claude-security-hooks/internal/hooks/preflight_test.go` | integration unit | request-response | RESEARCH.md Pattern 4 + Validation Architecture H2/H3 rows lines 891-892 | exact-pattern |
| 2 | `claude-security-hooks/internal/hooks/validate.go` | subcommand body | request-response | CON-validate-protocol (HAND_OFF §3.7) + RESEARCH.md Pattern 3 + Pitfall 4 (`tool_response.content` handling) lines 552-559 | role-match (compose Pattern 3 + Pitfall 4) |
| 2 | `claude-security-hooks/internal/hooks/validate_test.go` | integration unit | request-response | RESEARCH.md Pattern 4 + Validation Architecture H4/H7 rows lines 893, 896 | exact-pattern |
| 2 | `claude-security-hooks/internal/hooks/inject.go` | subcommand body (stub) | request-response | D-08 (CONTEXT.md lines 65-67) — read+validate `SubagentStartEvent`, exit 0 silent | derive from D-08 |
| 2 | `claude-security-hooks/internal/hooks/inject_test.go` | integration unit | request-response | RESEARCH.md Pattern 4 | exact-pattern |
| mod | `claude-security-hooks/go.mod` (modified) | manifest | none | existing 3 lines preserved; add `require github.com/google/go-cmp v0.7.0 // test-only` after `go get` | derive |
| new | `claude-security-hooks/go.sum` (generated) | lock | none | auto-generated by `go mod tidy` after Wave-0 `go get github.com/google/go-cmp@v0.7.0` | n/a |

## Pattern Assignments

### Wave 0 — Test infrastructure & shared types

---

### `internal/invariants/registry.go` (type-defs, pure-data)

**Canonical analog:** kubernetes apimachinery `field.ErrorList` (errors-as-values returned from imperative checks) — cited in RESEARCH.md line 392 as `[CITED: pkg.go.dev/k8s.io/apimachinery/pkg/util/validation/field]`. Pull excerpt verbatim from RESEARCH.md Pattern 2 § lines 330-348.

**Verbatim code to copy** (RESEARCH.md lines 330-348):
```go
// internal/invariants/registry.go
package invariants

type Severity string

const (
    SeverityCritical Severity = "critical"
    SeverityHigh     Severity = "high"
    SeverityMedium   Severity = "medium"
    SeverityLow      Severity = "low"
    SeverityInfo     Severity = "info"
)

// Violation is the shared violation shape; agents do NOT share Check signatures.
type Violation struct {
    Path     string // e.g. "path[0].step", or "" for boolean-style
    Expected string
    Actual   string
}
```

**No imports needed.** No I/O. Pure data types.

**Tests-first reminder (D-04 / CLAUDE.md TDD rule):** registry.go has no behavior to test directly, but its types are exercised by every per-agent `_test.go`. The TDD anchor is `internal/invariants/registry_test.go` (mentioned in CONTEXT.md D-04 line 55) which iterates each per-agent registry asserting ID set equals expected set — this is logically the same as the H6 coverage test but registry-shape focused. Decide in planning whether to merge it into `h6_coverage_test.go` or keep separate.

---

### `internal/invariants/fsutil.go` (utility, file-I/O)

**Canonical analog:** stdlib idiom — `os.Stat` for existence, `bufio.Scanner` over `os.Open` counting `Scan()` calls for line count. RESEARCH.md § "Don't Hand-Roll" line 502 explicitly prescribes `bufio.Scanner` over `os.ReadFile + strings.Count` because the former is memory-bounded.

**Required exports** (per RESEARCH.md Wave 0 Gaps line 909):
```go
package invariants

import (
    "bufio"
    "os"
)

// FileExists reports whether path resolves to an existing file (not a directory).
// Used by A7 / S1 (file-existence predicates).
func FileExists(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return !info.IsDir()
}

// LineCount returns the number of newline-delimited lines in the file at path.
// Used by A8 / T9 / AZ4 / OA4 / IC3 (every (file,line) citation predicate).
// Streams via bufio.Scanner; memory-bounded for large source files.
func LineCount(path string) (int, error) {
    f, err := os.Open(path)
    if err != nil {
        return 0, err
    }
    defer f.Close()

    n := 0
    sc := bufio.NewScanner(f)
    // Allow long lines (generated code, minified JSON in fixtures).
    sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
    for sc.Scan() {
        n++
    }
    return n, sc.Err()
}
```

**Workspace-root helper (D-11 + Pitfall 6, RESEARCH.md lines 571-580):** add a separate helper that prefers `$CLAUDE_PROJECT_DIR` then falls back to CWD-marker scan (`go.mod`, `.claude/`, `.planning/`). This helper is consumed by `internal/hooks/validate.go` to set CWD for fsutil reads.
```go
// ResolveWorkspaceRoot returns the workspace root or "" if neither
// $CLAUDE_PROJECT_DIR nor a recognizable CWD marker is found.
// Caller treats "" as: emit block reason "workspace_root_not_found".
func ResolveWorkspaceRoot() string {
    if dir := os.Getenv("CLAUDE_PROJECT_DIR"); dir != "" {
        if FileExists(dir + "/go.mod") || FileExists(dir + "/.claude") || FileExists(dir + "/.planning") {
            return dir
        }
    }
    // Fall back to CWD marker scan.
    if cwd, err := os.Getwd(); err == nil {
        if FileExists(cwd + "/go.mod") || FileExists(cwd + "/.claude") || FileExists(cwd + "/.planning") {
            return cwd
        }
    }
    return ""
}
```
(Note: `FileExists` works on regular files only; for directories like `.claude/` the planner may either add a `DirExists` companion or relax `FileExists` to "stat succeeds." Recommend a 2nd helper `DirExists` for clarity. The planner picks; either is sufficient.)

**Tests-first:** `fsutil_test.go` covers: existing file, missing file, directory (FileExists returns false), zero-byte file (LineCount = 0), single-line-no-newline file (LineCount = 1 via Scanner semantics — verify Scanner counts the unterminated final line), missing-file path (LineCount returns error). Workspace root: env-set+valid, env-set+missing-markers, env-unset+CWD-valid, env-unset+CWD-invalid.

---

### `internal/invariants/h5_deps_test.go` (meta-test, transform)

**Canonical analog:** RESEARCH.md § "Example: H5 meta-test" lines 655-708. **Use the example verbatim** — it is a complete, working meta-test for the H5 invariant.

**Verbatim source** (RESEARCH.md lines 657-707) — copy as-is into the file. Key contract the planner must preserve:
- `modulePrefix = "github.com/saghaulor/claude-security-hooks/"`
- `allowedTest = "github.com/google/go-cmp/"`
- Iterates `go list -deps -test -json ./...` output as a JSON stream (each package one object).
- Skips stdlib (`p.Standard == true`).
- Skips own module.
- Allows `go-cmp` ONLY if `p.ForTest != ""`; flags it as an H5 violation otherwise.
- Any other non-stdlib dep is an unconditional H5 violation.

**Tests-first:** this IS the test. Wave 0 lands h5_deps_test.go BEFORE any non-test code (per RESEARCH.md Pitfall 5 line 567: "Write the H5 meta-test FIRST"). The test will initially pass against an empty module (no deps); it remains green throughout Wave 1/2 as long as nobody imports a non-stdlib package outside `_test.go`.

**Risk:** the test runs `os/exec` to invoke `go list`. The H5 test itself uses `os/exec` from stdlib — that's stdlib, so it doesn't violate H5. But it does mean H5 needs Go installed at test time (CI invariant).

---

### `internal/invariants/h6_coverage_test.go` (meta-test, AST walk)

**Canonical analog:** RESEARCH.md § "Example: H6 meta-test sketch" lines 712-775. Use verbatim with minor extension.

**Verbatim source** (RESEARCH.md lines 714-774). Key shape:
- `expectedIDs()` returns `map[testfilename][]assertionID`. Hardcoded to A1..A11, T1..T11, AZ1..AZ6, OA1..OA7, IC1..IC4, S1..S6.
- Uses `ids("A", 1, 11)` helper (planner must define: `func ids(prefix string, lo, hi int) []string { ... }` returning `["A1", ..., "A11"]`).
- Parses each per-agent `_test.go` via `parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)`.
- For each ID, requires at least one `*ast.FuncDecl` whose name starts with `"Test<ID>_"`.
- Also asserts registry IDs match expected set via `assertRegistryIDs` + `idList` helpers (planner-implemented; trivial: `for _, inv := range reg { out = append(out, inv.ID) }`).

**Tests-first:** lands in Wave 0 as a SCAFFOLD (with empty `expectedIDs` map or all expected sets present-but-empty so it passes initially). As each Wave 1 task adds a registry + Test* functions, the expected set entries are flipped on. This is the H6 "ratchet" — once an ID is in `expectedIDs`, drift fails CI loudly.

**Alternative considered:** running `go test -list '.*'` and parsing stdout (rejected in RESEARCH.md "Don't Hand-Roll" line 500 — "AST walk is pure stdlib and runs in-process").

---

### `internal/invariants/testdata/workspace/` (fixture directory)

**No analog** — invent minimal fixture. Minimum contents per RESEARCH.md Validation Architecture rows for A7/A8/T9/AZ4/OA4/IC3:
- At least 2 `.go` files at known line counts (e.g., `handler.go` with exactly 25 lines, `service.go` with exactly 10 lines).
- One file referenced as "missing" — i.e., a fixture path `missing.go` that does NOT exist, used by A7-negative tests.
- One zero-byte file `empty.go` for edge-case LineCount = 0.
- One file with no trailing newline `noeol.go` (single line, no `\n`) for Scanner-semantic verification.

**Planner discretion:** exact line counts and contents. Document each file's expected line count in a `README.md` inside `testdata/workspace/` (single-purpose dir; OK to add a README even though the global rule prefers code) — call this `LINECOUNTS.txt` if README feels heavy.

---

### `internal/invariants/testdata/graphify-out/graph.json` (fixture, static JSON)

**No analog** — invent minimal valid graph for A6 tests (RESEARCH.md line 820: "Cited graph node IDs exist in `graphify-out/graph.json` — 1 pos + 2 neg (missing node, missing graph file)").

**Minimal shape** (cartographer's go-index/v1 will reference these node IDs):
```json
{
  "schema_version": "graphify/v1",
  "nodes": [
    {"id": "pkg/main.main", "kind": "func"},
    {"id": "pkg/handler.Login", "kind": "func"}
  ]
}
```

The planner extracts the actual schema from CON-schema-cartographer / HAND_OFF §3.1; the JSON above is illustrative shape only.

---

### `internal/invariants/testdata/synthesis_out/{review-report.json, review-report.md}` (fixture pair)

**No analog** — invent minimal pair for S1 (RESEARCH.md line 879: "Both `review-report.json` and `review-report.md` exist — predicate takes dir path; `testdata/synthesis_out/`").

**Minimal:** an empty `.md` file (0 bytes is fine — S1 only checks existence) and a `review-report.json` matching the review-report/v1 schema's minimum required fields per CON-schema-synthesis. Two sibling fixture dirs: one with both files (S1-positive), one missing the `.md` (S1-negative for "md missing"), one missing the `.json` (S1-negative for "json missing").

---

### `internal/hooks/doc.go` (package-doc)

**Canonical analog:** any stdlib `doc.go` (e.g., `os/doc.go`). Standard Go idiom: file-level comment ending with `package hooks`.

**Content driver:** D-14 two-class exit-code policy (CONTEXT.md lines 86-89) plus a brief one-line on each subcommand. Specifics referenced in CONTEXT.md line 89: "Documented in `internal/hooks/doc.go` as the binary's exit-code contract." Also reference Pitfall 2 (RESEARCH.md lines 532-540) — D-10 unknown-field strictness contract documented for schema bumps.

**Sketch** (planner refines wording):
```go
// Package hooks implements the three Claude Code lifecycle hook subcommands
// for the claude-security-hooks binary: preflight, validate, and inject-context.
//
// # Exit-code contract (D-14)
//
// Class A — validation block decisions:
//
//   Anything that says "the security agent did something I cannot accept"
//   emits {"decision":"block","reason":"..."} to stdout and exits 0.
//   Covers all 51 per-agent assertion violations (A*, T*, AZ*, OA*, IC*, S*),
//   schema parse failures (H7), preflight input mismatch (D-09),
//   unknown-field rejections (D-10), and workspace-root-not-found (D-11).
//
// Class B — infrastructure failures:
//
//   stdin read errors, broken pipes, OS-level failures (e.g., cannot stat
//   workspace at all when needed) write a human-readable line to stderr
//   and exit 1. These signal "the hook itself is broken," not "the agent
//   did something bad."
//
// # Strict JSON contract (D-10)
//
// All schema decoders use json.NewDecoder + DisallowUnknownFields. Any new
// field in an agent verdict triggers a block decision. Schema-bump protocol:
// when a verdict gains a field, both the agent prompt AND the matching
// internal/schema/<agent>.go struct must change in the same commit.
package hooks
```

**Tests-first:** doc.go has no testable behavior. The exit-code contract IS tested by `preflight_test.go`, `validate_test.go`, `inject_test.go`, and the H4/H7 integration tests.

---

### `internal/hooks/events.go` (type-defs, request-response)

**Canonical analog:** RESEARCH.md § "Example: events.go (the live payload shape)" lines 597-651. **Use verbatim.**

**Verbatim source** (RESEARCH.md lines 597-651). Critical field-name note from RESEARCH.md line 653: SubagentStartEvent uses `agent_type` (live docs), NOT `subagent_type` (HAND_OFF §8.1 stale). Also: HAND_OFF says `agent_name`; live docs say `agent_type`. **Use `agent_type` per RQ-1.**

**Test contract** (events_test.go): JSON round-trip for each of the three event structs, covering:
- PreToolUseEvent with `tool_name: "Task"`, `tool_input.subagent_type: "go-taint-tracer"`, `tool_input.prompt: "..."`.
- PostToolUseEvent same + `tool_response.content` as both a string AND an array (Pitfall 4, RESEARCH.md lines 552-559).
- SubagentStartEvent with `agent_type` (NOT `agent_name`).
- Unknown-field rejection: a payload with `tool_input.unexpected_field` must fail to decode (D-10 verification at events-layer too).

---

### `internal/hooks/decision.go` (utility, transform)

**No analog** — spec-derived. Formats per D-03 (CONTEXT.md lines 54, 168):

**Required exports:**
```go
package hooks

import (
    "encoding/json"
    "fmt"
    "io"
    "strings"

    "github.com/saghaulor/claude-security-hooks/internal/invariants"
)

// FormatViolation renders a single Violation to the D-03 block-reason format.
//   Path-style: "<ID>: <Path> expected '<Expected>' got '<Actual>'"
//   Boolean-style (Path == ""): "<ID>: <Description>"  (Description comes from the registry)
// The caller supplies (id, description); decision.go does not look them up.
func FormatViolation(id, description string, v invariants.Violation) string {
    if v.Path != "" {
        return fmt.Sprintf("%s: %s expected '%s' got '%s'", id, v.Path, v.Expected, v.Actual)
    }
    return fmt.Sprintf("%s: %s", id, description)
}

// FormatBlockReason joins per-violation strings with "; " in registration order.
func FormatBlockReason(parts []string) string {
    return strings.Join(parts, "; ")
}

// EmitBlock writes the {"decision":"block","reason":"..."} JSON to w.
// Caller is responsible for exiting 0 after the write (D-14 Class A).
func EmitBlock(w io.Writer, reason string) error {
    return json.NewEncoder(w).Encode(struct {
        Decision string `json:"decision"`
        Reason   string `json:"reason"`
    }{Decision: "block", Reason: reason})
}
```

**Tests-first:** decision_test.go covers: path-style format, boolean-style format (Path == ""), multi-violation join order is registration order (NOT sorted), EmitBlock writes single-line valid JSON with no trailing fields.

---

### `internal/hooks/agents.go` (type-defs / const)

**No analog** — D-07 spec verbatim (CONTEXT.md lines 62-63):

```go
package hooks

// SecurityAgentSet is the source of truth for which subagent_type values
// trigger validation. Hardcoded per D-07; drift test in agents_test.go.
var SecurityAgentSet = []string{
    "go-cartographer",
    "go-taint-tracer",
    "go-authz-tracer",
    "go-oauth-auditor",
    "invariant-checker",
    "synthesis",
}

// IsSecurityAgent reports whether the given subagent_type value is one
// the security-hooks binary is responsible for validating.
func IsSecurityAgent(subagentType string) bool {
    for _, name := range SecurityAgentSet {
        if name == subagentType {
            return true
        }
    }
    return false
}
```

**Tests-first:** `agents_test.go` defines `TestSecurityAgentSet_MatchesSpec` (RESEARCH.md line 169 + CONTEXT.md line 63):
- Hard-codes the expected slice verbatim (forces visibility: changing the set requires editing BOTH agents.go AND the test).
- Asserts `cmp.Diff(expected, hooks.SecurityAgentSet) == ""`.
- Also tests `IsSecurityAgent("general-purpose") == false` and `IsSecurityAgent("go-taint-tracer") == true`.

---

### `internal/schema/doc.go` (package-doc)

**Driver:** Pitfall 2 (RESEARCH.md lines 532-540) — "document this contract loudly." Mirror the D-10 strict-decode contract documented in `hooks/doc.go` but oriented to schema-bump protocol.

**Sketch:**
```go
// Package schema declares typed Go structs mirroring each security
// subagent's verdict JSON schema. All decoders in internal/hooks use
// json.NewDecoder(input).DisallowUnknownFields() per D-10; therefore
// any new field in an agent's verdict that is not declared on the
// corresponding struct here causes the validator to block at runtime.
//
// Schema-bump protocol: when a verdict gains a field, the agent prompt
// AND the matching struct here MUST change in the same commit.
//
// Field names are JSON-tag-driven; Go field names are CamelCase mirrors
// of the snake_case JSON keys defined by CON-schema-* in
// .planning/intel/constraints.md.
package schema
```

---

### `internal/schema/<agent>.go` (6 files: cartographer, taint_tracer, authz_tracer, oauth_auditor, invariant_checker, synthesis)

**Driver:** CON-schema-* constraints from HAND_OFF §3.1–§3.6. The planner reads each `§3.x` for the exact JSON schema and translates to Go structs with JSON tags. RESEARCH.md gives a fragment-level analog only for `TaintVerdict` (Pattern 2 line 433):
```go
in: &schema.TaintVerdict{
    Verdict: "exploitable",
    Path: []schema.TaintPathStep{{Step: "source"}, {Step: "call"}, {Step: "sink"}},
},
```

**Pattern to apply uniformly across all 6 schema files:**

```go
package schema

// <AgentName>Verdict mirrors CON-schema-<agent> per HAND_OFF §3.x.
// JSON tags use snake_case keys; Go names are CamelCase.
// json.NewDecoder + DisallowUnknownFields is enforced at the call site
// (internal/hooks/validate.go); do NOT add custom UnmarshalJSON methods.
type TaintVerdict struct {
    SchemaVersion string          `json:"schema_version"` // e.g. "taint/v1"
    Verdict       string          `json:"verdict"`        // exploitable | sanitized | unreachable | ambiguous | input_mismatch
    Confidence    string          `json:"confidence"`     // high | medium | low
    Path          []TaintPathStep `json:"path"`
    Semgrep       SemgrepEvidence `json:"semgrep"`
    Gopls         GoplsEvidence   `json:"gopls"`
    Notes         string          `json:"notes,omitempty"`
    // ... other fields per HAND_OFF §3.2
}

type TaintPathStep struct {
    Step string `json:"step"` // source | call | sink
    File string `json:"file"`
    Line int    `json:"line"`
    // ... other fields per HAND_OFF §3.2
}

// ... SemgrepEvidence, GoplsEvidence etc.
```

**Critical:** every nested struct (`TaintPathStep`, etc.) inherits the strict-decode behavior because `DisallowUnknownFields` recurses (verified RESEARCH.md line 409). Do NOT inline `map[string]any` — that defeats D-10. Always declare every field explicitly.

**Tests-first:** schema structs themselves don't need a dedicated test file; they are exercised by the per-agent registry tests in Wave 1 and the events_test.go round-trip in Wave 0. Planner may add a thin `schema/<agent>_test.go` that round-trips a minimal valid JSON to catch tag typos early — recommended for cartographer + taint_tracer (the most complex schemas) at minimum.

---

### `claude-security-hooks/Makefile`

**Driver:** D-13 (CONTEXT.md line 83). No external analog needed — Makefile is mechanical.

**Required targets:**
```makefile
.PHONY: build test install clean lint

BIN := bin/claude-security-hooks
HOOK_DIR := ../.claude/hooks/bin

build:
	CGO_ENABLED=0 go build -o $(BIN) ./cmd/claude-security-hooks

test:
	go test ./...

install: build
	mkdir -p $(HOOK_DIR)
	cp $(BIN) $(HOOK_DIR)/claude-security-hooks

clean:
	rm -rf bin

lint:
	go vet ./...
```

**Optional H1 smoke step** (RESEARCH.md line 902 recommends adding an `ldd` check to verify "not a dynamic executable" after `build`). Add as a separate target so the default `build` stays clean:
```makefile
.PHONY: verify-static
verify-static: build
	@ldd $(BIN) 2>&1 | grep -q "not a dynamic executable" \
		|| (echo "H1 FAILED: binary is dynamically linked" && exit 1)
```
(Note: `ldd` returns non-zero on static binaries on some Linux distros — alternative is `file $(BIN) | grep "statically linked"`. Planner picks based on CI environment.)

**No `go generate` step** (CONTEXT.md line 83: "No `go generate` step (would force Phase 3 dependency)"). Do NOT add one.

---

### Wave 1 — Per-agent registries (6 plans, parallelizable)

All 6 per-agent registries follow the **same shape** (`Pattern 2`). The only differences are:
- Registry type name (`TaintInvariant` vs `CartographerInvariant` vs …)
- Verdict struct type (`*schema.TaintVerdict` vs `*schema.CartographerIndex` vs …)
- ID prefix (`T`, `A`, `AZ`, `OA`, `IC`, `S`)
- Per-ID predicate body (driven by per-row Behavior column in RESEARCH.md Validation Architecture lines 813-885)

---

### `internal/invariants/taint_tracer.go` (registry + 11 predicates)

**Canonical analog:** RESEARCH.md Pattern 2 lines 350-388 — **use verbatim** for the registry shape and `checkT3` as the reference predicate. Other predicates (T1, T2, T4–T11) follow the same shape with different bodies.

**Verbatim source** (RESEARCH.md lines 350-388):
```go
// internal/invariants/taint_tracer.go
package invariants

import "github.com/saghaulor/claude-security-hooks/internal/schema"

type TaintInvariant struct {
    ID          string   // "T3"
    Description string   // "sanitized/exploitable requires non-empty path with source first and sink last"
    Severity    Severity
    Check       func(*schema.TaintVerdict) []Violation // typed per-agent: no `any`
}

var TaintTracerInvariants = []TaintInvariant{
    {ID: "T1", Description: "output is valid JSON conforming to verdict schema", Severity: SeverityCritical, Check: checkT1},
    {ID: "T2", Description: "verdict in {exploitable,sanitized,unreachable,ambiguous,input_mismatch}", Severity: SeverityCritical, Check: checkT2},
    // ... T3 .. T11
}

func checkT3(v *schema.TaintVerdict) []Violation {
    if v.Verdict != "sanitized" && v.Verdict != "exploitable" {
        return nil
    }
    var out []Violation
    if len(v.Path) == 0 {
        out = append(out, Violation{Path: "path", Expected: "non-empty", Actual: "empty"})
        return out
    }
    if v.Path[0].Step != "source" {
        out = append(out, Violation{Path: "path[0].step", Expected: "source", Actual: v.Path[0].Step})
    }
    if last := v.Path[len(v.Path)-1]; last.Step != "sink" {
        out = append(out, Violation{
            Path:     fmt.Sprintf("path[%d].step", len(v.Path)-1),
            Expected: "sink", Actual: last.Step,
        })
    }
    return out
}
```

(Note: the verbatim snippet uses `fmt.Sprintf` but the file doesn't import `fmt`. Planner must add `import "fmt"` alongside `internal/schema`.)

**Tests-first:** `taint_tracer_test.go` uses RESEARCH.md Pattern 4 lines 411-482 (`TestT3_PathNonEmptySourceFirstSinkLast`) **verbatim** as the canonical example. Repeat the shape for T1, T2, T4–T11. Each test uses `cmpopts.EquateEmpty()` per RESEARCH.md line 482.

**Per-ID test minimums** (RESEARCH.md lines 829-841): T1 (1 pos + 2 neg), T2 (5 pos + 1 neg), T3 (1 pos + 3 neg), T4 (1 pos), T5 (3 pos + 1 neg), T6 (2 pos + 1 neg + 1 pos exemption), T7 (joint input+output: 1 pos + 1 neg — predicate signature MUST accept both verdict AND input), T8 (joint: 1+1+1), T9 (filesystem: 1 pos + 1 neg using `testdata/workspace/`), T10 (2 pos + 1 neg), T11 (1 pos + 1 neg).

**Special case T7/T8:** RESEARCH.md says these are "joint input+output" — the predicate signature differs from the others. Planner decides: either (a) introduce a parallel type `type TaintInputJointInvariant struct { Check func(in *schema.TaintInput, out *schema.TaintVerdict) []Violation }` in a separate slice `TaintTracerJointInvariants`, or (b) overload the verdict struct to carry the input as a field. **Recommendation: option (a)** — keeps types narrow per D-02 no-`any` rule. The H6 meta-test expectedIDs map must then list T7/T8 in BOTH registries' ID sets to ensure coverage.

---

### `internal/invariants/cartographer.go` (registry + 11 A* predicates)

**Pattern:** identical to taint_tracer.go above; substitute `CartographerInvariant`, `*schema.CartographerIndex`, IDs A1..A11.

**Per-ID details** (RESEARCH.md lines 815-825):
- A1–A5, A9, A11: inline JSON — pure predicates over the verdict struct.
- A6: filesystem fixture (`testdata/graphify-out/graph.json`) — predicate reads the graph and checks cited node IDs.
- A7, A8: use `fsutil.FileExists` / `fsutil.LineCount` from registry.go's package. CWD assumption per D-11.
- A10: special — predicate inspects the verdict's optional `tool_calls` field (if the agent emitted one), asserts no Write/Edit tools called. If the field is absent from the verdict, A10 is satisfied vacuously (per RESEARCH.md line 824).

**Tests-first:** `cartographer_test.go` with 11 `TestA<n>_*` functions. A6 + A7 + A8 tests must point to fixtures under `testdata/`.

---

### `internal/invariants/authz_tracer.go` (registry + 6 AZ* predicates)

**Pattern:** identical shape; IDs AZ1..AZ6; `*schema.AuthzVerdict`.

**Joint-input predicates:** AZ2, AZ4, AZ5 are joint (take input + output per RESEARCH.md lines 848-852). Same option-(a) pattern as T7/T8: separate `AuthzJointInvariant` slice if planner prefers strict typing.

---

### `internal/invariants/oauth_auditor.go` (registry + 7 OA* predicates)

**Pattern:** identical; IDs OA1..OA7; `*schema.OAuthVerdict`. OA4–OA7 joint.

---

### `internal/invariants/invariant_checker.go` (registry + 4 IC* predicates)

**Pattern:** identical; IDs IC1..IC4; `*schema.InvariantCheckerVerdict`. IC2, IC4 joint.

---

### `internal/invariants/synthesis.go` (registry + 6 S* predicates)

**Pattern:** identical; IDs S1..S6; `*schema.SynthesisReport` (the review-report/v1 struct).

**Special case S1:** filesystem predicate that takes a directory path (not a verdict struct). Planner decides whether to extend `SynthesisInvariant.Check` to accept `(report *schema.SynthesisReport, dirPath string)` or introduce a parallel `SynthesisDirInvariant` slice. **Recommendation:** parallel slice (cleaner typing per D-02) — `SynthesisDirInvariants []SynthesisDirInvariant` with `Check func(dirPath string) []Violation`. The S1 predicate uses `fsutil.FileExists(filepath.Join(dirPath, "review-report.json"))` AND `fsutil.FileExists(filepath.Join(dirPath, "review-report.md"))`.

---

### Wave 2 — Wiring (subcommands + main + meta-tests)

---

### `cmd/claude-security-hooks/main.go` (entrypoint / dispatch)

**Canonical analog:** RESEARCH.md Pattern 1 lines 276-322 — **use verbatim.** Source cited as `gobyexample.com/command-line-subcommands` (line 325).

**Verbatim source** (RESEARCH.md lines 278-322):
```go
// cmd/claude-security-hooks/main.go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/saghaulor/claude-security-hooks/internal/hooks"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "usage: claude-security-hooks <preflight|validate|inject-context>")
        os.Exit(2)
    }
    sub, rest := os.Args[1], os.Args[2:]

    var (
        fs  *flag.FlagSet
        err error
    )
    switch sub {
    case "preflight":
        fs = flag.NewFlagSet("preflight", flag.ExitOnError)
        _ = fs.Parse(rest)
        err = hooks.Preflight(os.Stdin, os.Stdout, os.Stderr)
    case "validate":
        fs = flag.NewFlagSet("validate", flag.ExitOnError)
        _ = fs.Parse(rest)
        err = hooks.Validate(os.Stdin, os.Stdout, os.Stderr)
    case "inject-context":
        fs = flag.NewFlagSet("inject-context", flag.ExitOnError)
        _ = fs.Parse(rest)
        err = hooks.InjectContext(os.Stdin, os.Stdout, os.Stderr)
    default:
        fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", sub)
        os.Exit(2)
    }
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**Required signatures** on `internal/hooks` exports (constrains preflight/validate/inject .go shape):
- `func Preflight(stdin io.Reader, stdout io.Writer, stderr io.Writer) error`
- `func Validate(stdin io.Reader, stdout io.Writer, stderr io.Writer) error`
- `func InjectContext(stdin io.Reader, stdout io.Writer, stderr io.Writer) error`

All three return non-nil `error` for Class B failures (main exits 1). Class A blocks are written to `stdout` and return `nil`.

**Tests-first:** main.go is typically tested via `os.Exec`-style integration tests at the wave merge. A thin unit test (`cmd/claude-security-hooks/main_test.go` is fine to skip per Go convention) is optional. Coverage of the dispatch logic comes through preflight/validate/inject `_test.go` calling the package-level functions directly.

---

### `internal/hooks/preflight.go` (subcommand body)

**Drivers:** D-09 (CONTEXT.md lines 70-72) + RESEARCH.md Pattern 3 lines 396-407.

**Shape:**
```go
package hooks

import (
    "encoding/json"
    "fmt"
    "io"
    // ... per-agent schema imports
    "github.com/saghaulor/claude-security-hooks/internal/schema"
)

func Preflight(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
    // 1. Parse PreToolUseEvent from stdin with strict decoding (D-10).
    var ev PreToolUseEvent
    dec := json.NewDecoder(stdin)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&ev); err != nil {
        // Malformed hook event = Class A block (treat as agent did something bad).
        return EmitBlock(stdout, fmt.Sprintf("H7: PreToolUseEvent parse: %v", err))
    }

    // 2. H3 short-circuit: not a security agent → silent pass.
    if !IsSecurityAgent(ev.ToolInput.SubagentType) {
        return nil
    }

    // 3. D-09: parse ev.ToolInput.Prompt as JSON matching the per-agent input schema.
    //    On any parse failure, emit block.
    switch ev.ToolInput.SubagentType {
    case "go-taint-tracer":
        var in schema.TaintInput
        d := json.NewDecoder(strings.NewReader(ev.ToolInput.Prompt))
        d.DisallowUnknownFields()
        if err := d.Decode(&in); err != nil {
            return EmitBlock(stdout, fmt.Sprintf("D-09: taint input parse: %v", err))
        }
    case "go-cartographer":
        // ... same shape
    // ... 4 more agents
    }
    return nil
}
```

(Note: real implementation needs `strings.NewReader` import; planner adds.)

**Code-fence stripping:** D-09 doesn't strip fences in `preflight` (the prompt is supposed to be raw JSON dispatched by Claude Code, not an agent message). Code-fence stripping is `validate`-only per HAND_OFF §3.7 step 4 (CONTEXT.md line 100 "Claude's Discretion": "default to strip leading/trailing ```/```json fences").

**Tests-first:** `preflight_test.go` covers:
- H3: subagent_type not in set → returns nil, no stdout output.
- D-09 happy path: each of 6 security agents with valid input → returns nil.
- D-09 sad path: each of 6 security agents with malformed input → returns nil, stdout contains block decision with correct reason prefix.
- H7-at-event-layer: malformed PreToolUseEvent JSON → block emitted.
- Unknown-field rejection: PreToolUseEvent with extra top-level field → block emitted.

---

### `internal/hooks/validate.go` (subcommand body, the main logic)

**Drivers:** CON-validate-protocol (HAND_OFF §3.7) + RESEARCH.md Pattern 3 lines 396-407 + Pitfall 4 (`tool_response.content` polymorphism, RESEARCH.md lines 552-559) + Pitfall 6 (workspace root resolution, RESEARCH.md lines 571-580).

**Shape sketch:**
```go
func Validate(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
    // 1. Parse PostToolUseEvent.
    var ev PostToolUseEvent
    dec := json.NewDecoder(stdin)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&ev); err != nil {
        return EmitBlock(stdout, fmt.Sprintf("H7: PostToolUseEvent parse: %v", err))
    }

    // 2. H3 short-circuit.
    if !IsSecurityAgent(ev.ToolInput.SubagentType) {
        return nil
    }

    // 3. Resolve workspace root (D-11 + Pitfall 6).
    root := invariants.ResolveWorkspaceRoot()
    if root == "" {
        return EmitBlock(stdout, "workspace_root_not_found")
    }
    // (Optional: chdir to root so fsutil reads resolve. Or pass root through to predicates.)

    // 4. Extract verdict text from ev.ToolResponse.Content (Pitfall 4 — string OR array).
    verdictText, err := extractTextContent(ev.ToolResponse.Content)
    if err != nil {
        return EmitBlock(stdout, fmt.Sprintf("H7: tool_response.content not parseable as text: %v", err))
    }

    // 5. Strip code fences (Claude's Discretion, CONTEXT.md line 100).
    verdictText = stripCodeFences(verdictText)

    // 6. Per-agent dispatch: unmarshal verdict + run typed registry.
    var reasons []string
    switch ev.ToolInput.SubagentType {
    case "go-taint-tracer":
        var v schema.TaintVerdict
        d := json.NewDecoder(strings.NewReader(verdictText))
        d.DisallowUnknownFields()
        if err := d.Decode(&v); err != nil {
            return EmitBlock(stdout, fmt.Sprintf("T1: taint verdict parse: %v", err))
        }
        for _, inv := range invariants.TaintTracerInvariants {
            for _, vio := range inv.Check(&v) {
                reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
            }
        }
    case "go-cartographer":
        // ... same shape with CartographerInvariants
    // ... 4 more agents
    }

    if len(reasons) > 0 {
        return EmitBlock(stdout, FormatBlockReason(reasons))
    }
    return nil
}

// extractTextContent handles string-or-array per Pitfall 4 (RESEARCH.md lines 552-559).
func extractTextContent(raw json.RawMessage) (string, error) {
    // Try string first.
    var s string
    if err := json.Unmarshal(raw, &s); err == nil {
        return s, nil
    }
    // Try array of segments with .text or .content fields.
    var arr []map[string]json.RawMessage
    if err := json.Unmarshal(raw, &arr); err != nil {
        return "", fmt.Errorf("content is neither string nor array: %w", err)
    }
    var sb strings.Builder
    for _, seg := range arr {
        for _, key := range []string{"text", "content"} {
            if rawSeg, ok := seg[key]; ok {
                var sv string
                if err := json.Unmarshal(rawSeg, &sv); err == nil {
                    sb.WriteString(sv)
                }
            }
        }
    }
    return sb.String(), nil
}

// stripCodeFences strips leading/trailing ``` or ```json fences and whitespace.
// Default per CONTEXT.md line 100; recoverable so caller falls back to raw on parse failure.
func stripCodeFences(s string) string {
    s = strings.TrimSpace(s)
    s = strings.TrimPrefix(s, "```json")
    s = strings.TrimPrefix(s, "```")
    s = strings.TrimSuffix(s, "```")
    return strings.TrimSpace(s)
}
```

**Tests-first:** `validate_test.go` covers:
- H4: any failing invariant → exactly one `{"decision":"block","reason":"..."}` on stdout (RESEARCH.md line 893).
- H7 four cases (RESEARCH.md line 896): malformed JSON, valid JSON wrong type, empty content, multi-segment array.
- Workspace-root-not-found: env unset + cwd has no markers → block reason `"workspace_root_not_found"`.
- Code-fence stripping: ```` ```json {...} ``` ```` parses cleanly; raw `{...}` also parses cleanly.
- Per-agent happy path: one passing fixture per agent → no stdout, nil error.
- Per-agent failing path: one failing fixture per agent → block emitted with expected ID prefix.
- Multi-violation ordering: deterministic concatenation per D-03 (registration order, not sorted).

---

### `internal/hooks/inject.go` (subcommand stub)

**Driver:** D-08 (CONTEXT.md lines 65-67) — no-op stub: read SubagentStartEvent, validate it parses, exit 0 silent.

**Shape:**
```go
func InjectContext(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
    var ev SubagentStartEvent
    dec := json.NewDecoder(stdin)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&ev); err != nil {
        return EmitBlock(stdout, fmt.Sprintf("H7: SubagentStartEvent parse: %v", err))
    }
    // D-08: no further work in Phase 2. Phase 3 / Phase 5 will populate.
    _ = ev // suppress unused-var.
    return nil
}
```

**Tests-first:** `inject_test.go` covers:
- Happy path: valid SubagentStartEvent → nil error, no stdout output (the "silent" in D-08).
- Sad path: malformed JSON → block emitted (consistent with H7 contract across all three subcommands).
- Field-name verification: SubagentStartEvent uses `agent_type` (RESEARCH.md line 653); a fixture with `agent_name` (HAND_OFF stale name) MUST be rejected by `DisallowUnknownFields()`.

---

## Shared Patterns

### Pattern S1: Strict JSON decoding (D-10) — applies to every JSON parse site

**Source:** RESEARCH.md Pattern 3 lines 394-410. **Use everywhere we parse JSON** — both event-layer (PreToolUse/PostToolUse/SubagentStart) and verdict-layer (per-agent schema decode).

**Verbatim excerpt:**
```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()
if err := dec.Decode(&v); err != nil {
    return nil, fmt.Errorf("... parse: %w", err)
}
```

**Apply to:**
- `internal/hooks/events_test.go` (event round-trips)
- `internal/hooks/preflight.go` (PreToolUseEvent + each agent's input schema)
- `internal/hooks/validate.go` (PostToolUseEvent + each agent's verdict schema)
- `internal/hooks/inject.go` (SubagentStartEvent)

**Key fact** (RESEARCH.md line 409): `DisallowUnknownFields()` recurses into nested structs and array elements. Do NOT add per-field manual unknown-field guards.

---

### Pattern S2: Table-driven test with `cmp.Diff` + `cmpopts.EquateEmpty` (D-05, D-06)

**Source:** RESEARCH.md Pattern 4 lines 411-482. **Use verbatim shape for every `_test.go` in `internal/invariants/`** and the agents_test.go / decision_test.go in `internal/hooks/`.

**Verbatim test shape** (`TestT3_PathNonEmptySourceFirstSinkLast` lines 425-479):
- `cases := []struct { name string; in *schema.<X>Verdict; want []invariants.Violation }{ ... }`
- Locate the predicate from the registry by iterating the slice (NOT by directly calling `checkT3` — proves registry wiring at test time).
- `if checkT3 == nil { t.Fatal(...) }` — H6 link.
- `t.Run(tc.name, func(t *testing.T) { got := checkT3(tc.in); if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" { t.Errorf("...") } })`.

**Apply to:** all 6 per-agent `_test.go` files in `internal/invariants/`, plus `decision_test.go`, `events_test.go`, `agents_test.go`.

---

### Pattern S3: Workspace-root resolution (D-11 + Pitfall 6)

**Source:** RESEARCH.md Pitfall 6 lines 571-580. Used by every predicate that reads workspace files (A7, A8, T9, AZ4, OA4, IC3) and by `synthesis.S1` directory check.

**Apply via:** `invariants.ResolveWorkspaceRoot()` helper from `fsutil.go` (sketch above). `validate.go` calls it once at the top and either chdirs to the root or passes `root` into predicates. **Planner recommendation:** chdir at the top of `Validate()` (simpler) — predicates then use plain relative paths from the verdict citations.

---

### Pattern S4: Boundary rule — `internal/invariants/*.go` must NOT import `internal/hooks`

**Source:** RESEARCH.md line 99 (Architectural Responsibility Map "Boundary rule").

**Why:** Predicates are pure functions over typed verdict structs returning `[]Violation`. They never touch I/O. Importing `internal/hooks` would couple predicates to event parsing and JSON encoding — both irrelevant to the predicate contract.

**Enforcement:** add a simple test in `h5_deps_test.go` (or a sibling `h5_boundary_test.go`) that asserts no file under `internal/invariants/` imports `github.com/saghaulor/claude-security-hooks/internal/hooks`. Use `go/parser` for AST import inspection (same dependency as h6_coverage_test.go).

**Allowed imports for `internal/invariants/*.go`:** stdlib + `github.com/saghaulor/claude-security-hooks/internal/schema`. Nothing else.

---

### Pattern S5: Two-class exit policy (D-14)

**Source:** CONTEXT.md lines 86-89.

**Apply via:** all three subcommand functions (`Preflight`, `Validate`, `InjectContext`) return `error` for Class B (infrastructure) failures and `nil` for either "passed validation" or "Class A block emitted to stdout." `main.go` exits 1 on non-nil error, 0 on nil. Block emission is done via `decision.EmitBlock(stdout, reason)` — the function name documents the intent.

**Never panic.** REQ-hooks-H7. All errors are propagated as values; `extractTextContent` (validate.go) and JSON decoders must be wrapped in error returns, not assumed-safe calls.

---

### Pattern S6: Hardcoded sets with paired drift-detection tests (D-04, D-07)

**Source:** CONTEXT.md lines 55, 63; RESEARCH.md line 169.

**Apply to:**
- `SecurityAgentSet` (agents.go) + `TestSecurityAgentSet_MatchesSpec` (agents_test.go).
- Per-agent registry ID lists (TaintTracerInvariants, etc.) + the H6 meta-test (`h6_coverage_test.go`).
- The `expectedIDs()` map inside `h6_coverage_test.go` is itself a hardcoded set — updating any per-agent registry forces editing both the registry AND `expectedIDs`. "Forced visibility" pattern.

---

## No Analog Found (no-canonical-needed cases)

These files are spec-derived with no analog needed beyond CONTEXT.md / RESEARCH.md text:

| File | Spec source | Reason |
|------|-------------|--------|
| `internal/invariants/testdata/workspace/*` | RESEARCH.md A7/A8 rows + Wave 0 Gaps line 913 | Project-specific minimal fixture; planner invents 2–4 small Go files at known line counts |
| `internal/invariants/testdata/graphify-out/graph.json` | CON-schema-cartographer (HAND_OFF §3.1) | Minimal valid go-index/v1 instance; planner derives schema from §3.1 |
| `internal/invariants/testdata/synthesis_out/{review-report.json, review-report.md}` | CON-schema-synthesis (HAND_OFF §3.6) | Minimal review-report/v1 + an empty .md; planner derives schema from §3.6 |
| `claude-security-hooks/Makefile` | D-13 (CONTEXT.md line 83) | Mechanical Makefile; full text provided in pattern assignment above |
| `claude-security-hooks/go.mod` (modification) | RESEARCH.md line 142 (`go get github.com/google/go-cmp@v0.7.0`) | Trivial `go get` + `go mod tidy` produces the manifest delta |
| `claude-security-hooks/go.sum` (new) | auto-generated | `go mod tidy` writes this; no planner action |

## Metadata

**Analog search scope:**
- `/home/saghaulor/code/security_reviewer/claude-security-hooks/` (Phase 1 scaffold — empty except `go.mod`)
- `/home/saghaulor/code/security_reviewer/opengrep-mcp/` (sibling scaffold — empty)
- `/home/saghaulor/code/security_reviewer/examples/sample-vulnerable-service/` (empty)
- `/home/saghaulor/code/security_reviewer/` (root — only `HAND_OFF.md`, `README.md`, `CLAUDE.md`, `LICENSE`, settings)

**Files scanned:** 8 (entire non-`.git`, non-`.planning` Go-relevant footprint).
**In-repo Go files found:** 0.
**Canonical sources used:**
- RESEARCH.md § Code Examples (lines 591-775) — 5 verbatim Go snippets.
- RESEARCH.md § Recommended Project Structure (lines 223-274) — file layout.
- RESEARCH.md § Wave 0 Gaps (lines 904-922) — explicit infrastructure list.
- RESEARCH.md § Validation Architecture (lines 794-902) — per-ID test minimums.
- CONTEXT.md D-01 through D-15 — locked decisions driving spec-derived files.
- HAND_OFF.md §3.1–§3.7 + §8.1 — schema definitions and live event payload shapes (the planner reads these directly when filling per-agent struct fields).

**Pattern extraction date:** 2026-05-19.
