# Phase 9: opengrep-mcp server + infrastructure hardening

**Status:** In Progress — ⚠️ Plans blocked on research  
**Started:** 2026-05-26  
**Depends on:** Phase 8 (Automated E2E Testing)

---

## Goal

Ship the `opengrep-mcp` Go binary as a working MCP server that the security-review pipeline can invoke. At minimum it exposes `scan_with_rule` with tier-dispatched Docker containers (Semgrep Pro / OpenGrep intrafile / OpenGrep CE), read-only workspace mounts, timeout-based container kill, and `SEMGREP_APP_TOKEN` masking.

**Transport, tool surface, and entry point are TBD pending research.** Planning is blocked until the research phase resolves the open questions below.

Infrastructure work already committed as of Phase 9 start:
- Session ID threading through all tracers + orchestration
- `claude-security-hooks uuid` subcommand (replaces `uuidgen`)
- H5 predicate fix (two-invocation `go list` approach)
- Orchestration command hardening (stale file deletion, explicit Task JSON, mandatory fresh scan)
- govulncheck via Docker (no local install required)
- `.mcp.json` + `scripts/graphify-mcp.sh` MCP server registration

---

## ⚠️ Open Questions (Research Gate)

Planning cannot begin until these are resolved:

| # | Question | Why it matters |
|---|----------|----------------|
| Q1 | **Transport: SSE vs stdio?** Current `.mcp.json` points at `http://localhost:8000/sse`. SSE requires the server to be running before Claude Code starts; stdio is simpler and self-contained. Is concurrent multi-tool access actually needed? | Determines entire server architecture (HTTP listener vs stdin/stdout loop) |
| Q2 | **`get_ast` tool: include or drop?** Was in original HAND_OFF spec but no agent definition file currently lists `mcp__opengrep__get_ast` in its tool allowlist. | Tool surface size; build complexity |
| Q3 | **`scan_directory` vs merged into `scan_with_rule`?** Could be a `ruleset` param instead of a separate tool. | Schema surface; MCP registration complexity |

---

## Work Breakdown

### Already done (infrastructure — commits on `security_reviewers_initial_features`)

| Commit | What | Files |
|--------|------|-------|
| `4d86a5f` | Session ID threading: agent schema fields + orchestration command | `.claude/agents/*.md`, `.claude/commands/security-review.md` |
| `1d99906` | H5 test fix + orchestration command rewrite (stale files, fresh scan mandate) | `internal/invariants/h5_deps_test.go`, `security-review.md` |
| `aa7ed3a` | `uuid` subcommand on hooks binary (crypto/rand, stdlib-only) | `internal/uuidgen/`, `cmd/claude-security-hooks/main.go` |
| `ce0740d` | MCP config + govulncheck via Docker + graphify wrapper script | `.mcp.json`, `scripts/graphify-mcp.sh`, `security-review.md` |

### Still to implement (opengrep-mcp Go binary) — shape TBD

The `opengrep-mcp/` module has a `go.mod` stub and empty `cmd/`, `internal/` directories. The entire server must be built from scratch (TDD per project convention). Exact module layout and tool registration depend on research outcomes for Q1–Q3.

**Stable requirements (not affected by research):**

| Req | What |
|-----|------|
| REQ-mcp-O1 | Static `CGO_ENABLED=0` binary |
| REQ-mcp-O3 | Structured error on invalid `tier` |
| REQ-mcp-O5 | Workspace mounted read-only (`/src` with `ro` flag) |
| REQ-mcp-O6 | Container killed on `timeout_seconds` breach |
| REQ-mcp-O7 | Normalized findings shape regardless of engine |
| REQ-mcp-O8 | `SEMGREP_APP_TOKEN` never appears in stdout/stderr |

**Tier → image mapping (stable):**

| Tier | Docker image | Notes |
|------|-------------|-------|
| `pro` | `semgrep/semgrep:latest` (Pro token required) | Interprocedural taint analysis |
| `intrafile` | `opengrep-intrafile:latest` (local build) | Intrafile analysis only |
| `ce` | `opengrep/opengrep:latest` | Community edition |

---

## Success Criteria

1. `CGO_ENABLED=0 go build -o bin/opengrep-mcp ./cmd/opengrep-mcp` builds cleanly; health/ping check succeeds within 500 ms of startup.
2. MCP client probing the server receives at least `scan_with_rule` with a valid JSON Schema. *(Tool set may expand based on Q2/Q3 research outcomes.)*
3. `scan_with_rule(tier="bogus")` → structured error (not panic); `scan_with_rule(tier="intrafile")` → dispatches `opengrep-intrafile` container with `/src` mounted `ro`.
4. Integration test: `timeout_seconds: 1` on a slow rule kills the container and returns partial-results error.
5. `SEMGREP_APP_TOKEN=test-secret-do-not-log` never appears in server output during a Pro-tier invocation.
6. `go test ./...` passes (TDD — tests written before implementation for all new code).
7. *(Infrastructure — already passing)* `claude-security-hooks uuid` produces valid v4 UUID; H5 test passes; orchestration command enforces fresh scan with session ID threading.

---

## Plans

> **Blocked** — plan structure will be determined after Q1–Q3 are resolved in research.

Likely plan breakdown (provisional):
- [ ] `09-01-PLAN.md` — schema + runner (findings.go, runner.go + tests)
- [ ] `09-02-PLAN.md` — MCP server (transport TBD: SSE or stdio; tool registration)
- [ ] `09-03-PLAN.md` — cmd/main.go wiring + integration tests + O5/O6/O7/O8 hardening
- [ ] `09-04-PLAN.md` — Docker build verification + `.mcp.json` update to match final transport

---

## Notes

- The `docker/` directory exists in `opengrep-mcp/` — may already have a Dockerfile sketch from Phase 4.
- Use `github.com/modelcontextprotocol/go-sdk` for the MCP server (cited in original HAND_OFF §3.5).
- Token masking: at server init, read `SEMGREP_APP_TOKEN`; replace value with `[REDACTED]` in any log output before writing.
- If research chooses stdio transport, `.mcp.json` entry for `opengrep` switches from `type: "sse"` to `type: "stdio"` + `command:` pointing at the binary.
