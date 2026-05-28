# Phase 2: claude-security-hooks Go binary - Research

**Researched:** 2026-05-18
**Domain:** Go stdlib CLI, validator engine, table-driven testing, Claude Code hook protocol
**Confidence:** HIGH (most claims VERIFIED against live docs or local probes; per-assertion test sampling marked CITED to CONTEXT.md)

---

<user_constraints>
## User Constraints (from 02-CONTEXT.md)

### Locked Decisions

- **D-01** — Per-agent registries are structured slices of typed `Invariant` records, NOT a flat `map[string]func`. Each entry carries `ID`, `Description`, `Severity`, and a typed `Check` function. The slice is iterated in registration order for deterministic block-reason output.
- **D-02** — Per the user's global CLAUDE.md no-`any` rule, each agent has its OWN registry type with its OWN typed `Check` signature — no `func(v any)` shared signature. Six agent registries: `TaintTracerInvariants []TaintInvariant`, `CartographerInvariants []CartographerInvariant`, `AuthzInvariants []AuthzInvariant`, `OAuthInvariants []OAuthInvariant`, `InvariantCheckerInvariants []InvariantCheckerInvariant`, `SynthesisInvariants []SynthesisInvariant`.
- **D-03** — Block-reason string format: `"<ID>: <Path> expected '<Expected>' got '<Actual>'"` for path-style violations; `"<ID>: <Description>"` for boolean-style. Multiple violations join with `; ` in registration order.
- **D-04** — H6 enforcement is mechanical: `internal/invariants/registry_test.go` enumerates each per-agent registry, asserts the ID set equals the expected set (`{T1..T11}`, `{A1..A11}`, etc.), and asserts every ID has a matching `Test*` function in the agent's `_test.go` file.
- **D-05** — stdlib `testing` for runners, `github.com/google/go-cmp` v0.6.0+ for `cmp.Diff` on `[]Violation` comparisons in `_test.go` files only. `go-cmp` is the ONLY non-stdlib dependency permitted, only in `_test.go`. H5 unit test enforces this via `go list -deps -test ./...`.
- **D-06** — Tests are table-driven. One `*_test.go` file per agent in `internal/invariants/`. Each test function is `TestT3_PathNonEmpty(t *testing.T)` style (one test func per assertion ID).
- **D-07** — Hardcoded `var SecurityAgentSet = []string{"go-cartographer", "go-taint-tracer", "go-authz-tracer", "go-oauth-auditor", "invariant-checker", "synthesis"}` lives in `internal/hooks/agents.go`. Drift-detection test in `agents_test.go` lists the expected set verbatim.
- **D-08** — `inject-context` is a no-op stub in Phase 2: reads the `SubagentStart` event from stdin, validates it parses against §8.1 schema, exits 0 silently. No stdout output.
- **D-09** — `preflight` rejects (block decision, exit 0) when `tool_input.subagent_type` IS in `SecurityAgentSet` but `tool_input.prompt` does not parse as JSON matching the per-agent input contract. When NOT in `SecurityAgentSet`: exit 0 silently.
- **D-10** — All `internal/schema/*` unmarshalling uses `json.NewDecoder(input)` with `dec.DisallowUnknownFields()` enabled. Unknown fields trigger a block decision.
- **D-11** — A7/A8/T9/AZ4/OA4/IC3 read source files; Phase 2 assumes CWD = workspace root, uses `os.Stat(path)` for A7 and `bufio.Scanner` over `os.Open(path)` counting `\n` for A8. Guard in `validate` emits `block` reason `"workspace_root_not_found"` if CWD doesn't contain workspace markers (`go.mod`, `.claude/`, or `.planning/`). *(Research note: hooks ALSO receive `CLAUDE_PROJECT_DIR` env var — see RQ-8 below — planner may extend the guard.)*
- **D-12** — Project layout from Phase 1 scaffold preserved: `cmd/claude-security-hooks/`, `internal/{hooks,invariants,schema}/`. Module path `github.com/saghaulor/claude-security-hooks`, `go 1.22`.
- **D-13** — `Makefile` at `claude-security-hooks/Makefile`. Targets: `build`, `test`, `install` (copies to `../.claude/hooks/bin/`), `clean`, `lint` (`go vet ./...`).
- **D-14** — Two-class error policy: Class A (validation block) → `{"decision":"block","reason":"..."}` to stdout, exit 0; Class B (infrastructure failure) → stderr message, exit 1. Documented in `internal/hooks/doc.go`.
- **D-15** — Before encoding model identifier into any test fixture, fetch `https://docs.claude.com` to verify `claude-sonnet-4-6` is current. *(Research note: VERIFIED CURRENT — see RQ-2 below.)*

### Claude's Discretion

- Internal package import paths within the module.
- Specific Go struct field names within `internal/schema/*` (must mirror JSON keys per CON-schema-*).
- Test fixture data shapes inside `_test.go`.
- Specific error message wording beyond the D-03 format.
- Whether to add a `--version` flag (defer to Phase 6).
- Code-fence stripping algorithm in `validate` (default: strip leading/trailing ```/```json fences and whitespace, attempt JSON parse, fall back to raw stdin).
- File layout granularity inside `internal/invariants/` (one file per agent unless agent file > ~400 lines).

### Deferred Ideas (OUT OF SCOPE for Phase 2)

- `--version` flag → Phase 6.
- Real `inject-context` semantic injection → Phase 3 or Phase 5.
- Generating `SecurityAgentSet` from `.claude/agents/*.md` at build time → rejected for Phase 2.
- YAML support for `.claude/security-invariants/*.yaml` → deferred indefinitely.
- CI integration → out of scope for v1.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| REQ-hooks-H1 | Static `CGO_ENABLED=0` build produces `bin/claude-security-hooks` | RQ-7 cold-start probe confirms the binary builds statically and runs in 2–3 ms |
| REQ-hooks-H2 | All three subcommands accept Claude Code hook JSON on stdin | RQ-1 verified live PreToolUse / PostToolUse / SubagentStart schemas; RQ-9 stdlib `flag.NewFlagSet` dispatch pattern |
| REQ-hooks-H3 | When `tool_input.subagent_type` not in security set, exit 0 silently | RQ-1 confirms `tool_input.subagent_type` is the live field name for the Task tool |
| REQ-hooks-H4 | On invariant failure: emit `{"decision":"block","reason":"..."}` and exit 0 | Live PostToolUse decision schema includes `{"decision":"block","reason":"..."}` (verified RQ-1) |
| REQ-hooks-H5 | Core validator path: zero non-stdlib deps; test deps in `_test.go` only | RQ-9 + `go list -deps -test ./...` enforcement pattern |
| REQ-hooks-H6 | Every assertion has a check predicate + unit test | RQ-5 documents `go/parser` AST walk over `_test.go` files (testing.M does NOT expose runtime test enumeration) |
| REQ-hooks-H7 | Malformed `tool_response.content` → block decision, never panic | RQ-3 confirms stdlib JSON decoder surfaces errors as values, not panics |
| REQ-cartographer-A1..A11 | 11 predicates over `go-index/v1` schema | RQ-3 + RQ-4 (typed registry); A7/A8 need filesystem reads (RQ-8 CWD guard) |
| REQ-taint-T1..T11 | 11 predicates over taint verdict schema | RQ-3 + RQ-4; T9 needs filesystem reads (RQ-8) |
| REQ-authz-AZ1..AZ6 | 6 predicates over authz schema | RQ-3 + RQ-4; AZ4 needs source file read (RQ-8) |
| REQ-oauth-OA1..OA7 | 7 predicates over OAuth auditor schema | RQ-3 + RQ-4; OA4 needs source file read (RQ-8) |
| REQ-invariant-IC1..IC4 | 4 predicates over invariant-checker schema | RQ-3 + RQ-4; IC3 needs source file read (RQ-8) |
| REQ-synthesis-S1..S6 | 6 predicates over review-report/v1 schema | RQ-3 + RQ-4; S1 needs file-existence check |
</phase_requirements>

## Phase Boundary Recap

Phase 2 ships **one artifact**: a static `CGO_ENABLED=0` Go binary at `claude-security-hooks/bin/claude-security-hooks` that:

1. Exposes three subcommands — `preflight`, `validate`, `inject-context` — dispatched in `cmd/claude-security-hooks/main.go` using stdlib `flag.NewFlagSet`. Each subcommand reads a Claude Code hook event JSON payload on stdin.
2. Mirrors six agent verdict schemas as Go structs in `internal/schema/{cartographer,taint_tracer,authz_tracer,oauth_auditor,invariant_checker,synthesis}.go`, all decoded with `json.NewDecoder(...).DisallowUnknownFields()` per D-10.
3. Implements **51 per-agent assertion check predicates** (A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6) and **7 build/runtime invariants** (H1–H7) — total 52 unit-testable predicates plus 6 covered indirectly by build/CI checks. Per-agent typed registries live in `internal/invariants/<agent>.go` with no shared `any` signature (D-01/D-02 + user-global no-`any` rule).
4. Ships with a table-driven `_test.go` suite (D-06) co-located with each implementation file. `github.com/google/go-cmp` v0.7.0 is the ONLY non-stdlib dependency, used only in `_test.go`.
5. Builds via `Makefile` with `build|test|install|clean|lint` targets (D-13).

**Out of scope for Phase 2:** agent files in `.claude/agents/*.md` (Phase 3), `.claude/settings.json` (Phase 3), `/security-review` command (Phase 3), `opengrep-mcp` server (Phase 4), end-to-end smoke test (Phase 5), real semantic context injection in `inject-context` (deferred per D-08).

**Primary recommendation for the planner:** Structure the work as **6 plans in 3 waves**. Wave 0 lays the test infrastructure + schema structs (no behavior); Wave 1 implements the six per-agent invariant registries in parallel; Wave 2 wires the three subcommands, the H5/H6 meta-tests, and the Makefile. See "Recommendations to Planner" at the bottom.

## Architectural Responsibility Map

A single static Go binary is not "multi-tier" in the web-app sense, but it has internal architectural boundaries that the planner must keep crisp.

| Capability | Primary Layer | Secondary Layer | Rationale |
|------------|---------------|-----------------|-----------|
| Subcommand dispatch + stdin parsing | `cmd/claude-security-hooks/main.go` | — | Owns process I/O; thin shim that delegates to `internal/hooks/*` |
| Hook event payload structs (`PreToolUseEvent`, `PostToolUseEvent`, `SubagentStartEvent`) | `internal/hooks/events.go` | — | Lives next to subcommand bodies because shape is hook-protocol-specific, not agent-specific |
| Agent verdict schema structs (`TaintVerdict`, `CartographerIndex`, etc.) | `internal/schema/<agent>.go` | — | Pure data types mirroring CON-schema-* JSON keys; no behavior |
| Per-agent invariant predicates + registries | `internal/invariants/<agent>.go` | `internal/schema` (typed inputs) | Predicates are pure functions over typed verdict structs; they import `internal/schema` but NOT `internal/hooks` |
| Subcommand orchestration (parse payload → look up agent → unmarshal verdict → run registry → emit decision) | `internal/hooks/{preflight,validate,inject}.go` | `internal/schema`, `internal/invariants` | The wiring layer; the only layer that touches `os.Stdin`/`os.Stdout` (besides `main.go`) |
| Block-reason JSON encoding | `internal/hooks/decision.go` | — | Centralised so the D-03 format and exit-code policy (D-14) live in one file |
| Source-file inspection helpers (A7, A8, T9, AZ4, OA4, IC3) | `internal/invariants/fsutil.go` | `internal/hooks` (CWD guard) | Filesystem reads should be a single small helper used by all per-agent registries to avoid scattering `os.Open`/`bufio.Scanner` across 6 files |
| H5 dependency-policy meta-test | `internal/invariants/h5_deps_test.go` | — | Runs `go list -deps -test ./...` and asserts only stdlib + `github.com/google/go-cmp/...` appear |
| H6 coverage meta-test | `internal/invariants/h6_coverage_test.go` | — | Parses `_test.go` files via `go/parser`, asserts every registry ID has a matching `Test*` function |

**Boundary rule** (recommend the planner enforce in task ACs): `internal/invariants/*.go` files must NOT import `internal/hooks`. Predicates take typed verdict structs as input and return `[]Violation`; they never touch I/O. This keeps predicates trivially unit-testable with literal fixtures.

## Standard Stack

### Core

| Package | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | 1.22 (declared) / 1.24.13 (latest) | Compiler, build, test | Project go.mod already declares 1.22 (D-12); 1.24.13 is current stable (released 2026-02-04). [VERIFIED: go.dev/doc/devel/release] |
| `encoding/json` | stdlib | All JSON parsing | D-10 mandates `DisallowUnknownFields()`; verified to recurse into nested objects and array elements (RQ-3 local probe) [VERIFIED: local Go 1.24.4 run] |
| `flag` | stdlib | Subcommand dispatch via `flag.NewFlagSet` | Zero non-stdlib deps mandate (REQ-hooks-H5) [VERIFIED: gobyexample.com/command-line-subcommands] |
| `os` / `bufio` | stdlib | stdin/stdout/stderr; line-count scanning for A8 | Standard library |
| `go/parser` + `go/ast` + `go/token` | stdlib | Parse `_test.go` files at test time for the H6 coverage meta-test (D-04) | Only stdlib mechanism — `testing.M` does NOT expose runtime test enumeration [VERIFIED: pkg.go.dev/testing] |
| `fmt` / `strings` / `sort` | stdlib | Block-reason formatting per D-03 | Standard library |

### Test-only

| Package | Version | Purpose | When Used |
|---------|---------|---------|-----------|
| `testing` | stdlib | Test runner | Always |
| `github.com/google/go-cmp/cmp` | v0.7.0 (released 2024-02-21) | `cmp.Diff` for `[]Violation` comparisons | All table-driven tests [VERIFIED: github.com/google/go-cmp/releases] |
| `github.com/google/go-cmp/cmp/cmpopts` | v0.7.0 | `cmpopts.EquateEmpty()` for nil-vs-empty slice tolerance | Tests that assert "no violations" — distinguishes `nil` from `[]Violation{}` cleanly [VERIFIED: pkg.go.dev/github.com/google/go-cmp/cmp/cmpopts] |

**Version verification** (run before locking dependencies):

```bash
go version                                  # expect 1.22+ (project declares 1.22)
GOFLAGS=-mod=mod go list -m -versions github.com/google/go-cmp | tail -1   # latest tag
```

### Alternatives Considered (and rejected)

| Instead of stdlib | Could Use | Why rejected |
|-------------------|-----------|--------------|
| `flag.NewFlagSet` | `github.com/spf13/cobra` | Violates H5 zero-non-stdlib-deps. Cobra brings ~20 transitive deps. |
| `encoding/json` | `github.com/goccy/go-json` / `github.com/json-iterator/go` | Violates H5. Stdlib JSON is fast enough at our payload sizes (~few KB per verdict). |
| `cmp.Diff` for slice comparison | `reflect.DeepEqual` + custom diff | DeepEqual produces unreadable test failures for 51 distinct predicates. `cmp.Diff` is the one test-only dep we accept. |
| `go/parser` walk for H6 | `go test -list` + parse stdout | Would require running an external `go` command from inside a test — adds CI fragility and process overhead. AST walk is pure stdlib and runs in-process. |

**Installation:**

```bash
# From inside claude-security-hooks/
go get github.com/google/go-cmp@v0.7.0   # test-only dep
go mod tidy
```

(`go-cmp` will appear in `go.sum`; H5 meta-test confirms it appears in `Deps` only for `.test` binary import paths.)

## Architecture Patterns

### System Architecture Diagram

```
                ┌────────────────────────────────────────────────────────┐
                │ Claude Code (host)                                     │
                │                                                        │
                │   PreToolUse(Task)   PostToolUse(Task)   SubagentStart │
                │         │                  │                   │       │
                └─────────┼──────────────────┼───────────────────┼───────┘
                          │ JSON on stdin    │                   │
                          ▼                  ▼                   ▼
                   ┌─────────────┐   ┌─────────────┐     ┌─────────────┐
                   │ preflight   │   │ validate    │     │ inject-ctx  │
                   │ subcommand  │   │ subcommand  │     │ (stub D-08) │
                   └──────┬──────┘   └──────┬──────┘     └──────┬──────┘
                          │                 │                   │
                          ▼                 ▼                   ▼
              ┌──────────────────────────────────────────────────────────┐
              │  internal/hooks/events.go : parse PreToolUseEvent /      │
              │                            PostToolUseEvent /            │
              │                            SubagentStartEvent            │
              │  (json.NewDecoder + DisallowUnknownFields per D-10)      │
              └──────────────────────────────────────────────────────────┘
                          │                 │                   │
                  is subagent_type   is subagent_type      validate
                  in SecurityAgent   in SecurityAgent      schema only
                  Set? (D-07)        Set? (D-07)           (no-op)
                          │                 │                   │
              ┌───────────┴──────┐  ┌───────┴──────┐            │
              │ Yes → parse      │  │ Yes → strip  │            │
              │      prompt as   │  │      code-   │            │
              │      JSON per    │  │      fences, │            │
              │      per-agent   │  │      parse   │            │
              │      input ctx   │  │      verdict │            │
              │      (D-09)      │  │              │            │
              └───────┬──────────┘  └──────┬───────┘            │
                      │                    │                    │
                      ▼                    ▼                    │
              ┌──────────────────────────────────────┐          │
              │ internal/schema/<agent>.go           │          │
              │ Typed Go structs mirror JSON shapes  │          │
              │ (CON-schema-*)                       │          │
              └────────────────┬─────────────────────┘          │
                               │                                │
                               ▼                                │
              ┌──────────────────────────────────────┐          │
              │ internal/invariants/<agent>.go       │          │
              │ Per-agent typed registry:            │          │
              │   var TaintTracerInvariants =        │          │
              │     []TaintInvariant{T1, T2, …, T11} │          │
              │ Iterate in registration order →      │          │
              │   []Violation                        │          │
              └────────────────┬─────────────────────┘          │
                               │                                │
                               ▼                                ▼
              ┌──────────────────────────────────────────────────────────┐
              │ internal/hooks/decision.go : format {decision:block,    │
              │                                       reason:"…; …"}    │
              │ (D-03 format; D-14 two-class exit policy)               │
              └──────────────────────────────────────────────────────────┘
                               │                                │
                               ▼                                ▼
                       stdout (block) /                  exit 0 silent
                       silent exit 0 (pass)              (validation passed
                                                          or non-security agent)
```

**External I/O surface (intentionally minimal):**

- Reads: `os.Stdin`; selectively reads workspace source files for A7/A8/T9/AZ4/OA4/IC3 via the `fsutil.go` helper.
- Writes: `os.Stdout` (block decisions only); `os.Stderr` (Class B infrastructure failures only).
- Env: respects `$CLAUDE_PROJECT_DIR` if set (planner's choice — see RQ-8); falls back to CWD.

### Recommended Project Structure

```
claude-security-hooks/
├── cmd/
│   └── claude-security-hooks/
│       └── main.go                      # subcommand dispatch (flag.NewFlagSet)
├── internal/
│   ├── hooks/
│   │   ├── doc.go                       # package doc: D-14 exit-code contract
│   │   ├── agents.go                    # SecurityAgentSet + drift test target
│   │   ├── agents_test.go               # TestSecurityAgentSet_MatchesSpec
│   │   ├── events.go                    # PreToolUseEvent, PostToolUseEvent, SubagentStartEvent
│   │   ├── events_test.go               # roundtrip JSON parse tests
│   │   ├── decision.go                  # block-reason formatter (D-03)
│   │   ├── decision_test.go
│   │   ├── preflight.go                 # preflight subcommand body (D-09)
│   │   ├── preflight_test.go
│   │   ├── validate.go                  # validate subcommand body (CON-validate-protocol)
│   │   ├── validate_test.go
│   │   ├── inject.go                    # inject-context stub (D-08)
│   │   └── inject_test.go
│   ├── schema/
│   │   ├── cartographer.go              # go-index/v1 structs
│   │   ├── taint_tracer.go              # taint verdict structs
│   │   ├── authz_tracer.go
│   │   ├── oauth_auditor.go
│   │   ├── invariant_checker.go
│   │   └── synthesis.go                 # review-report/v1 structs
│   └── invariants/
│       ├── registry.go                  # shared Severity, Violation types
│       ├── fsutil.go                    # source-file existence + line-count helpers
│       ├── fsutil_test.go
│       ├── cartographer.go              # CartographerInvariant + A1..A11 checks
│       ├── cartographer_test.go         # TestA1_* .. TestA11_*
│       ├── taint_tracer.go              # TaintInvariant + T1..T11 checks
│       ├── taint_tracer_test.go
│       ├── authz_tracer.go              # AZ1..AZ6
│       ├── authz_tracer_test.go
│       ├── oauth_auditor.go             # OA1..OA7
│       ├── oauth_auditor_test.go
│       ├── invariant_checker.go         # IC1..IC4
│       ├── invariant_checker_test.go
│       ├── synthesis.go                 # S1..S6
│       ├── synthesis_test.go
│       ├── h5_deps_test.go              # H5 meta-test (go list parse)
│       └── h6_coverage_test.go          # H6 meta-test (go/parser walk)
├── Makefile
├── go.mod                                # github.com/saghaulor/claude-security-hooks, go 1.22
├── go.sum                                # added when go-cmp is pulled
└── README.md                             # (Phase 6 ships full README; Phase 2 ships stub)
```

### Pattern 1: Stdlib subcommand dispatch (REQ-hooks-H2, REQ-hooks-H5)

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

Source: gobyexample.com/command-line-subcommands [CITED].

### Pattern 2: Per-agent typed invariant registry (D-01, D-02)

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

Sibling registries (`CartographerInvariants`, `AuthzInvariants`, `OAuthInvariants`, `InvariantCheckerInvariants`, `SynthesisInvariants`) follow the exact same shape with their own typed `Check func(*schema.XVerdict) []Violation` signature. **Critical**: no shared `any` — the user-global CLAUDE.md ban on `any` is enforced by 6 typed registry types.

Source: pattern is the Kubernetes apimachinery `field.ErrorList` model (errors as values returned from imperative checks, not declarative tags) [CITED: pkg.go.dev/k8s.io/apimachinery/pkg/util/validation/field].

### Pattern 3: `json.NewDecoder + DisallowUnknownFields` (D-10)

```go
// internal/hooks/validate.go (sketch — agent dispatch loop)
func parseTaintVerdict(r io.Reader) (*schema.TaintVerdict, error) {
    var v schema.TaintVerdict
    dec := json.NewDecoder(r)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&v); err != nil {
        return nil, fmt.Errorf("taint verdict parse: %w", err)
    }
    return &v, nil
}
```

`DisallowUnknownFields()` **recurses into nested struct fields and into struct elements of arrays** — verified with a local Go 1.24.4 probe (RQ-3 below). On unknown field: returns `json: unknown field "<name>"` as an `error` (NOT a panic), satisfying REQ-hooks-H7.

### Pattern 4: Table-driven test with `cmp.Diff` (D-05, D-06)

```go
// internal/invariants/taint_tracer_test.go
package invariants_test

import (
    "testing"
    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"
    "github.com/saghaulor/claude-security-hooks/internal/invariants"
    "github.com/saghaulor/claude-security-hooks/internal/schema"
)

func TestT3_PathNonEmptySourceFirstSinkLast(t *testing.T) {
    cases := []struct {
        name string
        in   *schema.TaintVerdict
        want []invariants.Violation
    }{
        {
            name: "ok: source first, sink last, exploitable",
            in: &schema.TaintVerdict{
                Verdict: "exploitable",
                Path: []schema.TaintPathStep{{Step: "source"}, {Step: "call"}, {Step: "sink"}},
            },
            want: nil,
        },
        {
            name: "bad: first step is not source",
            in: &schema.TaintVerdict{
                Verdict: "exploitable",
                Path: []schema.TaintPathStep{{Step: "sink"}, {Step: "sink"}},
            },
            want: []invariants.Violation{{Path: "path[0].step", Expected: "source", Actual: "sink"}},
        },
        {
            name: "bad: empty path on sanitized verdict",
            in:   &schema.TaintVerdict{Verdict: "sanitized", Path: nil},
            want: []invariants.Violation{{Path: "path", Expected: "non-empty", Actual: "empty"}},
        },
        {
            name: "ok: input_mismatch permits empty path (T4 covers this)",
            in:   &schema.TaintVerdict{Verdict: "input_mismatch", Path: nil},
            want: nil,
        },
    }

    // Locate the T3 check from the registry — proves registry wiring at test time too.
    var checkT3 func(*schema.TaintVerdict) []invariants.Violation
    for _, inv := range invariants.TaintTracerInvariants {
        if inv.ID == "T3" {
            checkT3 = inv.Check
            break
        }
    }
    if checkT3 == nil {
        t.Fatal("T3 not registered in TaintTracerInvariants")
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got := checkT3(tc.in)
            if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
                t.Errorf("T3 check mismatch (-want +got):\n%s", diff)
            }
        })
    }
}
```

`cmpopts.EquateEmpty()` treats `nil` and `[]Violation{}` as equal — important because `nil`-returning checks should not be considered different from explicit-empty returns [VERIFIED: pkg.go.dev/github.com/google/go-cmp/cmp/cmpopts].

### Anti-Patterns to Avoid

- **Shared `Check func(v any)` signature.** Violates user-global CLAUDE.md and D-02. Forces type assertions inside every check function and loses compile-time safety.
- **Map-based registry `map[string]CheckFn`.** Map iteration order is non-deterministic in Go; D-03 requires deterministic block-reason ordering. Use a slice.
- **Calling `panic()` from a check function.** REQ-hooks-H7 requires graceful error handling. Checks return `[]Violation`; parse failures are caught by the decoder and turned into a single block reason at the dispatcher layer.
- **Reaching for `cobra`, `viper`, `gjson`, or any non-stdlib lib in non-test code.** Breaks H5 and the H5 meta-test will catch it in CI.
- **Encoding the model identifier (`claude-sonnet-4-6`) into the binary.** The binary is model-agnostic. Only test fixtures need it (D-15).
- **Cross-importing `internal/hooks` from `internal/invariants`.** Predicates must remain I/O-free and import only `internal/schema`. The H5 meta-test won't catch this, but a focused unit test (or `gopls` review) should.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Subcommand dispatch | Custom arg parser | `flag.NewFlagSet` (stdlib) | One-liner per subcommand; idiomatic Go |
| JSON unmarshalling | Custom decoder, `gjson` | `encoding/json.NewDecoder + DisallowUnknownFields()` | Stdlib recurses into nested objects and arrays correctly; verified |
| Slice-diff in tests | `reflect.DeepEqual + custom diff` | `cmp.Diff` (go-cmp v0.7.0) | Readable diffs essential for 51 distinct predicates |
| Test-name enumeration | Manual list of test functions | `go/parser` AST walk in H6 meta-test | Stdlib mechanism; `testing.M` does not expose runtime enumeration |
| Dependency policy enforcement | Manual `go.mod` audit | `go list -deps -test ./...` parse in H5 meta-test | Mechanical, runs in CI |
| Line-count for A8 | `os.ReadFile` then `strings.Count(data, "\n")` (loads whole file) | `bufio.Scanner` over `os.Open(path)` | Streams large files; memory-bounded |
| Workspace-root detection | Hardcoded `cwd := "."` | Check `$CLAUDE_PROJECT_DIR` env then fall back to CWD marker scan (`go.mod`/`.claude`/`.planning`) | Live Claude Code docs say env var is the recommended path (RQ-8); D-11 CWD guard is the fallback |

**Key insight:** Every "problem" in this table has a stdlib answer except slice-diff (one accepted test-only dep). The discipline of refusing to hand-roll AND refusing to add deps is what keeps the cold-start budget at 2–3 ms.

## Runtime State Inventory

Phase 2 is a **greenfield build**, not a rename/refactor. The binary is new code that did not previously exist. No data migrations, no service config changes, no OS-registered state, no secret rotation. Skipping is appropriate per the research instructions.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — greenfield binary | None |
| Live service config | None — Phase 3 owns `.claude/settings.json`; Phase 2 doesn't touch it | None |
| OS-registered state | None — binary is invoked synchronously by Claude Code; no services/timers | None |
| Secrets/env vars | None — Phase 2 binary doesn't read secrets. (`$CLAUDE_PROJECT_DIR` is set by Claude Code, not a secret) | None |
| Build artifacts | `claude-security-hooks/bin/claude-security-hooks` (built by `make build`) and `.claude/hooks/bin/claude-security-hooks` (copied by `make install`). Phase 1 scaffold already excluded `bin/` via `.gitignore`. | Verify `.gitignore` covers these paths (planner: re-confirm during Wave 0) |

## Common Pitfalls

### Pitfall 1: HAND_OFF spec drift on hook payload field names

**What goes wrong:** HAND_OFF §8.1 documents `tool_input.subagent_type`, but the field name could have changed in Claude Code's current docs. If we encode the wrong name, the validator silently no-ops on every invocation (REQ-hooks-H3 collapses everything to "not a security agent" → exit 0 silent).

**Why it happens:** HAND_OFF.md is a point-in-time design document. The live docs at https://code.claude.com/docs/en/hooks evolve.

**How to avoid:** RQ-1 below VERIFIED that `tool_input.subagent_type` is still correct for the `Task` tool as of 2026-05-18. The planner should write a test fixture that mirrors the live live payload shape verbatim and add a runtime check: if `tool_name != "Task"` AND `subagent_type` is absent, log a warning to stderr (Class B) so future spec drift surfaces loudly.

**Warning signs:** All `validate` invocations silently exit 0 with no stdout. Add a debug logging mode (gated by `$CLAUDE_SECURITY_HOOKS_DEBUG=1`) that writes "no-op: subagent_type=X" to stderr — invaluable for diagnosing dead validators.

### Pitfall 2: `DisallowUnknownFields()` rejecting valid agent output after schema evolution

**What goes wrong:** Phase 3 ships agent prompts; in Phase 4+ someone adds a new field to a verdict schema but forgets to update `internal/schema/<agent>.go`. The validator blocks every real invocation.

**Why it happens:** The strictness is intentional (D-10: ambiguity preferred over false confidence), but the cost is a tight coupling: every schema bump requires a coordinated commit touching both prompt and Go struct.

**How to avoid:** Document this contract loudly in `internal/schema/doc.go` AND in the Phase 6 README. The H6 meta-test won't catch schema bumps — only agent-side drift surfaces at runtime when the agent emits a new field.

**Warning signs:** Block-reason text `"json: unknown field \"<name>\""` showing up in real reviews. Treat this as a high-priority signal that prompt and schema have diverged.

### Pitfall 3: Conflating `preflight` (input validation) with `validate` (output validation)

**What goes wrong:** Reading `tool_response.content` in `preflight` (which has no `tool_response`) or reading `tool_input.prompt` in `validate` for input mismatch. Either causes nil-pointer dereferences or silently misses violations.

**Why it happens:** Both subcommands handle very similar JSON payloads. The schemas overlap (`tool_input.subagent_type` appears in both). D-09 makes the symmetry explicit but the implementer must internalize it.

**How to avoid:** Two clearly-separated event structs in `internal/hooks/events.go` (`PreToolUseEvent` has NO `tool_response` field; `PostToolUseEvent` has both). Use `DisallowUnknownFields()` to ensure each event-type decoder rejects fields from the wrong shape.

**Warning signs:** Tests for `preflight` accidentally pass a `PostToolUseEvent` fixture and still pass — clean event structs make this a compile error.

### Pitfall 4: `tool_response.content` is not always a string

**What goes wrong:** The live Claude Code docs (RQ-1) document `tool_response.content` as `"string or array (tool output)"`. If the validator assumes `string`, malformed inputs panic.

**Why it happens:** Tool output varies by tool. For the `Task` tool, content is the subagent's final message, which IS typically a string. But the validator must handle both cleanly per REQ-hooks-H7.

**How to avoid:** Model `tool_response.content` as `json.RawMessage` and attempt a `string` decode first; on failure, try `[]interface{}` and concatenate the text segments. Any failure → block decision `"H7: tool_response.content not parseable as text"`. Never panic.

**Warning signs:** Panic stack traces in real reviews. The H7 unit test must cover at least: malformed JSON, valid JSON but unexpected type, empty content, multi-segment array content.

### Pitfall 5: Forgetting `go-cmp` is allowed only in `_test.go`

**What goes wrong:** Someone imports `github.com/google/go-cmp/cmp` from `internal/invariants/registry.go` (non-test file) to compare violations during runtime. H5 meta-test fails; CI blocks the merge — but only if the H5 meta-test was written first.

**Why it happens:** It's a tempting one-liner. The H5 boundary is policy, not language enforcement.

**How to avoid:** Write the H5 meta-test FIRST (Wave 0, before any non-test code). Make `go-cmp` a "tripwire" — if it appears in a non-`.test` import path from `go list -deps -test`, fail the test with a clear message.

**Warning signs:** `go list -deps ./cmd/claude-security-hooks` shows any non-stdlib entry. (`go list -deps -test ./...` will show `go-cmp` under `.test` packages — those are fine.)

### Pitfall 6: CWD assumption breaks in environments where Claude Code launches hooks from a different directory

**What goes wrong:** A7/A8 try to `os.Stat("internal/foo.go")` but Claude Code launched the hook from `$HOME` or some temporary dir. The file is "missing" → false A7 violation on every cartographer output.

**Why it happens:** D-11 ASSUMES CWD = workspace root. The live docs (RQ-8) say hooks run in the session's CWD AND set `$CLAUDE_PROJECT_DIR`.

**How to avoid:** Use `$CLAUDE_PROJECT_DIR` if set (preferred per live docs); fall back to CWD + marker scan (`go.mod`/`.claude`/`.planning`) per D-11. If neither resolves to a directory containing `go.mod`, emit `block` with reason `"workspace_root_not_found"`. This is in scope for Phase 2 per D-11; the planner should make the env-var lookup explicit.

**Warning signs:** Every cartographer output blocks with `A7: <path> file_does_not_exist`. Real bug, but the most likely cause is workspace root resolution, not actual missing files.

### Pitfall 7: Non-deterministic block-reason ordering

**What goes wrong:** Multiple violations in one verdict produce a block reason like `"T3: ...; T6: ..."`. Different runs produce different orderings, breaking grep-based CI assertions.

**Why it happens:** Iterating a Go map is non-deterministic. Iterating a slice IS deterministic by definition.

**How to avoid:** D-01 mandates slice-based registries. Slice iteration is `for i := 0; i < len(s); i++ { ... }` (or `for _, x := range s`) — both fully deterministic. No `sort` needed; registration order IS the canonical order. Confirmed by Go spec (RQ-10 below).

**Warning signs:** Flaky CI tests on the block-reason format. If you ever see this, audit for accidental map use.

## Code Examples

Verified patterns referenced earlier in this document. All examples have been mentally type-checked against `go vet` rules. Code below is illustrative — the planner / implementer should pull final shapes from the live patterns in this document.

### Example: `internal/hooks/events.go` (the live payload shape)

```go
package hooks

import "encoding/json"

// PreToolUseEvent matches the live Claude Code hook input for PreToolUse on Task.
// Source: https://code.claude.com/docs/en/hooks (verified 2026-05-18, RQ-1).
type PreToolUseEvent struct {
    SessionID      string         `json:"session_id"`
    TranscriptPath string         `json:"transcript_path"`
    CWD            string         `json:"cwd"`
    HookEventName  string         `json:"hook_event_name"` // "PreToolUse"
    ToolName       string         `json:"tool_name"`        // "Task"
    ToolInput      TaskToolInput  `json:"tool_input"`
    ToolUseID      string         `json:"tool_use_id"`
    PermissionMode string         `json:"permission_mode,omitempty"`
}

// TaskToolInput is the tool_input shape when tool_name == "Task".
type TaskToolInput struct {
    SubagentType string `json:"subagent_type"`
    Prompt       string `json:"prompt"`
    Description  string `json:"description,omitempty"`
}

// PostToolUseEvent adds tool_response.
type PostToolUseEvent struct {
    SessionID      string          `json:"session_id"`
    TranscriptPath string          `json:"transcript_path"`
    CWD            string          `json:"cwd"`
    HookEventName  string          `json:"hook_event_name"` // "PostToolUse"
    ToolName       string          `json:"tool_name"`
    ToolInput      TaskToolInput   `json:"tool_input"`
    ToolUseID      string          `json:"tool_use_id"`
    ToolResponse   ToolResponse    `json:"tool_response"`
    PermissionMode string          `json:"permission_mode,omitempty"`
}

// ToolResponse.Content is RawMessage because per the live docs it may be string or array.
type ToolResponse struct {
    Content json.RawMessage `json:"content"`
    Type    string          `json:"type,omitempty"`
}

// SubagentStartEvent — note: agent_type (per live docs), not subagent_type.
type SubagentStartEvent struct {
    SessionID      string `json:"session_id"`
    TranscriptPath string `json:"transcript_path"`
    CWD            string `json:"cwd"`
    HookEventName  string `json:"hook_event_name"` // "SubagentStart"
    AgentType      string `json:"agent_type"`
    AgentID        string `json:"agent_id"`
    Prompt         string `json:"prompt"`
}
```

**Important note from RQ-1:** the field name differs between PreToolUse/PostToolUse (`tool_input.subagent_type`) and SubagentStart (`agent_type` at top-level). HAND_OFF.md §8.1 shows `agent_name` for SubagentStart — the live docs use `agent_type`. Use the live name (`agent_type`) for the SubagentStart struct.

### Example: H5 meta-test (`internal/invariants/h5_deps_test.go`)

```go
package invariants_test

import (
    "encoding/json"
    "os/exec"
    "strings"
    "testing"
)

// TestH5_CoreValidatorPathHasZeroNonStdlibDeps asserts:
//   1. Non-test packages import only stdlib.
//   2. Test packages may additionally import google/go-cmp.
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
        ForTest    string // non-empty for test-only packages
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
            continue // our own module is fine
        }
        if strings.HasPrefix(p.ImportPath, allowedTest) {
            // allowed but only if ForTest is set
            if p.ForTest == "" {
                t.Errorf("H5: %s imported by non-test code", p.ImportPath)
            }
            continue
        }
        t.Errorf("H5: disallowed dependency: %s (for-test=%q)", p.ImportPath, p.ForTest)
    }
}
```

Source: `go list -deps -test -json` produces one JSON object per package; the `Standard` field flags stdlib, and `ForTest` is set for test-only inflations [CITED: manpages.debian.org/testing/golang-go/go-list.1.en.html].

### Example: H6 meta-test sketch (`internal/invariants/h6_coverage_test.go`)

```go
package invariants_test

import (
    "go/ast"
    "go/parser"
    "go/token"
    "path/filepath"
    "strings"
    "testing"

    "github.com/saghaulor/claude-security-hooks/internal/invariants"
)

// expectedIDs returns the canonical set of assertion IDs that must each have
// at least one Test* function whose name starts with "Test<ID>_".
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
    // 1. Registry IDs match spec.
    assertRegistryIDs(t, "TaintTracer", taintIDs(), idList(invariants.TaintTracerInvariants))
    // ... repeat for each registry

    // 2. Every ID has a matching Test* function.
    fset := token.NewFileSet()
    for file, ids := range expectedIDs() {
        path := filepath.Join(".", file)
        f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
        if err != nil {
            t.Fatalf("parse %s: %v", path, err)
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
```

Source: `go/parser` AST walk is the only stdlib-only way to enumerate test names, because `testing.M` does not expose them at runtime [VERIFIED: pkg.go.dev/testing — no `List()` method on `*testing.M`].

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `flag.Parse()` + manual sub-arg munging | `flag.NewFlagSet(name, ExitOnError)` per subcommand | Go 1.x (stable since Go 1.0) | Idiomatic stdlib subcommand pattern |
| `json.Unmarshal` then walk for extra fields | `json.NewDecoder(...).DisallowUnknownFields()` | Go 1.10 (2018) | Strict-mode JSON detection in one line; recurses correctly into nested structs and array elements (verified RQ-3) |
| Hand-rolled diff helpers | `github.com/google/go-cmp` `cmp.Diff` | go-cmp v0.5.x → current v0.7.0 (released 2024-02-21) | De-facto standard for table-driven Go test diffs |
| `claude-sonnet-4-5-20250929` | `claude-sonnet-4-6` (dateless ID; current Sonnet, RQ-2) | Around Q1 2026 | Phase 2 fixtures use `claude-sonnet-4-6` |
| `subagent_type` only (HAND_OFF §8.1 SubagentStart) | `agent_type` (live docs) | Recent (post-HAND_OFF.md authorship) | Phase 2 `SubagentStartEvent` struct must use `agent_type` |

**Deprecated/outdated assumptions:**

- HAND_OFF §8.1 lists `agent_name` as the SubagentStart field. **Outdated** — live docs use `agent_type`. Use the live name.
- HAND_OFF §3.7 module path uses `github.com/sagjaulor/...` (typo). Phase 1 scaffold has corrected this to `github.com/saghaulor/claude-security-hooks` — use the corrected path everywhere.

## Validation Architecture

Phase 2's exit criterion IS the test suite. This section enumerates every assertion ID with its testing strategy. The planner uses this as the source of truth for which test files exist and what fixtures they need.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | stdlib `testing` (Go 1.22+) + `github.com/google/go-cmp` v0.7.0 (test-only) |
| Config file | none — `go test` defaults are fine |
| Quick run command (single agent) | `go test ./internal/invariants -run TestT3_` (or any ID prefix) |
| Per-agent suite | `go test ./internal/invariants -run TestT` (all taint tests), `TestA` (all cartographer), etc. |
| Full suite command | `go test ./...` from `claude-security-hooks/` |
| Coverage report | `go test -coverprofile=cover.out ./... && go tool cover -func=cover.out` |

### Phase Requirements → Test Map

#### Cartographer (A1–A11)

| ID | Behavior | Test Type | Min Cases (positive+adversarial) | Fixture Required |
|----|----------|-----------|-----|------------------|
| A1 | Output is valid `go-index/v1` JSON | unit (table-driven) | 1 pos + 2 neg (malformed, schema mismatch) | inline JSON string |
| A2 | `schema_version` literal `"go-index/v1"` | unit | 1 pos + 1 neg | inline |
| A3 | `entrypoints[*].handler.{fqn,file,line}` populated | unit | 1 pos + 4 neg (missing each field) | inline |
| A4 | `routers_detected` ⊆ known set | unit | 1 pos + 1 neg (unknown router) | inline |
| A5 | Non-empty routers + empty entrypoints ⇒ warning present | unit | 1 pos + 1 neg | inline |
| A6 | Cited graph node IDs exist in `graphify-out/graph.json` | unit + filesystem fixture | 1 pos + 2 neg (missing node, missing graph file) | `testdata/graphify-out/graph.json` fixture |
| A7 | Every cited file path exists in workspace | unit + filesystem | 1 pos + 1 neg | uses `fsutil`; fixtures under `testdata/workspace/` |
| A8 | Every cited line number within file bounds | unit + filesystem | 1 pos + 1 neg (line > file length) | `testdata/workspace/` with known line counts |
| A9 | `authz_primitives[*].blocking == true` (non-blocking excluded) | unit | 1 pos + 1 neg | inline |
| A10 | Agent did not modify `graphify-out/` or source tree | (Phase 5 concern — hook can't easily verify post-hoc; spec H6 requires a check predicate) | 1 pos (write-call count == 0 in agent tool log) + 1 neg | doc that this is a tool-allowlist check, predicate body asserts the agent's Write/Edit usage is absent from the verdict's `tool_calls` field if present, else `Violation{Path: "tool_calls", Expected: "no write tools", Actual: "<observed>"}` |
| A11 | `Bash` commands match `allowed_commands` | unit | 1 pos + 1 neg | inline |

#### Taint Tracer (T1–T11)

| ID | Behavior | Test Type | Min Cases | Fixture |
|----|----------|-----------|-----|---------|
| T1 | Valid JSON conforming to verdict schema | unit | 1 pos + 2 neg | inline |
| T2 | `verdict ∈ enum` | unit | 5 pos (one per enum value) + 1 neg | inline |
| T3 | `verdict ∈ {sanitized,exploitable}` ⇒ path non-empty + first=source + last=sink | unit | 1 pos + 3 neg | inline (sample shown in Pattern 4 above) |
| T4 | `verdict == input_mismatch` permits empty path | unit | 1 pos | inline |
| T5 | `confidence ∈ {high,medium,low}` | unit | 3 pos + 1 neg | inline |
| T6 | `semgrep.ran == true` OR `gopls.references_calls > 0` (unless input_mismatch) | unit | 2 pos (each branch) + 1 neg (neither, not input_mismatch) + 1 pos (input_mismatch exempt) | inline |
| T7 | `semgrep.tier == input.semgrep_tier` | unit (joint input+output) | 1 pos + 1 neg | predicate signature MUST accept both verdict AND input record |
| T8 | Interface-typed source ⇒ `implementation_calls > 0` OR notes explain skip | unit (joint) | 1 pos (implementation_calls > 0) + 1 pos (notes contains explanation) + 1 neg | inline |
| T9 | Every `(file,line)` in path exists in workspace | unit + filesystem | 1 pos + 1 neg | `testdata/workspace/` |
| T10 | `confidence == high` ⇒ (Semgrep Pro/intrafile finding OR full LSP path with no `sanitizers_unverified`) | unit | 2 pos (each branch) + 1 neg | inline |
| T11 | Agent didn't call Grep/Bash/Edit/Write (via tool_calls log if present in verdict) | unit | 1 pos + 1 neg | inline; same pattern as A10 |

#### Authz Tracer (AZ1–AZ6)

| ID | Behavior | Test Type | Min Cases | Fixture |
|----|----------|-----------|-----|---------|
| AZ1 | Valid JSON conforming to schema | unit | 1 pos + 2 neg | inline |
| AZ2 | `summary.routes_total == len(input.routes)` | unit (joint) | 1 pos + 1 neg | predicate takes input+output |
| AZ3 | Bucket sum equals routes_total (accounting invariant) | unit | 1 pos + 1 neg | inline |
| AZ4 | Every `findings[*].route` ∈ input.routes | unit (joint) | 1 pos + 1 neg | inline |
| AZ5 | Every `weak_primitives[*]` ∈ input.authz_primitives | unit (joint) | 1 pos + 1 neg | inline |
| AZ6 | Read-call count ≤ `len(authz_primitives) + 2*len(routes)` | unit | 1 pos + 1 neg | inline; predicate inspects tool-call log if present in verdict |

#### OAuth Auditor (OA1–OA7)

| ID | Behavior | Test Type | Min Cases | Fixture |
|----|----------|-----------|-----|---------|
| OA1 | Valid JSON conforming to schema | unit | 1 pos + 2 neg | inline |
| OA2 | `checklist[*].spec` matches recognized RFC/draft ID | unit | 1 pos + 1 neg | inline; checker uses a hardcoded recognizer (regex over `RFC|draft-`) |
| OA3 | `status == "pass"` ⇒ `evidence.{file,line}` populated | unit | 1 pos + 1 neg | inline |
| OA4 | `taint_pairs[*]` reference files+lines from `input.oauth_locations` | unit (joint) | 1 pos + 1 neg | inline |
| OA5 | `target_profile == oauth_2_1` AND `pkce in features_in_use` ⇒ PKCE checks present, non-NA | unit (joint) | 1 pos + 1 neg | inline |
| OA6 | `target_profile == oauth_2_0` AND `authorize_endpoint` set ⇒ implicit-flow-disallowed check evaluated | unit (joint) | 1 pos + 1 neg | inline |
| OA7 | Check whose target feature absent MUST be `not_applicable`, not `pass` | unit (joint) | 1 pos + 1 neg | inline |

#### Invariant Checker (IC1–IC4)

| ID | Behavior | Test Type | Min Cases | Fixture |
|----|----------|-----------|-----|---------|
| IC1 | Valid JSON conforming to schema | unit | 1 pos + 2 neg | inline |
| IC2 | One result per input invariant; IDs match input set | unit (joint) | 1 pos + 2 neg (missing ID, extra ID) | inline |
| IC3 | `status == "violated"` ⇒ `evidence.files` non-empty AND `evidence.explanation` non-empty | unit | 1 pos + 2 neg | inline |
| IC4 | Output does not introduce new invariant IDs not in input | unit (joint) | 1 pos + 1 neg | inline |

#### Synthesis (S1–S6)

| ID | Behavior | Test Type | Min Cases | Fixture |
|----|----------|-----------|-----|---------|
| S1 | Both `review-report.json` and `review-report.md` exist | filesystem | 1 pos + 2 neg (each missing) | predicate takes dir path; `testdata/synthesis_out/` |
| S2 | JSON validates against `review-report/v1` schema | unit | 1 pos + 1 neg | inline JSON |
| S3 | `summary.total_findings == len(findings)` | unit | 1 pos + 1 neg | inline |
| S4 | Sum of `summary.by_severity` == `total_findings` | unit | 1 pos + 1 neg | inline |
| S5 | Every finding has ≥ 1 `source_agents` | unit | 1 pos + 1 neg | inline |
| S6 | No two findings share `(file,line,class)` unless noted in `deduplication_notes` | unit | 1 pos + 1 neg | inline |

#### Hook Build/Runtime Invariants (H1–H7)

| ID | Behavior | Test Type | Min Cases | Fixture |
|----|----------|-----------|-----|---------|
| H1 | `CGO_ENABLED=0 go build` produces static binary | shell smoke (Makefile) + optional `runtime` package test | 1 pos | `make build` exits 0; verified by `ldd` showing "not a dynamic executable" |
| H2 | Three subcommands parse hook JSON on stdin | integration unit | 3 pos (one per subcommand) | inline JSON fixtures for PreToolUse, PostToolUse, SubagentStart |
| H3 | Non-security `subagent_type` ⇒ exit 0 silently | integration unit | 1 pos (e.g., subagent_type="general-purpose") | inline |
| H4 | Block ⇒ exactly one `{"decision":"block","reason":"..."}` JSON, exit 0 | integration unit | 1 pos (any failing invariant) | inline |
| H5 | Zero non-stdlib deps in core; go-cmp only in `_test.go` | meta-test (`h5_deps_test.go`) | 1 (asserts whole `go list -deps -test` output) | runs `go list` via `os/exec` |
| H6 | Every assertion has registry entry + Test* function | meta-test (`h6_coverage_test.go`) | 1 (asserts whole registry × test-file index) | parses _test.go files via `go/parser` |
| H7 | Malformed `tool_response.content` ⇒ block, never panic | integration unit | 4 pos (invalid JSON, valid JSON wrong type, empty, multi-segment array) | inline malformed strings |

### Sampling Rate

- **Per task commit (Wave 1 task development):** `go test ./internal/invariants -run Test<ID>_` — runs the targeted assertion's test cases in <100 ms.
- **Per wave merge:** `go test ./...` — full suite; expected <5 s on a dev machine.
- **Phase gate (before `/gsd-verify-work`):** `make build && make test` from inside `claude-security-hooks/`; H1 verified by `ldd bin/claude-security-hooks` showing "not a dynamic executable" (smoke step in Makefile recommended).

### Wave 0 Gaps

The following test infrastructure does NOT exist in the Phase 1 scaffold and must be created before any Wave 1 implementation work:

- [ ] `claude-security-hooks/internal/invariants/registry.go` — `Severity`, `Violation` types (D-01 shared definitions).
- [ ] `claude-security-hooks/internal/invariants/fsutil.go` — `FileExists(path) bool`, `LineCount(path) (int, error)` helpers used by A7/A8/T9/AZ4/OA4/IC3.
- [ ] `claude-security-hooks/internal/invariants/fsutil_test.go` — covers the helpers, including missing-file and zero-byte-file edge cases.
- [ ] `claude-security-hooks/internal/invariants/h5_deps_test.go` — H5 meta-test (verified pattern above).
- [ ] `claude-security-hooks/internal/invariants/h6_coverage_test.go` — H6 meta-test scaffold; expected-IDs map populated as each per-agent registry lands.
- [ ] `claude-security-hooks/internal/invariants/testdata/workspace/` — minimal directory with a few `.go` files at known line counts for A7/A8/T9 tests.
- [ ] `claude-security-hooks/internal/invariants/testdata/graphify-out/graph.json` — minimal valid graph for A6 tests.
- [ ] `claude-security-hooks/internal/hooks/events.go` — event structs (PreToolUseEvent, PostToolUseEvent, SubagentStartEvent) and TaskToolInput / ToolResponse types.
- [ ] `claude-security-hooks/internal/hooks/events_test.go` — fixtures from live docs verified in RQ-1; ensures JSON round-trips.
- [ ] `claude-security-hooks/internal/hooks/decision.go` + `_test.go` — D-03 block-reason formatter.
- [ ] `claude-security-hooks/internal/hooks/agents.go` + `agents_test.go` — `SecurityAgentSet` and drift test (D-07).
- [ ] `claude-security-hooks/internal/hooks/doc.go` — package doc explaining D-14 exit-code policy.
- [ ] `claude-security-hooks/Makefile` — `build`, `test`, `install`, `clean`, `lint` targets.
- [ ] `go.sum` will be created when `go-cmp` is `go get`'d during Wave 0.
- [ ] Add `bin/` to `.gitignore` if not already there (Phase 1 likely covered it; verify).

## Project Constraints (from user CLAUDE.md)

This project has no `./CLAUDE.md`. The user's GLOBAL `$HOME/.claude/CLAUDE.md` directives apply:

| Directive | How Phase 2 honors it |
|-----------|------------------------|
| Always use Go over other languages | The deliverable IS a Go binary. Makefile is the only shell. |
| Always use declared types; avoid `any` | D-02 codifies this: per-agent typed `Check` signatures. No `interface{}`/`any` in the core path. `json.RawMessage` (for `tool_response.content`) is a declared type, not `any`. |
| Always use gopls for code analysis | Implementing session uses gopls to navigate/edit; not enforced in the shipped binary (the binary doesn't analyze its own code). |
| Always use TDD; plans must have a test section before implementation | Wave 0 lays test infrastructure FIRST (D-04 H6 meta-test, H5 meta-test, fsutil tests, schema fixtures). Wave 1 tasks each pair a `_test.go` with implementation, with the test landing FIRST in the same task. The planner MUST honor this in task ordering. |

**No `./CLAUDE.md` was found** — no project-level overrides. **No `.claude/skills/` or `.agents/skills/` directories exist** — no project-level skills to honor.

---

## Research Question Findings

### RQ-1: Claude Code hook event payload schemas (CON-hook-payload-shapes)

**Status:** [VERIFIED: https://code.claude.com/docs/en/hooks, fetched 2026-05-18]

The live spec confirms HAND_OFF §8.1 for PreToolUse and PostToolUse but diverges on SubagentStart. Canonical fields:

**PreToolUse** (for `tool_name == "Task"`):

```json
{
  "session_id": "...",
  "transcript_path": "...",
  "cwd": "...",
  "permission_mode": "default|plan|acceptEdits|auto|dontAsk|bypassPermissions",
  "effort": {"level": "low|medium|high|xhigh|max"},
  "hook_event_name": "PreToolUse",
  "tool_name": "Task",
  "tool_input": {
    "subagent_type": "go-taint-tracer",
    "prompt": "<the JSON prompt the orchestrator sent>",
    "description": "..."
  },
  "tool_use_id": "..."
}
```

**PostToolUse** (for `tool_name == "Task"`):

```json
{
  "session_id": "...",
  "transcript_path": "...",
  "cwd": "...",
  "permission_mode": "...",
  "effort": {"level": "..."},
  "hook_event_name": "PostToolUse",
  "tool_name": "Task",
  "tool_input": {
    "subagent_type": "go-taint-tracer",
    "prompt": "...",
    "description": "..."
  },
  "tool_use_id": "...",
  "tool_response": {
    "content": "<string OR array — the subagent's final message text>",
    "type": "..."
  }
}
```

**SubagentStart** (live spec):

```json
{
  "session_id": "...",
  "transcript_path": "...",
  "cwd": "...",
  "hook_event_name": "SubagentStart",
  "agent_type": "go-taint-tracer",
  "agent_id": "...",
  "prompt": "<the prompt being sent to the subagent>"
}
```

**Important deltas from HAND_OFF §8.1:**

1. **HAND_OFF says `agent_name` on SubagentStart; live spec says `agent_type`.** Use `agent_type`. The Go struct field must be `AgentType string \`json:"agent_type"\``.
2. **Live payloads include `cwd`** — the validator can read it instead of relying on `os.Getwd()`. This is a tighter, more reliable signal than the D-11 marker scan. Recommend using `event.CWD` as the primary CWD signal AND $CLAUDE_PROJECT_DIR env var (see RQ-8) BEFORE falling back to `os.Getwd()` + marker scan.
3. **Live payloads include `permission_mode` and `effort`** — Phase 2 should accept (DisallowUnknownFields will fail otherwise!) but ignore. The structs in `internal/hooks/events.go` must include these fields with JSON tags even if unused.
4. **`tool_response.content` is documented as "string OR array (tool output)".** Phase 2 must model this as `json.RawMessage` and attempt string-decode first, array-decode as fallback. REQ-hooks-H7 covers the panic-protection requirement.

**Implication for D-10 (DisallowUnknownFields):** Phase 2 schema structs for hook events must include EVERY documented field listed above, even when unused. Otherwise `DisallowUnknownFields()` rejects every real payload. The planner must call this out explicitly in the Wave 0 events.go task.

**Decision control output shape** for PostToolUse hooks (what the validator emits on block):

```json
{
  "decision": "block",
  "reason": "<the concatenated violation reasons>"
}
```

This top-level `decision` shape is what Claude Code recognizes on PostToolUse stdin. HAND_OFF §3.7 already specifies this; the live docs confirm it.

### RQ-2: Current Claude Sonnet model identifier (Q9, D-15)

**Status:** [VERIFIED: https://platform.claude.com/docs/en/about-claude/models, fetched 2026-05-18]

**`claude-sonnet-4-6` is CURRENT as of 2026-05-18.** No newer Sonnet exists. The model overview table lists it as the recommended Sonnet model under "Latest models comparison":

- Description: "The best combination of speed and intelligence."
- Claude API ID: `claude-sonnet-4-6`
- Claude API alias: `claude-sonnet-4-6` (dateless format; pinned snapshot, not evergreen)
- Context: 1M tokens, max output 64k tokens.
- Pricing: $3/MTok input, $15/MTok output.
- Training data cutoff: Jan 2026; reliable knowledge cutoff: Aug 2025.

**Action for Phase 2:** Test fixtures may embed `claude-sonnet-4-6` for any agent-frontmatter-mimicking fixtures. The hook binary itself remains model-agnostic per D-15. Q9 is RESOLVED.

**Note for Phase 3:** When agent frontmatter is authored, this same identifier will be used in `.claude/agents/*.md` `model:` field. No new verification needed at Phase 3 boundary unless docs change between now and then.

### RQ-3: `json.NewDecoder(...).DisallowUnknownFields()` semantics

**Status:** [VERIFIED: local Go 1.24.4 probe + pkg.go.dev/encoding/json]

Local probe (full source committed in test fixtures during Wave 0) confirms:

```
[top-level extra]   {"name":"a","inner":{"foo":"x"},"extra":1}                                      → json: unknown field "extra"
[nested extra]      {"name":"a","inner":{"foo":"x","extra":1}}                                       → json: unknown field "extra"
[array element extra] {"name":"a","inner":{"foo":"x"},"items":[{"foo":"x","extra":1}]}              → json: unknown field "extra"
[all good]          {"name":"a","inner":{"foo":"x"},"items":[{"foo":"x"}]}                          → <nil>
```

**Conclusions:**

1. **Recurses correctly into nested struct fields.** Unknown field at any depth fails.
2. **Recurses into array element struct fields.** Per-element unknown fields fail.
3. **Errors are returned as values, NOT panics.** Satisfies REQ-hooks-H7's "never a panic" constraint.
4. **Error message:** `json: unknown field "<name>"`. Does NOT include the path to the field (known limitation — github.com/golang/go/issues/58649). For Phase 2's block reasons, the planner should wrap this with the schema name: e.g., `"H7: taint verdict parse: json: unknown field \"foo\""`. The field name in the message is enough to act on.
5. **Caveat for embedded structs:** When an embedded struct has a custom `UnmarshalJSON` method, the embedded type receives the FULL JSON object including parent fields, which can interact badly with `DisallowUnknownFields` (github.com/golang/go/issues/22533). Phase 2's schema structs should NOT define custom `UnmarshalJSON` — let stdlib handle decoding mechanically. If a future schema field needs polymorphic decoding (e.g., union types), use `json.RawMessage` and do a second-stage typed decode in the predicate, not in the struct.

[VERIFIED: local Go 1.24.4]
[CITED: pkg.go.dev/encoding/json]
[CITED: github.com/golang/go/issues/58649]
[CITED: github.com/golang/go/issues/22533]

### RQ-4: Per-agent invariant-registry idiom in Go

**Status:** [CITED: kubernetes/apimachinery field validation; pattern adapted]

Kubernetes apimachinery's `field.Error` / `field.ErrorList` is the closest established Go pattern. Their model:

- `Error` is a value type with fields `{Type, Field, BadValue, Detail, Origin}`.
- `ErrorList` is `[]*Error` (slice; ordering preserved).
- Validation functions take an object and return `ErrorList` — they are imperative functions, NOT declarative tags. (Kubernetes 1.36 added declarative validation via `+k8s:` comment tags processed by `validation-gen`, but for our scale that's over-engineered.)

**Adaptation for Phase 2:**

- Replace `field.Error` with our `Violation` (we don't need the Origin tracking).
- Replace `ErrorList` with `[]Violation`.
- Replace the "free functions registered ad-hoc" pattern with the **per-agent typed registry** (D-01/D-02).
- The registry is a `var` slice at package scope. The slice element struct includes the metadata (`ID`, `Description`, `Severity`) AND the typed `Check` function. This gives:
  - **Type safety:** `Check func(*schema.TaintVerdict) []Violation` — compiler enforces the signature per agent.
  - **Discoverability:** `for _, inv := range TaintTracerInvariants { ... }` lists every check at runtime AND at test time (H6 meta-test).
  - **Determinism:** slice iteration is deterministic by spec (Go language spec).

**Other Go projects using similar idioms:**

- **HashiCorp Terraform** uses `schema.Schema` declaratively but bakes validation into `ValidateFunc`/`ValidateDiagFunc` callbacks — slightly more declarative than ours, but again uses typed callbacks per field.
- **etcd's validator pattern** for raft config uses slices of typed check functions.

The Kubernetes pattern is the most directly applicable and cited at top of file.

[CITED: pkg.go.dev/k8s.io/apimachinery/pkg/util/validation/field]
[CITED: github.com/kubernetes/apimachinery/blob/master/pkg/util/validation/field/errors.go]
[CITED: kubernetes.io/docs/reference/using-api/declarative-validation/]

### RQ-5: Test-coverage meta-test technique (D-04, H6)

**Status:** [VERIFIED: pkg.go.dev/testing — `*testing.M` has no enumeration method]

Three options were considered:

1. **Reflection on `testing.InternalTest`.** The `InternalTest` struct is exported but is "internal" by name and is only populated inside the generated test main. There is no public API to retrieve `[]InternalTest` from `*testing.M`. Not viable.
2. **Shell out to `go test -list .` and parse stdout.** Works but adds: a subprocess at test time (slow, fragile), CI complexity (need `go` on PATH), and breaks IDE-only test runs. Not recommended.
3. **Parse `_test.go` files via `go/parser` AST walk inside the meta-test.** Pure stdlib. Runs in the same process as the rest of the test suite. Fast. This is the recommended approach. Sketch in Code Examples above.

**The H6 meta-test does two things:**

a. Assert each per-agent registry's ID set equals the expected set (e.g., `{T1..T11}`). This catches "registered too many" or "forgot one" drift between code and spec.

b. For each registered ID, assert at least one `Test*` function whose name starts with `Test<ID>_` exists in the corresponding `_test.go` file. This catches "added a check predicate but forgot the test."

**Naming convention enforced by H6:** `Test<ID>_<short_camelcase_description>`. Example: `TestT3_PathNonEmptySourceFirstSinkLast`. The `_` after the ID is the discriminator that lets us match by prefix.

[VERIFIED: pkg.go.dev/testing]
[CITED: pkg.go.dev/go/parser; pkg.go.dev/go/ast]

### RQ-6: Table-driven test patterns with `github.com/google/go-cmp`

**Status:** [VERIFIED: pkg.go.dev/github.com/google/go-cmp/cmp + cmpopts]

Established idiom:

```go
got := check(input)
if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
    t.Errorf("%s mismatch (-want +got):\n%s", caseName, diff)
}
```

**Recommendations for Phase 2 tests:**

- **Always include `cmpopts.EquateEmpty()`** when comparing `[]Violation`. A check that returns `nil` for a passing input and a check that returns `[]Violation{}` should both be considered "no violations." Without `EquateEmpty()`, these would compare as different.
- **Always use the order `cmp.Diff(want, got, opts...)`**. The `-` lines in the diff are `want`, `+` lines are `got`. This convention is documented in the go-cmp package docs and matches the failure message `(-want +got)`.
- **For ordered slices (like `[]Violation` where order encodes registration order per D-03):** do NOT use `cmpopts.SortSlices`. Order is part of the contract; sorting hides ordering bugs.
- **For `Violation.Path` field comparison:** straight string compare — `cmp.Diff` handles this trivially. No special transformer needed.

[VERIFIED: pkg.go.dev/github.com/google/go-cmp/cmp]
[VERIFIED: pkg.go.dev/github.com/google/go-cmp/cmp/cmpopts]
[CITED: dave.cheney.net/2019/05/07/prefer-table-driven-tests]

### RQ-7: Cold-start budget (CON-nfr-cold-start <5ms)

**Status:** [VERIFIED: local probe on Linux WSL2, Go 1.24.4]

A minimal `CGO_ENABLED=0 -ldflags='-s -w'` Go binary that does `json.NewDecoder(os.Stdin).DisallowUnknownFields(); dec.Decode(&p); os.Exit(0)` cold-starts in **2–3 ms** on a modest dev machine (WSL2 on Linux kernel 6.6.87). Eight runs, all <5ms:

```
run 1: 3ms (raw ns=3054957)
run 2: 2ms (raw ns=2778147)
run 3: 2ms (raw ns=2802537)
run 4: 2ms (raw ns=2588690)
run 5: 2ms (raw ns=2685094)
run 6: 2ms (raw ns=2677768)
run 7: 2ms (raw ns=2696474)
run 8: 2ms (raw ns=2680709)
```

Binary size 1.8 MB with `-ldflags='-s -w'` strip. `ldd cstest` returns "not a dynamic executable" — confirms static linkage.

**What dominates cold-start at this scale:** ELF load (kernel + libc-free), Go runtime init (scheduler, GC structures), one `os.Stdin` read system call. The 51 invariant checks contribute negligibly because the predicates are simple slice/string/numeric comparisons. Per-payload work (parse JSON + run 11 predicates for a typical agent) adds ~1–2 ms on top of the cold-start floor.

**Recommendation:** Strip with `-ldflags='-s -w'` in the Makefile `build` target. Skip UPX or other binary packing — they add startup decompression and aren't worth it.

**Caveat:** Numbers measured on WSL2; native Linux is usually 10–30% faster. macOS and Windows native are typically slower but still well within the 5 ms budget. CI hardware varies; if any individual measurement drifts above 5 ms, investigate before relaxing the budget.

[VERIFIED: local probe]
[CITED: eli.thegreenplace.net/2024/building-static-binaries-with-go-on-linux/]

### RQ-8: Workspace-root CWD assumption (D-11)

**Status:** [VERIFIED: live docs say hooks run in session CWD + set `$CLAUDE_PROJECT_DIR`]

From https://code.claude.com/docs/en/hooks:

> Handlers run in the current directory with Claude Code's environment. Both forms support the same path placeholders, and **both export them as the environment variables `CLAUDE_PROJECT_DIR`, `CLAUDE_PLUGIN_ROOT`, and `CLAUDE_PLUGIN_DATA` on the spawned process**, so a script can read `process.env.CLAUDE_PROJECT_DIR` regardless of how it was launched.

Also: every hook event payload now includes a top-level `cwd` field (RQ-1).

**Three signals available for workspace-root detection, in preferred order:**

1. **`$CLAUDE_PROJECT_DIR` environment variable** — explicit, official, set by Claude Code on every hook process. THIS IS THE RECOMMENDED SIGNAL.
2. **`event.CWD`** — included in the JSON payload. Same value typically as `$CLAUDE_PROJECT_DIR` but obtained from the input not the env.
3. **`os.Getwd()`** — the process's actual working directory at launch. Same as #2 normally.

**Recommendation for D-11 implementation:** Modify the workspace-root guard to:

```go
func workspaceRoot(eventCWD string) (string, error) {
    // 1. Prefer the documented env var.
    if d := os.Getenv("CLAUDE_PROJECT_DIR"); d != "" && hasWorkspaceMarkers(d) {
        return d, nil
    }
    // 2. Fall back to the event's CWD field.
    if eventCWD != "" && hasWorkspaceMarkers(eventCWD) {
        return eventCWD, nil
    }
    // 3. Last resort: process CWD with marker scan (D-11 spec).
    cwd, err := os.Getwd()
    if err != nil {
        return "", err
    }
    if !hasWorkspaceMarkers(cwd) {
        return "", errors.New("workspace_root_not_found")
    }
    return cwd, nil
}

func hasWorkspaceMarkers(dir string) bool {
    for _, marker := range []string{"go.mod", ".claude", ".planning"} {
        if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
            return true
        }
    }
    return false
}
```

This refines D-11 without changing its intent: the workspace-root assumption is now multi-source, with the official env var preferred. The block reason `"workspace_root_not_found"` still fires when all three signals fail.

[VERIFIED: code.claude.com/docs/en/hooks]

### RQ-9: Subcommand dispatch in Go stdlib

**Status:** [VERIFIED: gobyexample.com + dx13.co.uk + abhinavg.net pattern]

The idiomatic stdlib pattern uses `flag.NewFlagSet(name, flag.ExitOnError)` per subcommand and dispatches on `os.Args[1]`. Each subcommand parses `os.Args[2:]` independently and reads its own stdin. Full example in "Pattern 1" above.

**For Phase 2 specifically:**

- **No subcommand needs CLI flags.** Each subcommand reads exactly one JSON payload on stdin and writes either nothing or a single JSON object on stdout. So the `flag.NewFlagSet` calls in the dispatcher are essentially zero-flag placeholders — but using them anyway preserves the idiom and leaves the door open for future flags (e.g., `--debug`, `--strict`).
- **Each subcommand reads its own stdin.** Pass `os.Stdin` explicitly into each `hooks.Preflight(stdin, stdout, stderr)` etc. so the function is testable with strings.
- **Each subcommand handles its own errors.** Return `error` to `main`. Class A (block decisions) are written to stdout inside the function and return `nil`. Class B (infrastructure failures) return non-`nil` error which `main` writes to stderr + exits 1 (D-14).

[VERIFIED: gobyexample.com/command-line-subcommands]
[CITED: abhinavg.net/2022/08/13/flag-subcommand/]
[CITED: dx13.co.uk/articles/2022/10/12/cli-tools-in-go-with-flag/]

### RQ-10: Stable error-message ordering (D-03)

**Status:** [VERIFIED: Go language spec — slice iteration order is the source order]

The Go specification guarantees that iterating a slice with `for i := 0; i < len(s); i++` or `for i, v := range s` visits elements in index order. This is not an implementation detail; it's a language guarantee.

**Implication for D-03:** Block-reason concatenation walks each registry slice in registration order. Each predicate returns `[]Violation` in whatever order the predicate's body produces them (typically check-order within the predicate). Concatenated with `; ` between violations. No `sort` call required.

**Recommendation:** In `internal/hooks/decision.go`, write:

```go
func formatReason(violations []Violation) string {
    parts := make([]string, 0, len(violations))
    for _, v := range violations {
        if v.Path != "" {
            parts = append(parts, fmt.Sprintf("%s: %s expected '%s' got '%s'", v.ID, v.Path, v.Expected, v.Actual))
        } else {
            parts = append(parts, fmt.Sprintf("%s: %s", v.ID, v.Description))
        }
    }
    return strings.Join(parts, "; ")
}
```

(Where `Violation` is extended to carry `ID` and `Description` at the formatter boundary — the registry adds these in the wiring layer, so individual check functions keep returning the minimal struct shown in Pattern 2.)

**No flakiness possible.** Determinism is mechanical.

[CITED: go.dev/ref/spec — Go programming language specification, "For statements with range clause"]

### RQ-11: Validation Architecture

See "Validation Architecture" section above. Every assertion ID has a row with sampling strategy, minimum case counts, and required fixture. The planner will use this table to materialize VALIDATION.md and to derive the Wave-0 test scaffold.

### RQ-12: Project skill considerations

**Status:** No project-level `.claude/skills/` or `.agents/skills/` directories exist. No project-level `./CLAUDE.md` exists. Only the user's GLOBAL `$HOME/.claude/CLAUDE.md` applies, and its directives map directly to Phase 2 work as captured in "Project Constraints" above:

- "always use TDD" → Wave 0 ships tests/fixtures before Wave 1 ships implementation; within each Wave 1 task, the `_test.go` lands before or alongside the predicate, not after.
- "always use declared types" → D-02 captures this in the per-agent registry shape; ZERO `any` in `internal/invariants` or `internal/schema` non-test code.
- "always use gopls" → implementing session uses gopls for navigation/edits; not a deliverable invariant of the binary.
- "always use golang" → the entire deliverable IS Go; Makefile is the only shell.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `tool_input.subagent_type` is the live field name for the `Task` tool in PreToolUse/PostToolUse | RQ-1 | Validator silently no-ops on every invocation. Mitigated by RQ-1 verification + community blog confirmation (claude-howto, samuellawrentz.com) showing the field name in real hook code. |
| A2 | The "Task" tool and the "Agent" tool listed in `tools-reference` are the same tool (matcher `Task` matches subagent dispatch) | RQ-1 | If they're different, hook matcher won't fire. Mitigated by community hook code examples using `if tool_name === "Task"` in real PreToolUse handlers. |
| A3 | `tool_response.content` for the Task tool is typically a string (not array) but Phase 2 must handle both per H7 | RQ-1 Pitfall 4 | If always string, we waste a few lines of code. If we don't handle array, H7 fails on a real payload. Defensive coding wins. |
| A4 | A10/A11/T11/AZ6 (agent-behavior assertions) can be verified by inspecting fields the agent emits in its verdict (e.g., a `tool_calls` summary). HAND_OFF schemas do NOT explicitly define such a field. | Validation Architecture | If the agent's verdict has no such field, Phase 2 cannot verify these assertions purely from the JSON output — they would require inspecting Claude Code transcript or tool-call event log, which Phase 2 doesn't have access to. **The planner should resolve this with the user before Wave 1.** A pragmatic fallback: A10/A11/T11/AZ6 become "no-op pass" checks in Phase 2 with a TODO note, real enforcement deferred to Phase 5 (E2E smoke test). |
| A5 | Local cold-start probe (2–3 ms) generalizes to typical dev machines and CI runners | RQ-7 | If CI hardware is much slower, we may breach 5 ms. Mitigated by static-binary discipline + Makefile `-ldflags='-s -w'`. |
| A6 | Go 1.22 is sufficient for all required stdlib features (esp. `encoding/json`, `go/parser`, `flag`) | Standard Stack | All features used (DisallowUnknownFields, NewFlagSet, parser.ParseFile, ast.FuncDecl) have been in stdlib since Go 1.10–1.0. Very low risk. |
| A7 | The H5 meta-test using `go list -deps -test -json ./...` reliably identifies test-only deps via the `ForTest` field | Pattern in Code Examples | If `ForTest` semantics change in future Go versions, the meta-test may need adjustment. As of Go 1.24, the field is stable and documented. |

**If any of A1, A2, A4 turn out wrong, the impact is material.** The planner should flag them in discuss-phase before locking the plan if any uncertainty remains.

## Open Questions

1. **Are A10/A11/T11/AZ6 enforceable from agent verdict JSON alone?** (See Assumptions A4.)
   - **What we know:** These are "agent behavior" assertions (tool-usage compliance). HAND_OFF §3.7 says they're "for the hook validator." But the hook receives `tool_response.content` (the verdict JSON), NOT the agent's tool-call log.
   - **What's unclear:** Whether the verdict schema is supposed to include a self-reported `tool_calls_summary` or similar field for the validator to inspect. CON-schema-* doesn't mention one.
   - **Recommendation:** The planner should ask the user (via discuss-phase if needed) one of: (a) "should the verdict schemas grow a `tool_calls_summary` field?", (b) "should A10/A11/T11/AZ6 be deferred to Phase 5 with a no-op stub in Phase 2 (still has H6 test coverage but the check returns nil)?", or (c) "is there a Claude Code mechanism I'm missing that lets the PostToolUse hook see tool-call counts?" Option (b) is the lowest-risk Phase 2 path.

2. **Should the H5 meta-test allow `cmp/cmpopts` in addition to `cmp`?** D-05 says "go-cmp v0.6.0+". `cmpopts` is a subpackage. Verify allowlist in H5 meta-test covers `github.com/google/go-cmp/cmp/cmpopts` too — sketched as `strings.HasPrefix(p.ImportPath, "github.com/google/go-cmp/")` (covers both subpackages).

3. **What's the minimum `tool_response.content` array shape we must handle for H7?** Live docs say "string OR array" with no further detail.
   - **What we know:** The Task tool's typical content for a completing subagent is a string. Array form would arise if Claude Code ever wraps the message into a multi-block response.
   - **Recommendation:** Phase 2 handles "string", "array of {type:'text', text:string}", and "anything else → block with parse-error reason." This covers >99% of cases without overreach.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | Build, test, all dev tasks | ✓ | go1.24.4 (locally probed); go.mod requires ≥ 1.22 | — |
| `make` | Makefile targets | ✓ (presumed; standard on Linux/macOS) | — | If unavailable, document `go build`/`go test` commands as fallback in Phase 6 README |
| `ldd` (or equivalent) | H1 verification step (smoke check that binary is static) | ✓ on Linux/WSL | — | macOS uses `otool -L`; planner may make the smoke step OS-aware in Makefile |
| `git` | Phase commits (configured by gsd workflow) | ✓ | — | — |
| `$CLAUDE_PROJECT_DIR` env var | D-11 / RQ-8 workspace-root resolution | Set by Claude Code at hook process start | — | Falls back to event.CWD then os.Getwd() + marker scan |

**No missing dependencies block Phase 2.** All required tools are present in the dev environment.

## Security Domain

> Phase 2 produces a Go binary that runs inside Claude Code's hook lifecycle. Its threat model is constrained: it processes JSON input from a trusted producer (Claude Code itself), reads workspace source files read-only, and writes only block decisions to stdout. The project's security configuration includes ASVS requirements — the planner should determine whether `security_enforcement` is enabled in `.planning/config.json` (no config file exists yet — treat as enabled-by-default per research instructions).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Phase 2 binary has no auth surface. |
| V3 Session Management | No | Stateless per-invocation binary. |
| V4 Access Control | No | No user-facing access decisions. |
| V5 Input Validation | **Yes** | `encoding/json` + `DisallowUnknownFields()` per D-10. Every JSON input is strictly validated. Block on parse failure (H7). |
| V6 Cryptography | No | No crypto operations. |
| V7 Error Handling & Logging | **Yes** | Two-class error policy (D-14). Errors never leak internal state to stdout (block reasons are scoped to violation IDs + paths). Stderr is reserved for Class B infrastructure failures. |
| V8 Data Protection | **Yes (marginal)** | The binary reads workspace source files for A7/A8/T9/AZ4/OA4/IC3. It does NOT write source files. The file reads are bounded (one Stat + one line-count scan per cited path). No source content is echoed to stdout — only file existence and line-count results are exposed via block reasons. |
| V12 Files & Resources | **Yes** | All filesystem reads are read-only (`os.Stat`, `os.Open` + `bufio.Scanner`, never `os.Create`/`os.Truncate`). Phase 2 binary writes nothing to disk. |
| V14 Configuration | No | No config files in Phase 2 (Phase 1 scaffold; Phase 3 owns settings.json). |

### Known Threat Patterns for the Phase 2 stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via crafted `path` field in verdict (e.g., `"path/../../etc/passwd"`) read by A7/A8/T9/AZ4/OA4/IC3 | Information Disclosure / Tampering | `fsutil` helper must (a) reject absolute paths (`filepath.IsAbs`), (b) `filepath.Clean` then assert the cleaned path stays under the workspace root via `strings.HasPrefix(cleaned, root+string(filepath.Separator))` or `filepath.Rel` + no-`..`-prefix check. Reject with block reason `"path_escapes_workspace"`. Pattern matches §8.2 sanitizer table for `fs_path` sink kind — eat your own dogfood. |
| JSON deserialization gadget (e.g., huge nested arrays exhausting memory) | DoS | Bound input size with `io.LimitReader` wrapper before `json.NewDecoder`. Recommend 1 MB cap (verdicts are KB scale). Anything larger → block with reason `"input_too_large"`. |
| Symlink attack via `os.Open` on a path that resolves outside workspace | Information Disclosure | After `filepath.Clean` + prefix check, also `os.Lstat` and reject if symlink. Or: `filepath.EvalSymlinks` then re-check the resolved path stays under workspace root. |
| Resource exhaustion via deeply-nested JSON (recursive UnmarshalJSON) | DoS | stdlib `encoding/json` is iterative for arrays and objects; depth is bounded by goroutine stack (~10k). Combined with 1 MB input cap, not a practical risk. |
| Stdin starvation (slow producer holds the hook process open until timeout) | DoS | Claude Code already imposes per-hook timeouts (5s preflight/inject, 15s validate). Phase 2 binary need not implement its own deadline. |

### CLAUDE.md / user directives compliance

- "Always use Go": deliverable IS Go.
- "Always use declared types; avoid `any`": D-02 codifies this in the registry shape. The ONE use of `interface{}`/`any` allowed is `json.RawMessage` for `tool_response.content` (because the JSON Schema allows string OR array) — and even that is a declared type (`json.RawMessage` is `type RawMessage []byte`, not `any`).
- "Always use TDD": Wave 0 places test infrastructure before any implementation; within each Wave 1 task, the test file lands first.
- "Always use gopls for code analysis": applies to the implementing session, not the deliverable.

## Sources

### Primary (HIGH confidence — fetched live or probed locally)

- https://platform.claude.com/docs/en/about-claude/models — model identifier `claude-sonnet-4-6` is current as of 2026-05-18 [VERIFIED]
- https://code.claude.com/docs/en/hooks — hook event payload schemas verbatim; `tool_input.subagent_type` confirmed; SubagentStart uses `agent_type` (not `agent_name` as in HAND_OFF) [VERIFIED]
- https://code.claude.com/docs/en/hooks-guide — practical hook implementation guide; confirms `tool_name === "Task"` + `tool_input.subagent_type` pattern [VERIFIED]
- https://code.claude.com/docs/en/tools-reference — Agent/Task tool semantics [VERIFIED]
- Local Go 1.24.4 probe — `DisallowUnknownFields()` nested + array behavior [VERIFIED]
- Local Go 1.24.4 probe — static binary cold-start = 2–3 ms [VERIFIED]
- https://pkg.go.dev/encoding/json — `Decoder.DisallowUnknownFields()` API contract [VERIFIED]
- https://pkg.go.dev/testing — `*testing.M` has no enumeration method, confirming D-04 needs AST walk [VERIFIED]
- https://pkg.go.dev/github.com/google/go-cmp/cmp + cmpopts — `cmp.Diff` and `EquateEmpty` semantics [VERIFIED]
- https://github.com/google/go-cmp/releases — go-cmp v0.7.0 (2024-02-21) is latest [VERIFIED]
- https://go.dev/doc/devel/release — Go release history (1.22 stable, 1.24 latest) [VERIFIED]

### Secondary (MEDIUM confidence — official patterns and community blogs)

- https://gobyexample.com/command-line-subcommands — canonical `flag.NewFlagSet` subcommand idiom [CITED]
- https://abhinavg.net/2022/08/13/flag-subcommand/ — secondary subcommand pattern [CITED]
- https://dave.cheney.net/2019/05/07/prefer-table-driven-tests — table-driven test idioms [CITED]
- https://pkg.go.dev/k8s.io/apimachinery/pkg/util/validation/field — Kubernetes field error pattern [CITED]
- https://github.com/kubernetes/apimachinery/blob/master/pkg/util/validation/field/errors.go — implementation reference [CITED]
- https://kubernetes.io/docs/reference/using-api/declarative-validation/ — Kubernetes declarative validation (overkill for us but referenced for comparison) [CITED]
- https://manpages.debian.org/testing/golang-go/go-list.1.en.html — `go list -deps -test -json` reference [CITED]
- https://eli.thegreenplace.net/2024/building-static-binaries-with-go-on-linux/ — static binary build guide [CITED]

### Tertiary (LOW confidence — single-source, included for completeness)

- https://github.com/disler/claude-code-hooks-mastery — community hooks reference [CITED — used to cross-check field names]
- https://github.com/luongnv89/claude-howto/blob/main/06-hooks/README.md — community guide; confirms `tool_input.subagent_type` field name in `Task` tool hook handlers [CITED]
- https://github.com/golang/go/issues/22533 — embedded-struct + DisallowUnknownFields known bug [CITED — for the Pitfall 5 note]
- https://github.com/golang/go/issues/58649 — DisallowUnknownFields path-in-error feature request [CITED — for documenting the error-message limitation]

## Pitfalls & Landmines

(Aggregated from "Common Pitfalls" above and emphasized here so the planner can lift into VALIDATION.md or per-task verification steps.)

1. **`subagent_type` vs `agent_type`** — field name differs between PreToolUse/PostToolUse (`tool_input.subagent_type`) and SubagentStart (`agent_type` at top level). Don't conflate.
2. **`DisallowUnknownFields` rejects ANY extra field** — including `permission_mode`, `effort`, `tool_use_id` from the live hook payload. Phase 2 event structs MUST include every documented field even if unused, or every real payload blocks.
3. **`tool_response.content` is string OR array** — model as `json.RawMessage`, decode both shapes, never panic (H7).
4. **`go-cmp` only in `_test.go`** — write H5 meta-test FIRST (Wave 0), not last. Catches accidental imports immediately.
5. **A10/A11/T11/AZ6 are likely undecidable from verdict JSON alone** — open question #1. Planner should resolve with the user before Wave 1.
6. **Map iteration is non-deterministic** — never use a map for the invariant registry. Slice only (D-01).
7. **CWD assumption** — use `$CLAUDE_PROJECT_DIR` env var (preferred per live docs), then `event.CWD`, then process CWD + marker scan (D-11 fallback).
8. **Strict schema vs. evolving agent prompts** — Phase 4+ schema bumps require coordinated commit of both Go struct and agent prompt. Document loudly.
9. **`internal/invariants` MUST NOT import `internal/hooks`** — predicates are I/O-free; the wiring layer is the only place that touches stdin/stdout.
10. **Don't define custom `UnmarshalJSON` on schema structs** — interacts badly with `DisallowUnknownFields` for embedded types (RQ-3 caveat).

## Recommendations to Planner

### Suggested plan structure: 6 plans in 3 waves

**Wave 0 — Test infrastructure + shared types (single plan, blocks all of Wave 1):**

- **Plan 02-01: "Test infrastructure + schema scaffolding + meta-tests."**
  - Add `go-cmp@v0.7.0` to go.mod.
  - Create `internal/invariants/{registry.go,fsutil.go,fsutil_test.go,h5_deps_test.go,h6_coverage_test.go}` (latter two as scaffolds — the expected-IDs map starts empty and grows as Wave 1 lands).
  - Create `internal/hooks/{events.go,events_test.go,agents.go,agents_test.go,doc.go,decision.go,decision_test.go}`.
  - Create `internal/schema/{cartographer.go,taint_tracer.go,authz_tracer.go,oauth_auditor.go,invariant_checker.go,synthesis.go}` with structs only (no behavior).
  - Create `testdata/workspace/` fixture and `testdata/graphify-out/graph.json` minimal fixture.
  - Create `Makefile` with `build`, `test`, `install`, `clean`, `lint` targets.
  - **AC:** `go build ./...` succeeds; `go test ./internal/hooks/... ./internal/invariants/...` runs the meta-tests + fsutil tests successfully; H5 meta-test confirms zero non-stdlib deps in non-test paths; H6 meta-test scaffolded but vacuously passes (empty expected-IDs).

**Wave 1 — Six per-agent plans, fully parallel after Wave 0:**

These can be executed by different sessions in parallel because each plan touches only its agent's three files (`internal/schema/<agent>.go`, `internal/invariants/<agent>.go`, `internal/invariants/<agent>_test.go`) plus appending an entry to `h6_coverage_test.go`'s expected-IDs map.

- **Plan 02-02: "Cartographer schema + A1–A11 predicates + tests."** Owns: `internal/schema/cartographer.go` (full struct tree), `internal/invariants/cartographer.go` (11 predicates), `internal/invariants/cartographer_test.go` (≥ 22 test cases covering 11 IDs). Append `"cartographer_test.go": ids("A", 1, 11)` to h6 map.
- **Plan 02-03: "Taint tracer schema + T1–T11 predicates + tests."** Same shape for taint_tracer.
- **Plan 02-04: "Authz tracer schema + AZ1–AZ6 predicates + tests."**
- **Plan 02-05: "OAuth auditor schema + OA1–OA7 predicates + tests."**
- **Plan 02-06: "Invariant checker schema + IC1–IC4 predicates + tests."**
- **Plan 02-07: "Synthesis schema + S1–S6 predicates + tests + testdata/synthesis_out/."**

Each Wave 1 plan must:
1. Land the `_test.go` file FIRST in the task ordering (user CLAUDE.md TDD rule).
2. Have the test cases derived from the Validation Architecture table above.
3. Update the H6 meta-test's expected-IDs map; H6 stays green at the end of each plan.

**Wave 2 — Subcommand wiring + integration tests + H7 hardening (single plan, depends on ALL Wave 1):**

- **Plan 02-08: "Subcommand bodies + integration tests + H1/H7 hardening."**
  - `cmd/claude-security-hooks/main.go` (subcommand dispatch per Pattern 1).
  - `internal/hooks/preflight.go` + `_test.go` (D-09 logic).
  - `internal/hooks/validate.go` + `_test.go` (CON-validate-protocol).
  - `internal/hooks/inject.go` + `_test.go` (D-08 stub).
  - Integration tests with fixtures from each agent's verdict shape.
  - H7 hardening tests (malformed string, valid wrong-type, empty, array shape).
  - Makefile `build` step verified to produce static binary; smoke ldd step recommended.
  - **AC:** `go test ./...` is fully green; `make build` succeeds; `bin/claude-security-hooks` runs cold-start <5 ms; Phase 2 ROADMAP success criteria 1–5 all met.

### Suggested execution order with dependency edges

```
Wave 0:  [02-01]
              │
   ┌──────────┼──────────┬──────────┬──────────┬──────────┐
   │          │          │          │          │          │
Wave 1: [02-02] [02-03] [02-04] [02-05] [02-06] [02-07]
   │          │          │          │          │          │
   └──────────┴──────────┴──────────┴──────────┴──────────┘
              │
Wave 2:  [02-08]
```

### Resolve before Wave 1

The planner SHOULD pull these into discuss-phase or otherwise resolve with the user before Wave 1 starts:

1. **A10/A11/T11/AZ6 enforceability** (Open Question 1, Assumption A4). Without a decision, Wave 1 plans for affected agents can't write the predicates. Recommended default: implement as no-op pass with TODO note + H6 test that exercises the no-op shape; full enforcement deferred to Phase 5.
2. **D-11 update to also consult `$CLAUDE_PROJECT_DIR`** (RQ-8). This is a small refinement, not a controversial change. Planner can quietly fold it into Plan 02-01's events.go task without re-opening discuss-phase.
3. **HAND_OFF §8.1 SubagentStart field name correction (`agent_name` → `agent_type`)** (RQ-1). Same as above: small refinement to fold in.

### Don't overthink

- Each Wave 1 plan is **one file per agent for schema + one file per agent for invariants + one test file per agent**. Three files. Don't split per-assertion (e.g., one file per check) unless the agent file balloons past ~400 lines (D's Claude's-Discretion note).
- Tests are **table-driven**, one Test* function per assertion ID, named `Test<ID>_<short>` (RQ-5 / D-06). Don't invent fancier naming.
- Fixtures live in `testdata/` per Go convention (`go test` excludes it from compilation automatically).
- Wave 2 (subcommand wiring) is **last** because it's the only layer with stdin/stdout. Touching it first would require mocking the invariants you haven't yet written.

## Metadata

**Confidence breakdown:**

- **Standard stack:** HIGH — every package version verified against current releases.
- **Architecture / module layout:** HIGH — HAND_OFF §3.7 + Phase 1 scaffold + verified Go subcommand idiom.
- **Hook payload schemas:** HIGH for PreToolUse/PostToolUse/SubagentStart (verified at live docs); MEDIUM on the exact shape of `tool_response.content` array variant (live docs say "string or array" without further detail).
- **Per-assertion test sampling:** HIGH for predicate-level checks (A1–A11, T1–T10, AZ1–AZ5, OA1–OA7, IC1–IC4, S1–S6); MEDIUM for behavioral checks (A10, A11, T11, AZ6 — see Open Question 1).
- **Cold-start budget:** HIGH (locally measured 2–3 ms on representative hardware).
- **Model identifier (Q9):** HIGH (verified at platform.claude.com 2026-05-18).
- **`DisallowUnknownFields` semantics:** HIGH (locally probed across nested + array cases).
- **Test enumeration strategy:** HIGH (verified `testing.M` has no public enumeration API; `go/parser` is the standard fallback).

**Research date:** 2026-05-18
**Valid until:** 2026-06-17 (30 days for stable concerns; hook payload schemas could shift sooner — re-verify if Phase 2 work spans more than 30 days from research date).

## RESEARCH COMPLETE
