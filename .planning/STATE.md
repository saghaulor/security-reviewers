---
project_name: security-reviewer
milestone: Phase 11 — codegraph migration
milestone_version: 11
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
  - phase: 10
    name: Hook compatibility + code-ref schema
    completed_at: 2026-05-26
  - phase: 11
    name: codegraph migration
    completed_at: 2026-05-27
current_phase: 13
next_phase: null
total_phases: 13
paused_at: null
status: in_progress
---

# Project State: security-reviewer

**Phase 13 IN PROGRESS — Wave 0 + Wave 1 complete (4/6 plans done). Wave 2 (13-05) and Wave 3 (13-06) remain.**

Wave 0 (13-01 TDD RED): COMPLETE — 3 test files, 8 new failing tests for W4/W5/W7
Wave 1a (13-02 bootstrap): COMPLETE — bootstrap/pre-flight-checks.sh + make preflight target + .codegraph/ gitignore
Wave 1b (13-03 W2+W3): COMPLETE — JSON enforcement at Steps 4+6; Write added to 5 agents; A10-amended rules
Wave 1c (13-04 TDD GREEN): COMPLETE — validate.go S1 soft-fail, dispatch.go content preview, preflight.go schema doc, Handler fields
Remaining: Wave 2 (13-05 schema files) → Wave 3 (13-06 cartographer routes + first_param_read)

**Phase 11 COMPLETE — All 2 plans executed successfully.** Wave 1 (MCP transport layer): codegraph-mcp.sh created, graphify-mcp.sh deleted, .mcp.json updated, .gitignore updated. Wave 2 (agent/orchestration migration): go-cartographer.md fully migrated to mcp__codegraph__* tools (12 edits), security-review.md Step 3 updated to codegraph init/index. No mcp__graphify__* references remain in the pipeline. All tests pass.

## Current Phase: 11 — codegraph migration

**Phase status:** ALL PLANS COMPLETE — Wave 1 (transport layer), Wave 2 (agent migration) executed and verified

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

**Plan 03 completed (2026-05-26):** Configuration and documentation changes — B3, B4, E1 wiring
- B3: T9-PATH hard rule added to go-taint-tracer.md §6 (workspace-relative paths required)
- B4: Step 1.5 added to security-review.md (unconditional `make install` before agent dispatch)
- E1: code_ref computation + propagation wired through cartographer → tracers → synthesis
- Configuration: 5 agent definitions + 1 skill file updated with code_ref/code_ref_dirty specifications
- Commits: 7f4a7d4, 5756313, 3bd10eb, 4cc9417, 62af270, 5796725, ec31e87, c35bde8, 0080d26 (SUMMARY)

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
- Phase 10 Plan 03 complete: Configuration + documentation — B3/B4/E1 wiring (2026-05-26)
  - T9-PATH hard rule added to go-taint-tracer.md for workspace-relative path enforcement
  - Step 1.5 added to security-review.md for unconditional binary rebuild
  - code_ref computation in Step 2 + propagation through Steps 4/6 in security-review.md
  - code_ref echo instructions added to 4 tracer agent definitions
  - code_ref propagation step added to synthesis.md
  - 9 commits: 6 agent def updates + 1 skill update + SUMMARY
- **PHASE 10 COMPLETED:** All three waves executed; all blockers (B1–B4) unblocked; code_ref schema integrated end-to-end
- Phase 11 added: codegraph migration — replace graphify with codegraph as the graph/MCP layer (2026-05-27)
- Phase 11 planned: 2 plans across 2 waves — Wave 1 (script/MCP/gitignore), Wave 2 (agent/orchestration) (2026-05-27)
- Phase 11 Plan 01 complete: MCP transport layer (2026-05-27)
  - scripts/codegraph-mcp.sh created (stdio wrapper, reads .current-review, execs codegraph serve --mcp --path)
  - scripts/graphify-mcp.sh deleted (D-1 hard delete)
  - .mcp.json graphify key replaced with codegraph key (stdio, absolute path)
  - examples/sample-vulnerable-service/.gitignore gains .codegraph/ (D-7)
- Phase 11 Plan 02 complete: Agent/orchestration migration (2026-05-27)
  - go-cartographer.md: all 12 edits applied — mcp__graphify__* → mcp__codegraph__*, graphify-out refs removed
  - security-review.md: Step 2 updated, Step 3 codegraph init + index (2-step), govulncheck renumbered
  - grep -r "mcp__graphify__" .claude/ returns empty — complete removal
- **PHASE 11 COMPLETED:** Both waves executed; graphify fully replaced by codegraph end-to-end; all tests pass (2026-05-27)
- Phase 12 added: code review skill evaluation — evaluate two external code review skills for complementary coverage value (2026-05-27)
- Phase 13 added: pipeline reliability + bootstrap hardening — 7 failure modes from first scan run + bootstrap pre-flight gap addressed (2026-05-27)
