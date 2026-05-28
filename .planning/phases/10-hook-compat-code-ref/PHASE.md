# Phase 10: Hook compatibility + code-ref schema

**Status:** Planning  
**Started:** 2026-05-26  
**Depends on:** Phase 9 (opengrep-mcp integration)

---

## Goal

Unblock the Phase 5 smoke test by resolving four runtime blockers discovered during the first live `/security-review` run against `examples/sample-vulnerable-service`, and add a `code_ref` field to all review artifacts so they can be correlated to the exact code version they analyzed.

---

## Blockers Found During First Live Run

| ID | Where | Symptom | Impact |
|----|--------|---------|--------|
| B1 | `hooks/events.go` → `TaskToolInput` | `run_in_background` (and `model`, `isolation`) in Agent tool calls rejected by `DisallowUnknownFields` → "H7: PreToolUseEvent parse: json: unknown field" | Parallel agent dispatch is **impossible**; forced to run all specialists sequentially |
| B2 | `invariants/taint_tracer.go` → `checkT6` | T6 doesn't exempt `verdict="ambiguous"`; any verdict produced when Semgrep + gopls are both unavailable is blocked even when the agent correctly reports ambiguity | Taint verdicts with degraded tooling are always blocked regardless of how honest the agent is |
| B3 | `.claude/agents/go-taint-tracer.md` → Hard Rules | No rule requiring `path[*].file` to be workspace-relative; agent wrote absolute paths (`/home/saghaulor/...`) that fail T9's `FileExists` after `os.Chdir(root)` | All non-trivial taint verdicts blocked at PostToolUse |
| B4 | `.claude/commands/security-review.md` | Skill doesn't verify binary freshness before dispatching agents; stale binary silently blocks the pipeline until manually rebuilt | Silently broken after any commit to the hooks source without a rebuild |

---

## Schema Enhancement

| ID | What | Rationale |
|----|------|-----------|
| E1 | Add `code_ref string` and `code_ref_dirty bool` to `CartographerIndex` and `SynthesisReport` | Review assets (go-index.json, review-report.json) need to record what code version they were produced against. A UUID identifies the run; a git tree hash identifies the code. Multiple runs on the same commit get the same `code_ref` but different `review_id`. |

### code_ref design

`code_ref` is the **git tree object hash of the reviewed directory**:

```bash
git rev-parse HEAD:<target_dir_relative_to_repo_root>
# e.g. git rev-parse HEAD:examples/sample-vulnerable-service
# → 5a3f8c2de1b7f9c4a3d2e1f0...
```

Using the tree hash (not the commit hash) means:
- `code_ref` changes only when the reviewed directory changes, not when unrelated files do
- It directly identifies the exact tree that was analyzed
- It's reproducible: `git ls-tree HEAD <path>` gives the same hash

For **dirty working trees** (uncommitted changes), `code_ref_dirty: true` is set and `code_ref` records the HEAD tree hash with a note that local modifications were present. This matches the convention used by `goreleaser` and `go build -ldflags -X` version strings.

`code_ref` lives alongside `review_id` in each artifact — they serve different purposes:
- `review_id` (UUID) = run identity (links all artifacts from one run)
- `code_ref` (git tree hash) = code identity (links artifacts to one code state)

---

## Work Breakdown

### Plan 10-01 — Wave 0: Failing tests (RED)

Write tests that currently fail and will pass after Wave 1:
- `events_test.go`: 4 tests — Agent tool fields `run_in_background`, `model`, `isolation` currently rejected (B1)
- `taint_tracer_test.go`: T6 blocks `verdict="ambiguous"` (B2); T-CodeRef joint invariant missing (E1)
- `authz_tracer_test.go`, `oauth_auditor_test.go`, `invariant_checker_test.go`: AZ/OA/IC-CodeRef joint invariants missing (E1)
- Schema compile-error tests: `code_ref` + `code_ref_dirty` round-trip for all 10 structs — all 6 verdict structs (`CartographerIndex`, `TaintVerdict`, `AuthzVerdict`, `OAuthVerdict`, `InvariantCheckerVerdict`, `SynthesisReport`) and all 4 input structs (`TaintInput`, `AuthzInput`, `OAuthInput`, `InvariantCheckerInput`)

### Plan 10-02 — Wave 1: Go implementation (GREEN)

Make Wave 0 tests pass:
- Add `RunInBackground`, `Model`, `Isolation` (omitempty) to `TaskToolInput` in `events.go`
- Fix `checkT6` to exempt `verdict="ambiguous"` alongside `"input_mismatch"`
- Add `CodeRef string` + `CodeRefDirty bool` to all 10 schema structs (6 verdicts + 4 inputs)
- Add `T-CodeRef`, `AZ-CodeRef`, `OA-CodeRef`, `IC-CodeRef` joint invariants: "if input `code_ref` non-empty, verdict `code_ref` must match exactly"
- Rebuild and reinstall binary (`make install`)
- Run full test suite green

### Plan 10-03 — Wave 2: Agent definitions + skill (text changes)

Non-Go fixes:
- `go-taint-tracer.md`: Add hard rule T9-PATH — `path[*].file` MUST be workspace-relative (never absolute)
- `go-cartographer.md`: Instruct agent to write pre-computed `code_ref` + `code_ref_dirty` (from its input) into go-index.json
- `go-authz-tracer.md`, `go-oauth-auditor.md`, `invariant-checker.md`, `go-taint-tracer.md`: Instruct each tracer to echo `code_ref` + `code_ref_dirty` from its JSON input into its JSON output
- `synthesis.md`: Propagate `code_ref` + `code_ref_dirty` from `go-index.json` into `review-report.json`
- `security-review.md`: Step 1.5 (unconditional `make install`); Step 2 computes `CODE_REF` + `CODE_REF_DIRTY`; Steps 4 and 6 thread both into every Task input

---

## Success Criteria

1. `go test ./...` passes in `claude-security-hooks/` with all new Wave 0 tests green
2. An Agent call with `run_in_background: true` is NOT blocked by the preflight hook
3. A `TaintVerdict` with `verdict="ambiguous"`, `semgrep.ran=false`, `gopls.references_calls=0` passes T6
4. All 10 schema structs round-trip `code_ref` and `code_ref_dirty` correctly
5. A `TaintVerdict` with `code_ref` differing from its input's `code_ref` is blocked by T-CodeRef (same for AZ/OA/IC)
6. `go-taint-tracer.md` §6 contains T9-PATH requiring workspace-relative paths
7. `security-review.md` Step 1.5 runs `make install` before agent dispatch
8. All four tracer agent definitions instruct echoing `code_ref` from input to output
9. `review-report.json` contains `code_ref` propagated from `go-index.json`
10. A full `/security-review examples/sample-vulnerable-service` run completes without any hook-driven block on B1–B4

---

## Plans

- [x] 10-01-PLAN.md — Wave 0: Failing tests (RED) — events_test, 4 tracer invariant tests, 10 schema round-trips
- [x] 10-02-PLAN.md — Wave 1: Go implementation (GREEN) — TaskToolInput, checkT6, 10 schema fields, 4 code_ref joint invariants, binary rebuild
- [x] 10-03-PLAN.md — Wave 2: Agent defs + skill — T9-PATH, code_ref echo in 4 tracers, cartographer write, synthesis propagate, Step 1.5 + Step 2 + Steps 4/6 in skill
