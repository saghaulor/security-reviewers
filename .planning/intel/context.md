# Context Intel

Background, motivation, open questions, references, and implementation sequencing notes drawn from HAND_OFF.md sections that are not strict spec/decision/requirement material. Treat these as roadmap-shaping context for downstream `gsd-roadmapper`.

source-doc: /home/saghaulor/code/security_reviewer/HAND_OFF.md

---

## Topic: Problem statement and prior failure modes

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §1

The user is building a security review system as a Claude Code plugin. A first attempt — loading a security review skill into the main agent and instructing it to dispatch tracer subagents with LSP-using instructions — failed because subagent context isolation makes the main agent an unreliable relay. Instructions in the skill that said "tell the tracer to use LSP" were silently dropped or paraphrased away when the main agent composed the Task tool prompt.

The motivating bug class is an OAuth scope-tampering bug where the consent form submission was trusted as the source-of-truth for granted scopes rather than the original authorization request. A previous Claude session missed this bug entirely; that session conceded that a dedicated tracer agent would have caught it.

**Three failure modes the design must prevent:**

1. Instructions buried in a skill loaded into the main agent silently disappearing when subagents are spawned.
2. Agents choosing cheap heuristic tools (text grep, intuition) over expensive precise tools (Semgrep taint, gopls call hierarchy) when both are available.
3. Drift over time — six months from now, someone weakens an agent's system prompt or tool allowlist and no signal surfaces in normal output.

**Design answers (in order):**

1. Each specialist agent has its own definition file with its own tool allowlist. Instructions live where the agent reads them, not where the dispatcher reads them.
2. Tool allowlists exclude cheaper alternatives. Numbered protocols in the system prompt enforce ordering of expensive tools.
3. Lifecycle hooks validate every tracer's output mechanically: did it call Semgrep? did its verdict JSON conform to schema? On failure, block and force retry with explicit feedback.

---

## Topic: Why two phases of fan-out

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §2.2

The cartographer is sequential and runs once. Its output is small (a structural index of the codebase) and is consumed by every downstream agent. Running it in parallel with the tracers would force each tracer to re-discover codebase shape independently, wasting tokens.

The tracers fan out massively. A medium Go service might have ~200 entrypoints × 10 sink kinds = ~2,000 candidate source/sink pairs. Most are filtered cheaply by the cartographer (no SQL sink in this handler's call graph → skip the SQL tracer). The rest run in parallel, one Task call per pair.

---

## Topic: How to read the handoff (precedence on conflict)

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §0

The implementing session should treat the handoff as source of truth. Where it conflicts with prior memory, the handoff wins. Where it is silent, prefer simplicity and explicit code over cleverness. When uncertain, surface the uncertainty rather than guessing — the user prefers explicit gaps over inference. Every component spec contains assertions; treat them as test cases.

---

## Topic: Implementation plan (6 phases)

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §6

Recommended phase sequence. Each phase produces a committable artifact.

- **Phase 1 — Repo scaffold:** create GitHub repo, init two Go modules (`claude-security-hooks`, `opengrep-mcp`), create `.claude/{agents,hooks/bin,security-invariants}/`, add `.gitignore`, `LICENSE`, top-level `README.md`. Commit: `scaffold: initial project structure`.
- **Phase 2 — claude-security-hooks binary:** subcommand dispatch in `cmd/claude-security-hooks/main.go`; per-agent schema types in `internal/schema/`; per-agent invariants in `internal/invariants/` (every assertion A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6 as a Go predicate); preflight/validate/inject implementations in `internal/hooks/`; unit test per invariant; Makefile with `build|test|install` (install copies binary into `.claude/hooks/bin/`). Commit: `feat(hooks): claude-security-hooks binary with per-agent validation`.
- **Phase 3 — Agent definitions:** write `.claude/agents/{go-cartographer,go-taint-tracer,go-authz-tracer,go-oauth-auditor,invariant-checker,synthesis}.md` per §3.1–§3.6; embed §8.2 catalog into taint-tracer, §3.2.2 OAuth table into taint-tracer, §8.3 checklist into oauth-auditor; register hooks in `.claude/settings.json`. Commit: `feat(agents): security review agent definitions`.
- **Phase 4 — opengrep-mcp server:** MCP server skeleton using `github.com/modelcontextprotocol/go-sdk`; Docker runner using `github.com/docker/docker/client`; implement `scan_with_rule`, `scan_directory`, `get_ast`; build OpenGrep container image; integration tests with a simple Semgrep CE rule; per-component README. Commit: `feat(mcp): opengrep-mcp dual-engine scanner MCP server`.
- **Phase 5 — Smoke test:** build `examples/sample-vulnerable-service/` (SQLi + authz bypass + OAuth scope-tampering); run Graphify; run full review workflow; verify all three bugs caught. Commit: `test: end-to-end smoke test with sample vulnerable service`.
- **Phase 6 — Documentation:** top-level README, per-component READMEs, `CONTRIBUTING.md` on adding agents / extending catalog. Commit: `docs: README and contributor docs`.

---

## Topic: Open questions (integration details to resolve during build)

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §5 (Q1–Q10)

These are NOT design gaps. They are integration details that depend on environment specifics. The implementing session resolves each via the recommended path.

- **Q1 — Exact Graphify MCP tool schemas/names:** verify by running `python -m graphify.serve --help` and probing the MCP handshake. Update the cartographer allowlist if names differ from `{query_graph, get_node, get_neighbors, shortest_path}`.
- **Q2 — Custom router wrappers in the user's target Go services:** cartographer has a fallback heuristic for unknown routers. User said "most likely standard"; verify on first real review.
- **Q3 — Semgrep Pro license provisioning:** `SEMGREP_APP_TOKEN` env var to `opengrep-mcp` at startup; server forwards to Pro container. Document in README.
- **Q4 — OpenGrep build command for Dockerfile:** sketch in §3.9 is incomplete; pin a release tag and verify build commands against upstream README at https://github.com/opengrep/opengrep.
- **Q5 — Synthesis: JSON only or also Markdown:** designed for both. v1 may ship JSON-only with Markdown deferred.
- **Q6 — Concrete Semgrep rule shapes for OAuth audit checks:** each checklist item maps to a Semgrep pattern OR a code-read verification. Ship v1 with ~15 critical checks (PKCE, redirect_uri exact match, state validation, scope tampering) and grow.
- **Q7 — Orchestrator agent vs slash command:** recommend slash command (`/security-review`) for v1 — simpler, more transparent, easier to debug.
- **Q8 — Invariant storage:** recommend `.claude/security-invariants/<flow>.yaml` directory pattern, loaded by orchestrator and passed inline to the agent.
- **Q9 — Per-agent `model:` default:** default to `claude-sonnet-4-6` (verify current model identifier against https://docs.claude.com at implementation time). User can override to Opus per-agent if tracer accuracy proves insufficient.
- **Q10 — CI integration:** not specified in design. Out of scope for v1. Each agent's JSON outputs are machine-readable; future task can wire to GitHub Actions.

---

## Topic: References — OAuth / OIDC specifications

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §7.1

**Active / current (as of May 2026):**

- OAuth 2.1 draft-15 (March 2026): https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1
- RFC 6749 OAuth 2.0 Authorization Framework (legacy): https://www.rfc-editor.org/rfc/rfc6749
- RFC 6750 Bearer Token Usage (legacy): https://www.rfc-editor.org/rfc/rfc6750
- RFC 9700 OAuth 2.0 Security BCP: https://www.rfc-editor.org/rfc/rfc9700
- draft-ietf-oauth-security-topics-update: https://datatracker.ietf.org/doc/draft-ietf-oauth-security-topics-update/

**Code flow hardening:**

- RFC 7636 PKCE: https://www.rfc-editor.org/rfc/rfc7636
- RFC 9126 Pushed Authorization Requests (PAR): https://www.rfc-editor.org/rfc/rfc9126
- RFC 9207 Issuer Identification: https://www.rfc-editor.org/rfc/rfc9207
- RFC 9101 JWT-Secured Authorization Request (JAR): https://www.rfc-editor.org/rfc/rfc9101

**Native / cross-device:**

- RFC 8252 OAuth 2.0 for Native Apps: https://www.rfc-editor.org/rfc/rfc8252
- draft-ietf-oauth-cross-device-security

**AS metadata and discovery:**

- RFC 8414 Authorization Server Metadata: https://www.rfc-editor.org/rfc/rfc8414

**Token shape / sender-constraint / audience:**

- RFC 9068 JWT Profile for Access Tokens: https://www.rfc-editor.org/rfc/rfc9068
- RFC 8707 Resource Indicators: https://www.rfc-editor.org/rfc/rfc8707
- RFC 9449 DPoP: https://www.rfc-editor.org/rfc/rfc9449
- RFC 8705 OAuth 2.0 Mutual-TLS: https://www.rfc-editor.org/rfc/rfc8705

**Token exchange / assertions:**

- RFC 8693 Token Exchange (provided as project PDF): https://www.rfc-editor.org/rfc/rfc8693
- RFC 7521 Assertion Framework: https://www.rfc-editor.org/rfc/rfc7521
- RFC 7523 JWT Bearer Token Profile: https://www.rfc-editor.org/rfc/rfc7523

**Token operations:**

- RFC 7662 Introspection: https://www.rfc-editor.org/rfc/rfc7662
- RFC 7009 Revocation: https://www.rfc-editor.org/rfc/rfc7009

**Dynamic client management:**

- RFC 7591 Dynamic Client Registration: https://www.rfc-editor.org/rfc/rfc7591
- RFC 7592 Dynamic Client Registration Management: https://www.rfc-editor.org/rfc/rfc7592

**JOSE:**

- RFC 7515 JWS / 7516 JWE / 7517 JWK / 7518 JWA / 7519 JWT (https://www.rfc-editor.org/rfc/rfc7515 … rfc7519)

**OpenID Connect:**

- OIDC Core 1.0: https://openid.net/specs/openid-connect-core-1_0.html
- OIDC Discovery 1.0: https://openid.net/specs/openid-connect-discovery-1_0.html
- CIBA Core 1.0: https://openid.net/specs/openid-client-initiated-backchannel-authentication-core-1_0.html
- FAPI CIBA Profile (optional): https://openid.net/specs/openid-financial-api-ciba.html

---

## Topic: References — Tooling

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §7.2

**Claude Code:**

- Hooks reference: https://code.claude.com/docs/en/hooks
- Subagents: https://docs.claude.com/en/docs/claude-code/sub-agents
- Settings: https://docs.claude.com/en/docs/claude-code/settings
- Known limitation: `SubagentStop` payload lacks subagent ID with parallel subagents — https://github.com/anthropics/claude-code/issues/7881

**Graphify:**

- Repo: https://github.com/safishamsi/graphify
- Concept article: https://blog.gopenai.com/graphify-build-a-knowledge-graph-from-your-entire-codebase-without-sending-your-code-to-anyone-1b6924474b50

**Semgrep:**

- Docs: https://semgrep.dev/docs/
- Taint mode: https://semgrep.dev/docs/writing-rules/data-flow/taint-mode
- MCP repo (archived; migrated into main binary): https://github.com/semgrep/mcp
- Pro features: https://semgrep.dev/products/semgrep-code

**OpenGrep:**

- Repo: https://github.com/opengrep/opengrep
- Rules: https://github.com/opengrep/opengrep-rules

**gopls / Go LSP:**

- gopls: https://pkg.go.dev/golang.org/x/tools/gopls
- Built-in MCP server (v0.20.0+): https://github.com/golang/tools/tree/master/gopls/internal/mcp
- Third-party `mcp-gopls`: https://github.com/hloiseaufcms/mcp-gopls

**govulncheck:**

- Docs: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
- Vuln DB: https://pkg.go.dev/vuln/

**Model Context Protocol:**

- Spec: https://modelcontextprotocol.io/
- Go SDK (maintained with Google): https://github.com/modelcontextprotocol/go-sdk
- Reference servers: https://github.com/modelcontextprotocol/servers

**Docker Engine SDK for Go:** https://pkg.go.dev/github.com/docker/docker/client

---

## Topic: References — Go routing libraries for cartographer patterns

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §7.3

- `net/http` stdlib: https://pkg.go.dev/net/http
- chi: https://github.com/go-chi/chi
- gin: https://github.com/gin-gonic/gin
- gorilla/mux: https://github.com/gorilla/mux
- echo: https://github.com/labstack/echo
- fiber: https://github.com/gofiber/fiber
- httprouter: https://github.com/julienschmidt/httprouter

---

## Topic: Final notes from designer to implementing session

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §9

- **Verify everything that says "verify against current docs."** Model identifiers, hook payload field names, MCP tool names change.
- **Don't bundle the implementation into a single mega-commit.** The phase plan in §6 produces five-to-six commits; each is reviewable on its own.
- **The user prefers "ambiguous" over false confidence everywhere.** Apply to the implementer's own work: surface ambiguity to the user rather than guessing.
- **The user prefers compiled binaries over scripts.** Reaching for Python or shell → prefer Go.
- **Tests matter.** Every assertion is a test case.
- **You may discover the design is wrong in places.** If so, say so. Don't paper over a design flaw with implementation cleverness. The conversation that produced the handoff was adversarial; the user prefers being told the design is broken to being given a broken implementation.

---

## Topic: Cross-reference index (file paths referenced but not yet authored)

source: /home/saghaulor/code/security_reviewer/HAND_OFF.md cross_refs (from classification)

The handoff references several artifacts that DO NOT yet exist in the workspace. These are not pre-existing docs to ingest; they are paths the implementing session will create per the phase plan:

- `.claude/agents/go-cartographer.md` (Phase 3)
- `.claude/agents/go-taint-tracer.md` (Phase 3)
- `.claude/agents/go-authz-tracer.md` (Phase 3)
- `.claude/agents/go-oauth-auditor.md` (Phase 3)
- `.claude/agents/invariant-checker.md` (Phase 3)
- `.claude/agents/synthesis.md` (Phase 3)
- `graphify-out/graph.json` (Pre-pass artifact, produced by `graphify build` against target repo)
- `govulncheck.json` (Pre-pass artifact, produced by `govulncheck -json ./...`)
- `go-index.json` (Cartographer output, per-review)
- `review-report.json` / `review-report.md` (Synthesis output, per-review)
- `RFC_8693_OAuth_2_0_Token_Exchange.pdf` (project file, Token Exchange spec — see §7.4)

Cycle detection on the cross-reference graph: no cycles (single doc; all refs are outbound to artifacts-to-be-created or external URLs).
