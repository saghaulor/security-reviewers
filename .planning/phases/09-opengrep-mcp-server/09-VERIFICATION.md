---
phase: 09-opengrep-mcp-server
verified: 2026-05-26T15:40:00Z
status: human_needed
score: 4/5 must-haves verified
overrides_applied: 0
human_verification:
  - test: "Verify Claude Code can discover mcp__opengrep__scan_with_rule, mcp__opengrep__scan_directory, and mcp__opengrep__get_ast tools after restarting Claude Code with the current .mcp.json"
    expected: "All three tools appear in Claude Code's tool list under the mcp__opengrep__ prefix when opengrep-mcp is spawned as a stdio subprocess"
    why_human: "Tool discovery requires a live Claude Code runtime with working stdio subprocess spawning. The binary starts correctly and registers three tools in RegisterTools(), but the MCP protocol handshake across the stdio pipe cannot be verified by grep or a static code check alone."
---

# Phase 9: opengrep-mcp server + infrastructure hardening — Verification Report

**Phase Goal:** Integrate the standalone opengrep-mcp binary into the security-review pipeline. The standalone binary already exists at /home/saghaulor/code/opengrep-mcp (Phase 5 complete, tested). Work in this repo is integration-only: top-level Makefile with build/verify targets, .mcp.json switched from SSE to stdio, and orchestration command Step 3 (Docker bootstrap) removed.
**Verified:** 2026-05-26T15:40:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths (Phase 9 ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `make build-opengrep-mcp` produces a static binary at `.claude/hooks/bin/opengrep-mcp` by delegating to the standalone repo's build system | VERIFIED | Makefile line 28-29: `$(MAKE) -C /home/saghaulor/code/opengrep-mcp build` then `@cp .../opengrep-mcp .claude/hooks/bin/opengrep-mcp`. Binary confirmed present at `.claude/hooks/bin/opengrep-mcp`. |
| 2 | `make verify-opengrep-mcp` passes: binary exists, is executable, and is statically linked | VERIFIED | `make verify-opengrep-mcp` exits 0, prints "OK: opengrep-mcp binary verified". `readelf -d` shows "There is no dynamic section in this file." `ldd` reports "not a dynamic executable". |
| 3 | `.mcp.json` opengrep entry uses `type: "stdio"` pointing at `.claude/hooks/bin/opengrep-mcp` with `SAST_ENGINE=opengrep` env var; no SSE/localhost:8000 references remain | VERIFIED | `.mcp.json` confirmed: `"type": "stdio"`, `"command": "/home/saghaulor/code/security_reviewer/.claude/hooks/bin/opengrep-mcp"`, `"env": {"SAST_ENGINE": "opengrep"}`. Zero matches for "sse" and "localhost:8000" in .mcp.json. |
| 4 | `.claude/commands/security-review.md` Steps are numbered 0–8 with no gaps; no Docker health-check or container bootstrap logic remains | VERIFIED | 9 `## Step` headings in file (Steps 0–8). Zero matches for "Bootstrap", "localhost:8000", "Step 3: Bootstrap". The sole `docker run` remaining is the govulncheck pre-pass (Step 3, correct). Internal cross-references updated: "Step 6" used for parallel tracers in Key constraints. |
| 5 | Claude Code can discover and spawn opengrep-mcp as a stdio subprocess — the three MCP tools (scan_with_rule, scan_directory, get_ast) are accessible to agents via `mcp__opengrep__*` prefix | UNCERTAIN — needs human | Binary starts and waits for MCP stdio input (confirmed by process test). `server.RegisterTools()` registers all three tools in source code. However, live tool discovery through the Claude Code MCP client handshake cannot be verified programmatically without the Claude Code runtime. |

**Score:** 4/5 truths verified (1 requires human runtime verification)

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `Makefile` | Top-level build orchestration with `verify-opengrep-mcp` and `build-opengrep-mcp` targets | VERIFIED | 30-line Makefile exists. `.PHONY` declared for all three targets (`help`, `verify-opengrep-mcp`, `build-opengrep-mcp`). `OPENGREP_BIN := .claude/hooks/bin/opengrep-mcp` variable present. |
| `.claude/hooks/bin/opengrep-mcp` | Compiled static binary for MCP server | VERIFIED | Binary exists, executable bit set, no dynamic section (readelf), not a dynamic executable (ldd). Built from `/home/saghaulor/code/opengrep-mcp` standalone repo via `CGO_ENABLED=0 go build`. |
| `.mcp.json` | MCP server config with stdio transport for opengrep | VERIFIED | Valid JSON (python3 -m json.tool exits 0). opengrep entry: `type=stdio`, correct absolute path, `SAST_ENGINE=opengrep` env, no SEMGREP_APP_TOKEN. graphify entry unchanged. |
| `.claude/commands/security-review.md` | Orchestration command without Docker bootstrap step, steps 0–8 | VERIFIED | Step 3 (Bootstrap opengrep-mcp) fully removed. Steps renumbered 0–8 with no gaps. All cross-references updated (Step 7 → Step 6 for tracers). |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `Makefile` | `/home/saghaulor/code/opengrep-mcp/Makefile` | `$(MAKE) -C /home/saghaulor/code/opengrep-mcp build` | WIRED | Line 28 of Makefile confirmed. Pattern `MAKE.*-C.*opengrep-mcp` matches. |
| `Makefile` | `.claude/hooks/bin/opengrep-mcp` | `@cp /home/saghaulor/code/opengrep-mcp/bin/opengrep-mcp .claude/hooks/bin/opengrep-mcp` | WIRED | Line 29 of Makefile confirmed. Binary copy step present and correct. |
| `.mcp.json` | `.claude/hooks/bin/opengrep-mcp` | `"command"` field in stdio entry | WIRED | Absolute path `/home/saghaulor/code/security_reviewer/.claude/hooks/bin/opengrep-mcp` present in .mcp.json command field. |
| `verify-opengrep-mcp` Makefile target | static-link check | `readelf -d` / `ldd` fallback | WIRED | Lines 20–23 of Makefile: existence check, executable check, static-link check. All three checks sequential with `@` prefix. Note: `file` command (from PLAN) was correctly replaced with `readelf`/`ldd` (available in WSL2 environment). |

---

### Data-Flow Trace (Level 4)

Not applicable. Phase 9 artifacts are configuration and build orchestration files, not dynamic data-rendering components.

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| `make verify-opengrep-mcp` passes (binary exists, executable, static) | `make verify-opengrep-mcp` from repo root | Exit 0; "OK: opengrep-mcp binary verified" | PASS |
| Binary exists and is executable | `test -f .claude/hooks/bin/opengrep-mcp` / `test -x .claude/hooks/bin/opengrep-mcp` | Both exit 0 (silent success) | PASS |
| Binary is statically linked | `readelf -d .claude/hooks/bin/opengrep-mcp` | "There is no dynamic section in this file." | PASS |
| `.mcp.json` is valid JSON with stdio type | `python3 -m json.tool .mcp.json` | Exit 0, formatted output with `"type": "stdio"` | PASS |
| No SEMGREP_APP_TOKEN in .mcp.json | `grep -c SEMGREP_APP_TOKEN .mcp.json` | Exit 1, 0 matches | PASS |
| No SSE references in .mcp.json | `grep -c sse .mcp.json` | Exit 1, 0 matches | PASS |
| No localhost:8000 in security-review.md | `grep -c localhost:8000 security-review.md` | Exit 1, 0 matches | PASS |
| No Bootstrap heading in security-review.md | `grep -c Bootstrap security-review.md` | Exit 1, 0 matches | PASS |
| Steps 0–8 sequential in security-review.md | `grep "^## Step" security-review.md` | 9 headings: Steps 0, 1, 2, 3, 4, 5, 6, 7, 8 | PASS |
| Makefile delegates to standalone via $(MAKE) -C | `grep "MAKE.*-C.*opengrep-mcp" Makefile` | Line 28 matches | PASS |
| Binary starts as MCP stdio server | `SAST_ENGINE=opengrep timeout 3 .claude/hooks/bin/opengrep-mcp` | Process starts, blocks on stdin (expected stdio MCP behavior), exits cleanly on timeout | PASS |

---

### Probe Execution

No `probe-*.sh` files declared or found in `scripts/*/tests/`. Phase 9 verification gate is Makefile-based (`make verify-opengrep-mcp`), which was run directly as a behavioral spot-check above.

---

### Requirements Coverage

Phase 9 Plans 01 and 02 both declare REQ-mcp-O1 through REQ-mcp-O8. The REQUIREMENTS.md traceability table maps these requirements to Phase 4, not Phase 9. This is a documentation discrepancy, not a functional gap — Phase 9 is integration-only and the standalone binary from Phase 4 satisfies these requirements. The verifiable Phase 9 integration-layer properties are:

| Requirement | Evidence in Phase 9 scope | Status |
|-------------|--------------------------|--------|
| REQ-mcp-O1 (static binary) | Binary at `.claude/hooks/bin/opengrep-mcp` is statically linked (readelf/ldd confirmed) | SATISFIED by integration |
| REQ-mcp-O2 (three tools) | `server.RegisterTools()` in standalone registers `scan_with_rule`, `scan_directory`, `get_ast` | SATISFIED in standalone; accessible via MCP stdio |
| REQ-mcp-O3 (structured tier error) | `errTierText = "invalid_request: tier must be 'pro'"` in server.go; tier != "pro" returns IsError result | SATISFIED in standalone |
| REQ-mcp-O4 (container image caching) | Docker pull on first use — standalone behavior, not integration-layer concern | DEFERRED TO HUMAN: requires live Docker environment |
| REQ-mcp-O5 (read-only workspace mount) | runner.go line 104: `":ro"` flag in docker volume mount string | SATISFIED in standalone |
| REQ-mcp-O6 (timeout container kill) | runner.go: `exec.CommandContext(ctx, "docker", ...)` with `context.DeadlineExceeded` check | SATISFIED in standalone |
| REQ-mcp-O7 (normalized findings) | `ParseAndNormalize()` called on all scan output; returns `schema.ScanResult` | SATISFIED in standalone |
| REQ-mcp-O8 (SEMGREP_APP_TOKEN masking) | Token read at startup (main.go), passed via Docker env (not argv), never in .mcp.json, token_leak_test.go covers this | SATISFIED in standalone; token_leak test exists |

**Note on REQUIREMENTS.md traceability table:** REQ-mcp-O1..O8 are listed as Phase 4 requirements. Phase 9 Plans claiming these requirements is a docs-level redundancy but not a gap — Phase 9 integrates Phase 4's binary. No requirements are orphaned.

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None found | — | — | — | — |

No `TBD`, `FIXME`, `XXX`, `TODO`, or placeholder patterns found in any Phase 9 modified files (`Makefile`, `.mcp.json`, `.claude/commands/security-review.md`).

---

### Human Verification Required

#### 1. MCP Tool Discovery in Claude Code Runtime

**Test:** Restart Claude Code (or start a new session) with the current `.mcp.json` in place. Verify that the opengrep MCP server is listed in available tools.

**Expected:** Claude Code spawns `.claude/hooks/bin/opengrep-mcp` as a stdio subprocess and the tools `mcp__opengrep__scan_with_rule`, `mcp__opengrep__scan_directory`, and `mcp__opengrep__get_ast` appear in the tool list. No startup error for the opengrep MCP server in the Claude Code logs.

**Why human:** MCP tool discovery requires a live Claude Code runtime performing the actual stdio subprocess spawn and MCP protocol handshake (initialize → tools/list). This cannot be verified by reading source code or running CLI commands — it requires Claude Code to be running and configured with the .mcp.json. The binary is confirmed to start and block on stdin (correct behavior), and the source registers three tools, but the end-to-end handshake is only observable through the running Claude Code UI or MCP client.

---

### Gaps Summary

No blocking gaps identified. All four programmatically-verifiable ROADMAP Success Criteria (SCs 1–4) are satisfied with codebase evidence. SC5 (Claude Code tool discovery) requires runtime human verification — the code path is correct and complete but the behavior can only be confirmed by starting a Claude Code session.

**One minor note:** The `verify-opengrep-mcp` Makefile target was written in Plan 01 using `file $(OPENGREP_BIN) | grep -q "statically linked"`. Plan 02 correctly auto-fixed this to use `readelf -d` / `ldd` because the `file` command is not installed in WSL2. This was documented as a bug fix in the Plan 02 SUMMARY and the Makefile reflects the corrected implementation.

---

_Verified: 2026-05-26T15:40:00Z_
_Verifier: Claude (gsd-verifier)_
