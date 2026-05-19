---
phase: 2
slug: claude-security-hooks-go-binary
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-19
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
| **Quick run command** | `go test ./internal/invariants -run Test<ID>_` (from `claude-security-hooks/`) — single assertion (<100 ms) |
| **Per-agent suite** | `go test ./internal/invariants -run TestT` (taint), `TestA` (cartographer), `TestAZ`, `TestOA`, `TestIC`, `TestS` |
| **Full suite command** | `go test ./...` (from `claude-security-hooks/`) |
| **Build smoke** | `make build && ldd bin/claude-security-hooks` ⇒ `"not a dynamic executable"` (H1) |
| **Estimated runtime** | ~5 seconds full suite; ~100 ms targeted |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/invariants -run Test<ID>_` for the assertion(s) the task implemented.
- **After every plan wave:** `go test ./...` from `claude-security-hooks/`.
- **Before `/gsd-verify-work`:** `make build` (H1 smoke via `ldd`) + `make test` (full suite) — both green.
- **Max feedback latency:** **5 seconds** (full suite).

---

## Per-Task Verification Map

Task IDs are pending until plans are written. Each row maps an assertion predicate to its test technique, minimum case count, and fixture requirement (verbatim from RESEARCH.md § Validation Architecture).

### Cartographer (A1–A11) — REQ-cartographer-A1..A11

| ID | Test Type | Min Cases (pos+neg) | Fixture | Plan | Status |
|----|-----------|---------------------|---------|------|--------|
| A1 | unit (table-driven) | 1+2 | inline JSON | TBD | ⬜ pending |
| A2 | unit | 1+1 | inline | TBD | ⬜ pending |
| A3 | unit | 1+4 | inline | TBD | ⬜ pending |
| A4 | unit | 1+1 | inline | TBD | ⬜ pending |
| A5 | unit | 1+1 | inline | TBD | ⬜ pending |
| A6 | unit + filesystem | 1+2 | `testdata/graphify-out/graph.json` | TBD | ⬜ pending |
| A7 | unit + filesystem | 1+1 | `testdata/workspace/` | TBD | ⬜ pending |
| A8 | unit + filesystem | 1+1 | `testdata/workspace/` (known line counts) | TBD | ⬜ pending |
| A9 | unit | 1+1 | inline | TBD | ⬜ pending |
| A10 | unit (no-op stub) | 1+1 | inline; **no-op pass per Phase 2 decision; TODO(phase-5)** | TBD | ⬜ pending |
| A11 | unit | 1+1 | inline | TBD | ⬜ pending |

### Taint Tracer (T1–T11) — REQ-taint-T1..T11

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| T1 | unit | 1+2 | inline | TBD | ⬜ pending |
| T2 | unit | 5+1 | inline | TBD | ⬜ pending |
| T3 | unit | 1+3 | inline | TBD | ⬜ pending |
| T4 | unit | 1+0 | inline | TBD | ⬜ pending |
| T5 | unit | 3+1 | inline | TBD | ⬜ pending |
| T6 | unit | 3+1 | inline | TBD | ⬜ pending |
| T7 | unit (joint input+output) | 1+1 | signature accepts verdict AND input record | TBD | ⬜ pending |
| T8 | unit (joint) | 2+1 | inline | TBD | ⬜ pending |
| T9 | unit + filesystem | 1+1 | `testdata/workspace/` | TBD | ⬜ pending |
| T10 | unit | 2+1 | inline | TBD | ⬜ pending |
| T11 | unit (no-op stub) | 1+1 | inline; **no-op pass per Phase 2 decision; TODO(phase-5)** | TBD | ⬜ pending |

### Authz Tracer (AZ1–AZ6) — REQ-authz-AZ1..AZ6

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| AZ1 | unit | 1+2 | inline | TBD | ⬜ pending |
| AZ2 | unit (joint) | 1+1 | predicate takes input+output | TBD | ⬜ pending |
| AZ3 | unit | 1+1 | inline | TBD | ⬜ pending |
| AZ4 | unit (joint) | 1+1 | inline | TBD | ⬜ pending |
| AZ5 | unit (joint) | 1+1 | inline | TBD | ⬜ pending |
| AZ6 | unit (no-op stub) | 1+1 | inline; **no-op pass per Phase 2 decision; TODO(phase-5)** | TBD | ⬜ pending |

### OAuth Auditor (OA1–OA7) — REQ-oauth-OA1..OA7

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| OA1 | unit | 1+2 | inline | TBD | ⬜ pending |
| OA2 | unit | 1+1 | inline; checker uses regex over `RFC|draft-` | TBD | ⬜ pending |
| OA3 | unit | 1+1 | inline | TBD | ⬜ pending |
| OA4 | unit (joint) | 1+1 | inline | TBD | ⬜ pending |
| OA5 | unit (joint) | 1+1 | inline | TBD | ⬜ pending |
| OA6 | unit (joint) | 1+1 | inline | TBD | ⬜ pending |
| OA7 | unit (joint) | 1+1 | inline | TBD | ⬜ pending |

### Invariant Checker (IC1–IC4) — REQ-invariant-IC1..IC4

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| IC1 | unit | 1+2 | inline | TBD | ⬜ pending |
| IC2 | unit (joint) | 1+2 | inline | TBD | ⬜ pending |
| IC3 | unit | 1+2 | inline | TBD | ⬜ pending |
| IC4 | unit (joint) | 1+1 | inline | TBD | ⬜ pending |

### Synthesis (S1–S6) — REQ-synthesis-S1..S6

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| S1 | filesystem | 1+2 | `testdata/synthesis_out/` (both json + md) | TBD | ⬜ pending |
| S2 | unit | 1+1 | inline JSON | TBD | ⬜ pending |
| S3 | unit | 1+1 | inline | TBD | ⬜ pending |
| S4 | unit | 1+1 | inline | TBD | ⬜ pending |
| S5 | unit | 1+1 | inline | TBD | ⬜ pending |
| S6 | unit | 1+1 | inline | TBD | ⬜ pending |

### Hook Build/Runtime Invariants (H1–H7) — REQ-hooks-H1..H7

| ID | Test Type | Min Cases | Fixture | Plan | Status |
|----|-----------|-----------|---------|------|--------|
| H1 | Makefile smoke + optional Go runtime test | 1 | `make build` exits 0; `ldd` shows "not a dynamic executable" | TBD | ⬜ pending |
| H2 | integration unit | 3 | inline JSON: PreToolUse, PostToolUse, SubagentStart fixtures | TBD | ⬜ pending |
| H3 | integration unit | 1 | inline JSON: `subagent_type="general-purpose"` | TBD | ⬜ pending |
| H4 | integration unit | 1 | inline; any failing invariant ⇒ single block JSON, exit 0 | TBD | ⬜ pending |
| H5 | meta-test (`h5_deps_test.go`) | 1 | runs `go list -deps -test ./...` via `os/exec` | TBD | ⬜ pending |
| H6 | meta-test (`h6_coverage_test.go`) | 1 | parses `_test.go` via `go/parser`; asserts every registry ID has Test func | TBD | ⬜ pending |
| H7 | integration unit | 4 | inline: invalid JSON, valid-JSON-wrong-type, empty, multi-segment array | TBD | ⬜ pending |

*Status legend: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

**Coverage check:** 58 total IDs (51 assertion predicates + 7 H invariants). Every Phase 2 REQ-ID is represented above.

---

## Wave 0 Requirements

Per RESEARCH.md § "Wave 0 Gaps" (lines 904-923) — the following infrastructure does NOT exist in the Phase 1 scaffold and MUST be created before any Wave 1 implementation work:

- [ ] `claude-security-hooks/internal/invariants/registry.go` — shared `Severity` + `Violation` types (D-01).
- [ ] `claude-security-hooks/internal/invariants/fsutil.go` — `FileExists`, `LineCount` helpers (used by A6/A7/A8/T9/AZ4/OA4/IC3/S1).
- [ ] `claude-security-hooks/internal/invariants/fsutil_test.go` — fsutil edge cases (missing file, zero-byte file, symlink rejection).
- [ ] `claude-security-hooks/internal/invariants/h5_deps_test.go` — H5 meta-test scaffold.
- [ ] `claude-security-hooks/internal/invariants/h6_coverage_test.go` — H6 meta-test scaffold; expected-IDs map grows as each Wave 1 plan lands.
- [ ] `claude-security-hooks/internal/invariants/testdata/workspace/` — fixture tree with `.go` files of known line counts.
- [ ] `claude-security-hooks/internal/invariants/testdata/graphify-out/graph.json` — minimal valid graph for A6 tests.
- [ ] `claude-security-hooks/internal/invariants/testdata/synthesis_out/` — fixture dir for S1 (both `review-report.json` and `review-report.md`).
- [ ] `claude-security-hooks/internal/hooks/events.go` + `events_test.go` — PreToolUseEvent, PostToolUseEvent, SubagentStartEvent, TaskToolInput, ToolResponse structs honoring **live-docs deltas** (`agent_type` not `agent_name`; tolerate `cwd`, `permission_mode`, `effort`, `tool_use_id`).
- [ ] `claude-security-hooks/internal/hooks/decision.go` + `_test.go` — D-03 block-reason formatter.
- [ ] `claude-security-hooks/internal/hooks/agents.go` + `agents_test.go` — `SecurityAgentSet` const + drift test (D-07).
- [ ] `claude-security-hooks/internal/hooks/doc.go` — D-14 exit-code policy package comment.
- [ ] `claude-security-hooks/Makefile` — `build`, `test`, `install`, `clean`, `lint` targets (D-13).
- [ ] `go-cmp` declared as test dep (creates `go.sum`).
- [ ] Verify `claude-security-hooks/.gitignore` excludes `bin/`.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Compiled binary actually runs inside Claude Code hook lifecycle | REQ-hooks-H2 (end-to-end) | True E2E requires Claude Code to dispatch the hook; out of scope for Phase 2 (Phase 5 owns smoke run) | Phase 5 will run `examples/sample-vulnerable-service/` end-to-end. Phase 2 unit-level integration tests are sufficient for the phase exit criterion. |
| `inject-context` produces useful injected context | D-08 (deferred) | Per D-08 inject-context is a no-op stub in Phase 2 — useful injection is Phase 3 or Phase 5 concern | Verify only that the subcommand accepts SubagentStart JSON and exits 0 silently. |
| Workspace-root resolution under non-WSL platforms | D-11 / RQ-8 | Tests run on dev machine (WSL2); platform-specific CWD behavior on macOS/native Linux requires Claude Code installation there | Document fallback chain (`$CLAUDE_PROJECT_DIR` → event.cwd → marker scan) and rely on the runtime guard emitting `workspace_root_not_found` if assumption fails. |

---

## Validation Sign-Off

- [ ] All 58 IDs have an assigned Test Type + Min Cases (above) ✓ (drafted)
- [ ] Per-task automated verify command exists for every plan task (populated after planner writes PLAN.md files)
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify (checked post-planning)
- [ ] Wave 0 covers all MISSING references ✓ (all fsutil/registry/meta-test scaffolding listed above)
- [ ] No watch-mode flags ✓ (`go test` is single-shot)
- [ ] Feedback latency < 5s ✓ (full suite estimated ≤ 5s)
- [ ] `nyquist_compliant: true` set in frontmatter (after planner sign-off)

**Approval:** pending (awaiting planner to populate Plan column and assign Task IDs)
