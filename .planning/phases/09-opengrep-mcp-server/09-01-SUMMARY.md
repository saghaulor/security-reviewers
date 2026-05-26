---
phase: "09"
plan: "01"
subsystem: makefile-and-orchestration
tags: [tdd, makefile, orchestration, red-phase]
dependency_graph:
  requires: []
  provides:
    - verify-opengrep-mcp target (Makefile)
    - Step 3 Bootstrap removed from security-review.md
  affects:
    - .claude/commands/security-review.md
    - Makefile
tech_stack:
  added: []
  patterns:
    - TDD RED-phase gate: verify target fails before build target is implemented
    - .PHONY target declaration pattern (matching claude-security-hooks/Makefile style)
key_files:
  created:
    - Makefile
  modified:
    - .claude/commands/security-review.md
decisions:
  - Top-level Makefile uses .PHONY for all targets per claude-security-hooks/Makefile pattern
  - verify-opengrep-mcp checks binary exists, is executable, and is statically linked (three sequential @-prefixed recipe lines, no && chaining)
  - build-opengrep-mcp is a stub in Plan 01; full implementation deferred to Plan 02
  - Step 3 Docker bootstrap removed cleanly (no replacement); opengrep-mcp lifecycle managed by Claude Code via .mcp.json stdio
  - All Step cross-references inside security-review.md updated: Step 5 read-cartographer, Step 7 synthesis wait, Key constraints parallel tracers — all updated from old Step 7 to new Step 6
metrics:
  duration: "~10 minutes"
  completed_at: "2026-05-26"
  tasks_completed: 2
  files_changed: 2
requirements_met:
  - REQ-mcp-O1
  - REQ-mcp-O2
  - REQ-mcp-O3
---

# Phase 09 Plan 01: Makefile TDD RED Phase + security-review.md Cleanup Summary

**One-liner:** Top-level Makefile verify-opengrep-mcp TDD RED gate and Docker bootstrap step removal from orchestration command.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create top-level Makefile with TDD verify target (RED phase) | 8700257 | Makefile |
| 2 | Remove Step 3 Docker bootstrap from security-review.md and renumber steps | e44460c | .claude/commands/security-review.md |

## What Was Built

### Task 1: Top-level Makefile (RED phase)

Created `/home/saghaulor/code/security_reviewer/Makefile` with:

- `OPENGREP_BIN := .claude/hooks/bin/opengrep-mcp` variable for consistent path reference
- `verify-opengrep-mcp` target: three sequential checks (binary exists, binary is executable, binary is statically linked), each with an explicit FAIL message on failure
- `build-opengrep-mcp` stub target: prints not-implemented message, exits 0; will be replaced in Plan 02
- `help` target as the default (first target) listing available targets
- All targets declared `.PHONY`

RED state confirmed: `make verify-opengrep-mcp` exits non-zero with "FAIL: .claude/hooks/bin/opengrep-mcp does not exist. Run: make build-opengrep-mcp" before any build runs.

### Task 2: security-review.md Step Removal and Renumbering

Removed the entire Step 3 (Bootstrap opengrep-mcp) section from `.claude/commands/security-review.md`:
- Deleted `## Step 3: Bootstrap opengrep-mcp` heading and all content through to the next heading
- Content removed: GET health-check, docker stop/rm/run commands, 30-second polling loop, Docker-unavailability error block

Renumbered Steps 4-9 to Steps 3-8 sequentially. Updated all internal cross-references:
- Step 5 (Read cartographer): "for use in Step 7" → "for use in Step 6"
- Step 7 (Synthesis): "from Step 7" → "from Step 6"
- Key constraints: "tracer Tasks in Step 7" → "tracer Tasks in Step 6"

Final step sequence: 0, 1, 2, 3 (pre-pass), 4 (cartographer), 5 (read cartographer), 6 (tracers), 7 (synthesis), 8 (report).

## Verification Results

1. `make verify-opengrep-mcp` exits non-zero — RED state confirmed (binary not yet built)
2. No "Bootstrap" references in security-review.md
3. Steps 0-8 sequential with no gaps (9 `## Step` headings)
4. No `localhost:8000` references in security-review.md

## Deviations from Plan

None — plan executed exactly as written.

The plan noted the cross-reference in the Key constraints section about "Step 7: Run tracers (parallel)" becoming "Step 6". This was correctly identified and applied. Additionally, two more inline cross-references within the body of the step content (Read cartographer "for use in Step 7" and Synthesis "from Step 7") were also updated for consistency — these were necessary for correctness and treated as part of the renumbering task.

## Known Stubs

- `build-opengrep-mcp` in `Makefile` is an intentional stub per Plan 01 design. It will be implemented in Plan 02 with the actual `$(MAKE) -C /home/saghaulor/code/opengrep-mcp build` delegation and binary copy logic.

## Threat Flags

None. This plan only adds/edits a Makefile and a command markdown file. No new network endpoints, auth paths, file access patterns, or schema changes introduced.

## Self-Check: PASSED

- Makefile exists at `/home/saghaulor/code/security_reviewer/Makefile` — FOUND
- .claude/commands/security-review.md modified — FOUND
- Commit 8700257 exists — FOUND
- Commit e44460c exists — FOUND
