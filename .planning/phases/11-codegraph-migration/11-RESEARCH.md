# Phase 11: codegraph migration — Research

**Researched:** 2026-05-27
**Domain:** MCP server config migration — graphify → codegraph
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-1:** Delete `scripts/graphify-mcp.sh`. Do not keep it as a fallback — the migration is complete, not conditional.
- **D-2:** Replace the `graphify` key in `.mcp.json` with a `codegraph` key (stdio transport). The command is a new `scripts/codegraph-mcp.sh` wrapper.
- **D-3:** The `.current-review` mechanism is preserved. `codegraph-mcp.sh` reads `.current-review` exactly as `graphify-mcp.sh` does — same contract, same fallback to `examples/sample-vulnerable-service`.
- **D-4:** `go-cartographer.md` frontmatter `tools:` replaces all `mcp__graphify__*` entries with the correct `mcp__codegraph__*` names (to be confirmed by research — now confirmed).
- **D-5:** `go-cartographer.md` body §2 Preconditions replaces the `graphify-out/graph.json` existence check with the equivalent codegraph output artifact check.
- **D-6:** `security-review.md` Step 3 replaces `graphify update TARGET_DIR` with `codegraph index TARGET_DIR` (or equivalent command — now confirmed). The `.current-review` write in Step 2 is preserved unchanged.
- **D-7:** `examples/sample-vulnerable-service/.gitignore` gains `.codegraph/` (codegraph writes its index into the target project directory).
- **D-8:** codegraph is at `/home/saghaulor/code/codegraph` (TypeScript, already pulled). The `codegraph` binary must be on PATH for `codegraph-mcp.sh` to work — the wrapper does not pin an absolute path.
- **D-9:** No codegraph build step is added to the repo Makefile in this phase. The binary is assumed pre-installed by the developer.
- **D-11:** Remove the sentence "First, write TARGET_DIR to `.current-review` in the project root so the graphify MCP server wrapper knows which graph to serve" from Step 2 description — replace with "so the codegraph MCP server wrapper knows which graph to serve."

### Claude's Discretion
- Exact `mcp__codegraph__*` tool names — research must confirm the correct identifiers from `codegraph serve --mcp`.
- Exact `codegraph index` command syntax and output artifact path.
- Whether `go-cartographer.md` §4 (graph traversal steps) needs rewording to match codegraph query semantics.

### Deferred Ideas (OUT OF SCOPE)
- Integrating codegraph route detection into the go-cartographer index schema (would require schema changes — deferred to a future phase)
- Adding `codegraph` binary install to `scripts/install.sh` (nice-to-have for onboarding; deferred)
- Replacing `govulncheck` Docker step with a local equivalent (out of scope — no connection to graphify migration)
</user_constraints>

---

## Summary

This phase replaces the graphify graph/MCP layer with codegraph across five files. The scope is narrow and concrete: one script delete, one script create, one JSON key replacement, two agent definition updates, and one `.gitignore` line addition.

The key research finding is that codegraph's MCP and CLI surface differs significantly from graphify's. The graphify approach served a pre-built `graph.json` file via `graphify "$TARGET_DIR" --mcp`. The codegraph approach requires a **two-step workflow**: (1) initialize the project once with `codegraph init TARGET_DIR` (creates `.codegraph/codegraph.db`), then (2) index with `codegraph index TARGET_DIR` on each review run. The MCP server is launched separately via `codegraph serve --mcp` (optionally with `--path`). Critically, `codegraph serve --mcp` auto-discovers the project root by walking up from `process.cwd()` looking for `.codegraph/codegraph.db` — so `--path` is optional when the MCP server is launched from or resolved to the correct directory.

The output artifact is `.codegraph/codegraph.db` (SQLite) inside the target project directory — not a flat JSON file. There is no `graph.json` equivalent to hash as `graph_version`. The cartographer precondition check must be rewritten: instead of `Read graphify-out/graph.json`, it should verify reachability by calling `mcp__codegraph__codegraph_status` (the direct replacement for `mcp__graphify__graph_stats`).

**Primary recommendation:** The migration is a find-and-replace with two non-obvious decisions: (a) `codegraph-mcp.sh` needs `--path` flag because the MCP server is launched from the project root, not TARGET_DIR; and (b) `go-cartographer.md` needs a two-step pre-pass: `codegraph init` (idempotent if already initialized) then `codegraph index`.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Graph indexing (build index from source) | Orchestration command (Step 3) | — | `codegraph index` is a CLI pre-pass like govulncheck; owned by the orchestration skill |
| MCP server lifecycle | Claude Code runtime | — | `codegraph serve --mcp` is spawned as a stdio subprocess by Claude Code, same as opengrep-mcp |
| Graph query (entrypoints, callers, trace) | go-cartographer agent | go-taint-tracer (stretch) | Agent calls `mcp__codegraph__*` tools against the running MCP server |
| Index artifact (`.codegraph/`) | Target project directory | — | codegraph writes its SQLite DB into TARGET_DIR, not the security_reviewer repo |
| Precondition check in cartographer | go-cartographer §2 | — | Replaces `graphify-out/graph.json` existence check with `codegraph_status` reachability call |

---

## Standard Stack

### Core (this migration — no new packages installed in security_reviewer repo)

| Component | Version | Purpose | Source |
|-----------|---------|---------|--------|
| `@colbymchenry/codegraph` | 0.9.6 | Graph indexer + MCP server | [VERIFIED: /home/saghaulor/code/codegraph/package.json] |
| `codegraph` CLI binary | 0.9.6 | `init`, `index`, `serve --mcp` subcommands | [VERIFIED: source read] |

**No packages are installed into the security_reviewer repo in this phase.** The `codegraph` binary is an external developer tool assumed pre-installed on PATH (D-9).

### codegraph binary install (developer prerequisite, out of scope for Makefile in this phase)

```bash
# From the pulled source at /home/saghaulor/code/codegraph:
npm install       # install dependencies
npm run build     # compile TypeScript → dist/
npm link          # or: npm install -g .
```

---

## Package Legitimacy Audit

> No packages are installed in the security_reviewer repo by this phase. The only artifact being integrated is the `codegraph` binary assumed pre-installed (D-9). No legitimacy gate required for this phase.

---

## Architecture Patterns

### System Architecture Diagram

```
security-review.md Step 3 (orchestration)
  │
  ├── codegraph init TARGET_DIR   ← creates .codegraph/codegraph.db (idempotent)
  └── codegraph index TARGET_DIR  ← populates/refreshes the SQLite index

.mcp.json (Claude Code reads at startup)
  └── codegraph: { command: scripts/codegraph-mcp.sh }
        │
        └── reads .current-review → TARGET_DIR
              └── exec: codegraph serve --mcp --path TARGET_DIR
                    │
                    └── MCP stdio transport ← Claude Code runtime
                          │
                          └── mcp__codegraph__* tools available to agents

go-cartographer (agent)
  ├── §2 Preconditions:
  │     ├── codegraph_status  ← reachability check (replaces graph_stats)
  │     └── (no file existence check — index is a DB, not a JSON file)
  ├── §3 Step 1: call codegraph_status → graph_version = DB size or node count
  ├── §3 Step 5: codegraph_callers (authz primitives) + opengrep patterns
  ├── §3 Step 6: codegraph_search (OAuth surface)
  └── §3 Step 9: codegraph_trace / codegraph_callers (ambiguous edges)
```

### Recommended Project Structure (files changed)

```
security_reviewer/
├── scripts/
│   ├── graphify-mcp.sh       ← DELETE
│   └── codegraph-mcp.sh      ← CREATE (new)
├── .mcp.json                 ← replace graphify key with codegraph key
├── .claude/
│   ├── agents/
│   │   ├── go-cartographer.md ← update tools + body
│   │   └── go-taint-tracer.md ← OPTIONAL: add codegraph_trace note
│   └── commands/
│       └── security-review.md ← update Step 2 description + Step 3
└── examples/sample-vulnerable-service/
    └── .gitignore             ← add .codegraph/
```

### Pattern 1: codegraph-mcp.sh wrapper (replaces graphify-mcp.sh)

**What:** Reads `.current-review`, falls back to sample service, then execs `codegraph serve --mcp --path TARGET_DIR`.

**Key difference from graphify:** The `--path` flag is passed explicitly because the MCP server is launched from the security_reviewer project root (not TARGET_DIR), and codegraph's auto-discovery walks up from `process.cwd()` — which would find the wrong project (or no project) without an explicit `--path`.

```bash
#!/usr/bin/env bash
# codegraph-mcp.sh — Stdio MCP wrapper for codegraph.
#
# The target directory is read from .current-review (written by the
# /security-review orchestration command at Step 2). This lets the
# codegraph MCP server stay pointed at the currently-reviewed service
# without requiring a session restart.
#
# Falls back to examples/sample-vulnerable-service if the file is absent.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TARGET_FILE="$PROJECT_ROOT/.current-review"

if [ -f "$TARGET_FILE" ]; then
  TARGET_DIR="$(cat "$TARGET_FILE")"
else
  TARGET_DIR="$PROJECT_ROOT/examples/sample-vulnerable-service"
fi

exec codegraph serve --mcp --path "$TARGET_DIR"
```

[VERIFIED: source read of `/home/saghaulor/code/codegraph/src/bin/codegraph.ts` lines 1141-1193 — `serve` command with `-p, --path <path>` and `--mcp` flags confirmed]

### Pattern 2: codegraph init + index two-step in security-review.md Step 3

**What:** The graphify `update` command was a single idempotent call. codegraph requires `init` (first time only, idempotent on subsequent runs — outputs warning if already initialized, continues) then `index` (re-indexes every run).

```bash
# Step 3a: Initialize codegraph in target (idempotent — safe to re-run)
codegraph init TARGET_DIR

# Step 3b: Index the project (refreshes the index for this review run)
codegraph index TARGET_DIR
```

[VERIFIED: source read of `/home/saghaulor/code/codegraph/src/bin/codegraph.ts` lines 415-623 — `init` command confirmed idempotent (warns "Already initialized" but exits 0); `index` command confirmed as separate step]

### Pattern 3: go-cartographer precondition replacement

**Old precondition:**
1. `Read graphify-out/graph.json` — confirms file exists
2. SHA-256 hash of `graph.json` → `graph_version`
3. Call `mcp__graphify__graph_stats` — confirms MCP reachability

**New precondition:**
1. Call `mcp__codegraph__codegraph_status` — confirms MCP reachability AND confirms index exists
2. `graph_version` = derive from status output (e.g. files indexed count + total nodes as a fingerprint string, or omit entirely — the DB is not hashable like a flat file)
3. No file existence check needed — the DB is queried through the MCP, not read directly

**Recommendation for `graph_version` field:** Change from SHA-256 of a file to a string derived from `codegraph_status` output: `"codegraph:<fileCount>files/<nodeCount>nodes"`. This preserves the semantic intent (fingerprint the graph state) without requiring direct DB file access.

[VERIFIED: source read — `codegraph_status` handler at tools.ts lines 2144-2230; returns `fileCount`, `nodeCount`, `edgeCount`, `nodesByKind` stats]

### Pattern 4: mcp__codegraph__* tool name prefix

Claude Code derives the MCP tool prefix from the key name in `.mcp.json`. Changing the key from `"graphify"` to `"codegraph"` produces the prefix `mcp__codegraph__`. The tool suffix is the `name` field from `tools.ts`:

[VERIFIED: source read of `/home/saghaulor/code/codegraph/src/mcp/tools.ts` lines 357-570]

### Anti-Patterns to Avoid

- **Omitting `--path` from `codegraph serve`:** Without `--path`, the MCP server auto-discovers the project root by walking up from its working directory. Since the wrapper is launched by Claude Code (which runs from security_reviewer root), the server would try to find `.codegraph/codegraph.db` in the security_reviewer tree, not TARGET_DIR. Always pass `--path "$TARGET_DIR"`.
- **Running `codegraph serve --mcp` before `codegraph index`:** The server starts and responds to MCP calls but `codegraph_status` will show 0 files indexed and all tools return empty results. The index step in Step 3 must complete before the cartographer agent is dispatched.
- **Single-step `codegraph index` without `codegraph init`:** Running `codegraph index TARGET_DIR` on an uninitialised project fails with "CodeGraph not initialized in $path — Run 'codegraph init' first." Both steps are required for a fresh target.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| MCP server with codegraph tools | Custom server or wrapper | `codegraph serve --mcp` | Already implemented in codegraph source; stdio transport matches opengrep-mcp pattern |
| Graph version fingerprint | SHA-256 of DB file | Derive from `codegraph_status` output | `.codegraph/codegraph.db` is a live SQLite file (WAL mode); hashing it while the server runs is unreliable |
| Route entrypoint detection | Custom regex in cartographer | `codegraph_search` with `kind: "route"` filter | codegraph has a native `route` NodeKind (confirmed in types.ts line 39) — query `kind: "route"` directly |

**Key insight:** The `route` NodeKind in codegraph is a first-class concept, not inferred from regex. `codegraph_search` accepts a `kind` filter including `"route"` — this is more reliable than the current Semgrep router-detection pattern in cartographer Step 3. (Schema integration is deferred per D context, but the cartographer body can reference `codegraph_search` with `kind: "route"` for route discovery.)

---

## Confirmed MCP Tool Names

The following tool names are registered by `codegraph serve --mcp`. These become `mcp__codegraph__<name>` in Claude Code.

[VERIFIED: source read of `/home/saghaulor/code/codegraph/src/mcp/tools.ts` lines 357-570]

| Tool Name | MCP Identifier | Purpose | Security-Analysis Value |
|-----------|---------------|---------|------------------------|
| `codegraph_search` | `mcp__codegraph__codegraph_search` | Symbol search by name/kind | Find OAuth handlers, authz middleware, sink callsites |
| `codegraph_context` | `mcp__codegraph__codegraph_context` | Task-oriented context assembly | Broad first-pass for "how does auth work here?" |
| `codegraph_callers` | `mcp__codegraph__codegraph_callers` | Who calls a symbol | Find all callers of a sink (e.g. `db.Exec`) — replaces `mcp__graphify__get_neighbors` |
| `codegraph_callees` | `mcp__codegraph__codegraph_callees` | What a symbol calls | Trace forward from entrypoint |
| `codegraph_impact` | `mcp__codegraph__codegraph_impact` | Impact radius of a symbol change | Less useful for security; available |
| `codegraph_node` | `mcp__codegraph__codegraph_node` | Symbol details + call trail | Replaces `mcp__graphify__get_node` |
| `codegraph_explore` | `mcp__codegraph__codegraph_explore` | Multi-symbol source view | Batch inspection of related symbols |
| `codegraph_status` | `mcp__codegraph__codegraph_status` | Index stats | **Direct replacement for `mcp__graphify__graph_stats`** — reachability check |
| `codegraph_files` | `mcp__codegraph__codegraph_files` | Project file tree | File exploration |
| `codegraph_trace` | `mcp__codegraph__codegraph_trace` | Call path from A → B | **High value for security:** trace source → sink path statically; replaces manual graph traversal in §3 Steps 5-6-9 |

**Tools removed (no equivalent needed):**
- `mcp__graphify__query_graph` → replaced by `codegraph_search` + `codegraph_callers` + `codegraph_trace`
- `mcp__graphify__get_node` → replaced by `codegraph_node`
- `mcp__graphify__get_neighbors` → replaced by `codegraph_callers` + `codegraph_callees`
- `mcp__graphify__shortest_path` → replaced by `codegraph_trace`
- `mcp__graphify__god_nodes` → no direct equivalent; `codegraph_impact` provides similar insight
- `mcp__graphify__get_community` → no direct equivalent; `codegraph_explore` covers community-like grouping

**Recommended `go-cartographer.md` tools list (frontmatter):**

```
mcp__codegraph__codegraph_search, mcp__codegraph__codegraph_callers, mcp__codegraph__codegraph_trace, mcp__codegraph__codegraph_node, mcp__codegraph__codegraph_status, mcp__gopls__go_search, mcp__gopls__go_workspace, mcp__gopls__go_package_api, mcp__gopls__go_references, mcp__opengrep__scan_with_rule, Bash, Read, Glob
```

---

## Complete File-by-File Change Map

### 1. `scripts/graphify-mcp.sh` → DELETE

No replacement fallback. D-1 is a hard delete.

### 2. `scripts/codegraph-mcp.sh` → CREATE

Same contract as `graphify-mcp.sh`. Only the final `exec` line changes:
- Old: `exec graphify "$TARGET_DIR" --mcp`
- New: `exec codegraph serve --mcp --path "$TARGET_DIR"`

File must be `chmod +x`.

### 3. `.mcp.json` → UPDATE

```json
{
  "mcpServers": {
    "codegraph": {
      "type": "stdio",
      "command": "/home/saghaulor/code/security_reviewer/scripts/codegraph-mcp.sh"
    },
    "opengrep": {
      "type": "stdio",
      "command": "/home/saghaulor/code/security_reviewer/.claude/hooks/bin/opengrep-mcp",
      "env": {
        "SAST_ENGINE": "opengrep"
      }
    }
  }
}
```

[VERIFIED: current `.mcp.json` read — only the `graphify` key needs replacement]

### 4. `.claude/agents/go-cartographer.md` → UPDATE

**Frontmatter `tools:` line:** Replace all six `mcp__graphify__*` entries with the five security-relevant `mcp__codegraph__*` tools:
- Remove: `mcp__graphify__query_graph, mcp__graphify__get_node, mcp__graphify__get_neighbors, mcp__graphify__shortest_path, mcp__graphify__god_nodes, mcp__graphify__get_community`
- Add: `mcp__codegraph__codegraph_search, mcp__codegraph__codegraph_callers, mcp__codegraph__codegraph_trace, mcp__codegraph__codegraph_node, mcp__codegraph__codegraph_status`

**Body §2 Preconditions:** Replace both graphify precondition checks:
- Remove: "1. `graphify-out/graph.json` exists in the working directory. If it does not exist, the user must run `graphify build .` from the project root before invoking this agent."
- Remove: "2. A Graphify MCP server is running and reachable against the built graph. Test reachability by calling `mcp__graphify__graph_stats` with no arguments."
- Replace with: "1. The codegraph index is initialized and populated. Test reachability by calling `mcp__codegraph__codegraph_status` with no arguments. If the call fails with a connection error or returns 0 files indexed, return a structured error and stop — the user must run `codegraph init TARGET_DIR` then `codegraph index TARGET_DIR` before invoking this agent."

**Body §2 error response shape:** Update `detail` field:
- Old: `"detail": "graphify-out/graph.json not found. Run: graphify build ."`
- New: `"detail": "codegraph index not found or empty. Run: codegraph init TARGET_DIR then: codegraph index TARGET_DIR"`

**Body §3 Step 1:** Replace graphify-specific checks:
- Old: "Read `graphify-out/graph.json` to confirm it exists. Compute its SHA-256 hash — this becomes `graph_version` in the output. Call `mcp__graphify__graph_stats` to confirm the MCP server is reachable."
- New: "Call `mcp__codegraph__codegraph_status` to confirm the MCP server is reachable and to retrieve index statistics. Extract `fileCount` and `nodeCount` from the response — format as `'codegraph:<fileCount>files/<nodeCount>nodes'` and use this as `graph_version` in the output."

**Body §3 Step 5 (authz primitives):** Replace `mcp__graphify__query_graph` call:
- Old: "Use `mcp__graphify__query_graph` to search for function symbols that appear on paths between entrypoints and business logic"
- New: "Use `mcp__codegraph__codegraph_callers` to find all symbols that call entrypoint handlers, then use `mcp__codegraph__codegraph_trace` with `from: <entrypoint>` and `to: <business_logic_symbol>` to confirm paths between entrypoints and business logic."

**Body §3 Step 6 (OAuth surface):** Replace `mcp__graphify__query_graph` call:
- Old: "use `mcp__graphify__query_graph` to find the handler symbols that use these packages"
- New: "use `mcp__codegraph__codegraph_search` with the OAuth-related symbol names to find handlers, then `mcp__codegraph__codegraph_node` to confirm file/line locations"

**Body §3 Step 7 (payment surface):** Replace `mcp__graphify__get_community` and `mcp__graphify__query_graph`:
- Old: "Use `mcp__graphify__get_community` or `mcp__graphify__query_graph` to find clusters of symbols related to payment processing"
- New: "Use `mcp__codegraph__codegraph_search` to find symbols related to payment processing (search for symbol names containing: stripe, braintree, paypal, square, adyen, payment, billing, charge, invoice, subscription)"

**Body §3 Step 9 (ambiguous edges):** Replace `mcp__graphify__query_graph` and AMBIGUOUS edge concept:
- codegraph does not have an AMBIGUOUS edge type; dynamic dispatch is instead surfaced as a gap in `codegraph_trace` output
- New: "Call `mcp__codegraph__codegraph_trace` with `from: <entrypoint>` and `to: <sink>` for each entrypoint-sink pair where the call path is not already confirmed. Where `codegraph_trace` reports 'no static path' or 'breaks at dynamic dispatch', record those (entrypoint, sink) pairs in `ambiguous_nodes` with reason `dynamic_dispatch_break`."

**Body §5 Output Schema:** Update `graph_version` comment and Field notes:
- Old schema comment: `"graph_version": "<sha256 of graphify-out/graph.json>"`
- New schema comment: `"graph_version": "<codegraph status fingerprint: 'codegraph:<N>files/<M>nodes'>"`
- Old field note: `- \`graph_version\`: SHA-256 hex string of \`graphify-out/graph.json\` at analysis time.`
- New field note: `- \`graph_version\`: codegraph status fingerprint string at analysis time: \`"codegraph:<N>files/<M>nodes"\``

**Body §6 Hard Rules:** Update A6 and A10:
- A6 now references codegraph node IDs (from `codegraph_node` or `codegraph_search` responses), not `graphify-out/graph.json`
- A10: change "Read-only operation: You MUST NOT modify any file in `graphify-out/`" → "You MUST NOT modify any file in `.codegraph/` or in the source tree."

### 5. `.claude/commands/security-review.md` → UPDATE

**Step 2 description:** Update one sentence per D-11:
- Old: "First, write TARGET_DIR to `.current-review` in the project root so the graphify MCP server wrapper knows which graph to serve"
- New: "First, write TARGET_DIR to `.current-review` in the project root so the codegraph MCP server wrapper knows which graph to serve"

**Step 3:** Replace graphify commands with codegraph two-step:
- Old: `Bash: graphify update TARGET_DIR` → "This produces `TARGET_DIR/graphify-out/graph.json`."
- New:
  ```
  1. `Bash: codegraph init TARGET_DIR` (idempotent — safe to re-run if already initialized)
  2. `Bash: codegraph index TARGET_DIR` — This populates/refreshes `.codegraph/codegraph.db` inside TARGET_DIR. If it fails, stop with error.
  ```

### 6. `examples/sample-vulnerable-service/.gitignore` → UPDATE

Add one line: `.codegraph/`

---

## Common Pitfalls

### Pitfall 1: `codegraph serve --mcp` without `--path` resolves wrong project root
**What goes wrong:** The MCP server is launched via `scripts/codegraph-mcp.sh`, which Claude Code spawns with its own working directory (security_reviewer root). Without `--path`, `findNearestCodeGraphRoot(process.cwd())` walks up from the security_reviewer root looking for `.codegraph/codegraph.db`. If security_reviewer itself has been initialized with codegraph (it may have, since codegraph ships with a `.claude/` installer), the server serves the security_reviewer graph instead of TARGET_DIR.
**Why it happens:** codegraph's auto-discovery walks parent dirs; the MCP server inherits the Claude Code process's cwd, not TARGET_DIR.
**How to avoid:** Always pass `--path "$TARGET_DIR"` explicitly in `codegraph-mcp.sh`. [VERIFIED: source read of `codegraph serve` action — `const projectPath = options.path ? resolveProjectPath(options.path) : undefined`]
**Warning signs:** `codegraph_status` shows files from the security_reviewer codebase (TypeScript files) instead of the target Go service.

### Pitfall 2: Running `codegraph index` without `codegraph init` on a fresh target
**What goes wrong:** `codegraph index TARGET_DIR` exits with "CodeGraph not initialized in $path — Run 'codegraph init' first" and Step 3 fails.
**Why it happens:** The `index` command calls `isInitialized(projectPath)` and exits non-zero if `.codegraph/codegraph.db` does not exist.
**How to avoid:** Always run `codegraph init TARGET_DIR` before `codegraph index TARGET_DIR` in Step 3. `init` is idempotent — it logs a warning and exits 0 if already initialized. [VERIFIED: source read of `init` and `index` commands in codegraph.ts]

### Pitfall 3: Node.js version incompatibility (Node 25+ blocked)
**What goes wrong:** codegraph hard-exits on Node.js 25+ due to a V8 turboshaft WASM JIT bug that crashes tree-sitter grammar compilation.
**Why it happens:** codegraph's CLI includes an explicit version gate: `if (nodeMajor >= 25) { process.exit(1) }` unless `CODEGRAPH_ALLOW_UNSAFE_NODE` is set. [VERIFIED: source read — codegraph.ts lines 63-71]
**How to avoid:** Use Node.js 20–24. Current installed version is `v24.15.0` (confirmed via `npm list -g`). This version is within the supported range (`"node": ">=20.0.0 <25.0.0"` in package.json).
**Warning signs:** `codegraph: Node.js 25` banner on stderr, immediate exit.

### Pitfall 4: `codegraph` binary not on PATH
**What goes wrong:** `codegraph-mcp.sh` silently fails when Claude Code tries to spawn the MCP server. Claude Code may show the server as unavailable or show a connection error.
**Why it happens:** The source repo at `/home/saghaulor/code/codegraph` is TypeScript and has no built `dist/` directory. The binary is not globally installed. `which codegraph` returns nothing.
**How to avoid:** The developer must build and install codegraph before running a review: `npm install` + `npm run build` + `npm link` (or `npm install -g .`) from `/home/saghaulor/code/codegraph`. This is out of scope for the phase Makefile (D-9) but must be documented in the verification step.
**Warning signs:** `command not found: codegraph` in the MCP server spawn error.

### Pitfall 5: A6 hard rule references non-existent graphify node IDs
**What goes wrong:** After migration, go-cartographer still has A6 referencing `graphify-out/graph.json` as the authoritative source for node IDs. An agent following the old rule would cite non-existent node IDs or reference a deleted file.
**Why it happens:** A6 in the current body reads: "Every graph node ID cited in `ambiguous_nodes` or elsewhere MUST exist in `graphify-out/graph.json`."
**How to avoid:** Update A6 to: "Every node ID cited in `ambiguous_nodes` MUST be a real node ID returned by a `mcp__codegraph__codegraph_search` or `mcp__codegraph__codegraph_node` call in this session."

---

## Code Examples

### Confirmed: `codegraph serve` command signature
```typescript
// Source: /home/saghaulor/code/codegraph/src/bin/codegraph.ts lines 1141-1193
program
  .command('serve')
  .description('Start CodeGraph as an MCP server for AI assistants')
  .option('-p, --path <path>', 'Project path (optional for MCP mode, uses rootUri from client)')
  .option('--mcp', 'Run as MCP server (stdio transport)')
  .option('--no-watch', 'Disable the file watcher (no auto-sync)')
```

### Confirmed: `codegraph index` requires prior `codegraph init`
```typescript
// Source: /home/saghaulor/code/codegraph/src/bin/codegraph.ts line 569
if (!isInitialized(projectPath)) {
  error(`CodeGraph not initialized in ${projectPath}`);
  info('Run "codegraph init" first');
  process.exit(1);
}
```

### Confirmed: `.codegraph/codegraph.db` is the output artifact
```typescript
// Source: /home/saghaulor/code/codegraph/src/directory.ts lines 26-33
export function isInitialized(projectRoot: string): boolean {
  const codegraphDir = getCodeGraphDir(projectRoot);  // projectRoot/.codegraph
  // Must have codegraph.db, not just .codegraph folder
  const dbPath = path.join(codegraphDir, 'codegraph.db');
  return fs.existsSync(dbPath);
}
```

### Confirmed: All 10 MCP tool names
```typescript
// Source: /home/saghaulor/code/codegraph/src/mcp/tools.ts lines 357-570
// Full list:
'codegraph_search', 'codegraph_context', 'codegraph_callers', 'codegraph_callees',
'codegraph_impact', 'codegraph_node', 'codegraph_explore', 'codegraph_status',
'codegraph_files', 'codegraph_trace'
```

### Confirmed: `route` is a native NodeKind
```typescript
// Source: /home/saghaulor/code/codegraph/src/types.ts lines 18-41
export const NODE_KINDS = [
  // ... (other kinds)
  'route',    // ← native first-class node kind
  'component',
] as const;
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `graphify "$TARGET_DIR" --mcp` (single command) | `codegraph serve --mcp --path TARGET_DIR` (needs separate `init` + `index` first) | Phase 11 | Two-step pre-pass required; init is idempotent |
| `graphify-out/graph.json` (flat JSON, hashable) | `.codegraph/codegraph.db` (SQLite, not hashable while live) | Phase 11 | `graph_version` field must change derivation strategy |
| `mcp__graphify__graph_stats` (reachability check) | `mcp__codegraph__codegraph_status` (direct replacement) | Phase 11 | Same role, different tool name |
| `mcp__graphify__shortest_path` (explicit path query) | `mcp__codegraph__codegraph_trace` (from/to symbol names) | Phase 11 | More expressive: returns code at each hop |
| AMBIGUOUS edge type in graphify | `codegraph_trace` "no static path" / "dynamic dispatch break" | Phase 11 | No edge-type concept in codegraph; dynamic dispatch surfaced differently |

**Deprecated/removed:**
- `graphify-out/` directory: no longer produced; `.gitignore` entry can stay (it's harmless but won't be populated)
- `mcp__graphify__god_nodes`: no direct equivalent in codegraph; remove from cartographer body
- `mcp__graphify__get_community`: no direct equivalent; replace with `codegraph_explore` or `codegraph_search`

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `codegraph serve --mcp --path TARGET_DIR` with an uninitialized path (no `.codegraph/codegraph.db`) starts the server but returns empty results, not a startup error | Pitfall 1 / Pattern 1 | If server errors on startup instead, the wrapper script needs error handling |
| A2 | `codegraph init` exits 0 (not non-zero) when the project is already initialized, so the two-step in Step 3 is safe to re-run | Pattern 2 / Pitfall 2 | If init exits non-zero on re-run, Step 3 would need an existence check or `2>/dev/null` |
| A3 | `.mcp.json` `"codegraph"` key name produces `mcp__codegraph__` tool prefix in Claude Code | Pattern 4 / Tool Name section | If Claude Code uses a different prefix derivation, all tool references in agents must change |

**A1 note:** The source shows `MCPServer` constructor accepts `undefined` for `projectPath` and handles lazy init — the server likely starts without error and surfaces "not initialized" only on individual tool calls. But the exact error behavior for an uninit path with explicit `--path` was not traced through `MCPServer.start()` completely.

**A2 note:** Source read of `init` command (line 429-431) confirms: `if (isInitialized(projectPath)) { clack.log.warn('Already initialized'); ... clack.outro(''); return; }` — this `return`s without calling `process.exit(1)`, so it exits 0. HIGH confidence.

**A3 note:** This is how Claude Code MCP prefix derivation works for all other servers (e.g. `"opengrep"` → `mcp__opengrep__`). ASSUMED based on observable behavior from existing setup.

---

## Open Questions (RESOLVED)

1. **Does `codegraph serve --mcp --path <uninitialized-path>` error at startup or on first tool call?**
   - What we know: `MCPServer` constructor accepts optional `projectPath`; `start()` resolves daemon root via `findNearestCodeGraphRoot`; if no root found, falls back to direct mode
   - What's unclear: In direct mode with an explicit uninit path, does the server start but fail on tool calls, or fail at startup?
   - RESOLVED: The Step 3 `init` + `index` sequence prevents this in practice. go-cartographer §2 precondition calls `codegraph_status` and fails fast if 0 files indexed, surfacing a clear error before any analysis proceeds.

2. **Should `graphify-out/` entry remain in `.gitignore` of `examples/sample-vulnerable-service/`?**
   - What we know: The current `.gitignore` has `graphify-out/`. After migration, graphify will not run and `graphify-out/` will never be created.
   - RESOLVED: Leave `graphify-out/` in `.gitignore` (harmless). Add `.codegraph/` as a new line. D-7 only requires adding `.codegraph/` — no removal required.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `codegraph` binary | `scripts/codegraph-mcp.sh`, Step 3 | NOT on PATH | — | None — D-8 states binary must be on PATH |
| Node.js 20–24 | codegraph runtime | v24.15.0 | 24.15.0 | — |
| `npm` | Building codegraph from source | v10.x (with nvm) | present | — |

**Missing dependencies with no fallback:**
- `codegraph` binary: Not on PATH and not globally installed. Source at `/home/saghaulor/code/codegraph` has no built `dist/` directory (TypeScript source only). The planner must include a prerequisite step: build and install codegraph before running the migration's smoke test (but NOT in the Makefile — D-9).

**Missing dependencies with fallback:**
- None.

---

## Validation Architecture

> This phase is a pure config/script/markdown migration — no Go code changes. No Go tests are added or modified. The only testable exit criterion is a manual smoke test: `/security-review examples/sample-vulnerable-service` completes with codegraph as the graph source and `review-report.json` still contains non-ambiguous findings.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | None (no Go changes) |
| Config file | N/A |
| Quick run command | `codegraph_status` call via MCP (manual verification) |
| Full suite command | `/security-review examples/sample-vulnerable-service` (human-run) |

### Phase Requirements → Test Map

| Req | Behavior | Test Type | Automated Command | File Exists? |
|-----|----------|-----------|-------------------|-------------|
| SC-1 | `scripts/codegraph-mcp.sh` exists and is executable | smoke | `test -x scripts/codegraph-mcp.sh` | ❌ Wave 0 (file created by phase) |
| SC-2 | `scripts/graphify-mcp.sh` is deleted | smoke | `test ! -f scripts/graphify-mcp.sh` | ❌ Wave 0 |
| SC-3 | `.mcp.json` has `codegraph` key, no `graphify` key | smoke | `jq '.mcpServers.codegraph' .mcp.json` | ❌ Wave 0 |
| SC-4 | No `mcp__graphify__` references remain in agent defs | smoke | `grep -r mcp__graphify__ .claude/` (should be empty) | ❌ Wave 0 |
| SC-5 | `examples/sample-vulnerable-service/.gitignore` has `.codegraph/` | smoke | `grep -c .codegraph .gitignore` | ❌ Wave 0 |
| SC-7 | Full review run completes with codegraph | integration | manual `/security-review examples/sample-vulnerable-service` | manual |

### Wave 0 Gaps
- `scripts/codegraph-mcp.sh` — created by phase
- All file changes in scope — created/updated by phase; no test file scaffolding needed (no Go code)

---

## Security Domain

> This phase makes no code changes to security-enforcing logic. It replaces one graph tool with another. No new attack surface is introduced. The MCP server pattern (stdio transport, local process) matches the existing opengrep-mcp pattern already in use.

### ASVS Categories Applicable

| ASVS Category | Applies | Notes |
|---------------|---------|-------|
| V5 Input Validation | No | No new input paths |
| V6 Cryptography | No | No crypto changes |
| All others | No | Config/script migration only |

---

## Sources

### Primary (HIGH confidence)
- `/home/saghaulor/code/codegraph/src/bin/codegraph.ts` — CLI commands: `serve`, `init`, `index`; flag names; error handling
- `/home/saghaulor/code/codegraph/src/mcp/tools.ts` — All 10 MCP tool names and their `inputSchema`
- `/home/saghaulor/code/codegraph/src/directory.ts` — `.codegraph/codegraph.db` as the output artifact; `isInitialized()` contract
- `/home/saghaulor/code/codegraph/src/types.ts` — `route` NodeKind confirmed
- `/home/saghaulor/code/codegraph/package.json` — version 0.9.6; Node engine constraint `>=20.0.0 <25.0.0`
- `/home/saghaulor/code/security_reviewer/scripts/graphify-mcp.sh` — existing wrapper pattern to replicate
- `/home/saghaulor/code/security_reviewer/.mcp.json` — current graphify entry structure
- `/home/saghaulor/code/security_reviewer/.claude/agents/go-cartographer.md` — full current body; all change targets identified
- `/home/saghaulor/code/security_reviewer/.claude/commands/security-review.md` — Step 2 and Step 3 change targets

### Secondary (MEDIUM confidence)
- `/home/saghaulor/code/security_reviewer/migration_HANDOFF.md` — intent summary; flag names confirmed by source read

---

## Metadata

**Confidence breakdown:**
- MCP tool names: HIGH — read directly from `tools.ts` source
- CLI command syntax (`serve`, `init`, `index`): HIGH — read directly from `codegraph.ts` source
- Output artifact (`.codegraph/codegraph.db`): HIGH — read from `directory.ts`
- `codegraph init` idempotency (exits 0): HIGH — traced through source, no `process.exit(1)` on re-init
- `mcp__codegraph__` prefix derivation: ASSUMED — based on observable pattern from existing `mcp__opengrep__` setup
- `codegraph serve --mcp` behavior with uninit path: LOW — MCPServer start() not fully traced

**Research date:** 2026-05-27
**Valid until:** Stable until codegraph source changes — re-verify if `/home/saghaulor/code/codegraph` is updated before Phase 11 executes.
