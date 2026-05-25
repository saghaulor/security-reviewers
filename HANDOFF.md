# Security Reviewer — Handoff Document

**Date:** 2026-05-25  
**Branch:** `security_reviewers_initial_features`  
**Last commit:** `ce0740d`

---

## Current State

The end-to-end `/security-review` pipeline exists and is structurally complete through Phase 9. **One blocker remains before the first successful full run:** the `opengrep-mcp` server does not yet expose an MCP transport endpoint, so `mcp__opengrep__*` tools are unavailable to subagents.

### What works today

| Component | Status |
|---|---|
| `claude-security-hooks` binary | ✅ Built, installed at `.claude/hooks/bin/` |
| H7 fix (Status field in PostToolUseEvent) | ✅ In binary |
| Phase 9 schema (ReviewSessionID in all 5 output types + 4 input types) | ✅ Complete |
| Joint invariants T12/AZ7/OA8/IC5 (session ID matching) | ✅ Implemented + tested |
| H5 test (stdlib-only dep check) | ✅ Fixed — was using wrong `ForTest` heuristic |
| H6 test (all 51 assertion IDs covered) | ✅ Passing |
| `claude-security-hooks uuid` subcommand | ✅ RFC 4122 v4, `crypto/rand`, no external deps |
| Orchestration command (`/security-review`) | ✅ Rewritten — explicit steps, no shortcuts |
| Stale-file cleanup in orchestration | ✅ Step 2 deletes all prior-run intermediates |
| SESSION_ID threading | ✅ All agents receive `review_session_id` |
| govulncheck | ✅ Runs via `docker run golang:latest` |
| graphify MCP | ✅ Configured via `scripts/graphify-mcp.sh` wrapper |
| gopls/lsp MCP | ✅ Provided by globally-installed `gopls-lsp` plugin |
| opengrep MCP SSE transport | ❌ **MISSING — see below** |

---

## One Remaining Blocker: opengrep-mcp SSE Transport

### Problem

`opengrep-mcp` (`~/code/opengrep-mcp`) runs as a Docker container on port 8000. It currently only exposes:
- `GET /health` → `{"status": "ok"}`

It does **not** expose an MCP protocol endpoint. The project's `.mcp.json` configures Claude Code to connect to it at `http://localhost:8000/sse`, but that endpoint doesn't exist yet, so subagents cannot call `mcp__opengrep__scan_with_rule` or `mcp__opengrep__get_ast`.

### What needs to be built

Add MCP SSE transport to `~/code/opengrep-mcp`. The MCP SSE protocol requires:

**`GET /sse`** — Client connects; server immediately sends:
```
event: endpoint
data: /message?sessionId=<uuid>

```
Then holds the connection open and streams JSON-RPC responses as SSE events.

**`POST /message?sessionId=<uuid>`** — Client sends JSON-RPC requests. Server dispatches them and writes responses to the SSE stream for that session.

**JSON-RPC methods to handle:**
- `initialize` — MCP handshake (return server info + capabilities)
- `tools/list` — return the three tool definitions
- `tools/call` — dispatch to `ScanWithRule`, `GetAST` (ScanDirectory is currently a stub)

### Implementation approach (TDD)

**Use the official Go MCP SDK:** `github.com/modelcontextprotocol/go-sdk`

The SDK provides the SSE transport (`server.NewSSETransport`) and the JSON-RPC dispatch layer. You only need to wire the three existing handler methods into it — no protocol implementation from scratch.

**Rough steps:**
1. `go get github.com/modelcontextprotocol/go-sdk` in `~/code/opengrep-mcp`
2. Write failing tests first (`internal/mcp/transport_test.go`): test that `GET /sse` returns 200 with `Content-Type: text/event-stream`, that `tools/list` returns all three tools, that `tools/call scan_with_rule` dispatches correctly.
3. Create `internal/mcp/handler.go` wrapping the SDK server
4. Register `/sse` and `/message` handlers in `cmd/opengrep-mcp/main.go`
5. `make build` + `docker build -t opengrep-mcp:latest .` + restart container

**Note on H5:** The hooks binary has an H5 test enforcing zero non-stdlib deps. That's for `claude-security-hooks`, not for `opengrep-mcp` — they're separate modules. opengrep-mcp can freely add the SDK dep.

### Verification

After implementing, run:
```bash
# Rebuild and restart container
docker stop opengrep-mcp-e2e-phase8
docker rm opengrep-mcp-e2e-phase8
docker build -t opengrep-mcp:latest ~/code/opengrep-mcp
docker run -d --name opengrep-mcp-e2e-phase8 -p 8000:8000 opengrep-mcp:latest

# Confirm SSE endpoint is live
curl -N http://localhost:8000/sse
# Expected: event: endpoint\ndata: /message?sessionId=...
```

Then start a new Claude Code session (to reload `.mcp.json`) and run:
```
/security-review ./examples/sample-vulnerable-service
```

---

## Architecture Reference

### Pipeline stages (from orchestration command)

```
Step 0  Resolve TARGET_DIR (realpath $ARGUMENTS)
Step 1  Generate SESSION_ID  →  claude-security-hooks uuid
Step 2  Write .current-review; delete stale intermediate files
Step 3  Bootstrap opengrep-mcp  →  curl http://localhost:8000/health
Step 4  Pre-pass:
          graphify update TARGET_DIR  →  graphify-out/graph.json
          docker run golang:latest govulncheck -json ./...  →  govulncheck.json
Step 5  Cartographer (sequential Task)
          in: {working_directory, review_session_id}
          out: TARGET_DIR/go-index.json
Step 6  Read go-index.json to extract routes, sinks, authz_primitives, oauth_locations
Step 7  Four tracers (parallel Tasks):
          go-authz-tracer       →  authz-findings.json
          go-oauth-auditor      →  oauth-checklist.json
          invariant-checker     →  invariant-results.json
          go-taint-tracer (×N)  →  taint-verdict-<handler>-sqli.json
        All four receive review_session_id
Step 8  Synthesis (sequential Task, waits for all Step 7)
          in: {working_directory, review_id: SESSION_ID}
          out: review-report.json, review-report.md
Step 9  Report findings summary
```

### MCP servers (`.mcp.json`)

| Name | Type | Command/URL | Tools exposed |
|---|---|---|---|
| `graphify` | stdio | `scripts/graphify-mcp.sh` (wraps `graphify TARGET --mcp`) | `mcp__graphify__query_graph`, `get_node`, `get_neighbors`, `shortest_path`, `god_nodes`, `get_community` |
| `opengrep` | sse | `http://localhost:8000/sse` | `mcp__opengrep__scan_with_rule`, `mcp__opengrep__get_ast` (**not live until SSE transport is added**) |
| `gopls` | (plugin) | globally-installed `gopls-lsp` plugin | `mcp__gopls__go_references`, `go_search`, `go_workspace`, `go_package_api`, `go_file_context`, `go_symbol_references` |
| `lsp` | (plugin) | globally-installed `gopls-lsp` plugin | `mcp__lsp__textDocument_definition`, `textDocument_implementation`, `callHierarchy_outgoingCalls`, `callHierarchy_incomingCalls` |

### graphify MCP — dynamic target

`scripts/graphify-mcp.sh` reads `.current-review` (written by orchestration Step 2) to determine which directory's `graphify-out/graph.json` to serve. `.current-review` is gitignored. Falls back to `examples/sample-vulnerable-service` if absent.

**Known limitation:** The graphify MCP server starts once per Claude Code session. If you change the review target mid-session, you need to restart the session for the new `.current-review` value to take effect.

### Hooks binary

Located at `.claude/hooks/bin/claude-security-hooks`. Source at `claude-security-hooks/`. Three hooks registered in `.claude/settings.json`:

| Hook | Event | Matcher |
|---|---|---|
| `preflight` | PreToolUse | Task |
| `validate` | PostToolUse | Task |
| `inject-context` | SubagentStart | `go-cartographer\|go-taint-tracer\|go-authz-tracer\|go-oauth-auditor\|invariant-checker\|synthesis` |

**Rebuild after code changes:**
```bash
cd claude-security-hooks
go test ./...          # all must pass (including H5, H6)
make build
cp bin/claude-security-hooks ../.claude/hooks/bin/
```

---

## Key files

```
.claude/
  commands/security-review.md   # orchestration command — the pipeline definition
  agents/
    go-cartographer.md           # structural indexer
    go-taint-tracer.md           # SQLi/injection tracer (1 per source-sink pair)
    go-authz-tracer.md           # authorization checker
    go-oauth-auditor.md          # OAuth RFC conformance
    invariant-checker.md         # business-logic invariants
    synthesis.md                 # deduplicator + report generator
  hooks/bin/claude-security-hooks  # installed binary
  settings.json                    # hook registrations

.mcp.json                          # MCP server config (graphify + opengrep)
scripts/graphify-mcp.sh            # graphify --mcp wrapper

claude-security-hooks/             # hooks binary source (Go module)
  internal/
    hooks/
      events.go                    # PreToolUseEvent / PostToolUseEvent (Status field for H7)
      validate.go                  # PostToolUse validation dispatcher
      dispatch.go                  # per-agent invariant runners
    invariants/
      taint_tracer.go              # T1-T11 + T12 (session ID joint)
      authz_tracer.go              # AZ1-AZ6 + AZ7 (session ID joint)
      oauth_auditor.go             # OA1-OA7 + OA8 (session ID joint)
      invariant_checker.go         # IC1-IC4 + IC5 (session ID joint)
      cartographer.go              # A1-A11
      synthesis.go                 # S1-S6
    schema/                        # all agent I/O structs (ReviewSessionID in all)
    uuidgen/
      uuid.go                      # RFC 4122 v4 via crypto/rand (no deps)
  cmd/claude-security-hooks/
    main.go                        # preflight | validate | inject-context | uuid

~/code/opengrep-mcp/               # MCP SAST server (separate module — needs SSE transport)
  cmd/opengrep-mcp/main.go         # HTTP server (only /health currently)
  internal/server/server.go        # tool definitions + handlers
  internal/mcp/                    # (directory created, empty — SSE transport goes here)

examples/sample-vulnerable-service/ # test target (deliberately vulnerable Go service)
  handlers.go                       # SQLi, authz bypass, OAuth issues
  main.go
  (no intermediate files — cleaned up, will be regenerated on next run)
```

---

## Session history

| Phase | What was built |
|---|---|
| Phase 8 | Full pipeline: all 6 agents, 51 hook assertions, graphify/govulncheck pre-pass, synthesis |
| Phase 9 | Session ID threading: ReviewSessionID in all schemas; T12/AZ7/OA8/IC5 joint invariants; H7 fix (Status field); uuid subcommand; orchestration rewrite |
| This session | H5 test fix; confirmed Phase 9 code correct; diagnosed orchestration shortcut bug; fixed it; govulncheck → Docker; MCP config (.mcp.json + graphify wrapper) |

---

## Next session checklist

- [ ] Implement MCP SSE transport in `~/code/opengrep-mcp` using the Go MCP SDK
- [ ] Rebuild Docker image and restart container
- [ ] Start a **new Claude Code session** (required to reload `.mcp.json`)
- [ ] Run `/security-review ./examples/sample-vulnerable-service`
- [ ] Verify all 6 agents appear in the agents view
- [ ] Verify `review-report.json` has `review_session_id` matching the run's UUID
- [ ] Verify a second run with a new UUID ignores the first run's intermediate files
