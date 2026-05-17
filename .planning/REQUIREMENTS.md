# Requirements: security-reviewer

**Defined:** 2026-05-17
**Core Value:** Catch security bugs that previous Claude sessions miss — specifically the OAuth scope-tampering class — by isolating each specialist agent in its own context with its own tool allowlist and mechanically validating every output against a strict JSON schema before the verdict reaches synthesis.

## v1 Requirements

61 verifiable requirements lifted verbatim from HAND_OFF.md §3.1–§3.10. Each maps to exactly one roadmap phase. IDs are stable across the project lifecycle.

(The user's success metric refers to "51 per-agent assertion unit tests" — that label corresponds to the 45 per-agent runtime assertions A1–A11/T1–T11/AZ1–AZ6/OA1–OA7/IC1–IC4/S1–S6 plus 6 of the 7 hooks-H invariants implemented as unit-testable check predicates. H6 is the meta-coverage assertion; H1/H5 are build/dependency policy verified by build script not unit test. The remaining REQ-IDs — REQ-agents-prompt-shape and REQ-mcp-O1..O8 — are also v1 scope and are tracked below for full traceability, though they are not part of the "51 assertion" unit test suite specifically.)

### claude-security-hooks (binary contract)

- [ ] **REQ-hooks-H1** — `CGO_ENABLED=0 go build -o bin/claude-security-hooks ./cmd/claude-security-hooks` produces a single static binary. (HAND_OFF §3.7 H1)
- [ ] **REQ-hooks-H2** — All three subcommands (`preflight`, `validate`, `inject-context`) accept Claude Code hook event JSON on stdin matching the documented schemas (§8.1). (§3.7 H2)
- [ ] **REQ-hooks-H3** — When `tool_input.subagent_type` is not in the security agent set, the binary exits 0 with no stdout output. (§3.7 H3)
- [ ] **REQ-hooks-H4** — On any invariant failure, the binary writes exactly one JSON object `{"decision":"block","reason":"..."}` to stdout and exits 0. (§3.7 H4)
- [ ] **REQ-hooks-H5** — The core validator path has zero non-stdlib Go dependencies. Test framework deps allowed only in `_test.go` files. (§3.7 H5)
- [ ] **REQ-hooks-H6** — Every assertion A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6 is implemented as a check in `internal/invariants/<agent>.go` and exercised by a unit test. (§3.7 H6)
- [ ] **REQ-hooks-H7** — Malformed JSON in `tool_response.content` causes the validator to emit a block decision with a parse-error reason — never a panic. (§3.7 H7)

### go-cartographer (agent contract)

- [ ] **REQ-cartographer-A1** — Output is valid JSON parseable as `go-index/v1`. (§3.1 A1)
- [ ] **REQ-cartographer-A2** — `schema_version` is exactly the literal string `"go-index/v1"`. (§3.1 A2)
- [ ] **REQ-cartographer-A3** — `entrypoints` is an array (possibly empty); every entry has `router`, `method`, `path`, `handler.fqn`, `handler.file`, `handler.line` populated. (§3.1 A3)
- [ ] **REQ-cartographer-A4** — Every value in `routers_detected` is in `{net/http, chi, gin, gorilla/mux, echo, fiber, httprouter, custom}`. (§3.1 A4)
- [ ] **REQ-cartographer-A5** — If `routers_detected` is non-empty AND `entrypoints` is empty, `warnings` contains the literal `"router_detected_but_no_routes"`. (§3.1 A5)
- [ ] **REQ-cartographer-A6** — Any graph node ID cited in output MUST exist in `graphify-out/graph.json`. (§3.1 A6)
- [ ] **REQ-cartographer-A7** — Every file path cited in output exists in the workspace. (§3.1 A7)
- [ ] **REQ-cartographer-A8** — Every line number cited is within the corresponding file's line count. (§3.1 A8)
- [ ] **REQ-cartographer-A9** — `authz_primitives` contains only entries with `blocking: true`. (§3.1 A9)
- [ ] **REQ-cartographer-A10** — Agent MUST NOT modify any file in `graphify-out/` or the source tree (verified by absence of `Edit`/`Write` from allowlist + post-run hash). (§3.1 A10)
- [ ] **REQ-cartographer-A11** — Agent invokes `Bash` only with commands matching its `allowed_commands` list (`govulncheck -json ./...`, `govulncheck -json -mode=source ./...`). (§3.1 A11)

### go-taint-tracer (agent contract)

- [ ] **REQ-taint-T1** — Output parses as valid JSON conforming to the verdict schema in §3.2. (§3.2 T1)
- [ ] **REQ-taint-T2** — `verdict ∈ {exploitable, sanitized, unreachable, ambiguous, input_mismatch}`. (§3.2 T2)
- [ ] **REQ-taint-T3** — If `verdict ∈ {sanitized, exploitable}`, `path` is non-empty; first entry `step="source"`, last entry `step="sink"`. (§3.2 T3)
- [ ] **REQ-taint-T4** — If `verdict == "input_mismatch"`, `path` MAY be empty (no error). (§3.2 T4)
- [ ] **REQ-taint-T5** — `confidence ∈ {high, medium, low}`. (§3.2 T5)
- [ ] **REQ-taint-T6** — Either `semgrep.ran == true` OR `gopls.references_calls > 0` (or both), UNLESS `verdict == "input_mismatch"`. (§3.2 T6)
- [ ] **REQ-taint-T7** — `semgrep.tier` in output exactly equals the `semgrep_tier` passed in input. (§3.2 T7)
- [ ] **REQ-taint-T8** — If input source kind is interface-typed, EITHER `gopls.implementation_calls > 0` OR `notes` contains an explicit explanation for skipping the implementation walk. (§3.2 T8)
- [ ] **REQ-taint-T9** — Every `(file, line)` cited in `path` corresponds to a real workspace location. (§3.2 T9)
- [ ] **REQ-taint-T10** — `confidence == "high"` requires EITHER a Semgrep Pro/intrafile finding OR full LSP path verification with no `sanitizers_unverified` entries. (§3.2 T10)
- [ ] **REQ-taint-T11** — Agent MUST NOT call `Grep`, `Bash`, `Edit`, or `Write` (enforced via frontmatter + hook validator). (§3.2 T11)

### go-authz-tracer (agent contract)

- [ ] **REQ-authz-AZ1** — Output parses as valid JSON conforming to the `go-authz-tracer` output schema in §3.3. (§3.3 AZ1)
- [ ] **REQ-authz-AZ2** — `summary.routes_total == len(input.routes)`. (§3.3 AZ2)
- [ ] **REQ-authz-AZ3** — `summary.protected + summary.missing + summary.weak + summary.idor_risk + summary.public_intentional` accounts for all routes; `findings` de-dupes per route. (§3.3 AZ3)
- [ ] **REQ-authz-AZ4** — Every `route` value in `findings` corresponds to a route in `input.routes`. (§3.3 AZ4)
- [ ] **REQ-authz-AZ5** — Every entry in `weak_primitives` was present in `input.authz_primitives`. (§3.3 AZ5)
- [ ] **REQ-authz-AZ6** — Agent's `Read` call count ≤ `len(input.authz_primitives) + len(input.routes) * 2`. Each authz primitive body read at most once. (§3.3 AZ6)

### go-oauth-auditor (agent contract)

- [ ] **REQ-oauth-OA1** — Output parses as valid JSON conforming to the auditor schema in §3.4. (§3.4 OA1)
- [ ] **REQ-oauth-OA2** — Every `checklist[].spec` value matches a recognized RFC number or IETF/OIDF draft ID. (§3.4 OA2)
- [ ] **REQ-oauth-OA3** — No checklist entry may have `status == "pass"` without populated `evidence.file` AND `evidence.line`. (§3.4 OA3)
- [ ] **REQ-oauth-OA4** — Every `taint_pairs` entry references files and lines present in input `oauth_locations`. (§3.4 OA4)
- [ ] **REQ-oauth-OA5** — If `target_profile == "oauth_2_1"` and `features_in_use` contains `"pkce"`, all PKCE checks present with non-`not_applicable` status. (§3.4 OA5)
- [ ] **REQ-oauth-OA6** — If `target_profile == "oauth_2_0"` (with BCP) and `oauth_locations.authorize_endpoint` is set, implicit-flow-disallowed check evaluated and reported. (§3.4 OA6)
- [ ] **REQ-oauth-OA7** — A check whose target feature is absent MUST be `not_applicable`, never `pass`. (§3.4 OA7)

### invariant-checker (agent contract)

- [ ] **REQ-invariant-IC1** — Output parses as valid JSON conforming to the invariant-checker schema in §3.5. (§3.5 IC1)
- [ ] **REQ-invariant-IC2** — `results` has exactly one entry per input invariant; `invariant_id` values are the same set as input IDs. (§3.5 IC2)
- [ ] **REQ-invariant-IC3** — `status == "violated"` requires `evidence.files` non-empty AND `evidence.explanation` citing a specific code path. (§3.5 IC3)
- [ ] **REQ-invariant-IC4** — Agent only verifies input list; does not emit new invariants. (§3.5 IC4)

### synthesis (agent contract)

- [ ] **REQ-synthesis-S1** — Produces exactly two files: `review-report.json` and `review-report.md`. (§3.6 S1)
- [ ] **REQ-synthesis-S2** — `review-report.json` parses and validates against the `review-report/v1` schema. (§3.6 S2)
- [ ] **REQ-synthesis-S3** — `summary.total_findings == len(findings)`. (§3.6 S3)
- [ ] **REQ-synthesis-S4** — Sum of `summary.by_severity` values equals `summary.total_findings`. (§3.6 S4)
- [ ] **REQ-synthesis-S5** — Every entry in `findings` has at least one element in its `source_agents` array. (§3.6 S5)
- [ ] **REQ-synthesis-S6** — No two findings share `(file, line, class)` unless an explicit cross-reference note exists in `deduplication_notes`. (§3.6 S6)

### Agent definition format

- [ ] **REQ-agents-prompt-shape** — Every agent definition file in `.claude/agents/*.md` follows the common structure: (1) role statement, (2) input contract JSON shape, (3) numbered protocol, (4) reference tables where applicable, (5) output schema JSON shape, (6) hard rules (negative guidance, ambiguity preference, no-invent rule). Frontmatter declares `name`, `description`, `model`, `tools`, and optionally `allowed_commands`. (HAND_OFF §3.10)

### opengrep-mcp (server contract)

- [ ] **REQ-mcp-O1** — `CGO_ENABLED=0 go build` produces a working `opengrep-mcp` binary. (§3.8 O1)
- [ ] **REQ-mcp-O2** — Server registers `scan_with_rule`, `scan_directory`, `get_ast`; each tool schema validates against MCP JSON Schema spec. (§3.8 O2)
- [ ] **REQ-mcp-O3** — `scan_with_rule` returns a structured error when `tier` is not in `{pro, intrafile, ce}`. (§3.8 O3)
- [ ] **REQ-mcp-O4** — Server pulls container images on first use and caches them locally. (§3.8 O4)
- [ ] **REQ-mcp-O5** — Container invocation mounts workspace at `/src` read-only. (§3.8 O5)
- [ ] **REQ-mcp-O6** — When `timeout_seconds` exceeded, server kills container and returns a partial-results error. (§3.8 O6)
- [ ] **REQ-mcp-O7** — Server output uses the normalized schema in `internal/schema/findings.go` regardless of which scanner produced raw findings. (§3.8 O7)
- [ ] **REQ-mcp-O8** — `SEMGREP_APP_TOKEN` (or equivalent) is read from environment at startup and forwarded to the Pro container as env var; never logged to stdout/stderr. (§3.8 O8)

## v2 Requirements

Per D17 (Go-first sequential per-stack expansion), the v1 milestone is Go-only. Cross-stack support and CI integration are explicitly deferred.

### Cross-stack expansion

- **STACK-PY-01** — Python tracer/auditor agent set (post-Go validation)
- **STACK-TS-01** — TypeScript/JavaScript tracer/auditor agent set
- **STACK-SHARED-01** — Promote shared schema/hook layer to a cross-stack core module

### Synthesis enhancements

- **SYN-MD-01** — Rich Markdown report formatting (Q5 in HAND_OFF §5 — v1 may ship JSON-only with Markdown deferred)

### Invariant discovery

- **INV-DISC-01** — Auto-discovery of invariants from code patterns (deferred per D16 — out of scope for v1)

### CI integration

- **CI-GHA-01** — GitHub Actions wrapper consuming `review-report.json` (Q10 — out of scope for v1)

## Out of Scope

| Feature | Reason |
|---------|--------|
| `gosec`, `staticcheck` in pre-pass | D10 — redundant with custom Semgrep rules; `govulncheck` is the only deterministic pre-pass tool |
| RFC 6819 in OAuth checklist | D7 — superseded by RFC 9700 |
| Security review packaged as a Claude Code skill | D15 — skills load into main agent; instructions silently disappear across subagent boundary |
| `SubagentStop` hook for output validation | D5 — payload lacks subagent identification under parallel fan-out (anthropics/claude-code#7881) |
| Invariant discovery by `invariant-checker` | D16 — verifies caller-supplied invariants only; discovery is harder and out of scope for v1 |
| Parameterized cross-stack agent files | D1 / D17 — per-stack files keep prompt size manageable and accuracy high; abstract after one stack works end-to-end |
| Free-form prose verdicts | D13 — breaks mechanical parsing and deduplication |
| Two separate repos (one per Go module) | Operational — shared schema types and single-PR surface are lower friction during v1 |
| Python scripts or shell pipelines in shipped artifacts | User principle (compiled binaries over scripts) — Python startup latency is the explicit reason Go was chosen for the hook binary |

## Traceability

Each v1 requirement maps to exactly one roadmap phase. Updated 2026-05-17 during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| REQ-hooks-H1 | Phase 2 | Pending |
| REQ-hooks-H2 | Phase 2 | Pending |
| REQ-hooks-H3 | Phase 2 | Pending |
| REQ-hooks-H4 | Phase 2 | Pending |
| REQ-hooks-H5 | Phase 2 | Pending |
| REQ-hooks-H6 | Phase 2 | Pending |
| REQ-hooks-H7 | Phase 2 | Pending |
| REQ-cartographer-A1 | Phase 2 | Pending |
| REQ-cartographer-A2 | Phase 2 | Pending |
| REQ-cartographer-A3 | Phase 2 | Pending |
| REQ-cartographer-A4 | Phase 2 | Pending |
| REQ-cartographer-A5 | Phase 2 | Pending |
| REQ-cartographer-A6 | Phase 2 | Pending |
| REQ-cartographer-A7 | Phase 2 | Pending |
| REQ-cartographer-A8 | Phase 2 | Pending |
| REQ-cartographer-A9 | Phase 2 | Pending |
| REQ-cartographer-A10 | Phase 2 | Pending |
| REQ-cartographer-A11 | Phase 2 | Pending |
| REQ-taint-T1 | Phase 2 | Pending |
| REQ-taint-T2 | Phase 2 | Pending |
| REQ-taint-T3 | Phase 2 | Pending |
| REQ-taint-T4 | Phase 2 | Pending |
| REQ-taint-T5 | Phase 2 | Pending |
| REQ-taint-T6 | Phase 2 | Pending |
| REQ-taint-T7 | Phase 2 | Pending |
| REQ-taint-T8 | Phase 2 | Pending |
| REQ-taint-T9 | Phase 2 | Pending |
| REQ-taint-T10 | Phase 2 | Pending |
| REQ-taint-T11 | Phase 2 | Pending |
| REQ-authz-AZ1 | Phase 2 | Pending |
| REQ-authz-AZ2 | Phase 2 | Pending |
| REQ-authz-AZ3 | Phase 2 | Pending |
| REQ-authz-AZ4 | Phase 2 | Pending |
| REQ-authz-AZ5 | Phase 2 | Pending |
| REQ-authz-AZ6 | Phase 2 | Pending |
| REQ-oauth-OA1 | Phase 2 | Pending |
| REQ-oauth-OA2 | Phase 2 | Pending |
| REQ-oauth-OA3 | Phase 2 | Pending |
| REQ-oauth-OA4 | Phase 2 | Pending |
| REQ-oauth-OA5 | Phase 2 | Pending |
| REQ-oauth-OA6 | Phase 2 | Pending |
| REQ-oauth-OA7 | Phase 2 | Pending |
| REQ-invariant-IC1 | Phase 2 | Pending |
| REQ-invariant-IC2 | Phase 2 | Pending |
| REQ-invariant-IC3 | Phase 2 | Pending |
| REQ-invariant-IC4 | Phase 2 | Pending |
| REQ-synthesis-S1 | Phase 2 | Pending |
| REQ-synthesis-S2 | Phase 2 | Pending |
| REQ-synthesis-S3 | Phase 2 | Pending |
| REQ-synthesis-S4 | Phase 2 | Pending |
| REQ-synthesis-S5 | Phase 2 | Pending |
| REQ-synthesis-S6 | Phase 2 | Pending |
| REQ-agents-prompt-shape | Phase 3 | Pending |
| REQ-mcp-O1 | Phase 4 | Pending |
| REQ-mcp-O2 | Phase 4 | Pending |
| REQ-mcp-O3 | Phase 4 | Pending |
| REQ-mcp-O4 | Phase 4 | Pending |
| REQ-mcp-O5 | Phase 4 | Pending |
| REQ-mcp-O6 | Phase 4 | Pending |
| REQ-mcp-O7 | Phase 4 | Pending |
| REQ-mcp-O8 | Phase 4 | Pending |

**Coverage:**

| Phase | Owned Requirements | Count |
|-------|-------------------|-------|
| Phase 1 (Repo scaffold) | (none — pre-implementation support) | 0 |
| Phase 2 (claude-security-hooks binary) | REQ-hooks-H1..H7, REQ-cartographer-A1..A11, REQ-taint-T1..T11, REQ-authz-AZ1..AZ6, REQ-oauth-OA1..OA7, REQ-invariant-IC1..IC4, REQ-synthesis-S1..S6 | 52 |
| Phase 3 (Agent definitions + settings.json) | REQ-agents-prompt-shape | 1 |
| Phase 4 (opengrep-mcp server + container) | REQ-mcp-O1..O8 | 8 |
| Phase 5 (E2E smoke test) | (none — exit criterion is the 3-bug detection in `examples/sample-vulnerable-service/`) | 0 |
| Phase 6 (Documentation) | (none — docs phase) | 0 |
| **Total** | | **61** |

- v1 requirements: **61 total**
- Mapped to phases: **61**
- Unmapped: **0** ✓

---
*Requirements defined: 2026-05-17*
*Last updated: 2026-05-17 after initial definition*
