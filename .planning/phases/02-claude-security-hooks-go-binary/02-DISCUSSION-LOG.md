# Phase 2: claude-security-hooks Go binary - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-18
**Phase:** 02-claude-security-hooks-go-binary
**Areas discussed:** Invariant registry shape, Test framework choice, Subagent-set source of truth, inject-context v1 scope, Derived defaults (preflight strictness + JSON strictness + workspace-CWD assumption)

---

## Invariant registry shape

| Option | Description | Selected |
|--------|-------------|----------|
| Structured slice with metadata | `type Invariant struct{ ID, Description string; Severity; Check func(v any) []Violation }`. Registry is a slice carrying metadata; block-reasons include path + expected/actual. H6 enumerable via `range`. | ✓ |
| Flat map of ID → predicate | `map[string]func(v any) error`. Simpler; harder to enforce message format. Map iteration order non-deterministic. | |
| Interface with typed predicates per agent | One struct per assertion implementing a typed interface. Type-safe but many files. | |

**User's choice:** Structured slice with metadata (Recommended).
**Notes:** Per the user's global CLAUDE.md no-`any` rule, refined the slice option to be PER-AGENT registries each with its OWN typed `Check` function (e.g., `func(*schema.TaintVerdict) []Violation`). No shared `any`-typed signature. Captured as D-01 and D-02 in CONTEXT.md. Block-reason format standardized as D-03; mechanical H6 coverage test as D-04.

---

## Test framework choice

| Option | Description | Selected |
|--------|-------------|----------|
| stdlib `testing` + table-driven tests | Pure stdlib. Zero deps even in tests. Easiest audit. | |
| stdlib `testing` + `github.com/google/go-cmp` | Stdlib runners; `cmp.Diff` for better Violation-slice diffs in failure output. One non-stdlib test dep. | ✓ |
| stdlib `testing` + `testify/require` | Familiar ergonomics; less boilerplate. Pulls testify+objx+pmezard. | |

**User's choice:** stdlib `testing` + `github.com/google/go-cmp`.
**Notes:** Captured as D-05 (go-cmp is the only permitted non-stdlib dep, confined to `_test.go` — enforced by an H5 unit test parsing `go list -deps -test ./...`) and D-06 (table-driven, one file per agent, `Test<ID>_<short-description>` naming so `-run TestT3_` selects a single assertion).

---

## Subagent-set source of truth

| Option | Description | Selected |
|--------|-------------|----------|
| Hardcoded const slice in `internal/hooks` | Zero coupling to Phase 3, fastest cold start, trivial test. Drift risk mitigated by spec-match unit test. | ✓ |
| Generated from `.claude/agents/*.md` at build time | `go generate` discipline. Forces Phase 3 dependency before Phase 2 builds. | |
| Read `.claude/agents/*.md` at hook startup | Zero drift. Adds 1–3ms I/O per hook call (concerning vs <5ms budget). | |

**User's choice:** Hardcoded const slice in `internal/hooks` (Recommended).
**Notes:** Captured as D-07. The drift-mitigation test `TestSecurityAgentSet_MatchesSpec` lives in `internal/hooks/agents_test.go` and asserts the const matches the documented set verbatim. Updating the set after Phase 3 adds an agent requires editing BOTH the const AND the test — visible signal.

---

## `inject-context` v1 scope

| Option | Description | Selected |
|--------|-------------|----------|
| No-op stub for Phase 2 | Subcommand exists, reads + validates stdin, exits 0 silently. Satisfies H2 but defers real injection until protocol stabilizes. | ✓ |
| Static deterministic context now | Emit fixed per-agent JSON (semgrep_tier, schema_versions, etc.) on stdout. Locks in choices before Phase 4 ships scanner. | |
| Read from `.claude/security-invariants/*.yaml` | Per-agent YAML files. Adds YAML parsing (third-party dep — violates CON-nfr-zero-deps) or forces rename to `.json`. | |

**User's choice:** No-op stub for Phase 2 (Recommended).
**Notes:** Captured as D-08. Rationale: no IC*/H* assertion targets `inject-context`; how Claude Code consumes injected context is not fully specified in HAND_OFF.md. Defer to Phase 3 or Phase 5. Phase 2 still implements the subcommand (REQ-hooks-H2) — it just contributes nothing semantically.

---

## Derived defaults (preflight strictness + JSON strictness + workspace CWD)

| Option | Description | Selected |
|--------|-------------|----------|
| Preflight REJECTs malformed dispatch (matches security agent name but bad prompt JSON) | Symmetric with validate H7. Without this, malformed dispatches reach the subagent and waste tokens. | ✓ |
| JSON decoder uses `DisallowUnknownFields` for verdict parsing | Strict schema. New unknown fields → block. Aligned with D13/D14. Schema migrations require coordinated commits. | ✓ |
| A7/A8 file-existence and line-count checks assume CWD = workspace root | Use `os.Stat` + `bufio.Scanner`. Guards block-reason emitted if workspace markers (`go.mod`, `.claude/`, `.planning/`) missing. | ✓ |

**User's choice:** All three selected (multi-select).
**Notes:** Captured as D-09 (preflight strictness), D-10 (JSON unknown-fields), D-11 (CWD assumption + workspace-marker guard). The CWD assumption is the most likely thing to need adjustment during execution; D-11 encodes the assumption AND the guard so any divergence surfaces as a structured block decision rather than a silent failure.

---

## Claude's Discretion

The user delegated these to Claude:
- Internal Go package import paths within the `claude-security-hooks` module.
- Specific Go struct field names within `internal/schema/*` (constrained to mirror JSON keys per CON-schema-*).
- Test fixture data shapes inside `_test.go` files.
- Specific error message wording for block reasons beyond the D-03 format pattern.
- Whether to add a `--version` flag (deferred to Phase 6).
- Code-fence stripping algorithm in `validate`.
- Whether each `Check` function lives in `internal/invariants/<agent>_check_<id>.go` (one file per ID) or `internal/invariants/<agent>.go` (one file per agent); planner picks based on resulting file size.

## Deferred Ideas

- `--version` flag → Phase 6.
- Real `inject-context` semantic injection → Phase 3 or Phase 5.
- Build-time generation of `SecurityAgentSet` from `.claude/agents/*.md` → post-v1 if drift becomes painful.
- YAML for `.claude/security-invariants/*.yaml` → indefinitely deferred; prefer `.json` to preserve CON-nfr-zero-deps.
- CI integration (Q10 from HAND_OFF §5) → out of scope for v1 per PROJECT.md.
