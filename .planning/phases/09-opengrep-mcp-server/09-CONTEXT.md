# Phase 9: opengrep-mcp server + infrastructure hardening - Context

**Gathered:** 2026-05-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 9 integrates the `opengrep-mcp` binary into the security-review pipeline as the SAST backend. The binary already exists as a fully-implemented standalone module at `/home/saghaulor/code/opengrep-mcp` (Phase 5 complete, tested). The work in **this** repo (security_reviewer) is **integration only** — no new Go binary code is written here.

**In scope (security_reviewer):**
- Makefile target that builds `opengrep-mcp` binary from the standalone repo and installs it to a local path
- Update `.mcp.json`: switch `opengrep` entry from SSE to stdio transport, pointing at the installed binary
- Update `security-review.md`: remove Step 3 (Docker health-check + container bootstrap for opengrep-mcp)
- TDD for the Makefile target (verify binary exists and is executable after `make build-opengrep-mcp`)

**Out of scope for Phase 9 (security_reviewer):**
- Writing Go source code for opengrep-mcp (it's in the standalone repo)
- Adding HTTP/SSE transport to the server
- Three-tier container dispatch (pro/intrafile/ce per-call)
- Building the opengrep-intrafile Docker image
- Stubbing the `security_reviewer/opengrep-mcp/` module — that stub stays as-is (go.mod only)

</domain>

<decisions>
## Implementation Decisions

### Architecture: Standalone as External Dep

- **D-01:** The `opengrep-mcp` binary lives in a **separate standalone repo** at `/home/saghaulor/code/opengrep-mcp`. security_reviewer does NOT duplicate or fork the source code.
- **D-02:** Integration is binary-level: security_reviewer's Makefile builds the standalone and copies the binary locally. The source of truth for opengrep-mcp code is the standalone repo.
- **D-03:** The `security_reviewer/opengrep-mcp/` stub (go.mod only) is intentionally left as-is — it's a placeholder that records the module relationship without containing implementation.

### Transport: stdio

- **D-04:** MCP transport is **stdio** — Claude Code spawns `opengrep-mcp` as a subprocess via `.mcp.json` `type: "stdio"`. No HTTP server, no port 8000, no SSE.
- **D-05:** The standalone already uses `mcp.StdioTransport` from `github.com/modelcontextprotocol/go-sdk`. No transport changes needed in the standalone.
- **D-06:** `security-review.md` Step 3 (Docker health-check + container start for opengrep-mcp) is **removed entirely**. The MCP server lifecycle is managed by Claude Code, not the orchestration command.

### Binary Reference: Makefile Build Target

- **D-07:** Add `build-opengrep-mcp` target to security_reviewer's `Makefile`. The target runs `go build` in the standalone repo and copies the binary to `.claude/hooks/bin/opengrep-mcp`.
  ```makefile
  build-opengrep-mcp:
      CGO_ENABLED=0 go build -o .claude/hooks/bin/opengrep-mcp ./cmd/opengrep-mcp
  ```
  The `go build` is run with `$(MAKE) -C /home/saghaulor/code/opengrep-mcp build` or equivalent, referencing the standalone's Makefile.
- **D-08:** `.mcp.json` `opengrep` entry points to `.claude/hooks/bin/opengrep-mcp` (relative to project root):
  ```json
  "opengrep": {
    "type": "stdio",
    "command": "/home/saghaulor/code/security_reviewer/.claude/hooks/bin/opengrep-mcp",
    "env": {
      "SAST_ENGINE": "opengrep"
    }
  }
  ```

### Backend / Tier Model: Simplified env-var

- **D-09:** Backend is selected via **`SAST_ENGINE` env var** at subprocess startup. Default is `opengrep` (open-source, no token needed). Set explicitly in `.mcp.json` env block.
- **D-10:** MCP tier parameter is **always `"pro"`** from callers. The standalone server accepts only `"pro"`. No server-side tier aliases needed.
- **D-11:** Agent internal evidence weighting (pro=high confidence, intrafile=moderate) is **preserved as-is** in agent definitions — this logic uses the `semgrep_tier` field from the orchestration input for confidence interpretation, not for routing. The actual `mcp__opengrep__scan_with_rule` calls always pass `tier: "pro"`.
- **D-12:** `SAST_ENGINE=opengrep` is the default; reviewers with a Semgrep Pro token can override via environment before invoking Claude Code.

### Standalone Repo State

- **D-13:** The standalone at `/home/saghaulor/code/opengrep-mcp` is Phase 5 complete:
  - Three tools: `scan_with_rule`, `scan_directory`, `get_ast` — all registered
  - Two backends: `semgrep-pro` (returntocorp/semgrep:1.55.0), `opengrep` (opengrep/opengrep:1.20.0)
  - Docker CLI only (no SDK), read-only `/src` mounts, timeout enforcement, token masking via `SEMGREP_APP_TOKEN` env var (ps-safe: passed as `-e SEMGREP_APP_TOKEN` without value in argv)
  - Tests written and passing (unit + integration)
- **D-14:** No changes to the standalone are in scope for Phase 9 of security_reviewer.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Standalone opengrep-mcp (source of truth)
- `/home/saghaulor/code/opengrep-mcp/cmd/opengrep-mcp/main.go` — Binary entrypoint; stdio transport setup, SAST_ENGINE reading, token handling
- `/home/saghaulor/code/opengrep-mcp/internal/server/server.go` — Three MCP tools, tier validation (only "pro"), path validation
- `/home/saghaulor/code/opengrep-mcp/internal/runner/runner.go` — Docker CLI orchestration, backend selection, token masking
- `/home/saghaulor/code/opengrep-mcp/internal/config/backend.go` — Backend enum: `semgrep-pro` → `returntocorp/semgrep:1.55.0`, `opengrep` → `opengrep/opengrep:1.20.0`
- `/home/saghaulor/code/opengrep-mcp/Makefile` — Build targets to reference for Makefile integration
- `/home/saghaulor/code/opengrep-mcp/go.mod` — Module path: `github.com/security_reviewer/opengrep-mcp`; dep: `github.com/modelcontextprotocol/go-sdk v1.4.0`

### Security Reviewer integration files to modify
- `/home/saghaulor/code/security_reviewer/.mcp.json` — Must switch `opengrep` entry from SSE to stdio
- `/home/saghaulor/code/security_reviewer/.claude/commands/security-review.md` — Must remove Step 3 (Docker health-check + bootstrap)
- `/home/saghaulor/code/security_reviewer/Makefile` — Must add `build-opengrep-mcp` target

### Phase 9 PHASE.md (full requirements reference)
- `.planning/phases/09-opengrep-mcp-server/PHASE.md` — Contains REQ-mcp-O1..O8 and success criteria

### Prior phase orchestration decisions
- `.planning/phases/02-claude-security-hooks-go-binary/02-CONTEXT.md` — Prior decisions on hooks binary patterns (TDD, stdlib-only)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `.claude/hooks/bin/` — Existing bin directory where `claude-security-hooks` lives; `opengrep-mcp` binary goes here too (consistent location for all project binaries)
- `security_reviewer/Makefile` — Already has build targets for `claude-security-hooks`; extend with `build-opengrep-mcp` using same CGO_ENABLED=0 pattern
- `scripts/graphify-mcp.sh` — Example of how stdio MCP servers are wired in `.mcp.json`; opengrep follows the same `type: "stdio"` pattern

### Established Patterns
- **Binary placement**: `[module]/bin/[name]` in standalone, `.claude/hooks/bin/[name]` in security_reviewer — Phase 9 copies from standalone to local bin
- **Docker CLI only**: Standalone uses `exec.Command("docker", ...)` — no Docker SDK dependency (matching existing project pattern)
- **TDD**: Tests before implementation. For Phase 9, this means testing the Makefile target produces a valid binary before writing the target.
- **No compound shell commands**: Run each bash command as a separate call (CLAUDE.md requirement)

### Integration Points
- `.mcp.json` controls how Claude Code discovers and spawns MCP servers — the `opengrep` entry is the single connection point
- `security-review.md` Step 3 currently manages the opengrep-mcp lifecycle (to be removed)
- Agent definitions use `mcp__opengrep__scan_with_rule` and `mcp__opengrep__get_ast` — these route through `.mcp.json` `opengrep` server entry automatically
- `SAST_ENGINE` env in `.mcp.json` env block is inherited by the subprocess — no runtime configuration changes needed

### Key discovery from code scout
The standalone `server.go` validates `tier != "pro"` and returns an error. This means agents calling `mcp__opengrep__scan_with_rule` MUST pass `tier: "pro"`. The taint tracer's internal `semgrep_tier: "intrafile"` evidence weighting is applied after getting results back — it doesn't affect what tier is sent to the MCP server.

</code_context>

<specifics>
## Specific Ideas

- Binary install path: `.claude/hooks/bin/opengrep-mcp` — same directory as `claude-security-hooks`, consistent lookup
- `.mcp.json` env block sets `SAST_ENGINE=opengrep` explicitly (not relying on binary default) for clarity
- The Makefile target delegates to the standalone repo's build system: `$(MAKE) -C /home/saghaulor/code/opengrep-mcp build` copies binary from `opengrep-mcp/bin/opengrep-mcp` to `.claude/hooks/bin/opengrep-mcp`
- Step 3 removal from security-review.md is a clean delete — no replacement step needed; the health-check loop and Docker container management for opengrep-mcp goes away entirely
- Phase 9 has no new Go code in security_reviewer — all changes are config files, Makefile, and orchestration command

</specifics>

<deferred>
## Deferred Ideas

- **Three-tier dispatch** (pro/intrafile/ce per-call routing): Original ROADMAP requirement; simplified to env-var model in standalone. Could be revisited in a future phase if different scan tiers are needed per-invocation.
- **Standalone repo tier aliases** (`"intrafile"` → accepted as `"pro"` in server): Not needed since all agent calls use `tier: "pro"`. If tier diversity is restored, patch the standalone server then.
- **opengrep-intrafile local build**: ROADMAP mentioned building a local `opengrep-intrafile:latest` image. Not needed with the env-var backend model.
- **Embedding opengrep-mcp source in security_reviewer monorepo**: User chose to keep it as an external dep. Revisit if the standalone repo needs to be archived or moved.

</deferred>

---

*Phase: 9-opengrep-mcp-server*
*Context gathered: 2026-05-26*
