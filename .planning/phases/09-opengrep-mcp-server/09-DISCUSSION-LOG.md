# Phase 9: opengrep-mcp server + infrastructure hardening - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-26
**Phase:** 9-opengrep-mcp-server
**Areas discussed:** Code source / architecture, Transport, Tier model, Binary reference, SAST_ENGINE default, Phase 9 scope confirmation

---

## Codebase Scout Finding

Before presenting gray areas, a full code walk of `/home/saghaulor/code/opengrep-mcp` revealed the standalone module is Phase 5 complete — not a stub. This changed the discussion from "how do we build it?" to "how do we integrate it?" All gray areas below stem from this discovery.

---

## Code Source / Architecture

| Option | Description | Selected |
|--------|-------------|----------|
| Port it in | Copy/port standalone into security_reviewer/opengrep-mcp/ stub | |
| Adopt standalone as external dep | Keep standalone as separate repo; security_reviewer references binary | ✓ |
| Build fresh in stub | Implement from scratch using standalone as reference | |

**User's choice:** Adopt standalone as external dep
**Notes:** The standalone at `/home/saghaulor/code/opengrep-mcp` stays authoritative. security_reviewer builds the binary from there and installs it locally.

---

## Transport

| Option | Description | Selected |
|--------|-------------|----------|
| stdio | Claude Code spawns binary as subprocess via `.mcp.json`. No server management, removes Step 3 from orchestration. | ✓ |
| SSE / HTTP | Keep Docker container + health-check loop. Add HTTP layer to standalone. | |

**User's choice:** stdio (recommended)
**Notes:** The standalone already uses `mcp.StdioTransport`. Aligns with how graphify is wired (also stdio). Simpler lifecycle — Claude Code manages subprocess, not the orchestration command.

---

## Tier Model

| Option | Description | Selected |
|--------|-------------|----------|
| Simplified env-var model | SAST_ENGINE selects backend at startup. tier="pro" is the only MCP tier. | ✓ |
| Three-tier dispatch | Accept pro/intrafile/ce per-call, route to different Docker images. | |

**User's choice:** Keep simplified env-var model (recommended)
**Notes:** Agents keep internal tier-based evidence weighting (pro=high confidence, intrafile=moderate) for confidence scoring, but all MCP calls to the server use `tier: "pro"`. The server's tier validation stays as-is.

---

## Binary Reference

| Option | Description | Selected |
|--------|-------------|----------|
| Absolute path to pre-built binary | `.mcp.json` points to standalone's bin/; requires pre-build step outside security_reviewer. | |
| Makefile build target | `make build-opengrep-mcp` builds from standalone and copies to `.claude/hooks/bin/`. | ✓ |
| Install via PATH | Assume binary on PATH via `make install`. | |

**User's choice:** Makefile build target (recommended)
**Notes:** Consistent with how claude-security-hooks is built. Target delegates to standalone's Makefile. Binary lands in `.claude/hooks/bin/opengrep-mcp`.

---

## SAST_ENGINE Default

| Option | Description | Selected |
|--------|-------------|----------|
| Set in .mcp.json env block | Explicit: `"env": {"SAST_ENGINE": "opengrep"}` | ✓ |
| Leave unset | Binary defaults to opengrep when unset. Less explicit. | |

**User's choice:** Set explicitly in .mcp.json env block
**Notes:** Explicit over implicit; makes the default backend visible to anyone reading the config.

---

## Phase 9 Scope Confirmation

| Item | Selected |
|------|----------|
| Makefile target: build-opengrep-mcp | ✓ |
| Update .mcp.json to stdio transport | ✓ |
| Remove Step 3 from security-review.md | ✓ |
| Patch standalone server: accept 'intrafile'/'ce' tiers | — (not in scope) |

**Notes:** Tier alias patch is not needed since all MCP calls will use `tier: "pro"`. The internal agent tier weighting is advisory, not a routing signal.

---

## Claude's Discretion

- Exact Makefile syntax for delegating to standalone's build system (e.g. `$(MAKE) -C` vs `cd && make`)
- Whether to add a `check-opengrep-mcp` target that verifies the binary exists and is executable
- Test strategy for Makefile target (shell-based test vs Go test binary existence check)

## Deferred Ideas

- Three-tier dispatch (pro/intrafile/ce per-call): Deferred — simplified env-var model is sufficient
- opengrep-intrafile local Docker image build: Deferred — not needed with env-var backend model
- Embedding opengrep-mcp source in security_reviewer monorepo: Deferred — user chose external dep pattern
- Standalone server tier aliases: Deferred — not needed since all calls use tier="pro"
