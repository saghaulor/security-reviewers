---
phase: "09"
plan: "02"
subsystem: makefile-and-mcp-config
tags: [tdd, makefile, green-phase, mcp, stdio, opengrep]
dependency_graph:
  requires:
    - 09-01 (verify-opengrep-mcp RED gate, stub build target)
  provides:
    - build-opengrep-mcp target (real implementation)
    - .claude/hooks/bin/opengrep-mcp (static binary installed)
    - .mcp.json opengrep stdio entry
  affects:
    - Makefile
    - .mcp.json
    - .claude/hooks/bin/opengrep-mcp
tech_stack:
  added: []
  patterns:
    - TDD GREEN phase: verify target passes after build target implemented
    - $(MAKE) -C delegation to standalone repo build system
    - stdio MCP transport for Claude Code subprocess spawning
key_files:
  created:
    - .claude/hooks/bin/opengrep-mcp
  modified:
    - Makefile
    - .mcp.json
decisions:
  - build-opengrep-mcp delegates to $(MAKE) -C /home/saghaulor/code/opengrep-mcp build then copies binary
  - verify-opengrep-mcp static-link check uses readelf -d / ldd fallback (file cmd not available in WSL)
  - SAST_ENGINE=opengrep set in .mcp.json env block; SEMGREP_APP_TOKEN excluded (user-supplied secret)
  - Binary path in .mcp.json is absolute: /home/saghaulor/code/security_reviewer/.claude/hooks/bin/opengrep-mcp
metrics:
  duration: "~8 minutes"
  completed_at: "2026-05-26"
  tasks_completed: 2
  files_changed: 2
requirements_met:
  - REQ-mcp-O1
  - REQ-mcp-O2
  - REQ-mcp-O3
  - REQ-mcp-O4
  - REQ-mcp-O5
  - REQ-mcp-O6
  - REQ-mcp-O7
  - REQ-mcp-O8
---

# Phase 09 Plan 02: build-opengrep-mcp GREEN Phase + .mcp.json stdio wiring Summary

**One-liner:** TDD GREEN phase — build-opengrep-mcp target delegates to standalone repo, copies static binary, and .mcp.json opengrep entry switches from SSE to stdio subprocess.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Implement build-opengrep-mcp Makefile target (GREEN phase) | ade319e | Makefile, .claude/hooks/bin/opengrep-mcp |
| 2 | Switch .mcp.json opengrep entry from SSE to stdio | 249c118 | .mcp.json |

## What Was Built

### Task 1: build-opengrep-mcp Makefile target (GREEN phase)

Replaced stub `build-opengrep-mcp` target in `/home/saghaulor/code/security_reviewer/Makefile` with real implementation:

- `@mkdir -p .claude/hooks/bin` — idempotent destination directory creation
- `$(MAKE) -C /home/saghaulor/code/opengrep-mcp build` — delegates to standalone repo's Makefile `build` target (CGO_ENABLED=0 static binary)
- `@cp /home/saghaulor/code/opengrep-mcp/bin/opengrep-mcp .claude/hooks/bin/opengrep-mcp` — installs binary to project-local path

GREEN state confirmed: `make verify-opengrep-mcp` exits 0 with "OK: opengrep-mcp binary verified".

Also fixed a bug in the `verify-opengrep-mcp` target (written in Plan 01): the `file` command is not installed in this WSL environment. Replaced `file $(OPENGREP_BIN) | grep -q "statically linked"` with `readelf -d $(OPENGREP_BIN) 2>&1 | grep -q "no dynamic section" || ldd $(OPENGREP_BIN) 2>&1 | grep -q "not a dynamic executable"` which correctly detects static binaries using tools available in this environment.

### Task 2: .mcp.json opengrep entry switch from SSE to stdio

Replaced the `opengrep` entry in `/home/saghaulor/code/security_reviewer/.mcp.json`:

**Before:**
```json
"opengrep": {
  "type": "sse",
  "url": "http://localhost:8000/sse"
}
```

**After:**
```json
"opengrep": {
  "type": "stdio",
  "command": "/home/saghaulor/code/security_reviewer/.claude/hooks/bin/opengrep-mcp",
  "env": {
    "SAST_ENGINE": "opengrep"
  }
}
```

- `graphify` entry unchanged (still `type: stdio` pointing at `scripts/graphify-mcp.sh`)
- `SEMGREP_APP_TOKEN` excluded per T-09-04 (user-supplied secret; binary reads from shell env and warns if missing for pro backend)
- Valid JSON confirmed via `python3 -m json.tool`

## Verification Results

1. `make build-opengrep-mcp` exits 0 — binary built and copied
2. `make verify-opengrep-mcp` exits 0 — "OK: opengrep-mcp binary verified"
3. Binary is statically linked (readelf -d shows no dynamic section; ldd shows "not a dynamic executable")
4. `python3 -m json.tool .mcp.json` exits 0 — valid JSON
5. `.mcp.json` opengrep `type` is `stdio`
6. No `localhost:8000` references in `.mcp.json`
7. No `SEMGREP_APP_TOKEN` in `.mcp.json`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] verify-opengrep-mcp used `file` command not available in WSL**

- **Found during:** Task 1 acceptance criteria run
- **Issue:** The `verify-opengrep-mcp` target (written in Plan 01) used `file $(OPENGREP_BIN) | grep -q "statically linked"` to check that the binary is statically linked. The `file` command is not installed in this WSL2 environment.
- **Fix:** Replaced the static-link check with `readelf -d $(OPENGREP_BIN) 2>&1 | grep -q "no dynamic section" || ldd $(OPENGREP_BIN) 2>&1 | grep -q "not a dynamic executable"`. Both `readelf` and `ldd` are available (`/usr/bin/readelf`, `/usr/bin/ldd`). A static binary has no dynamic section (readelf) and ldd reports "not a dynamic executable".
- **Files modified:** `Makefile` (line 22)
- **Commit:** ade319e

## Known Stubs

None — all stubs from Plan 01 resolved. `build-opengrep-mcp` is now fully implemented.

## Threat Flags

None. This plan installs a local static binary (built from developer-controlled source) and updates an MCP config file. No new network endpoints, auth paths, or schema changes.

## Self-Check: PASSED

- Makefile at `/home/saghaulor/code/security_reviewer/Makefile` — FOUND (modified)
- `.mcp.json` at `/home/saghaulor/code/security_reviewer/.mcp.json` — FOUND (modified)
- `.claude/hooks/bin/opengrep-mcp` binary — FOUND
- Commit ade319e exists — FOUND
- Commit 249c118 exists — FOUND
