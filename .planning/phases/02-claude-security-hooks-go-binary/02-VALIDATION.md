---
phase: 2
slug: claude-security-hooks-go-binary
status: planned
nyquist_compliant: true
wave_0_complete: false
created: 2026-05-19
plans_assigned: 2026-05-19
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Source: `02-RESEARCH.md` § "Validation Architecture" (lines 794-922).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | stdlib `testing` (Go 1.22+) + `github.com/google/go-cmp` v0.7.0 (test-only, per D-05) |
| **Config file** | none — `go test` defaults are sufficient |
| **Quick run command** | `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./internal/invariants -run Test<ID>_` — single assertion (<100 ms) |
| **Per-agent suite** | `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./internal/invariants -run TestT` (taint), `TestA` (cartographer), `TestAZ`, `TestOA`, `TestIC`, `TestS` |
| **Full suite command** | `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./...` |
| **Build smoke** | `make -C /home/saghaulor/code/security_reviewer/claude-security-hooks build` then `make -C /home/saghaulor/code/security_reviewer/claude-security-hooks verify-static` (uses `file` command per portability) — run each as a separate Bash call |
| **Estimated runtime** | ~5 seconds full suite; ~100 ms targeted |

---

## Sampling Rate

- **After every task commit:** `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./internal/invariants -run Test<ID>_` for the assertion(s) the task implemented.
- **After every plan wave:** `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./...`.
- **Before `/gsd-verify-work`:** `make -C ... build` then `make -C ... test` (separate calls) — both green.
- **Max feedback latency:** **5 seconds** (full suite).

---

## Per-Task Verification Map

Each assertion is owned by the plan listed in the **Plan** column. Plans land
tests FIRST per the TDD discipline, so the test rows below describe what the
plan ships.

### Cartographer (A1–A11) — REQ-cartographer-A1..A11

| ID | Test Type | Min Cases (pos+neg) | Fixture | Plan | Status |
|----|-----------|---------------------|---------|------|--------|
| A1 | unit (table-driven) | 1+2 | inline JSON | 02-02 | ⬜ pending |
| A2 | unit | 1+1 | inline | 02-02 | ⬜ pending |
| A3 | unit | 1+4 | inline | 02-02 | ⬜ pending |
| A4 | unit | 1+1 | inline | 02-02 | ⬜ pending |
| A5 | unit | 1+1 | inline | 02-02 | ⬜ pending |
| A6 | unit + filesystem | 1+2 | `testdata/graphify-out/graph.json` | 02-02 | ⬜ pending |
| A7 | unit + filesystem | 1+1 | `testdata/workspace/` | 02-02 | ⬜ pending |
| A8 | unit + filesystem | 1+1 | `testdata/workspace/` (known line counts) | 02-02 | ⬜ pending |
| A9 | unit | 1+1 | inline | 02-02 | ⬜ pending |
| A10 | unit (no-op stub) | 1+1 | inline; **no-op pass per Phase 2 decision; TODO(phase-5)** | 02-02 | ⬜ pending |
| A11 | unit (no-op stub) | 1+1 | inline; **no-op pass per Phase 2 decision; TODO(phase-5)** | 02-02 | ⬜ pending |

### Taint Tracer (T1–T11) — REQ-taint-T1..T11

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| T1 | unit | 1+2 | inline | 02-03 | ⬜ pending |
| T2 | unit | 5+1 | inline | 02-03 | ⬜ pending |
| T3 | unit | 1+3 | inline | 02-03 | ⬜ pending |
| T4 | unit | 1+0 | inline | 02-03 | ⬜ pending |
| T5 | unit | 3+1 | inline | 02-03 | ⬜ pending |
| T6 | unit | 3+1 | inline | 02-03 | ⬜ pending |
| T7 | unit (joint input+output) | 1+1 | signature accepts verdict AND input record | 02-03 | ⬜ pending |
| T8 | unit (joint) | 2+1 | inline | 02-03 | ⬜ pending |
| T9 | unit + filesystem | 1+1 | `testdata/workspace/` | 02-03 | ⬜ pending |
| T10 | unit | 2+1 | inline | 02-03 | ⬜ pending |
| T11 | unit (no-op stub) | 1+1 | inline; **no-op pass per Phase 2 decision; TODO(phase-5)** | 02-03 | ⬜ pending |

### Authz Tracer (AZ1–AZ6) — REQ-authz-AZ1..AZ6

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| AZ1 | unit | 1+2 | inline | 02-04 | ⬜ pending |
| AZ2 | unit (joint) | 1+1 | predicate takes input+output | 02-04 | ⬜ pending |
| AZ3 | unit | 1+1 | inline | 02-04 | ⬜ pending |
| AZ4 | unit (joint) | 1+1 | inline | 02-04 | ⬜ pending |
| AZ5 | unit (joint) | 1+1 | inline | 02-04 | ⬜ pending |
| AZ6 | unit (no-op stub) | 1+1 | inline; **no-op pass per Phase 2 decision; TODO(phase-5)** | 02-04 | ⬜ pending |

### OAuth Auditor (OA1–OA7) — REQ-oauth-OA1..OA7

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| OA1 | unit | 1+2 | inline | 02-05 | ⬜ pending |
| OA2 | unit | 1+1 | inline; checker uses regex over `RFC|draft-` | 02-05 | ⬜ pending |
| OA3 | unit | 1+1 | inline | 02-05 | ⬜ pending |
| OA4 | unit (joint) | 1+1 | inline | 02-05 | ⬜ pending |
| OA5 | unit (joint) | 1+1 | inline | 02-05 | ⬜ pending |
| OA6 | unit (joint) | 1+1 | inline | 02-05 | ⬜ pending |
| OA7 | unit (joint) | 1+1 | inline | 02-05 | ⬜ pending |

### Invariant Checker (IC1–IC4) — REQ-invariant-IC1..IC4

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| IC1 | unit | 1+2 | inline | 02-06 | ⬜ pending |
| IC2 | unit (joint) | 1+2 | inline | 02-06 | ⬜ pending |
| IC3 | unit | 1+2 | inline | 02-06 | ⬜ pending |
| IC4 | unit (joint) | 1+1 | inline | 02-06 | ⬜ pending |

### Synthesis (S1–S6) — REQ-synthesis-S1..S6

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| S1 | filesystem | 1+2 | `testdata/synthesis_out/{both,no_md,no_json}/` | 02-07 | ⬜ pending |
| S2 | unit | 1+1 | inline JSON | 02-07 | ⬜ pending |
| S3 | unit | 1+1 | inline | 02-07 | ⬜ pending |
| S4 | unit | 1+1 | inline | 02-07 | ⬜ pending |
| S5 | unit | 1+1 | inline | 02-07 | ⬜ pending |
| S6 | unit | 1+1 (+recovery) | inline; includes unnoted-duplicate negative AND noted-duplicate recovery positive | 02-07 | ⬜ pending |

### Hook Build/Runtime Invariants (H1–H7) — REQ-hooks-H1..H7

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| H1 | Makefile smoke (`make verify-static`) | 1 | `make -C ... build` exits 0; `file bin/...` reports "statically linked" | 02-08 (Wave 0 02-01 lays Makefile) | ⬜ pending |
| H2 | integration unit | 3 | inline JSON: PreToolUse, PostToolUse, SubagentStart fixtures (Wave 0 events tests + Wave 2 subcommand tests) | 02-01 (events) + 02-08 (subcommands) | ⬜ pending |
| H3 | integration unit + end-to-end smoke | 1+1 | inline JSON: `subagent_type="general-purpose"`; smoke via piped printf to compiled binary, asserts empty stdout | 02-08 | ⬜ pending |
| H4 | integration unit + end-to-end smoke | 1+1 | inline; any failing invariant ⇒ single block JSON, exit 0; smoke via piped printf, asserts `"decision":"block"` present | 02-08 | ⬜ pending |
| H5 | meta-test (`h5_deps_test.go`) | 1 | runs `go list -deps -test ./...` via `os/exec` | 02-01 | ⬜ pending |
| H6 | meta-test (`h6_coverage_test.go`) | 1 | parses `_test.go` via `go/parser`; expectedIDs map FULLY PRE-POPULATED in Wave 0 (sealed); Wave 1 plans contribute test FUNCTIONS only (no map edits) | 02-01 (full map seal) + 02-02..02-07 (test funcs) | ⬜ pending |
| H7 | integration unit + end-to-end smoke | 4+1 | inline: invalid JSON, valid-JSON-wrong-type, empty, multi-segment array; smoke via piped printf with malformed `tool_response.content` | 02-08 | ⬜ pending |

*Status legend: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

**Coverage check:** 58 total IDs (51 assertion predicates + 7 H invariants). Every Phase 2 REQ-ID is represented above. Every assertion has an owning plan.

---

## Wave 0 Requirements

Per RESEARCH.md § "Wave 0 Gaps" (lines 904-923) — the following infrastructure does NOT exist in the Phase 1 scaffold and is created in Plan 02-01:

- [ ] `claude-security-hooks/internal/invariants/registry.go` — shared `Severity` + `Violation` types (D-01). **Plan 02-01 Task 01.2**
- [ ] `claude-security-hooks/internal/invariants/fsutil.go` — `FileExists`, `DirExists`, `LineCount`, `ResolveWorkspaceRoot` helpers. **Plan 02-01 Task 01.2**
- [ ] `claude-security-hooks/internal/invariants/fsutil_test.go` — covers helpers including traversal/symlink rejection (security T-02-01-01/T-02-01-05). **Plan 02-01 Task 01.2**
- [ ] `claude-security-hooks/internal/invariants/h5_deps_test.go` — H5 meta-test. **Plan 02-01 Task 01.5**
- [ ] `claude-security-hooks/internal/invariants/h6_coverage_test.go` — H6 meta-test scaffold with FULL pre-populated expectedIDs map (sealed Wave 0; Wave 1 plans do not modify it). **Plan 02-01 Task 01.5**
- [ ] `claude-security-hooks/internal/invariants/testdata/workspace/` — fixtures with known line counts. **Plan 02-01 Task 01.2**
- [ ] `claude-security-hooks/internal/invariants/testdata/graphify-out/graph.json` — minimal valid graph for A6. **Plan 02-01 Task 01.5**
- [ ] `claude-security-hooks/internal/invariants/testdata/synthesis_out/{both,no_md,no_json}/` — fixture trees for S1. **Plan 02-01 Task 01.5**
- [ ] `claude-security-hooks/internal/hooks/events.go` + `events_test.go` — live-docs deltas (D-15 fold-in). **Plan 02-01 Task 01.3**
- [ ] `claude-security-hooks/internal/hooks/decision.go` + `_test.go` — D-03 formatter. **Plan 02-01 Task 01.3**
- [ ] `claude-security-hooks/internal/hooks/agents.go` + `agents_test.go` — `SecurityAgentSet` + drift test (D-07). **Plan 02-01 Task 01.3**
- [ ] `claude-security-hooks/internal/hooks/doc.go` — D-14 exit-code package doc. **Plan 02-01 Task 01.3**
- [ ] `claude-security-hooks/internal/schema/{cartographer,taint_tracer,authz_tracer,oauth_auditor,invariant_checker,synthesis}.go` — 6 verdict + 5 input struct files mirroring HAND_OFF schemas. **Plan 02-01 Task 01.4**
- [ ] `claude-security-hooks/internal/schema/doc.go` — D-10 strict-decode contract doc. **Plan 02-01 Task 01.4**
- [ ] `claude-security-hooks/Makefile` — `build`, `test`, `install`, `clean`, `lint`, `verify-static`. **Plan 02-01 Task 01.1**
- [ ] `go-cmp` declared as test dep (creates `go.sum`). **Plan 02-01 Task 01.1**

---

## Wave 2 Additions (Plan 02-08)

Files added in Plan 02-08 that enable end-to-end H1–H7 coverage:

- [ ] `claude-security-hooks/cmd/claude-security-hooks/main.go` — subcommand dispatch via `switch os.Args[1]` (NO `flag.NewFlagSet` — Phase 2 subcommands take no flags)
- [ ] `claude-security-hooks/internal/hooks/preflight.go` + `_test.go` — D-09 input strict decode (inline typed per agent; no interface{} helper)
- [ ] `claude-security-hooks/internal/hooks/validate.go` + `_test.go` — CON-validate-protocol + Pitfall 4 string/array content
- [ ] `claude-security-hooks/internal/hooks/inject.go` + `_test.go` — D-08 stub
- [ ] `claude-security-hooks/internal/hooks/dispatch.go` + `_test.go` — per-agent dispatch with inline typed decode (no interface{} parameter; preserves D-02)
- [ ] `claude-security-hooks/internal/hooks/codefences.go` + `_test.go` — code-fence stripping helper
- [ ] `claude-security-hooks/internal/invariants/boundary_test.go` — enforces PATTERNS.md S4 (no internal/hooks import from invariants)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Compiled binary actually runs inside Claude Code hook lifecycle | REQ-hooks-H2 (end-to-end) | True E2E requires Claude Code to dispatch the hook; out of scope for Phase 2 (Phase 5 owns smoke run) | Phase 5 will run `examples/sample-vulnerable-service/` end-to-end. Phase 2 unit-level integration tests + the H3/H4/H7 piped-printf smokes are sufficient for the phase exit criterion. |
| `inject-context` produces useful injected context | D-08 (deferred) | Per D-08 inject-context is a no-op stub in Phase 2 — useful injection is Phase 3 or Phase 5 concern | Verify only that the subcommand accepts SubagentStart JSON and exits 0 silently. |
| Workspace-root resolution under non-WSL platforms | D-11 / RQ-8 | Tests run on dev machine (WSL2); platform-specific CWD behavior on macOS/native Linux requires Claude Code installation there | Document fallback chain (`$CLAUDE_PROJECT_DIR` → event.cwd → marker scan) and rely on the runtime guard emitting `workspace_root_not_found` if assumption fails. |

---

## Validation Sign-Off

- [x] All 58 IDs have an assigned Test Type + Min Cases + Plan column ✓
- [ ] Per-task automated verify command exists for every plan task (every Plan 02-* PLAN.md task has at least one `<automated>` verify block — Plan 02-08 uses MULTIPLE separate blocks to cover H1, static linkage, full test, H3 smoke, H4/H7 smoke independently — verified during planner write)
- [x] Sampling continuity: no 3 consecutive tasks without automated verify ✓ (every task has automated_verify per plan structure)
- [x] Wave 0 covers all MISSING references ✓ (all fsutil/registry/meta-test scaffolding listed above, owned by Plan 02-01; h6_coverage_test.go expectedIDs map is SEALED at Wave 0)
- [x] No watch-mode flags ✓ (`go test` is single-shot)
- [x] Feedback latency < 5s ✓ (full suite estimated ≤ 5s)
- [x] `nyquist_compliant: true` set in frontmatter ✓
- [x] No compound shell commands in any plan `<automated>` block (per CLAUDE.md) ✓
- [x] `h6_coverage_test.go` SEALED in Wave 0 with full pre-populated expectedIDs map; Wave 1 plans (02-02..02-07) contribute test FUNCTIONS only — eliminates the parallel-write race ✓

**Approval:** Plan column populated; ready for execution.
