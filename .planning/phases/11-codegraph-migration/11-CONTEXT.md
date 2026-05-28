# Phase 11: codegraph migration — Context

**Gathered:** 2026-05-27
**Status:** Ready for planning
**Source:** PRD Express Path (migration_HANDOFF.md)

<domain>
## Phase Boundary

Replace graphify with codegraph as the graph/MCP layer in the security-review pipeline. This phase is a targeted find-and-replace of the graphify integration surface: the MCP wrapper script, the `.mcp.json` entry, the `go-cartographer` agent definition (tools list + body), and the orchestration command Step 3. The `go-taint-tracer` update is optional (stretch goal). No new agent types, no new hook predicates, no schema changes.

**What ships at the end of this phase:**
- `scripts/codegraph-mcp.sh` (replaces `scripts/graphify-mcp.sh`)
- `.mcp.json` graphify entry replaced with codegraph entry
- `go-cartographer.md` updated to `mcp__codegraph__*` tools
- `security-review.md` Step 3 updated from `graphify update` to `codegraph index`
- `examples/sample-vulnerable-service/.gitignore` gains `.codegraph/`

**Out of scope:** New hook predicates, schema changes, new agents, changes to govulncheck step, changes to the tracer fan-out.

</domain>

<decisions>
## Implementation Decisions

### Replacement Target — graphify surface
- **D-1 (LOCKED):** Delete `scripts/graphify-mcp.sh`. Do not keep it as a fallback — the migration is complete, not conditional.
- **D-2 (LOCKED):** Replace the `graphify` key in `.mcp.json` with a `codegraph` key (stdio transport). The command is a new `scripts/codegraph-mcp.sh` wrapper.
- **D-3 (LOCKED):** The `.current-review` mechanism is preserved. `codegraph-mcp.sh` reads `.current-review` exactly as `graphify-mcp.sh` does — same contract, same fallback to `examples/sample-vulnerable-service`.
- **D-4 (LOCKED):** `go-cartographer.md` frontmatter `tools:` replaces all `mcp__graphify__*` entries with the correct `mcp__codegraph__*` names (to be confirmed by research).
- **D-5 (LOCKED):** `go-cartographer.md` body §2 Preconditions replaces the `graphify-out/graph.json` existence check with the equivalent codegraph output artifact check.
- **D-6 (LOCKED):** `security-review.md` Step 3 replaces `graphify update TARGET_DIR` with the correct `codegraph index TARGET_DIR` (or equivalent command — to be confirmed by research). The `.current-review` write in Step 2 is preserved unchanged.
- **D-7 (LOCKED):** `examples/sample-vulnerable-service/.gitignore` gains `.codegraph/` (codegraph writes its index into the target project directory).

### codegraph source
- **D-8 (LOCKED):** codegraph is at `/home/saghaulor/code/codegraph` (TypeScript, already pulled). The `codegraph` binary must be on PATH for `codegraph-mcp.sh` to work — the wrapper does not pin an absolute path.
- **D-9 (LOCKED):** No codegraph build step is added to the repo Makefile in this phase. The binary is assumed pre-installed by the developer.

### go-taint-tracer update
- **D-10 (Discretion):** Optionally add a note to `go-taint-tracer.md` recommending use of `mcp__codegraph__codegraph_trace` for data-flow confirmation. This is a stretch goal — do not block phase completion on it.

### Security-review Step 2 cleanup
- **D-11 (LOCKED):** Remove the sentence "First, write TARGET_DIR to `.current-review` in the project root so the graphify MCP server wrapper knows which graph to serve" from Step 2 description — replace with "so the codegraph MCP server wrapper knows which graph to serve."

### Claude's Discretion
- Exact `mcp__codegraph__*` tool names — research must confirm the correct identifiers from `codegraph serve --mcp`.
- Exact `codegraph index` command syntax and output artifact path.
- Whether `go-cartographer.md` §4 (graph traversal steps) needs rewording to match codegraph query semantics.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Files being modified
- `.mcp.json` — MCP server config; graphify entry to be replaced with codegraph entry
- `scripts/graphify-mcp.sh` — to be deleted; new `scripts/codegraph-mcp.sh` modeled after it
- `.claude/agents/go-cartographer.md` — agent definition; `tools:` list and body updated for codegraph
- `.claude/commands/security-review.md` — orchestration command; Step 2 description + Step 3 graphify call updated
- `examples/sample-vulnerable-service/.gitignore` — gains `.codegraph/`

### codegraph source (read-only)
- `/home/saghaulor/code/codegraph` — TypeScript source; read to confirm MCP tool names, index command, output artifact location, and `--path` flag semantics

### Optional files
- `.claude/agents/go-taint-tracer.md` — stretch: add `codegraph_trace` reference if research confirms it is useful

</canonical_refs>

<specifics>
## Specific Ideas

- `codegraph serve --mcp --path TARGET_DIR` — the equivalent of `graphify "$TARGET_DIR" --mcp` (from handoff; confirm exact flag spelling in research)
- `codegraph_trace`, `codegraph_callers` — codegraph MCP tools with security analysis value; replace `mcp__graphify__query_graph` traversal steps in cartographer body
- codegraph has a "native route node kind" — may improve entrypoint detection in `go-cartographer.md`
- No LLM API key required — codegraph is 100% local analysis; this is why it works on Bedrock

</specifics>

<deferred>
## Deferred Ideas

- Integrating codegraph route detection into the go-cartographer index schema (would require schema changes — deferred to a future phase)
- Adding `codegraph` binary install to `scripts/install.sh` (nice-to-have for onboarding; deferred)
- Replacing `govulncheck` Docker step with a local equivalent (out of scope — no connection to graphify migration)

</deferred>

---

*Phase: 11-codegraph-migration*
*Context gathered: 2026-05-27 via PRD Express Path (migration_HANDOFF.md)*
