---
project_name: security-reviewer
milestone: Phase 9 — opengrep-mcp SSE server + infrastructure hardening
milestone_version: 9
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
current_phase: 9
next_phase: null
total_phases: 9
paused_at: null
status: in_progress
---

# Project State: security-reviewer

Phase 9 in progress. Phase 8 (Automated E2E Testing) completed 2026-05-25.

## Current Phase: 9 — opengrep-mcp SSE server + infrastructure hardening

**Started:** 2026-05-26

**Infrastructure work already committed (done during/after Phase 8):**
- Session ID threading across all tracers + orchestration command
- `claude-security-hooks uuid` subcommand (stdlib crypto/rand, replaces `uuidgen`)
- H5 predicate fix (two-invocation `go list` approach correctly separating test vs non-test deps)
- Orchestration command hardening: stale file deletion before agents run, explicit Task JSON, mandatory SESSION_ID threading to all four tracers
- govulncheck via Docker (no local install required; non-zero exit = vulns found, not error)
- `.mcp.json` + `scripts/graphify-mcp.sh`: graphify MCP via stdio wrapper, opengrep MCP via SSE at localhost:8000

**Plan 01 completed (2026-05-26):**
- Top-level Makefile created with `verify-opengrep-mcp` TDD RED gate and `build-opengrep-mcp` stub
- `security-review.md` Step 3 Docker bootstrap removed; steps renumbered 0-8 sequentially
- Commits: 8700257 (Makefile RED gate), e44460c (security-review.md cleanup)

**Still to implement (the main Phase 9 deliverable):**
- Plan 02: `build-opengrep-mcp` Makefile target (delegate to standalone repo, copy binary to .claude/hooks/bin)
- Plan 03: `.mcp.json` switch from SSE to stdio transport

## Decisions

- D-04 (stdio transport): MCP transport is stdio — Claude Code spawns opengrep-mcp as subprocess
- D-06 (Step 3 removal): opengrep-mcp lifecycle managed by Claude Code runtime, not orchestration command
- D-07 (binary path): .claude/hooks/bin/opengrep-mcp is the install target

## Roadmap Evolution

- Phase 9 added: opengrep-mcp SSE server + infrastructure hardening (2026-05-26)
- Phase 9 Plan 01 complete: Makefile TDD RED gate + security-review.md cleanup (2026-05-26)
