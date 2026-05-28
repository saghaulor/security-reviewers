# Constraints Intel

Technical contracts (schemas, protocols, NFRs, allowlists) lifted from HAND_OFF.md. These are SPEC-grade constraints that bind implementation choices.

source-doc: /home/saghaulor/code/security_reviewer/HAND_OFF.md
classification: SPEC, precedence 0 (manifest override)

---

## CON-arch-three-layers — Three-layer review architecture

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §2.1
- type: protocol
- scope: review-time orchestration
- constraint:
  1. **Pre-pass (once per repo, on dep change):** `govulncheck` produces `govulncheck.json`; `graphify build` produces `graphify-out/`.
  2. **Sequential cartographer (per review):** `go-cartographer` consumes pre-pass artifacts via Graphify/gopls/Semgrep MCPs and emits `go-index.json`.
  3. **Parallel fan-out (per review):** `go-taint-tracer` (one per source/sink pair), `go-authz-tracer` (one per middleware chain scope), `go-oauth-auditor` (one per OAuth surface), `invariant-checker` (one per declared invariant) run concurrently from cartographer output.
  4. **Synthesis (after fan-out completes):** `synthesis` reads the directory of specialist outputs and produces `review-report.{json,md}`.

## CON-engine-tier-abstraction — Single MCP tool with tier parameter

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §2.3, §2.4
- type: api-contract
- scope: `opengrep-mcp` server
- constraint: Agents call `mcp__opengrep__scan_with_rule(rule_yaml, paths, tier)` where `tier ∈ {"pro", "intrafile", "ce"}`. The MCP server dispatches to the correct container image based on tier. Agents do not select containers directly.

## CON-tool-assignment-matrix — Engine assignment by analysis layer

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §2.3
- type: protocol
- scope: per-agent tool selection
- constraint:
  - Cartographer structural map → Graphify MCP
  - Cartographer precise resolution → gopls MCP (`go_search`, `go_workspace`, `go_package_api`)
  - Cartographer CVE reachability → `govulncheck` (deterministic, no LLM)
  - Cartographer entrypoint enumeration → Semgrep/OpenGrep pattern mode (not taint)
  - Tracer data flow → Semgrep Pro taint (preferred) → OpenGrep intrafile (fallback) → gopls cross-check (always)
  - Tracer interface dispatch → gopls `textDocument/implementation` (LSP) — required for interface-typed flows
  - Authz tracer → gopls call hierarchy + middleware-chain walk (control-flow dominator problem)
  - OAuth auditor checklist → Read + gopls + Semgrep patterns
  - OAuth taint pairs → reuse `go-taint-tracer` with OAuth source/sink table

## CON-schema-go-index-v1 — go-index/v1 JSON schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (Output schema block)
- type: schema
- scope: `go-cartographer` output
- constraint: Output document must conform to the `go-index/v1` schema with required top-level fields: `schema_version`, `graph_version`, `routers_detected`, `entrypoints`, `sinks_by_kind`, `authz_primitives`, `oauth_locations`, `payment_surface`, `vuln_deps`, `ambiguous_nodes`, `warnings`. Sink kinds enumeration: `sql_exec | cmd_exec | http_client | fs_path | deserialize | template_render | xml_parse`. Authz primitive kinds: `middleware | guard | decorator`. Confidence enum: `extracted | inferred`.

## CON-schema-taint-verdict — Taint tracer verdict schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (Output schema block)
- type: schema
- scope: `go-taint-tracer` output
- constraint: Top-level `verdict ∈ {exploitable | sanitized | unreachable | ambiguous | input_mismatch}`. Top-level `confidence ∈ {high | medium | low}`. Path step enum: `source | assign | call | return | sanitize | sink | iface_dispatch | chan_send | chan_recv`. Required telemetry counters under `gopls`: `references_calls`, `definition_calls`, `implementation_calls`, `call_hierarchy_calls`, `branches_explored`, `branches_unexplored`, `interface_fanout_max`, `goroutine_boundaries_crossed`. Source kinds: `http_query | http_form | http_body | http_header | http_cookie | grpc_arg | env | file | stdin | channel_recv`. Sink kinds: `sql_exec | cmd_exec | http_client | fs_path | deserialize | template_render | xml_parse | sql_query_raw | reflect_call`.

## CON-schema-authz — Authz tracer schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.3
- type: schema
- scope: `go-authz-tracer` I/O
- constraint: Input includes `routes[]`, `authz_primitives[]`, `sensitive_operations[]`. Output `summary` has buckets `routes_total`, `protected`, `missing`, `weak`, `idor_risk`, `public_intentional`. Finding `issue` enum: `missing_authz | weak_authz | idor_risk | bypass_path | non_blocking_middleware`. Router enum: `chi | gin | mux | net_http | echo | fiber`.

## CON-schema-oauth-auditor — OAuth auditor schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4
- type: schema
- scope: `go-oauth-auditor` I/O
- constraint: Input `target_profile ∈ {oauth_2_1 | oauth_2_0 | oauth_2_0_with_9700_bcp}`. Input `features_in_use[]` enum: `pkce | par | dpop | ciba | token_exchange | dynamic_client_registration | introspection | revocation`. Output `checklist[].status ∈ {pass | fail | not_applicable | unverified}`. Severity enum: `critical | high | medium | low | info`. Output produces `taint_pairs[]` to be dispatched downstream to `go-taint-tracer`.

## CON-schema-invariant — Invariant checker schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.5
- type: schema
- scope: `invariant-checker` I/O
- constraint: Input contains `flow_name` and `invariants[]` (each with `id`, `statement`, `anchor_symbols[]`). Output `results[].status ∈ {holds | violated | unverifiable}`. The agent does NOT mint new invariant IDs in output.

## CON-schema-review-report — review-report/v1 schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.6
- type: schema
- scope: `synthesis` output
- constraint: Top-level: `review_id`, `timestamp`, `scma_version` (literal `"review-report/v1"`), `summary`, `findings[]`, `deduplication_notes[]`. Severity enum: `critical | high | medium | low | info`. Class enum: `injection | authz | oauth | invariant | vuln_dep`. Synthesis produces both `review-report.json` and `review-report.md`.

## CON-input-taint-tracer — Taint tracer input contract

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (Input contract block)
- type: api-contract
- scope: orchestrator → `go-taint-tracer` Task prompt
- constraint: The orchestrator's Task prompt body must be valid JSON of the shape `{source: {file, line, expr, kind}, sink: {file, line, expr, kind}, max_depth: int, semgrep_tier: "pro"|"intrafile"|"ce"}`. Verified by `claude-security-hooks preflight` on `PreToolUse`.

## CON-tracer-protocol — Numbered behavior protocol (taint tracer)

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (Behavior contract)
- type: protocol
- scope: `go-taint-tracer` system prompt
- constraint: Steps 1–9 in order: (1) verify input + read source/sink ±10 lines, (2) resolve types via `go_file_context`, (3) build Semgrep YAML rule, (4) run Semgrep scoped ≤50 files, (5) interpret Semgrep result, (6) gopls cross-check with recursive depth-1 propagation and interface-implementation walks, (7) stop conditions (depth=0, 60 LSP calls, 5 unexplored branches, reflection/unsafe), (8) verdict per truth table, (9) never substitute Grep for symbol queries.

## CON-tool-allowlists — Per-agent tool allowlists

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1–§3.6 (Tool allowlist blocks)
- type: protocol
- scope: `.claude/agents/*.md` frontmatter
- constraint:
  - **go-cartographer:** `mcp__graphify__*` (query_graph, get_node, get_neighbors, shortest_path), `mcp__gopls__*` (go_search, go_workspace, go_package_api, go_references), `mcp__opengrep__scan_with_rule`, `Bash` (restricted via `allowed_commands` to `govulncheck -json ./...` and `govulncheck -json -mode=source ./...`), `Read`, `Glob`.
  - **go-taint-tracer:** `mcp__opengrep__scan_with_rule`, `mcp__opengrep__get_ast`, `mcp__gopls__go_references`, `mcp__gopls__go_symbol_references`, `mcp__gopls__go_search`, `mcp__gopls__go_package_api`, `mcp__gopls__go_file_context`, `mcp__lsp__textDocument_implementation`, `mcp__lsp__callHierarchy_outgoingCalls`, `mcp__lsp__callHierarchy_incomingCalls`, `mcp__lsp__textDocument_definition`, `Read`, `Glob`. **Deliberately omitted:** `Grep`, `Bash`, `Edit`, `Write`.
  - **go-authz-tracer:** `mcp__gopls__go_references`, `mcp__gopls__go_symbol_references`, `mcp__gopls__go_search`, `mcp__lsp__callHierarchy_outgoingCalls`, `mcp__lsp__textDocument_definition`, `mcp__lsp__textDocument_implementation`, `Read`, `Glob`.
  - **go-oauth-auditor:** `mcp__gopls__go_search`, `mcp__gopls__go_references`, `mcp__gopls__go_file_context`, `mcp__opengrep__scan_with_rule`, `Read`, `Glob`.
  - **invariant-checker:** `mcp__gopls__go_search`, `mcp__gopls__go_references`, `mcp__gopls__go_file_context`, `mcp__lsp__callHierarchy_outgoingCalls`, `mcp__lsp__textDocument_definition`, `Read`, `Glob`.
  - **synthesis:** `Read`, `Glob` only.

## CON-hook-registration — settings.json hook registration shape

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7
- type: api-contract
- scope: `.claude/settings.json`
- constraint: Three hook registrations:
  - `PreToolUse` matcher `Task` → `.claude/hooks/bin/claude-security-hooks preflight` (timeout 5s)
  - `PostToolUse` matcher `Task` → `.claude/hooks/bin/claude-security-hooks validate` (timeout 15s)
  - `SubagentStart` matcher `go-taint-tracer|go-authz-tracer|go-oauth-auditor|go-cartographer|invariant-checker|synthesis` → `.claude/hooks/bin/claude-security-hooks inject-context` (timeout 5s)
  - **`SubagentStop` is explicitly NOT used** (payload lacks subagent identification with parallel runs).

## CON-validate-protocol — Hook validate subcommand flow

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (Validate subcommand logic)
- type: protocol
- scope: `claude-security-hooks validate`
- constraint: (1) Read `PostToolUse` JSON from stdin. (2) If `tool_input.subagent_type` not in security agent set, exit 0 silently. (3) Extract `tool_response.content`. (4) Strip optional code-fence wrapping. (5) Parse JSON into the per-agent Go struct. (6) Run invariant list for that agent. (7) On any failure emit `{"decision":"block","reason":"..."}` to stdout and exit 0. (8) On success exit 0 with no output.

## CON-nfr-cold-start — Hook binary cold-start latency

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 D2
- type: nfr
- scope: `claude-security-hooks`
- constraint: Binary cold start must be <5ms (the explicit reason Go was chosen over Python). Hook timeouts are 5–15s; cold start must not erode that budget meaningfully under parallel fan-out.

## CON-nfr-zero-deps — Hooks zero non-stdlib dependencies

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H5) and §3.7 dependency note
- type: nfr
- scope: `claude-security-hooks` core validator
- constraint: The core validator path must compile with zero non-stdlib imports. Test framework dependencies allowed in `_test.go` only.

## CON-nfr-static-binary — Static, CGO-disabled binaries

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (build target), §3.8 (assertion O1)
- type: nfr
- scope: `claude-security-hooks`, `opengrep-mcp`
- constraint: Both binaries build with `CGO_ENABLED=0` and run on `gcr.io/distroless/static-debian12` (or equivalent) without dynamic library dependencies.

## CON-nfr-container-isolation — Read-only workspace mount

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O5), §3.8 container strategy
- type: nfr
- scope: `opengrep-mcp` container invocations
- constraint: Workspaces mounted into scanner containers must be read-only (`:ro`). Rule YAML mounted read-only at `/tmp/rule.yaml`. Container is killed on timeout exceedance.

## CON-source-sink-catalog — Go source/sink/sanitizer catalog

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §8.2
- type: protocol
- scope: `go-taint-tracer` system prompt
- constraint: Canonical Go source kinds, sink kinds, and recognized sanitizers per sink kind (full table in §8.2). The agent's system prompt must embed this catalog verbatim. Unrecognized sanitizers (third-party, native, generated bodies unreadable) classify as `unverified` — not credited.

## CON-oauth-source-sink-table — OAuth-specific source/sink/sanitizer table

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2.2
- type: protocol
- scope: `go-taint-tracer` system prompt (OAuth extension)
- constraint: When source/sink kind is OAuth-specific, the tracer applies the OAuth table in addition to the baseline catalog. Pairs include scope tampering, redirect_uri tampering, client authentication, state CSRF, code reuse, audience binding, CIBA `login_hint`, CIBA `auth_req_id`, CIBA backchannel push notification target.

## CON-oauth-checklist-taxonomy — OAuth auditor checklist IDs

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §8.3
- type: protocol
- scope: `go-oauth-auditor` system prompt
- constraint: The auditor checklist contains the stable check IDs listed in §8.3 (OAUTH21-*, RFC9700-*, RFC9126-*, RFC8252-*, RFC9068-*, RFC8707-*, RFC9449-*, RFC8693-*, JOSE-*, RFC7591-*, RFC7662-*, RFC7009-*, CIBA-*, OIDC-*). Findings reference these IDs verbatim. Implementation v1 ships `critical`-tagged checks first, then `high`.

## CON-stop-conditions-taint — Taint tracer stop conditions

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (behavior step 7)
- type: nfr
- scope: `go-taint-tracer`
- constraint: Hard stops: traversal depth = 0, 60 LSP calls budget, 5 unexplored branches, reflection or `unsafe` package encountered. Semgrep scan capped at 50-file source-reachable set.

## CON-hook-payload-shapes — Hook event payload field names

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §8.1
- type: api-contract
- scope: `claude-security-hooks` stdin parsing
- constraint: The binary parses Claude Code hook event JSON with the field names documented in §8.1 (PreToolUse, PostToolUse, SubagentStart payloads). Implementing session must verify against current docs at https://code.claude.com/docs/en/hooks because schema may evolve. Q1/Q9 of §5 reinforce this verification duty.
