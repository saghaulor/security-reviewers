---
project_name: security-reviewer
milestone: Phase 10 — Hook compatibility + code-ref schema
milestone_version: 10
completed_phases:
  - phase: 1
    name: Repo scaffold
    completed_at: 2026-05-19
  - phase: 2
    name: claude-security-hooks binary
    completed_at: 2026-05-19
  - phase: 3
    name: Agent definitions + hook registration
    completed_at: 2026-05-23
  - phase: 4
    name: opengrep-mcp server + OpenGrep container
    completed_at: 2026-05-24
  - phase: 5
    name: End-to-end smoke test
    completed_at: 2026-05-24
  - phase: 6
    name: Documentation
    completed_at: 2026-05-24
  - phase: 7
    name: Extended test cases
    completed_at: 2026-05-24
  - phase: 8
    name: Automated E2E Testing
    completed_at: 2026-05-25
  - phase: 9
    name: opengrep-mcp server + infrastructure hardening
    completed_at: 2026-05-26
current_phase: 10
next_phase: null
total_phases: 10
paused_at: null
status: ready_to_execute
---

# Project State: security-reviewer

**Phase 10 Wave 1 (Plan 10-02) complete — All GREEN implementations done.** All Wave 0 tests (20 total: B1 x4, B2 x1, E1 x15) now pass. TaskToolInput, T6 amendment, 4 code-ref joint invariants, and 11 schema fields all implemented. Binary rebuilt and functional.

## Current Phase: 10 — Hook compatibility + code-ref schema

**Phase status:** Wave 0 (TDD RED) complete; Wave 1 Plan 02 (GREEN) complete; Plan 03 (config) pending

**Plan 01 completed (2026-05-26):** TDD RED phase — 20 failing tests written
- B1 tests (4): Agent tool extra fields rejection tests (events_test.go)
- B2 tests (1): T6 ambiguous verdict blocking test (taint_tracer_test.go)
- E1 tests (15): code_ref schema mismatch tests (5 joint invariant + 10 round-trip)
  - Joint invariants: T-CodeRef, AZ-CodeRef, OA-CodeRef, IC-CodeRef
  - Round-trip tests: all 10 structs that gain CodeRef/CodeRefDirty fields
- Commits: 236e1d9, 73d2d53, 409e243, 0eab613, 4f2ffdd, 347404b, 9b26c3e (SUMMARY)

**Plan 02 completed (2026-05-26):** GREEN phase — all B1, B2, E1 fixes implemented
- TaskToolInput: +RunInBackground, +Model, +Isolation (all omitempty)
- T6 amendment: now exempts verdict="ambiguous"
- Schema fields: CodeRef + CodeRefDirty added to 10 structs + SynthesisReport
- Joint invariants: T-CodeRef, AZ-CodeRef, OA-CodeRef, IC-CodeRef implemented
- All Wave 0 tests now GREEN: `go test ./...` 100% pass
- Binary rebuilt and verified functional
- Commits: 1e17858, 781ac46, e8bee3b, c5cb5a3, 367cb36, 8b5356f, cadb67b (SUMMARY)

**Plan 03 pending (Wave 1):** Configuration changes — B3, B4, additional text/config updates
- B3: T9 absolute path validation
- B4: Binary freshness validation
- Additional configuration and documentation updates

## Decisions

- D-04 (stdio transport): MCP transport is stdio — Claude Code spawns opengrep-mcp as subprocess
- D-06 (Step 3 removal): opengrep-mcp lifecycle managed by Claude Code runtime, not orchestration command
- D-07 (binary path): .claude/hooks/bin/opengrep-mcp is the install target
- D-08 (token security): SEMGREP_APP_TOKEN excluded from .mcp.json; user must supply via shell env
- D-09 (static-link check): readelf/ldd used instead of file cmd (not available in WSL2)

## Roadmap Evolution

- Phase 9 added: opengrep-mcp SSE server + infrastructure hardening (2026-05-26)
- Phase 9 Plan 01 complete: Makefile TDD RED gate + security-review.md cleanup (2026-05-26)
- Phase 9 Plan 02 complete: build-opengrep-mcp GREEN phase + .mcp.json stdio wiring (2026-05-26)
- Phase 10 Plan 01 complete: 20 RED tests for B1/B2/E1 blockers (2026-05-26)
  - 4 B1 tests: Agent tool extra fields rejection
  - 1 B2 test: T6 ambiguous verdict blocking
  - 5 E1 joint invariant tests: code_ref mismatch detection
  - 10 E1 schema round-trip tests: compile-error RED placeholders
- Phase 10 Plan 02 complete: GREEN implementation — all Wave 0 tests passing (2026-05-26)
  - TaskToolInput expanded with 3 Agent tool fields
  - T6 amended to exempt ambiguous verdicts
  - 4 code-ref joint invariants implemented
  - 11 schema structs expanded with CodeRef + CodeRefDirty fields
  - Binary rebuilt and functional
  - 7 commits: schema, invariants, tests, binary
