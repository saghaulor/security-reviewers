# Security Review Agent System — Implementation Handoff

**Audience:** Claude Code session that will implement this design and commit to GitHub.
**Author context:** Produced collaboratively between sagjaulor and a Claude web session over an extended design conversation.
**Status:** Design-complete for the Go-first slice. Ready to implement.

---

## 0. How to read this document

The implementing Claude Code session should treat this document as the source of truth. Where this document and prior memory conflict, this document wins. Where this document is silent on an implementation detail, prefer simplicity and explicit code over cleverness. When uncertain, surface the uncertainty rather than guessing — the human who initiated this work prefers explicit gaps over inference.

The document is structured as:

1. **Problem statement** — what we're solving and why prior attempts failed.
2. **Architecture** — the full system shape.
3. **Component specifications** — one section per artifact to be built.
4. **Decisions log** — the choices made during design and their reasoning.
5. **Open questions** — items the implementing session must resolve before or during build.
6. **Implementation plan** — recommended sequencing.
7. **References** — RFCs, tool docs, MCP specs, Go libraries.

Every component spec contains **assertions** — invariants that must be true of the finished artifact. Treat these as test cases.

---

## 1. Problem statement

The user is building a security review system as a Claude Code plugin. The first attempt — loading a security review skill into the main agent and instructing it to dispatch tracer subagents with LSP-using instructions — failed because subagent context isolation makes the main agent an unreliable relay. Instructions in the skill that said "tell the tracer to use LSP" were silently dropped or paraphrased away when the main agent composed the Task tool prompt.

The user also identified a specific bug class that motivated this work: an OAuth scope-tampering bug where the consent form submission was trusted as the source-of-truth for granted scopes rather than the original authorization request. A previous Claude session missed this bug entirely. When pressed, that session said a dedicated tracer agent would have caught it. This system is the design response.

**Three failure modes the design must prevent:**

- Instructions buried in a skill loaded into the main agent silently disappearing when subagents are spawned.
- Agents choosing cheap heuristic tools (text grep, intuition) over expensive precise tools (Semgrep taint analysis, gopls call hierarchy) when both are available.
- Drift over time — six months from now, someone weakens an agent's system prompt or tool allowlist and no signal surfaces in normal output.

**The design's answers, in order:**

- Each specialist agent has its own definition file with its own tool allowlist. Instructions live where the agent reads them, not where the dispatcher reads them.
- Tool allowlists exclude cheaper alternatives. If you don't want the agent to use `Grep`, don't grant `Grep`. Numbered protocols in the system prompt enforce ordering of expensive tools.
- Lifecycle hooks validate every tracer's output mechanically: did it call Semgrep? did its verdict JSON conform to schema? On failure, block and force retry with explicit feedback.

---

## 2. Architecture

### 2.1 System layers

```
[Once per repo, on dep change]
  govulncheck         →  govulncheck.json
  graphify build      →  graphify-out/

[Per review, sequential]
  go-cartographer
    inputs: graphify-out, govulncheck.json
    queries: graphify MCP, gopls MCP, semgrep MCP
    outputs: go-index.json

[Per review, parallel fan-out from cartographer output]
  ┌─────────────────┬──────────────────┬───────────────────┬─────────────────────┐
  ↓                 ↓                  ↓                   ↓                     ↓
  go-taint-tracer   go-authz-tracer    go-oauth-auditor    invariant-checker     (one instance
  (one per          (one per scope     (one instance per   (one per declared      per OAuth-
   source/sink      — e.g. per         OAuth surface)      invariant)             specific
   pair)             middleware                                                    source/sink
                     chain)                                                        pair, via
                                                                                   go-taint-tracer)

[Per review, after all fan-out completes]
  synthesis agent → final findings report (JSON + Markdown)
```

### 2.2 Why two phases of fan-out

The cartographer is sequential and runs once. Its output is small (a structural index of the codebase) and is consumed by every downstream agent. Running it in parallel with the tracers would force each tracer to discover codebase shape independently, wasting tokens.

The tracers fan out massively. A medium Go service might have 200 entrypoints × 10 sink kinds = 2,000 candidate source/sink pairs. Most are filtered cheaply by the cartographer (no SQL sink in this handler's call graph → skip the SQL tracer). The rest run in parallel, one Task call per pair.

### 2.3 Tool / engine assignment by layer

| Layer | Primary tool | Reason |
|---|---|---|
| Cartographer structural map | Graphify MCP | Built for this. Confidence-tagged edges. Cheap. |
| Cartographer precise resolution | gopls MCP (`go_search`, `go_workspace`, `go_package_api`) | Type-aware. Authoritative when Graphify is AMBIGUOUS. |
| Cartographer CVE reachability | `govulncheck` (deterministic pre-pass, no LLM) | SSA-based reachability against the Go vuln DB. Not replaceable by patterns. |
| Cartographer entrypoint enumeration | Semgrep/OpenGrep pattern mode (not taint) | Imperative router registration → pattern match. |
| Tracer data flow | Semgrep Pro taint mode (preferred) → OpenGrep intrafile (fallback) → gopls cross-check (always) | Pro does interprocedural Go cross-file taint. OpenGrep does intrafile. gopls validates and catches what Semgrep misses (custom framework wrappers, dynamic dispatch). |
| Tracer interface dispatch | gopls `textDocument/implementation` (LSP) | Go's biggest taint-analysis blind spot. Required for any trace through interface params. |
| Authz tracer | gopls call hierarchy + middleware-chain walk | This is a control-flow dominator problem, not a data-flow problem. LSP-only is fine. |
| OAuth auditor checklist | Read + gopls + Semgrep patterns | Spec-conformance checking. Pattern + spot read. |
| OAuth taint pairs | Reuses `go-taint-tracer` with OAuth source/sink table | Scope tampering, redirect_uri tampering, etc. are data-flow bugs. |

### 2.4 Engine tier abstraction

The user has Semgrep Pro available and has built OpenGrep with intrafile analysis. Both should be available behind one MCP interface so the tracer agent calls a single tool with a `tier` parameter. This is the `opengrep-mcp` Go server (component 5 below). Its `scan_with_rule` tool dispatches to the right container based on tier.

```
agent → mcp__opengrep__scan_with_rule(rule_yaml, paths, tier="pro"|"intrafile"|"ce")
                                                  ↓
opengrep-mcp Go binary → docker run --rm -v ${paths}:/src:ro
                                  semgrep-pro-image | opengrep-image
                                  --config /tmp/rule.yaml /src
                         → parse JSON → return findings
```

---

## 3. Component specifications

Every component below has: purpose, inputs, outputs, tool/dependency list, behavior contract, and **assertions**. Treat assertions as acceptance tests.

### 3.1 Component: `go-cartographer` (Claude subagent)

**File:** `.claude/agents/go-cartographer.md`
**Purpose:** Build a structural index of the Go codebase for downstream tracers. Runs once per review.

**Tool allowlist:**

- `mcp__graphify__query_graph`
- `mcp__graphify__get_node`
- `mcp__graphify__get_neighbors`
- `mcp__graphify__shortest_path`
- `mcp__gopls__go_search`
- `mcp__gopls__go_workspace`
- `mcp__gopls__go_package_api`
- `mcp__gopls__go_references`
- `mcp__opengrep__scan_with_rule`
- `Bash` (restricted via `allowed_commands` to `govulncheck -json ./...` and `govulncheck -json -mode=source ./...`)
- `Read`
- `Glob`

**Inputs:** Working directory containing a Go module. Assumes `graphify-out/graph.json` exists and a Graphify MCP server is running against it.

**Outputs:** A single JSON document conforming to `go-index/v1` schema. Schema below.

**Behavior contract:**

1. Verify preconditions (graph exists, MCP reachable). On failure, return structured error and stop.
2. Detect routers via parallel `go_search` for known router imports. Multiple routers can coexist in one repo (e.g. `chi` for the API and `net/http` for `/healthz`).
3. For each detected router, run a router-specific Semgrep pattern to enumerate entrypoints (path, method, handler symbol, middleware chain).
4. For each sink kind in the taxonomy, run a Semgrep pattern scan to enumerate callsites.
5. Detect authz primitives by Graphify query + Semgrep middleware-shape pattern. Cross-reference for confidence. Read the body of each detected primitive to confirm it has a failure-return path; non-blocking "middleware" is excluded.
6. Detect OAuth surface via Graphify symbol search + import detection (`golang.org/x/oauth2`, `github.com/coreos/go-oidc`, `github.com/golang-jwt/jwt`, `github.com/ory/fosite`).
7. Detect payment surface via Graphify symbol cluster query.
8. Run `govulncheck`. Record each finding with the tool's `reachable` flag.
9. Surface Graphify AMBIGUOUS edges that touch entrypoints, sinks, or authz primitives — downstream tracers need to know what to verify.

**Output schema (`go-index/v1`):**

```json
{
  "schema_version": "go-index/v1",
  "graph_version": "<sha256 of graphify-out/graph.json>",
  "routers_detected": ["chi", "net/http"],
  "entrypoints": [
    {
      "router": "chi",
      "method": "POST",
      "path": "/api/orders/{id}",
      "handler": {"fqn": "...", "file": "...", "line": 0},
      "middleware_chain": [
        {"fqn": "...", "kind": "global|group|route", "file": "...", "line": 0}
      ]
    }
  ],
  "sinks_by_kind": {
    "sql_exec": [{"file": "...", "line": 0, "callee": "..."}],
    "cmd_exec": [],
    "http_client": [],
    "fs_path": [],
    "deserialize": [],
    "template_render": [],
    "xml_parse": []
  },
  "authz_primitives": [
    {"fqn": "...", "kind": "middleware|guard|decorator", "confidence": "extracted|inferred", "blocking": true}
  ],
  "oauth_locations": {
    "authorize_endpoint": {"fqn": "...", "file": "...", "line": 0},
    "token_endpoint": {...},
    "callback_handler": {...},
    "token_storage": {...},
    "refresh_path": {...}
  },
  "payment_surface": {"files": ["..."], "cluster_id": "...", "confidence": "..."},
  "vuln_deps": {
    "available": true,
    "findings": [{"osv_id": "GO-2023-...", "package": "...", "symbol": "...", "reachable": true, "call_stack": ["..."]}]
  },
  "ambiguous_nodes": [{"from": "...", "to": "...", "reason": "..."}],
  "warnings": ["unknown_router", "router_detected_but_no_routes"]
}
```

**Assertions:**

- A1. Output is valid JSON parseable as `go-index/v1`.
- A2. `schema_version` equals `"go-index/v1"`.
- A3. `entrypoints` is a (possibly empty) array; every entry has `router`, `method`, `path`, `handler.fqn`, `handler.file`, `handler.line`.
- A4. `routers_detected` is a (possibly empty) array of strings drawn from a known set: `{"net/http", "chi", "gin", "gorilla/mux", "echo", "fiber", "httprouter", "custom"}`.
- A5. If `routers_detected` is non-empty and `entrypoints` is empty, `warnings` contains `"router_detected_but_no_routes"`.
- A6. If a graph node is cited in the output, that node MUST exist in `graphify-out/graph.json` (no fabricated node IDs).
- A7. Every file path cited MUST exist in the workspace.
- A8. Every line number cited MUST be within the file's bounds.
- A9. `authz_primitives` entries with `blocking: false` are excluded from the array (only blocking primitives are listed).
- A10. The agent MUST NOT modify any file in `graphify-out/` or the source tree.
- A11. The agent invokes `Bash` only with commands matching its `allowed_commands` list.

### 3.2 Component: `go-taint-tracer` (Claude subagent)

**File:** `.claude/agents/go-taint-tracer.md`
**Purpose:** Verify whether untrusted input from one source reaches one sink along an exploitable path. One invocation per (source, sink) pair. Use for injection-class flaws in Go.

**Scope (in):** SQLi, command injection, SSRF, path traversal, unsafe deserialization (`gob`, unsafe `encoding/xml` configs, `yaml.v2.Unmarshal` of attacker data), template injection (`text/template` rendering into HTML context), XXE.

**Scope (out):** Authorization checks (use `go-authz-tracer`), business invariants (use `invariant-checker`), OAuth conformance (use `go-oauth-auditor`). OAuth taint pairs DO use this agent — the auditor produces (source, sink) pairs and dispatches them here.

**Tool allowlist:**

- `mcp__opengrep__scan_with_rule`
- `mcp__opengrep__get_ast`
- `mcp__gopls__go_references`
- `mcp__gopls__go_symbol_references`
- `mcp__gopls__go_search`
- `mcp__gopls__go_package_api`
- `mcp__gopls__go_file_context`
- `mcp__lsp__textDocument_implementation` (critical for interface dispatch)
- `mcp__lsp__callHierarchy_outgoingCalls`
- `mcp__lsp__callHierarchy_incomingCalls`
- `mcp__lsp__textDocument_definition`
- `Read`
- `Glob`

**Deliberately omitted:** `Grep` (forces use of LSP references, not text search), `Bash`, `Edit`, `Write`.

**Input contract (the prompt the orchestrator sends MUST be valid JSON of this shape):**

```json
{
  "source": {
    "file": "path/to/file.go",
    "line": 42,
    "expr": "r.URL.Query().Get(\"id\")",
    "kind": "http_query|http_form|http_body|http_header|http_cookie|grpc_arg|env|file|stdin|channel_recv"
  },
  "sink": {
    "file": "path/to/other.go",
    "line": 88,
    "expr": "db.Exec(query)",
    "kind": "sql_exec|cmd_exec|http_client|fs_path|deserialize|template_render|xml_parse|sql_query_raw|reflect_call"
  },
  "max_depth": 8,
  "semgrep_tier": "pro|intrafile|ce"
}
```

**Output schema:**

```json
{
  "verdict": "exploitable|sanitized|unreachable|ambiguous|input_mismatch",
  "confidence": "high|medium|low",
  "path": [
    {"file": "...", "line": 0, "expr": "...", "step": "source"},
    {
      "file": "...", "line": 0, "expr": "...",
      "step": "assign|call|return|sanitize|sink|iface_dispatch|chan_send|chan_recv",
      "callee": "fqn-or-null",
      "sanitizer_kind": "kind-or-null",
      "interface": "fqn-or-null",
      "implementer": "fqn-or-null"
    }
  ],
  "semgrep": {"tier": "pro|intrafile|ce", "rule_id": "...", "ran": true, "finding": false, "error": null},
  "gopls": {
    "references_calls": 0,
    "definition_calls": 0,
    "implementation_calls": 0,
    "call_hierarchy_calls": 0,
    "branches_explored": 0,
    "branches_unexplored": 0,
    "interface_fanout_max": 0,
    "goroutine_boundaries_crossed": 0
  },
  "sanitizers_unverified": [],
  "notes": "..."
}
```

**Behavior contract (numbered protocol — agent system prompt enforces this order):**

1. Verify input. Read source ±10 lines and sink ±10 lines. If expressions don't match, return `verdict="input_mismatch"`.
2. Resolve types via `go_file_context`. Identify static type of `source.expr`. Flag interface-typed flows for special handling.
3. Build Semgrep taint-mode YAML rule with source/sink/sanitizer entries from the canonical tables (see 3.2.1).
4. Run Semgrep via `mcp__opengrep__scan_with_rule` scoped to source-reachable file set (cap 50 files).
5. Interpret Semgrep result. Pro tier finding → high-confidence verdict. CE-tier negative result → weak signal, proceed to step 6 regardless.
6. gopls cross-check. References → classify each (assignment, call arg, return, channel send, interface call). For each propagating reference, recurse with depth-1. Interface calls require `textDocument_implementation` walk over concrete implementers.
7. Apply stop conditions: depth=0, 60 LSP calls, 5 unexplored branches, reflection/unsafe encountered.
8. Verdict per the truth table in component spec.
9. Never substitute Grep for symbol queries.

**3.2.1 Canonical source/sink/sanitizer tables for Go**

These live in the agent's system prompt. The full table is in section 8 of this document (References → "Go source/sink/sanitizer catalog"). The agent's system prompt copy-includes that catalog.

**3.2.2 OAuth-specific extensions**

When the source or sink kind is OAuth-specific, the tracer applies the OAuth source/sink table in addition to the baseline table. OAuth pairs are dispatched by `go-oauth-auditor` after it locates the OAuth surface. The OAuth table:

| Source | Sink | Required sanitizer |
|---|---|---|
| `r.PostFormValue("scope")` / equivalent on consent endpoint | Token scope claim assignment | Server-side lookup of original `scope` from authorization request keyed by session/auth-code |
| `r.PostFormValue("redirect_uri")` / query equivalent | HTTP redirect call, `Location` header set, callback URL resolution | Exact-match comparison against client's registered `redirect_uris` allowlist (RFC OAuth 2.1 §2.3.1) |
| `r.PostFormValue("client_id")` on token endpoint | Client record lookup leading to token issuance | Client authentication: confirm presenter holds `client_secret` (RFC 6749 §2.3.1) or has signed assertion (RFC 7521/7523) |
| `r.URL.Query().Get("state")` on callback | Equality comparison sink | Must reach `==` against session-stored state; absence of comparison is the bug |
| `r.PostFormValue("code")` on token endpoint | Authorization code lookup leading to token mint | Atomic single-use enforcement: lookup must mark code consumed in the same transaction |
| `r.PostFormValue("resource")` or `audience` | Token `aud` claim assignment | Validation against client's registered allowed resources (RFC 8707) |
| `login_hint` on CIBA backchannel endpoint | User resolution sink | Must be nonce-like OR paired with `user_code` (CIBA core §13, FAPI-CIBA) |
| `backchannel_client_notification_endpoint` (push mode) | Outbound notification URL | Exact-match allowlist against client's registered URI |
| `auth_req_id` on token endpoint | Token mint | Single-use enforcement + binding to original `client_id` |

**Assertions:**

- T1. Output is valid JSON conforming to the verdict schema.
- T2. `verdict` is one of `{exploitable, sanitized, unreachable, ambiguous, input_mismatch}`.
- T3. If `verdict ∈ {sanitized, exploitable}`, `path` is non-empty and contains `step="source"` as first entry and `step="sink"` as last entry.
- T4. If `verdict == "input_mismatch"`, `path` MAY be empty.
- T5. `confidence ∈ {high, medium, low}`.
- T6. Either `semgrep.ran == true` OR `gopls.references_calls > 0` (or both), UNLESS `verdict == "input_mismatch"`.
- T7. `semgrep.tier` matches the tier passed in input.
- T8. If the input source kind is interface-typed, `gopls.implementation_calls > 0` OR `notes` contains an explanation for why implementation walk was skipped.
- T9. Every file path and line number cited in `path` corresponds to a real location in the workspace.
- T10. `confidence == "high"` requires either Semgrep Pro/intrafile finding OR full LSP path verification with no `sanitizers_unverified` entries.
- T11. The agent MUST NOT call `Grep`, `Bash`, `Edit`, or `Write` (these are not in its allowlist; this assertion is for the hook validator).

### 3.3 Component: `go-authz-tracer` (Claude subagent)

**File:** `.claude/agents/go-authz-tracer.md`
**Purpose:** Verify every route in the application passes through recognized authorization middleware before reaching the handler, and that handlers don't internally escape the authz check.

**Tool allowlist:**

- `mcp__gopls__go_references`
- `mcp__gopls__go_symbol_references`
- `mcp__gopls__go_search`
- `mcp__lsp__callHierarchy_outgoingCalls`
- `mcp__lsp__textDocument_definition`
- `mcp__lsp__textDocument_implementation`
- `Read`
- `Glob`

**Input contract:**

```json
{
  "routes": [
    {
      "router": "chi|gin|mux|net_http|echo|fiber",
      "method": "GET|POST|...",
      "path": "/api/...",
      "handler": {"file": "...", "line": 0, "fqn": "..."},
      "middleware_chain": [{"fqn": "...", "kind": "global|group|route"}]
    }
  ],
  "authz_primitives": [{"fqn": "...", "kind": "middleware|guard|decorator"}],
  "sensitive_operations": [{"file": "...", "line": 0, "kind": "db_write|external_api|privileged_op"}]
}
```

**Output schema:**

```json
{
  "summary": {
    "routes_total": 0,
    "protected": 0,
    "missing": 0,
    "weak": 0,
    "idor_risk": 0,
    "public_intentional": 0
  },
  "findings": [
    {
      "route": "POST /api/...",
      "issue": "missing_authz|weak_authz|idor_risk|bypass_path|non_blocking_middleware",
      "evidence": {"file": "...", "line": 0, "explanation": "..."},
      "confidence": "high|medium|low"
    }
  ],
  "weak_primitives": [{"fqn": "...", "reason": "..."}]
}
```

**Behavior contract:**

1. For each route: walk middleware chain in registration order. Classify as `protected` (authz primitive present), `public_intentional` (handler or package marked public/health/metrics), or `missing_authz`.
2. For each protected route: `callHierarchy_outgoingCalls` on handler. Verify handler reads identity from request context (not from request parameters). Reading `userID := chi.URLParam(r, "userID")` and using it for a database lookup of that user's data is the IDOR pattern — flag.
3. For each authz primitive used: read its body once. Verify it has a failure-return path (no proceeding to `next.ServeHTTP` on auth failure), sets identity in request context, has no bypass mode (env-var conditional, debug flag).

**Assertions:**

- AZ1. Output JSON conforms to schema.
- AZ2. `summary.routes_total == len(input.routes)`.
- AZ3. `summary.protected + summary.missing + summary.weak + summary.idor_risk + summary.public_intentional` accounts for all routes (sum equals routes_total; a route may appear in multiple categories if it has multiple findings — in which case findings are de-duped per route in the `findings` array but counted in each summary bucket they trigger).
- AZ4. Every `route` value in `findings` corresponds to a route present in input.
- AZ5. Every `weak_primitive` referenced was present in input's `authz_primitives`.
- AZ6. The agent reads each authz primitive body at most once (cache reads in agent's working memory; this is an efficiency assertion, validated by `Read` call count not exceeding `len(input.authz_primitives) + len(input.routes)*2`).

### 3.4 Component: `go-oauth-auditor` (Claude subagent)

**File:** `.claude/agents/go-oauth-auditor.md`
**Purpose:** Conformance checking of Go OAuth/OIDC implementation against current RFCs and drafts. Produces two outputs: a checklist findings report, and a list of OAuth-specific (source, sink) pairs to be dispatched to `go-taint-tracer`.

**Tool allowlist:**

- `mcp__gopls__go_search`
- `mcp__gopls__go_references`
- `mcp__gopls__go_file_context`
- `mcp__opengrep__scan_with_rule`
- `Read`
- `Glob`

**Input contract:**

```json
{
  "oauth_locations": {
    "authorize_endpoint": {"fqn": "...", "file": "...", "line": 0},
    "token_endpoint": {...},
    "callback_handler": {...},
    "token_storage": {...},
    "refresh_path": {...}
  },
  "target_profile": "oauth_2_1|oauth_2_0|oauth_2_0_with_9700_bcp",
  "features_in_use": ["pkce", "par", "dpop", "ciba", "token_exchange", "dynamic_client_registration", "introspection", "revocation"]
}
```

**Output schema:**

```json
{
  "profile": "oauth_2_1",
  "checklist": [
    {
      "check_id": "OAUTH21-2.3.1-redirect-exact-match",
      "spec": "draft-ietf-oauth-v2-1-15 §2.3.1",
      "status": "pass|fail|not_applicable|unverified",
      "evidence": {"file": "...", "line": 0, "explanation": "..."},
      "severity": "critical|high|medium|low|info"
    }
  ],
  "taint_pairs": [
    {
      "source": {"file": "...", "line": 0, "expr": "...", "kind": "oauth_scope_param"},
      "sink": {"file": "...", "line": 0, "expr": "...", "kind": "token_scope_claim"},
      "rationale": "scope tampering — verify server re-reads from authoritative state"
    }
  ]
}
```

**Checklist taxonomy (organized by spec):**

The full checklist is in section 8.3. The auditor's system prompt embeds it. Examples:

- OAuth 2.1 / RFC 6749: PKCE present (`code_challenge` on authorize, `code_verifier` on token); redirect_uri exact-match; state parameter present and verified; implicit flow not used; ROPC not used; refresh token rotation; one-time-use authorization codes; client authentication on token endpoint.
- RFC 9700 (Security BCP): PKCE for all clients including confidential; sender-constrained tokens for high-risk APIs; mix-up defense via `iss` parameter.
- RFC 9207 (Issuer Identification): `iss` parameter returned and validated.
- RFC 9126 (PAR): authorization request via PAR if claimed; `request_uri` validation.
- RFC 7636 (PKCE): `S256` method, not `plain`; verifier entropy.
- RFC 8252 (Native Apps): claimed-https-scheme or loopback or private-use-URI; no embedded user-agents.
- draft-ietf-oauth-cross-device-security-15: phishing resistance for cross-device flows.
- RFC 9068 (JWT access tokens): correct `typ` header (`at+jwt`); audience validation.
- RFC 8707 (Resource Indicators): `resource` parameter handling; audience binding.
- RFC 9449 (DPoP): `htm`/`htu` claim validation; nonce handling; replay protection.
- RFC 8693 (Token Exchange): `subject_token` validation; `actor_token` validation; impersonation/delegation policy enforcement.
- JOSE family (RFC 7515-7519): `alg=none` rejection; alg confusion (HS vs RS); `kid` injection; `jku`/`x5u` SSRF; type confusion.
- RFC 7591/7592 (Dynamic Client Registration): authentication on registration endpoint; registered metadata validation.
- RFC 7662 (Introspection): authentication required on introspection endpoint.
- RFC 7009 (Revocation): revocation propagation; refresh-token-revokes-access-token.
- CIBA Core 1.0: backchannel endpoint authentication; `login_hint` handling; `binding_message` display path; push mode disabled or audited; `auth_req_id` single-use.
- OpenID Connect Core 1.0: `nonce` for implicit/hybrid (where used); ID token signature validation; `at_hash`/`c_hash` validation.

**Behavior contract:**

1. For each item in the checklist applicable to `target_profile` and `features_in_use`: locate the relevant code via `oauth_locations` + `go_search`, verify or report.
2. Each check returns one of: `pass`, `fail`, `not_applicable`, `unverified`. `unverified` means the auditor couldn't determine the answer (e.g., the relevant code wasn't located).
3. Each check's `evidence` cites file:line. No invented locations.
4. After the checklist pass, generate a `taint_pairs` list — every OAuth source on a located endpoint paired with the corresponding OAuth sink. The orchestrator dispatches these to `go-taint-tracer`.

**Assertions:**

- OA1. Output JSON conforms to schema.
- OA2. Every checklist entry has a `spec` reference that matches a recognized RFC or draft ID.
- OA3. No checklist entry has `status="pass"` without `evidence.file` and `evidence.line` populated (proof of where it passed).
- OA4. `taint_pairs` entries reference files and lines present in `oauth_locations`.
- OA5. If `target_profile == "oauth_2_1"` and `features_in_use` contains `"pkce"`, the PKCE-related checks are evaluated.
- OA6. If `target_profile == "oauth_2_0"` and `oauth_locations.authorize_endpoint` is set, the implicit-flow-disallowed check is evaluated (still applies; 2.0 with BCP forbids implicit).
- OA7. The auditor never claims a check passes based on the absence of the corresponding feature — that's `not_applicable`, not `pass`.

### 3.5 Component: `invariant-checker` (Claude subagent)

**File:** `.claude/agents/invariant-checker.md`
**Purpose:** Verify human-stated business-logic invariants over Go code. Used for payment/checkout/ordering flows.

**Critical design note:** This agent does NOT invent invariants. Its caller (orchestrator, or human via CLI) provides them. The agent's job is verification, not discovery. Discovery is a different problem and out of scope for the current design.

**Tool allowlist:**

- `mcp__gopls__go_search`
- `mcp__gopls__go_references`
- `mcp__gopls__go_file_context`
- `mcp__lsp__callHierarchy_outgoingCalls`
- `mcp__lsp__textDocument_definition`
- `Read`
- `Glob`

**Input contract:**

```json
{
  "flow_name": "checkout",
  "invariants": [
    {
      "id": "checkout-server-price-authority",
      "statement": "The value passed to payment_gateway.charge(amount) MUST equal sum(server_lookup_price(item.sku) for item in cart). Client-submitted prices MUST NOT influence the charged amount.",
      "anchor_symbols": ["payment_gateway.charge", "server_lookup_price"]
    }
  ]
}
```

**Output schema:**

```json
{
  "flow_name": "checkout",
  "results": [
    {
      "invariant_id": "checkout-server-price-authority",
      "status": "holds|violated|unverifiable",
      "evidence": {"files": ["..."], "explanation": "..."},
      "confidence": "high|medium|low"
    }
  ]
}
```

**Behavior contract:**

1. For each invariant: locate the `anchor_symbols` via `go_search`. If any cannot be located, mark invariant `unverifiable` with reason.
2. Read the body of the function containing each anchor symbol. Trace dependencies via `callHierarchy_outgoingCalls` and `textDocument_definition`.
3. Verify the invariant statement against the code path. Status:
   - `holds` — invariant is satisfied; explain why.
   - `violated` — invariant is violated; cite the specific path that violates it.
   - `unverifiable` — anchors located but the invariant requires runtime semantics the agent cannot determine (e.g., correctness depends on data values not visible statically).
4. Output is conservative: `unverifiable`eferred over false `holds`.

**Assertions:**

- IC1. Output JSON conforms to schema.
- IC2. `results` has one entry per input invariant; `invariant_id` values match input.
- IC3. `status == "violated"` requires `evidence.files` to be non-empty AND `evidence.explanation` to cite a specific code path.
- IC4. The agent does not generate invariants; it only verifies the input list.

### 3.6 Component: `synthesis` (Claude subagent)

**File:** `.claude/agents/synthesis.md`
**Purpose:** After all specialist agents complete, consume their outputs, deduplicate findings, rank by severity, and produce a final review report.

**Tool allowlist:**

- `Read`
- `Glob`

**Input contract:** Path to a directory containing all specialist outputs (one JSON file per Task invocation).

**Output:** Two files — `review-report.json` (machine-readable, conforms to schema below) and `review-report.md` (human-readable summary).

**Output schema (`review-report.json`):**

```json
{
  "review_id": "uuid",
  "timestamp": "2026-...",
  "scma_version": "review-report/v1",
  "summary": {
    "total_findings": 0,
    "by_severity": {"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 0},
    "by_class": {"injection": 0, "authz": 0, "oauth": 0, "invariant": 0, "vuln_dep": 0}
  },
  "findings": [
    {
      "id": "finding-001",
      "class": "injection|authz|oauth|invariant|vuln_dep",
      "severity": "critical|high|medium|low|info",
      "title": "...",
      "description": "...",
      "evidence": {"files": ["..."], "lines": [0], "path": [...]},
      "source_agents": ["go-taint-tracer", "go-oauth-auditor"],
      "confidence": "high|medium|low",
      "spec_references": ["RFC ...", "..."]
    }
  ],
  "deduplication_notes": ["finding-001 and finding-007 merged: same source/sink, different tracers"]
}
```

**Behavior contract:**

1. Read all JSON files in input directory.
2. Validate each against the expected schema for that agent.
3. Deduplicate: findings that share (file, line, class) are merged. Source agents are listed in `source_agents`; the highest confidence wins.
4. Rank by severity (critical > high > medium > low > info). Within severity, rank by class (oauth and authz top, then injection, then invariant, then vuln_dep).
5. Generate Markdown report grouped by severity.

**Assertions:**

- S1. Output is two files: `review-report.json` and `review-report.md`.
- S2. JSON conforms to `review-report/v1` schema.
- S3. `summary.total_findings == len(findings)`.
- S4. Severity counts in `summary.by_severity` sum to `total_findings`.
- S5. Every finding has at least one `source_agent`.
- S6. No two findings share `(file, line, class)` without being explicitly noted as cross-references in `deduplication_notes`.

### 3.7 Component: `claude-security-hooks` (Go binary)

**Module path (recommended):** `github.com/sagjaulor/claude-security-hooks` — adjust for your actual GitHub username/org.
**Binary location:** `.claude/hooks/bin/claude-security-hooks`
**Purpose:** Lifecycle hooks that validate security agent invocations and outputs. Single biny with subcommands.

**Subcommands:**

- `preflight` — invoked on `PreToolUse` for `Task` tool calls. Validates that security agents are dispatched with proper structured input.
- `validate` — invoked on `PostToolUse` for `Task` tool calls. Validates security agents' output JSON against schema and invariants.
- `inject-context` — invoked on `SubagentStart` for security agents. Optional: injects deterministic context (Semgrep tier, schema versions).

**Build target:** Static binary, no CGO, distroless-ly.

```go
// go.mod minimum
module github.com/sagjaulor/claude-security-hooks
go 1.22

// Recommended dependencies (minimal):
//   encoding/json (stdlib)
//   fmt (stdlib)
//   os (stdlib)
//   No external deps required for the core validator.
```

**Project layout (recommended):**

```
claude-security-hooks/
├── cmd/
│   └── claude-security-hooks/
│       └── main.go             # subcommand dispatch
├── internal/
│   ├── hooks/
│   │   ├── preflight.go
│   │   ├── validate.go
│   │   └── inject.go
│   ├── schema/
│   │   ├── taint_tracer.go     # Go structs mirroring the verdict schema
│   │   ├── cartographer.go
│   │   ├── authz_tracer.go
│   │   ├── oauth_auditor.go
│   │   ├── invariant_checker.go
│   │   └── synthesis.go
│   └── invariants/
│       ├── registry.go         # per-agent invariant definitions
│       ├── taint_tracer.go     # T1-T11     ├── cartographer.go     # A1-A11
│       └── ...
├── Makefile
├── Dockerfile                  # optional; binary is the artifact
├── go.mod
└── README.md
```

**`main.go` shape:**

```go
package main

import (
    "fmt"
    "os"

    "github.com/sagjaulor/claude-security-hooks/internal/hooks"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "usage: claude-security-hooks <preflight|validate|inject-context>")
        os.Exit(2)
    }
    var ror
    switch os.Args[1] {
    case "preflight":
        err = hooks.Preflight(os.Stdin, os.Stdout, os.Stderr)
    case "validate":
        err = hooks.Validate(os.Stdin, os.Stdout, os.Stderr)
    case "inject-context":
        err = hooks.InjectContext(os.Stdin, os.Stdout, os.Stderr)
    default:
        fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", os.Args[1])
        os.Exit(2)
    }
    if err != nil {
        // Errors that are NOT block decisions go to stderr and exit 1.
        // Block decisions are emitted as JSON to stdout with exit 0.
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**Validate subcommand logic:**

The validate function:
1. Reads PostToolUse JSON payload from stdin. The payload shape is documented at https://code.claude.com/docs/en/hooks.
2. Checks `tool_input.subagent_type` — if not in security agent set, exits 0 silently.
3. Extracts `tool_response.content` as the subagent's final message.
4. Strips optional code-fence wrapping.
5. Parses as JSON into t appropriate Go struct based on `subagent_type`.
6. Runs the invariant list for that agent.
7. On any invariant failure, emits `{"decision":"block","reason":"..."}` to stdout, exits 0. (Per Claude Code hooks spec: PostToolUse block surfaces the reason to Claude as a tool error, prompting retry.)
8. On all invariants passing, exits 0 with no output.

**Settings.json hook registration:**

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Task",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/bin/claude-security-hooks preflight",
            "timeout": 5
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Task",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/bin/claude-security-hooks validate",
            "timeout": 15
          }
        ]
      }
    ],
    "SubagentStart": [
      {
        "matcher": "go-taint-tracer|go-authz-tracer|go-oauth-auditor|go-cartographer|invariant-checker|synthesis",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/bin/claude-security-hooks inject-context",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

**Why `SubagentStop` is NOT used:**

`SubagentStop` payload does not include `agent_name` or `subagent_type`. With parallel subagents, the hook cannot route the validation. See GitHub issue anthropics/claude-code#7881. Hooking the `Task` tool's `PostToolUse` gives per-invocation routing and full access to `tool_input` and `tool_response`. This is the better mechanism.

**Assertions:**

- H1. `claude-security-hooks` is a single static binary buildable with `CGO_ENABLED=0 go build -o bin/claude-security-hooks ./cmd/claude-security-hooks`.
- H2. The binary's three subcommands accept JSON payloads on stdin matching the documented Claude Code hook event schemas.
- H3. On no-op cases (non-security agent), the binary exits 0 silently with no stdout output.
- H4. On block decisions, the binary writes a single JSON object `{"decision":"block","reason":"..."}` to stdout and exits 0.
- H5. The binary has zero external (non-stdlib) Go dependencies for the core validator path. (Test framework dependencies allowed in `_test.go` files.)
- H6. Every per-agent invariant from sections 3.1–3.6 (A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6) is implemented as a check in `internal/invariants/<agent>.go` and exercised by a unit test.
- H7. The validator handleON in `tool_response.content` by emitting a block decision with a parse-error reason, not by panicking.

### 3.8 Component: `opengrep-mcp` (Go binary, MCP server)

**Module path (recommended):** `github.com/sagjaulor/opengrep-mcp`
**Purpose:** Single MCP server fronting both Semgrep Pro and OpenGrep. Agents call one tool with a tier parameter; the server dispatches to the right container.

**SDK:** Use the official Go MCP SDK at `github.com/modelcontextprotocol/go-sdk`. As of December 2025 this is maintained in collaboration with Google. See https://github.com/modelcontextprotocol/go-sdk.

**Container strategy:** Use the Docker Engine SDK for Go (`github.com/docker/docker/client`) to invoke containers. Avoid shelling out to `docker` CLI. This gives typed errors, structured output, and avoids PATH dependency.

**Tools exposed:**

- `scan_with_rule(rule_yaml: string, paths: []string, tier: "pro"|"intrafile"|"ce", timeout_seconds: int) -> findings`
- `scan_directory(rule_dir: string, paths: []string, tier: ...) -> findings`
- `get_ast(file_path: string, language: string) -> ast_json`

**Behavior contract:**

1. On startup, verify Docker daemon reachable.
2. For each tool call, pull the appropriate container image if not cached locally.
3. Mount the workspace read-only at `/src` inside the container.
4. Write the rule YAML to a temp file, mount read-only at `/tmp/rule.yaml`.
5. Run the scanner. Stream stderr to the MCP server's log; parse stdout JSON.
6. Return structured findings.
7. On timeout, kill the container and return a partial-results error.

**Container image plan:**

- `pro`: official Semgrep image with Pro license env var (user provides via MCP server config).
- `intrafile`: a self-built `opengrep` image. Dockerfile in this repo (see component 3.10).
- `ce`: official Semgrep CE image.

**Project layout:**

```
opengrep-mcp/
├── cmd/opengrep-mcp/main.go     # MCP server entrypoint
├── internal/
│   ├── server/server.go         # MCP tool registration
│   ├── runner/
│   │   ├── runner.go            # Docker invocation
│   │   ├── parser.go            # Semgrep/OpenGrep output parsing
│   │   └── tier.go              # tier → image mapping
│   └── schema/findings.go       # normalized finding type
├── docker/
│   └── opengrep/Dockerfile      # OpenGrep image build
├── go.mod
├── Makefile
└── README.md
```

**Assertions:**

- O1. The MCP server binary builds with `CGO_ENABLED=0 go build`.
- O2. The server registthree tools: `scan_with_rule`, `scan_directory`, `get_ast`. Tool schemas validate against MCP JSON Schema spec.
- O3. `scan_with_rule` rejects requests where `tier` is not one of `{pro, intrafile, ce}` with a structured error.
- O4. The server pulls container images on demand and caches them.
- O5. The server invokes containers with the workspace mounted **read-only**.
- O6. The server kills the container on timeout exceedance.
- O7. The server's findings output uses a normalized schema (defined in `internal/schema/findings.go`) regardless of which scanner ran.
- O8. The server passes through Semgrep Pro license credentials only via environment variable at server startup, never logged.

### 3.9 Component: Container infrastructure

**Dockerfile for `opengrep-mcp` MCP server itself (optional — the binary can run unwrapped):**

```dockerfile
FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /opengrep-mcp ./cmd/opengrep-mcp

FROM gcr.io/distroless/static-debian12
COPY --from=build /ongrep-mcp /opengrep-mcp
ENTRYPOINT ["/opengrep-mcp"]
```

**Dockerfile for OpenGrep scanner image (in `opengrep-mcp/docker/opengrep/Dockerfile`):**

```dockerfile
# Build OpenGrep from source per the upstream README at github.com/opengrep/opengrep.
# Pin to a specific release tag for reproducibility.
FROM ocaml/opam:debian-12-ocaml-5.1 AS build
ARG OPENGREP_VERSION=v1.16.0
RUN git clone --depth 1 --branch ${OPENGREP_VERSION} https://github.com/opengrep/opengrep /src
WORKDIR /src
# Follow the build instructions in the upstream README.
# (Implementing session: verify the exact build commands against the upstream
#  README at the chosen release tag.)

FROM gcr.io/distroless/cc-debian12
COPY --from=build /src/_build/default/src/main/Main.exe /usr/local/bin/opengrep
ENTRYPOINT ["/usr/local/bin/opengrep"]
```

**Note for implementing session:** The exact OpenGrep build command is not pinned here because the upstream README is the source of truth and may change. Read https://github.com/opengrep/opengrep at the release tag you choose and replicate the documented build.

**Semgrep Pro container:** Use the official `semgrep/semgrep` image. Pass the Pro license via `SEMGREP_APP_TOKEN` env var (provisioned to `opengrep-mcp` via its own startup config, then forwarded to the container).

### 3.10 Component: Agent definition files (markdown)

All agent definitions live in `.claude/agents/`. Each file is a markdown document with YAML frontmatter. The frontmatter declares `name`, `description`, `model`, `tools`, and (optionally) `allowed_commands`. The body is the agent's system prompt.

**Default model:** Sonnet (current latest stable Sonnet variant available in the Claude Code session). The user noted earlier that Opus may be overkill for parallel tracer dispatch. The implementing session should set `model: claude-sonnet-4-6` (or whichever is current at implementation time — verify against Claude Code's product docs) and let the user override per-agent if needed.

**Common structure of each agent's system prompt:**

1. Role stement (one paragraph).
2. Input contract (JSON shape).
3. Numbered protocol (mandatory order of operations).
4. Reference tables (sources, sinks, sanitizers, RFC checklist — whichever applies).
5. Output schema (JSON shape).
6. Hard rules (negative guidance, ambiguity preference, no-invent rule).

The full text of each agent's system prompt is derived from the corresponding section in this handoff. The implementing session should generate each agent's `.md` file by combining the section's contract + behaor + assertions into the format above.

---

## 4. Decisions log

This section captures key choices and why, so the implementing session and future maintainers understand the rationale.

| # | Decision | Rationale |
|---|---|---|
| D1 | Agent definitions are per-stack, not parameterized | Go source/sink/sanitizer tables are large enough that combining stacks blows up prompt size and degrades accuracy. Per-stack files are easier to maintain. |
| D2 | Hooks are written in Go, not Python | Cold-start latency. Python interpreter startup compounds across parallel fan-outs. Static Go binary starts in <5ms. Aligns with user preference for compiled binaries. |
| D3 | Semgrep and OpenGrep are containerized | User preference. Reduces host pollution, simplifies CI parity. |
| D4 | Single `opengrep-mcp` server fronts both engines via tier parameter | Simpler agent prompt (one tool, not two). Engine selection becomes a runtime decision. |
| D5 | `Task` tool's PostToolUse hook validates outputs, NOT `SubagentStop` | `SubagentStop` payload doesn't identify which subagent finished when multiple run in parallel. `Task` tool hooks give per-invocation routing. See issue anthropics/claude-code#7881. |
| D6 | OAuth bug classes are split across an auditor (checklist) and the taint tracer (flow analysis) | Scope-tampering bugs are data-flow bugs, not checklist items. The auditor finds OAuth endpoints; the tracer verifies the flows. |
| D7 | RFC 6819 dropped from the audit checklist | Superseded by RFC 9700. Per user note. |
| D8 | OAuth 2.1 draft (currently draft-15) included as a target profile | Latest IETF draft, March 2026. Obsoletes 6749/6750 once finalized. Auditor handles both regimes via `target_profile` input. |
| D9 | CIBA included as part of OIDC family | User explicitly requested. CIBA introduces decoupled-flow attack surface not covered by baseline OAuth. |
| D10 | gosec and staticcheck dropped from pre-pass | Per user: redundant with custom Semgrep rules. **govulncheck kept** — it's an SSA-based reachability check again the Go vuln DB, not pattern matching, and cannot be replicated by Semgrep rules. |
| D11 | Cartographer runs sequentially before tracers, tracers run in parallel | Cartographer output is small and consumed by every tracer; redundant per-tracer discovery would waste tokens. Tracers are independent and parallelizable. |
| D12 | Tool allowlists deliberately omit cheap alternatives | If `Grep` is in the allowlist, the agent will use it instead of LSP `references`. Removing `Grep` from tracer allowlists is the cheapest enforcement mechanism. |
| D13 | Verdict schemas are strict JSON, prose disallowed | The orchestrator and synthesis agent must parse outputs mechanically. Prose verdicts force English parsing, which scales badly and breaks deduplication. |
| D14 | "Ambiguous" verdict is encouraged over false confidence | Bias toward honest uncertainty. False sanitized verdicts are worse than honest ambiguous ones — they create blind spots. |
| D15 | Skills not used for the security review skill | Skills load intohe main agent. Subagents start with empty context. The earlier failure mode (instructions disappearing across context boundary) is exactly the bug this design avoids. The orchestrator is a slash command or a manually-invoked workflow, not a skill. |
| D16 | Invariant checker does not generate invariants | Discovery is a different (harder) problem. The current design verifies user-stated invariants. Discovery can be a future agent. |
| D17 | Per-stack expansion is sequential, Go first | User confirmed: ship Go well first, then port. Building cross-stack abstraction before having one stack working produces leaky abstractions. The shared layer is the JSON schemas and hook validators; agents themselves remain per-stack. |
| D18 | Synthesis agent reads from a directory of JSON files | Decouples synthesis from in-flight Task results. The orchestrator dumps each subagent's verdict to disk; synthesis reads the whole directory at the end. Simpler than streaming state. |

---

## 5. Open questions

These items require resolution by the implementing session, or by the human user, during build. They are NOT design gaps — they're integration details that depend on environment specifics.

| # | Question | Recommended path |
|---|---|---|
| Q1 | Exact Graphify MCP tool schemas and parameter names | Verify by running `python -m graphify.serve --help` and `python -m graphify.serve graphify-out/graph.json` then probing the MCP handshake. If the tool names differ from `{query_graph, get_node, get_neighbors, shortest_path}`, upde the cartographer's allowlist. |
| Q2 | Whether the user's target Go services use a custom router wrapper | The cartographer has a fallback heuristic for unknown routers. If wrapper patterns are common, add explicit Semgrep patterns for them. The user said this is "most likely standard"; verify on first real review. |
| Q3 | Semgrep Pro license provisioning to containers | User has Semgrep Pro. Provision via `SEMGREP_APP_TOKEN` env var to the `opengrep-mcp` server at startup. The server forwards it to the Pro container. Document this in the README. |
| Q4 | OpenGrep build command for the Dockerfile | The Dockerfile in 3.9 is a sketch. Implementing session: pin a release tag and verify build commands against the upstream README at https://github.com/opengrep/opengrep. |
| Q5 | Whether to ship synthesis as JSON-only or also Markdown | Designed for both. Implementing session may start with JSON-only if Markdown rendering is non-trivial; defer Markdown to a follow-up. |
| Q6 | Concrete Semgrep rule shapes for OAuth audit checks | Each checklist item maps to a Semgrep pattern OR a code-read verification. Implementing session: write the rules iteratively, starting with the high-severity items (PKCE, redirect_uri exact match, state validation, scope tampering). The full checklist is large; ship a v1 with ~15 critical checks and grow. |
| Q7 | Whether to add an orchestrator agent or use a slash command | The orchestrator pattern (cartographer → fan-out → synthesis) can be a slash command (`/security-review`) that kes the agents in sequence via Task calls. Alternatively, an `orchestrator` agent could automate this. Recommend slash command for v1 — simpler, more transparent, easier to debug. |
| Q8 | Whether `invariant-checker` invariants are stored in a file or passed inline | Both are valid. Recommend a `.claude/security-invariants/<flow>.yaml` directory pattern for declared invariants, loaded by the orchestrator and passed inline to the agent. |
| Q9 | The `model:` field default per agent | The user prefers efficncy. Default to `claude-sonnet-4-6` for all agents at implementation time, with a note that the user can override to Opus on a per-agent basis if tracer accuracy proves insufficient in practice. The implementing session must verify this model string is current — check https://docs.claude.com for the latest model identifiers. |
| Q10 | CI integration | Not specified in design. Each agent's outputs are JSON to a directory; the synthesis produces machine-readable results. A future task can wire this to GitHuActions. Out of scope for v1. |

---

## 6. Implementation plan

Recommended sequencing for the Claude Code session. Each phase produces something committable.

### Phase 1: Repo scaffold

- Create the GitHub repo.
- Initialize Go module for `claude-security-hooks`.
- Initialize Go module for `opengrep-mcp` (separate module — these are distinct binaries with different dependency profiles).
- Create `.claude/` directory tree: `.claude/agents/`, `.claude/hooks/bin/`, `.claude/security-invariants/`.
- Add `.tignore`, `LICENSE` (the user's choice), `README.md` at repo root with a brief description of the system.
- Commit: "scaffold: initial project structure"

### Phase 2: claude-security-hooks binary

- Implement `cmd/claude-security-hooks/main.go` with subcommand dispatch.
- Implement schema types in `internal/schema/` for all six agents.
- Implement `internal/invariants/` with the assertion lists from sections 3.1–3.6 as Go predicates.
- Implement `internal/hooks/preflight.go`, `validate.go`, `inject.go`.
Write unit tests for every invariant (one test case per assertion).
- Add `Makefile` with `build`, `test`, `install` targets. `install` copies the binary to `.claude/hooks/bin/`.
- Commit: "feat(hooks): claude-security-hooks binary with per-agent validation"

### Phase 3: Agent definitions

- Write `.claude/agents/go-cartographer.md` per spec 3.1.
- Write `.claude/agents/go-taint-tracer.md` per spec 3.2, including the full Go source/sink/sanitizer catalog (section 8.2 of this handoff) and the OAuth source/sink table (section 3.2.2).
- Write `.claude/agents/go-authz-tracer.md` per spec 3.3.
- Write `.claude/agents/go-oauth-auditor.md` per spec 3.4, including the RFC checklist taxonomy (section 8.3).
- Write `.claude/agents/invariant-checker.md` per spec 3.5.
- Write `.claude/agents/synthesis.md` per spec 3.6.
- Add `.claude/settings.json` registering the hooks per spec 3.7.
- Commit: "feat(agents): security review agent definitions"

### Phase 4: opengrep-mcp server

- Implement MCP server skeleton using `github.com/modelcontextprotocol/go-sdk`.
- Implement Docker runner using `github.com/docker/docker/client`.
- Implement `scan_with_rule`, `scan_directory`, `get_ast` tools.
- Build OpenGrep container image (Dockerfile in `opengrep-mcp/docker/opengrep/`).
- Write integration tests that invoke the server end-to-end with a simple Semgrep CE rule.
- Document the MCP server in its own README, including how to register it with Claude Code.
- Commit: "feat(mcp): opengrep-mcp dual-engine scanner MCP server"

### Phase 5: Smoke test

- Build a small test Go service (in `examples/sample-vulnerable-service/`) with one obvious SQLi bug, one authz bypass, and one OAuth scope-tampering bug.
- Run Graphify against it: `graphify build examples/sample-vulnerable-service`.
- Run the full review workflow against it via a slash command or manual orchestration.
- Verify the system catches all three bugs.
- Commit: "test: end-to-end smoke test with sample vulnerable service"

### Phase 6: Documentation

- Top-level README explaining the system architecture (link to this handoff document).
- Per-component READMEs (one in each of `claude-security-hooks/`, `opengrep-mcp/`, `.claude/`).
- A `CONTRIBUTING.md` explaining how to add a new agent or extend the source/sink catalog.
- Commit: "docs: README and contributor docs"

---

## 7. References

### 7.1 OAuth / OIDC specifications

**Active / current (as of May 2026):**

- OAuth 2.1 draft: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1 (latest at time of writing: draft-15, March 2026)
- RFC 6749 — OAuth 2.0 Authorization Framework: https://www.rfc-editor.org/rfc/rfc6749 (legacy; obsoleted by OAuth 2.1 once finalized)
- RFC 6750 — Bearer Token Usage: https://www.rfc-editor.org/rfc/rfc6750 (legacy; obsoleted by OAuth 2.1)
- RFC 9700 — OAuth 2.0 Security Best Current Practice: https://www.rfc-editor.org/rfc/rfc9700
- draft-ietf-oauth-security-topics-update — Updates to OAuth 2.0 Security BCP: https://datatracker.ietf.org/doc/draft-ietf-oauth-security-topics-update/

**Code flning:**

- RFC 7636 — PKCE: https://www.rfc-editor.org/rfc/rfc7636
- RFC 9126 — Pushed Authorization Requests (PAR): https://www.rfc-editor.org/rfc/rfc9126
- RFC 9207 — Issuer Identification: https://www.rfc-editor.org/rfc/rfc9207
- RFC 9101 — JWT-Secured Authorization Request (JAR): https://www.rfc-editor.org/rfc/rfc9101

**Native / cross-device:**

- RFC 8252 — OAuth 2.0 for Native Apps: https://www.rfc-editor.org/rfc/rfc8252
- draft-ietf-oauth-cross-device-security: https://datatracker.ietf.org-ietf-oauth-cross-device-security/

**AS metadata and discovery:**

- RFC 8414 — Authorization Server Metadata: https://www.rfc-editor.org/rfc/rfc8414

**Token shape / sender-constraint / audience:**

- RFC 9068 — JWT Profile for Access Tokens: https://www.rfc-editor.org/rfc/rfc9068
- RFC 8707 — Resource Indicators: https://www.rfc-editor.org/rfc/rfc8707
- RFC 9449 — DPoP: https://www.rfc-editor.org/rfc/rfc9449
- RFC 8705 — OAuth 2.0 Mutual-TLS: https://www.rfc-editor.org/rfc/rfc8705

**Token exchertion:**

- RFC 8693 — OAuth 2.0 Token Exchange: https://www.rfc-editor.org/rfc/rfc8693 (provided in project files)
- RFC 7521 — Assertion Framework: https://www.rfc-editor.org/rfc/rfc7521
- RFC 7523 — JWT Bearer Token Profile: https://www.rfc-editor.org/rfc/rfc7523

**Token operations:**

- RFC 7662 — Token Introspection: https://www.rfc-editor.org/rfc/rfc7662
- RFC 7009 — Token Revocation: https://www.rfc-editor.org/rfc/rfc7009

**Dynamic client management:**

- RFC 7591 — Dynamic Client Regitps://www.rfc-editor.org/rfc/rfc7591
- RFC 7592 — Dynamic Client Registration Management: https://www.rfc-editor.org/rfc/rfc7592

**JOSE:**

- RFC 7515 — JSON Web Signature (JWS): https://www.rfc-editor.org/rfc/rfc7515
- RFC 7516 — JSON Web Encryption (JWE): https://www.rfc-editor.org/rfc/rfc7516
- RFC 7517 — JSON Web Key (JWK): https://www.rfc-editor.org/rfc/rfc7517
- RFC 7518 — JSON Web Algorithms (JWA): https://www.rfc-editor.org/rfc/rfc7518
- RFC 7519 — JSON Web Token (JWT): https://www.rfc-fc/rfc7519

**OpenID Connect:**

- OpenID Connect Core 1.0: https://openid.net/specs/openid-connect-core-1_0.html
- OpenID Connect Discovery 1.0: https://openid.net/specs/openid-connect-discovery-1_0.html
- CIBA Core 1.0 (final): https://openid.net/specs/openid-client-initiated-backchannel-authentication-core-1_0.html
- FAPI CIBA Profile (optional, financial-grade): https://openid.net/specs/openid-financial-api-ciba.html

### 7.2 Tooling documentation

**Claude Code:**

- Hooks reference: https://code.claude.com/docs/en/hooks
- Subagents: https://docs.claude.com/en/docs/claude-code/sub-agents
- Settings reference: https://docs.claude.com/en/docs/claude-code/settings
- Plugin / skill model: https://docs.claude.com (search "skills" and "agents")
- Hook event payloads — `PreToolUse`, `PostToolUse`, `SubagentStart`, `SubagentStop`, `Stop`: documented in the hooks reference above

**Known limitation:**
- `SubagentStop` payload lacks subagent identification with parallel subagents: https://github.com/anthropics/cude-code/issues/7881

**Graphify:**

- Repo: https://github.com/safishamsi/graphify
- Concept article: https://blog.gopenai.com/graphify-build-a-knowledge-graph-from-your-entire-codebase-without-sending-your-code-to-anyone-1b6924474b50

**Semgrep:**

- Semgrep documentation: https://semgrep.dev/docs/
- Semgrep taint mode: https://semgrep.dev/docs/writing-rules/data-flow/taint-mode
- Semgrep MCP repo (archived; migrated into main binary): https://github.com/semgrep/mcp
- Semgrep Pro features: https://semgrep.dev/products/semgrep-code

**OpenGrep:**

- Repo: https://github.com/opengrep/opengrep
- Rules: https://github.com/opengrep/opengrep-rules
- Intrafile taint tutorial: search the repo's docs for "Intrafile Tainting Tutorial"

**gopls / Go LSP:**

- gopls overview: https://pkg.go.dev/golang.org/x/tools/gopls
- gopls MCP server (built-in v0.20.0+): https://github.com/golang/tools/tree/master/gopls/internal/mcp
- Third-party `mcp-gopls` (richer LSP surface): https://github.com/hloiseaufcms/mcp-gopls

**govulncheck:**

- Documentation: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
- Go vulnerability database: https://pkg.go.dev/vuln/

**Model Context Protocol:**

- Specification: https://modelcontextprotocol.io/
- Go SDK (maintained with Google): https://github.com/modelcontextprotocol/go-sdk
- Reference servers: https://github.com/modelcontextprotocol/servers

**Docker Engine SDK for Go:**

- Documentation: https://pkg.go.dev/github.com/docker/docker/client

### 7.3 Go routing libraries (for cartographer router patterns)

- `net/http` stdlib: https://pkg.go.dev/net/http
- chi: https://github.com/go-chi/chi
- gin: https://github.com/gin-gonic/gin
- gorilla/mux: https://github.com/gorilla/mux
- echo: https://github.com/labstack/echo
- fiber: https://github.com/gofiber/fiber
- httprouter: https://github.com/julienschmidt/httprouter

### 7.4 Project file from this conversation

- `RFC_8693_OAuth_2_0_Token_Exchange.pdf` — present in the project files. Token Exchange spec.

---

## 8. Appendices

### 8.1 Hook pload shapes (reference)

`PreToolUse` payload (relevant fields):

```json
{
  "session_id": "...",
  "transcript_path": "...",
  "hook_event_name": "PreToolUse",
  "tool_name": "Task",
  "tool_input": {
    "subagent_type": "go-taint-tracer",
    "prompt": "...",
    "description": "..."
  }
}
```

`PostToolUse` payload (relevant fields):

```json
{
  "session_id": "...",
  "transcript_path": "...",
  "hook_event_name": "PostToolUse",
  "tool_name": "Task",
  "tool_input": {
    "subagent_type": "go-taint-tracer",
    "prompt": "...",
    "description": "..."
  },
  "tool_response": {
    "content": "<the subagent's final message text>"
  }
}
```

`SubagentStart` payload (relevant fields):

```json
{
  "session_id": "...",
  "transcript_path": "...",
  "hook_event_name": "SubagentStart",
  "agent_name": "go-taint-tracer"
}
```

Implementing session: verify these field names against the current Claude Code hooks docs at https://code.claude.com/docs/en/hooks. The schema may have evolved.

### 8.2 Go source/sink/sanitizer catalog (full)

This is the canonical table that ships in `go-taint-tracer`'s system prompt. Reproduced here as the authoritative reference.

**Sources (per kind):**

| Kind | Canonical Go expressions |
|---|---|
| `http_query` | `r.URL.Query().Get(_)`, `r.URL.Query()[_]`, `gin.Context.Query`, `echo.Context.QueryParam`, `chi.URLParam`, `mux.Vars(r)[_]` |
| `http_form` | `r.FormValue(_)`, `r.PostFormValue(_)`, `r.MultipartForm`, `gin.Context.PostForm`, `echo.Context.FormValue` |
| `http_body` | `io.ReadAll(r.Body)`, `json.NewDecoder(r.Body).Decode(_)`, `gin.Context.Bind*`, `echo.Context.Bind`, `c.ShouldBindJSON` |
| `http_header` | `r.Header.Get(_)`, `r.Header[_]`, `gin.Context.GetHeader` |
| `http_cookie` | `r.Cookie(_).Value`, `r.Cookies()` |
| `grpc_arg` | Any field of a generated `*Request` struct in an RPC handler |
| `env` | `os.Getenv(_)`, `os.LookupEnv(_)` |
| `file` | `os.ReadFile`, `io.ReadAll` on `os.Open` result |
| `stdin` | `bufio.NewReader(os.Stdin)` |
| `channel_recv` | `<-ch` where `ch` was tainted |

**Sinks (per kind):**

| Kind | Canonical Go expressions |
|---|---|
| `sql_exec` | `db.Exec(query, ...)`, `db.Query(query, ...)` where `query` is built via concat/`Sprintf`/`+` rather than placeholders. `sqlx`, `gorm db.Raw(query)`, `gorm db.Where("col = " + tainted, ...)` |
| `sql_query_raw` | `gorm` `Raw`/`Exec` with non-placeholder substitution, `sqlx.NamedExec` with hand-built string |
| `cmd_exec` | `exec.Command(name, args...)` where `name` is tainted, `exec.CommandContext` same, `os/exec.LookPath` with taint |
| `http_client` | `http.Get/Post/Do` where URL is tainted without allowlist, `net/http.NewRequestWithContext` same |
| `fs_path` | `os.Open`, `os.ReadFile`, `os.Create`, `filepath.Join` when result passed to file op, `http.ServeFile` with tainted path |
| `deserialize` | `gob.NewDecoder(_).Decode`, `encoding/xml` `Unmarshal`/`Decode` without security flags, `gopkg.in/yaml.v2.Unmarshal` |
| `template_render` | `text/template.Execute` with user data rendered into HTML context |
| `xml_parse` | `encoding/xml.NewDecoder` without security flags set |
| `reflect_call` | `reflect.Value.Call`, `reflect.ValueOf(_).MethodByName(_).Call` |

**Sanitizers (per sink kind; trust only these):**

| Sink kind | Recognized sanitizers (Go) |
|---|---|
| SQL | `database/sql` with placeholders (`$1`, `?`); `sqlx.NamedExec` with bound `:name`; `gorm` `Where("col = ?", val)`, `First(&obj, id)`, `Model().Updates(map)`; `ent`, `sqlc`, `sqlboiler` generated query methods |
| Cmd | `exec.Command(constName, tainted...)` where `constName` is a constant string AND tainted is passed as a separate argv element (not joined); allowlist match against constant `[]string` before exec |
| SSRF | After `url.Parse`, hostname checked against constant `[]string` allowlist; OR resolved via `net.LookupIP` then each IP checked against private-range blocklist (with TOCTOU note); same resolved IP used for the actual request |
| Path | `filepath.Clean` THEN `strings.HasPrefix(cleaned, baseDir+string(filepath.Separator))` where `baseDir` is constant; OR `filepath.Rel` with result-not-starting-with-`..` check |
| Deserialize | `json.Unmarshal` into strongly-typed struct; `encoding/xml` with `d.Strict = true` AND `DisallowUnknownFields`; `yaml.v3` with strict types. **NEVER**: `gob` with attacker data; `yaml.v2.Unmarshal` with attacker data |
| Template | `html/template` (auto-escapes for HTML context). **NEVER**: `text/template` rendered into HTML response |

**Rule:** If a function looks like a sanitizer but its body cannot be read (third-party, native, generated), classify as `unverified`. Do not credit.

### 8.3 OAuth auditor checklist (taxonomy with check IDs)

The auditor's system prompt embeds this list. Implementing session: each check needs a concrete verification strategy (Semgrep pattern or code-read + LSP). The strategies can be developed iteratively. The check IDs are the stable identifiers; findings reference them.

Format: `<check_id> | <spec ref> | <severity> | <one-line description>`

```
OAUTH21-2.3.1-redirect-exact-match     | OAuth 2.1 §2.3.1                       | critical | redirect_uri compared by exact string match
OAUTH21-2.3.3-csrf-state               | OAuth 2.1 §2.3.3                       | critical | state parameter present in authz request, validated on callback
OAUTH21-2.3.4-mix-up-iss               | OAuth 2.1 §2.3.4 / RFC 9207            | high     | iss parameter returned in authz response, validated by client
OAUTH21-4.1-pkce-required              | OAuth 2.1 §4.1 / RFC 7636              | critic PKCE required: code_challenge on authz, code_verifier on token
OAUTH21-pkce-s256                      | RFC 7636                                | high     | code_challenge_method=S256, not plain
OAUTH21-implicit-removed               | OAuth 2.1 (removed grant)               | critical | implicit flow (response_type=token) not in use
OAUTH21-ropc-removed                   | OAuth 2.1 (removed grant)               | critical | resource owner password credentials grant not in use
OAUTH21-code-one-time-use              | OAuth 2.1 §4.1.3 / RFC 6749 §10.5      | critical | authorization code is one-time-use, atomic consumption
OAUTH21-client-auth-token-endpoint     | OAuth 2.1 §2.4 / RFC 6749 §2.3         | high     | confidential clients authenticate on token endpoint
OAUTH21-scope-server-authoritative     | OAuth 2.1 §3.2.3 / RFC 6749 §10.6      | critical | granted scope derived from server-side authz request, not form input (this is the scope-tampering bug class)
OAUTH21-refresh-token-rotation         h 2.1 §4.3                          | high     | refresh tokens rotated on use, prior token invalidated
OAUTH21-refresh-token-binding          | OAuth 2.1 §4.3.2 / RFC 9700            | high     | refresh tokens bound to client; rejected if presented by different client
RFC9700-pkce-confidential-clients      | RFC 9700                                | medium   | PKCE applied to confidential clients too, not only public
RFC9700-sender-constrained             | RFC 9700                                | medi   | high-risk APIs use sender-constrained tokens (DPoP or mTLS)
RFC9126-par-request-uri                | RFC 9126                                | high     | if PAR used, request_uri properly validated and single-use
RFC8252-native-redirect                | RFC 8252                                | high     | native apps use claimed-https-scheme, loopback, or private-use-URI
RFC8252-no-embedded-useragent          | RFC 8252                                | high     | native apps do not use embedded user-agents (webviews) for authz
RFC9068-jwt-typ-header                 | RFC 9068                                | medium   | JWT access tokens have typ=at+jwt header
RFC9068-jwt-audience                   | RFC 9068                                | high     | JWT access token audience claim validated by resource server
RFC8707-resource-binding               | RFC 8707                                | high     | resource parameter handling: token aud bound to requested resource
RFC9449-dpop-htm-htu                   | RFC 9449                                | high     | DPoP proof: htm and htu claims match request
RFC9449-dpop-nonce                     | RFC 9449                                | medium   | DPoP nonce supported and replay-protected
RFC9449-dpop-jti                       | RFC 9449                                | high     | DPoP jti claim tracked for replay prevention
RFC8693-token-exchange-subject         | RFC 8693                                | high     | subject_token validated before issuing exchanged token
RFC8693-token-exchange-actor           | RFC 8693                                | high     | actor_token validated; impersonation/delegation policy enforced
RFC8693-token-exchange-may-act         | RFC 8693 §4.4                          | high     | may_act claim consulted when issuing delegated tokens
JOSE-alg-none-rejected                 | RFC 7515                                | critical | alg=none rejected at verification
JOSE-alg-confusion                     | RFC 7515 / RFC 7517                   | critical | HS-signed tokens rejected when key material is RSA/EC (confusion attack)
JOSE-kid-injection                     | RFC 7517                                | high     | kid header value not used in unsafe lookups (path traversal, SQL)
JOSE-jku-ssrf                          | RFC 7515 §4.1.2                        | high     | jku/x5u URLs allowlisted; not fetched from untrusted source
JOSE-typ-confusion                     | RFC 7519 / RFC 9068                    | high     | ty header validated to distinguish token classes
RFC7591-dcr-auth                       | RFC 7591                                | critical | dynamic client registration endpoint authenticated or rate-limited
RFC7591-dcr-metadata-validation        | RFC 7591                                | high     | registered redirect_uris validated against policy
RFC7662-introspect-auth                | RFC 7662                                | high     | introspection endpoint requires authentication
RFC7009-revocation-propagation         | RFC 7009                                | medium   | refresh token revocation also revokes derived access tokens
CIBA-backchannel-auth                  | CIBA Core §7                            | critical | backchannel auth endpoint requires confidential client authentication
CIBA-login-hint-nonce                  | CIBA Core §13                           | high     | login_hint is nonce-like or paired with user_code (anti-phishing)
CIBA-binding-message-display           | CIBA Core .1                          | high     | binding_message from request displayed on auth device
CIBA-push-mode-audited                 | CIBA Core §10.3                         | medium   | if push mode used, notification endpoint allowlist enforced
CIBA-auth-req-id-single-use            | CIBA Core §10.1                         | critical | auth_req_id single-use and bound to original client_id
OIDC-nonce-implicit-hybrid             | OIDC Core §3.2.2.11                    | high     | nonce required andlidated in implicit/hybrid flows
OIDC-id-token-signature                | OIDC Core §3.1.3.7                     | critical | ID token signature validated against issuer's published keys
OIDC-at-hash-validation                | OIDC Core §3.2.2.9 / §3.3.2.9          | high     | at_hash and c_hash validated when present
```

This list is intentionally not exhaustive. The implementing session ships v1 with the `critical`-tagged checks first, then adds `high`-tagged, and grows from there.

---

## 9. Finaltes to the implementing session

- **Verify everything that says "verify against current docs."** Model identifiers, hook payload field names, MCP tool names — these change. The handoff captures the design at a point in time; the implementation must reconcile against the current state.
- **Don't bundle the implementation into a single mega-commit.** The phase plan in section 6 produces five-to-six commits. Each is reviewable on its own.
- **The user prefers "ambiguous" over false confidence everywhere.** ply this principle to your own work: if an instruction in this document is ambiguous, surface that to the user rather than guessing.
- **The user prefers compiled binaries over scripts.** If you find yourself reaching for a Python or shell script during implementation, prefer Go.
- **Tests matter.** Every assertion in this document is a test case. The user values verification.
- **You may discover this design is wrong in places.** If you do, say so. Don't paper over a design flaw with implementation cleverness. The conversation that produced this document was explicitly adversarial; the user prefers being told the design is broken to being given a broken implementation.

End of handoff.

