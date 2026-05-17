# Requirements Intel

Requirements derived from HAND_OFF.md §3 component specifications. Each numbered assertion in a component contract maps to one verifiable requirement. Sources are SPEC-classified — these are normative behavior contracts the implementation must satisfy.

source-doc: /home/saghaulor/code/security_reviewer/HAND_OFF.md
sections: §3.1–§3.10 (component specifications)

---

## REQ-cartographer-A1 — Output is valid JSON parseable as go-index/v1

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A1)
- component: `go-cartographer`
- scope: cartographer output contract
- acceptance: The agent's final message text, after optional code-fence stripping, parses without error against the `go-index/v1` schema.

## REQ-cartographer-A2 — schema_version equals "go-index/v1"

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A2)
- component: `go-cartographer`
- scope: schema versioning
- acceptance: The output JSON's `schema_version` field is exactly the literal string `"go-index/v1"`.

## REQ-cartographer-A3 — entrypoints fully populated

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A3)
- component: `go-cartographer`
- scope: entrypoint enumeration
- acceptance: `entrypoints` is an array (possibly empty); every entry has `router`, `method`, `path`, `handler.fqn`, `handler.file`, `handler.line` populated.

## REQ-cartographer-A4 — routers_detected drawn from known set

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A4)
- component: `go-cartographer`
- scope: router detection
- acceptance: Every value in `routers_detected` is one of `{net/http, chi, gin, gorilla/mux, echo, fiber, httprouter, custom}`.

## REQ-cartographer-A5 — warn when router detected but no routes

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A5)
- component: `go-cartographer`
- scope: warning surface
- acceptance: If `routers_detected` is non-empty AND `entrypoints` is empty, `warnings` contains the literal string `"router_detected_but_no_routes"`.

## REQ-cartographer-A6 — no fabricated graph node IDs

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A6)
- component: `go-cartographer`
- scope: provenance
- acceptance: Any graph node ID cited in the output MUST exist in `graphify-out/graph.json`.

## REQ-cartographer-A7 — every file path exists

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A7)
- component: `go-cartographer`
- scope: provenance
- acceptance: Every file path cited in the output exists in the workspace.

## REQ-cartographer-A8 — every line number within bounds

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A8)
- component: `go-cartographer`
- scope: provenance
- acceptance: Every line number cited is within the corresponding file's line count.

## REQ-cartographer-A9 — non-blocking authz primitives excluded

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A9)
- component: `go-cartographer`
- scope: authz primitive filtering
- acceptance: `authz_primitives` contains only entries with `blocking: true`. Non-blocking middleware is excluded from the array.

## REQ-cartographer-A10 — no source modifications

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A10)
- component: `go-cartographer`
- scope: read-only guarantee
- acceptance: The agent MUST NOT modify any file in `graphify-out/` or the source tree. Verified by absence of `Edit`/`Write` from allowlist and by post-run file hash comparison.

## REQ-cartographer-A11 — Bash restricted to allowed_commands

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.1 (assertion A11)
- component: `go-cartographer`
- scope: tool execution policy
- acceptance: The agent invokes `Bash` only with commands matching its frontmatter `allowed_commands` list (`govulncheck -json ./...`, `govulncheck -json -mode=source ./...`).

---

## REQ-taint-T1 — Verdict JSON conforms to schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T1)
- component: `go-taint-tracer`
- scope: verdict contract
- acceptance: Output parses as valid JSON conforming to the verdict schema declared in §3.2.

## REQ-taint-T2 — verdict enum constrained

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T2)
- component: `go-taint-tracer`
- scope: verdict enum
- acceptance: `verdict ∈ {exploitable, sanitized, unreachable, ambiguous, input_mismatch}`.

## REQ-taint-T3 — path bookends for resolved verdicts

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T3)
- component: `go-taint-tracer`
- scope: path integrity
- acceptance: If `verdict ∈ {sanitized, exploitable}`, `path` is non-empty; first entry has `step="source"`, last entry has `step="sink"`.

## REQ-taint-T4 — input_mismatch may have empty path

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T4)
- component: `go-taint-tracer`
- scope: path integrity
- acceptance: If `verdict == "input_mismatch"`, `path` MAY be empty (no error).

## REQ-taint-T5 — confidence enum constrained

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T5)
- component: `go-taint-tracer`
- scope: confidence enum
- acceptance: `confidence ∈ {high, medium, low}`.

## REQ-taint-T6 — at least one analysis engine ran

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T6)
- component: `go-taint-tracer`
- scope: methodology proof
- acceptance: Either `semgrep.ran == true` OR `gopls.references_calls > 0` (or both), UNLESS `verdict == "input_mismatch"`.

## REQ-taint-T7 — semgrep.tier matches input

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T7)
- component: `go-taint-tracer`
- scope: tier integrity
- acceptance: `semgrep.tier` in the output exactly equals the `semgrep_tier` passed in input.

## REQ-taint-T8 — interface dispatch requires implementation walk

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T8)
- component: `go-taint-tracer`
- scope: interface fan-out
- acceptance: If the input source kind is interface-typed, EITHER `gopls.implementation_calls > 0` OR `notes` contains an explicit explanation for why the implementation walk was skipped.

## REQ-taint-T9 — every cited location is real

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T9)
- component: `go-taint-tracer`
- scope: provenance
- acceptance: Every `(file, line)` cited in `path` corresponds to a real location in the workspace.

## REQ-taint-T10 — high-confidence preconditions

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T10)
- component: `go-taint-tracer`
- scope: confidence gating
- acceptance: `confidence == "high"` requires EITHER a Semgrep Pro/intrafile finding OR full LSP path verification with no `sanitizers_unverified` entries.

## REQ-taint-T11 — forbidden tools never invoked

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.2 (assertion T11)
- component: `go-taint-tracer`
- scope: tool allowlist enforcement
- acceptance: The agent MUST NOT call `Grep`, `Bash`, `Edit`, or `Write`. Enforced by absence from frontmatter `tools:` list and verified by the hook validator.

---

## REQ-authz-AZ1 — Output JSON conforms to schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.3 (assertion AZ1)
- component: `go-authz-tracer`
- scope: output contract
- acceptance: Output parses as valid JSON conforming to the `go-authz-tracer` output schema in §3.3.

## REQ-authz-AZ2 — summary.routes_total matches input

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.3 (assertion AZ2)
- component: `go-authz-tracer`
- scope: route accounting
- acceptance: `summary.routes_total == len(input.routes)`.

## REQ-authz-AZ3 — summary buckets account for every route

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.3 (assertion AZ3)
- component: `go-authz-tracer`
- scope: route accounting
- acceptance: `summary.protected + summary.missing + summary.weak + summary.idor_risk + summary.public_intentional` accounts for all routes. A route may appear in multiple buckets if it has multiple findings; the `findings` array de-dupes per route.

## REQ-authz-AZ4 — finding routes correspond to input

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.3 (assertion AZ4)
- component: `go-authz-tracer`
- scope: provenance
- acceptance: Every `route` value in `findings` corresponds to a route present in `input.routes`.

## REQ-authz-AZ5 — weak_primitives reference input

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.3 (assertion AZ5)
- component: `go-authz-tracer`
- scope: provenance
- acceptance: Every entry in `weak_primitives` was present in `input.authz_primitives`.

## REQ-authz-AZ6 — read budget enforced

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.3 (assertion AZ6)
- component: `go-authz-tracer`
- scope: efficiency
- acceptance: The agent's `Read` call count does not exceed `len(input.authz_primitives) + len(input.routes) * 2`. Each authz primitive body is read at most once.

---

## REQ-oauth-OA1 — Output JSON conforms to schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4 (assertion OA1)
- component: `go-oauth-auditor`
- scope: output contract
- acceptance: Output parses as valid JSON conforming to the auditor schema in §3.4.

## REQ-oauth-OA2 — checklist entries cite recognized specs

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4 (assertion OA2)
- component: `go-oauth-auditor`
- scope: spec attribution
- acceptance: Every `checklist[].spec` value matches a recognized RFC number or IETF/OIDF draft ID.

## REQ-oauth-OA3 — pass requires concrete evidence

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4 (assertion OA3)
- component: `go-oauth-auditor`
- scope: evidence requirement
- acceptance: No checklist entry may have `status == "pass"` without populated `evidence.file` AND `evidence.line`.

## REQ-oauth-OA4 — taint_pairs reference oauth_locations

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4 (assertion OA4)
- component: `go-oauth-auditor`
- scope: taint pair provenance
- acceptance: Every `taint_pairs` entry references files and lines present in the input `oauth_locations`.

## REQ-oauth-OA5 — PKCE checks evaluated for 2.1+pkce

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4 (assertion OA5)
- component: `go-oauth-auditor`
- scope: conditional check coverage
- acceptance: If `target_profile == "oauth_2_1"` and `features_in_use` contains `"pkce"`, all PKCE-related checks are present in the checklist with a non-`not_applicable` status.

## REQ-oauth-OA6 — implicit-disallowed evaluated on 2.0+BCP

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4 (assertion OA6)
- component: `go-oauth-auditor`
- scope: conditional check coverage
- acceptance: If `target_profile == "oauth_2_0"` (with BCP) and `oauth_locations.authorize_endpoint` is set, the implicit-flow-disallowed check is evaluated and reported.

## REQ-oauth-OA7 — absence is not_applicable, not pass

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.4 (assertion OA7)
- component: `go-oauth-auditor`
- scope: status semantics
- acceptance: A check whose target feature is absent from the codebase MUST be reported as `not_applicable`, never `pass`.

---

## REQ-invariant-IC1 — Output JSON conforms to schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.5 (assertion IC1)
- component: `invariant-checker`
- scope: output contract
- acceptance: Output parses as valid JSON conforming to the invariant-checker schema in §3.5.

## REQ-invariant-IC2 — results align with input invariants

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.5 (assertion IC2)
- component: `invariant-checker`
- scope: result accounting
- acceptance: `results` has exactly one entry per input invariant; `invariant_id` values are the same set as input IDs.

## REQ-invariant-IC3 — violated requires cited evidence

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.5 (assertion IC3)
- component: `invariant-checker`
- scope: evidence requirement
- acceptance: `status == "violated"` requires `evidence.files` non-empty AND `evidence.explanation` citing a specific code path.

## REQ-invariant-IC4 — agent does not invent invariants

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.5 (assertion IC4)
- component: `invariant-checker`
- scope: scope discipline
- acceptance: The agent only verifies the input list. It does not emit new invariants in its output.

---

## REQ-synthesis-S1 — Two-file output

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.6 (assertion S1)
- component: `synthesis`
- scope: output contract
- acceptance: The agent produces exactly two files: `review-report.json` and `review-report.md`.

## REQ-synthesis-S2 — JSON conforms to review-report/v1

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.6 (assertion S2)
- component: `synthesis`
- scope: schema conformance
- acceptance: `review-report.json` parses and validates against the `review-report/v1` schema.

## REQ-synthesis-S3 — total_findings consistency

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.6 (assertion S3)
- component: `synthesis`
- scope: accounting
- acceptance: `summary.total_findings == len(findings)`.

## REQ-synthesis-S4 — severity counts sum to total

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.6 (assertion S4)
- component: `synthesis`
- scope: accounting
- acceptance: Sum of `summary.by_severity` values equals `summary.total_findings`.

## REQ-synthesis-S5 — every finding has source_agents

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.6 (assertion S5)
- component: `synthesis`
- scope: provenance
- acceptance: Every entry in `findings` has at least one element in its `source_agents` array.

## REQ-synthesis-S6 — no silent duplicates

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.6 (assertion S6)
- component: `synthesis`
- scope: deduplication policy
- acceptance: No two findings share `(file, line, class)` unless an explicit cross-reference note exists in `deduplication_notes`.

---

## REQ-hooks-H1 — Single static binary

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H1)
- component: `claude-security-hooks`
- scope: build target
- acceptance: `CGO_ENABLED=0 go build -o bin/claude-security-hooks ./cmd/claude-security-hooks` produces a single static binary.

## REQ-hooks-H2 — Subcommand stdin contract

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H2)
- component: `claude-security-hooks`
- scope: I/O contract
- acceptance: All three subcommands (`preflight`, `validate`, `inject-context`) accept Claude Code hook event JSON on stdin matching the documented schemas (see §8.1).

## REQ-hooks-H3 — Silent no-op for non-security agents

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H3)
- component: `claude-security-hooks`
- scope: scoping
- acceptance: When `tool_input.subagent_type` is not in the security agent set, the binary exits 0 with no stdout output.

## REQ-hooks-H4 — Block decision JSON shape

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H4)
- component: `claude-security-hooks`
- scope: block protocol
- acceptance: On any invariant failure, the binary writes exactly one JSON object `{"decision":"block","reason":"..."}` to stdout and exits 0.

## REQ-hooks-H5 — Zero external runtime dependencies

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H5)
- component: `claude-security-hooks`
- scope: dependency policy
- acceptance: The core validator path has zero non-stdlib Go dependencies. Test framework deps are allowed only in `_test.go` files.

## REQ-hooks-H6 — Every per-agent assertion is a hook check

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H6)
- component: `claude-security-hooks`
- scope: invariant coverage
- acceptance: Every assertion A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6 is implemented as a check in `internal/invariants/<agent>.go` and exercised by a unit test.

## REQ-hooks-H7 — Parse errors become block decisions

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.7 (assertion H7)
- component: `claude-security-hooks`
- scope: failure mode
- acceptance: Malformed JSON in `tool_response.content` causes the validator to emit a block decision with a parse-error reason — never a panic.

---

## REQ-mcp-O1 — MCP binary builds without CGO

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O1)
- component: `opengrep-mcp`
- scope: build target
- acceptance: `CGO_ENABLED=0 go build` produces a working `opengrep-mcp` binary.

## REQ-mcp-O2 — Three tools registered with MCP schemas

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O2)
- component: `opengrep-mcp`
- scope: tool surface
- acceptance: The server registers `scan_with_rule`, `scan_directory`, `get_ast`. Each tool schema validates against the MCP JSON Schema spec.

## REQ-mcp-O3 — Reject unknown tier

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O3)
- component: `opengrep-mcp`
- scope: input validation
- acceptance: `scan_with_rule` returns a structured error when `tier` is not in `{pro, intrafile, ce}`.

## REQ-mcp-O4 — On-demand image pull with caching

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O4)
- component: `opengrep-mcp`
- scope: container management
- acceptance: The server pulls container images on first use and caches them locally for subsequent calls.

## REQ-mcp-O5 — Workspace mounted read-only

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O5)
- component: `opengrep-mcp`
- scope: isolation
- acceptance: The container invocation mounts the workspace at `/src` with read-only flag set.

## REQ-mcp-O6 — Timeout kills container

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O6)
- component: `opengrep-mcp`
- scope: timeout handling
- acceptance: When `timeout_seconds` is exceeded, the server kills the container and returns a partial-results error.

## REQ-mcp-O7 — Normalized findings schema

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O7)
- component: `opengrep-mcp`
- scope: output normalization
- acceptance: Server output uses the normalized schema defined in `internal/schema/findings.go` regardless of which scanner produced the raw findings.

## REQ-mcp-O8 — Pro license never logged

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.8 (assertion O8)
- component: `opengrep-mcp`
- scope: secret handling
- acceptance: `SEMGREP_APP_TOKEN` (or equivalent) is read from the server's environment at startup and forwarded to the Pro container as an env var. It is never written to logs or stderr.

---

## REQ-agents-prompt-shape — Common agent system prompt structure

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §3.10
- component: `.claude/agents/*.md`
- scope: agent definition format
- acceptance: Every agent definition file follows the common structure: (1) role statement, (2) input contract JSON shape, (3) numbered protocol, (4) reference tables where applicable, (5) output schema JSON shape, (6) hard rules (negative guidance, ambiguity preference, no-invent rule). Frontmatter declares `name`, `description`, `model`, `tools`, and optionally `allowed_commands`.
