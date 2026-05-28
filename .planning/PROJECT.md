# security-reviewer

## What This Is

A Claude Code plugin that performs security review of Go services by dispatching a sequential cartographer agent followed by parallel specialist tracer agents (taint, authorization, OAuth/OIDC, invariant) and synthesizing their machine-readable verdicts into a single review report. It is for the maintainer of this repo (the user) operating inside Claude Code; v1 ships Go-only and runs through a `/security-review` slash command.

## Core Value

Catch security bugs that previous Claude sessions miss — specifically the OAuth scope-tampering class — by isolating each specialist agent in its own context with its own tool allowlist and mechanically validating every output against a strict JSON schema before the verdict is allowed to reach synthesis.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

(None yet — v1 in progress.)

### Active

<!-- v1 scope. Full list with REQ-IDs in REQUIREMENTS.md. -->

- [ ] `claude-security-hooks` Go binary that mechanically validates every specialist agent's JSON verdict against per-agent invariants (H1–H7, plus 44 per-agent assertion checks)
- [ ] Six agent definition files under `.claude/agents/` (go-cartographer, go-taint-tracer, go-authz-tracer, go-oauth-auditor, invariant-checker, synthesis) wired to the hooks binary via `.claude/settings.json`
- [ ] `opengrep-mcp` Go MCP server fronting Semgrep Pro / OpenGrep intrafile / OpenGrep CE via a single `scan_with_rule` tool with a `tier` parameter (O1–O8)
- [ ] End-to-end smoke test against `examples/sample-vulnerable-service/` containing SQLi + authz bypass + OAuth scope-tampering bugs, all three flagged with non-ambiguous verdicts
- [ ] Top-level + per-component READMEs + `CONTRIBUTING.md`

### Out of Scope

- **Cross-stack agents (Python, TypeScript, Java, etc.)** — per D17, ship Go end-to-end before adding another language. Cross-stack abstractions before having one stack working produce leaky abstractions.
- **Invariant discovery** — per D16, `invariant-checker` only verifies caller-supplied invariants; it does not mint new ones.
- **`gosec` / `staticcheck` in the cartographer pre-pass** — per D10, both are redundant with custom Semgrep rules. `govulncheck` is the only deterministic pre-pass tool.
- **RFC 6819 in the OAuth checklist** — per D7, superseded by RFC 9700.
- **Security review as a Claude Code skill** — per D15, instructions in skills silently disappear across subagent boundaries; security review is a slash command + per-agent definition files.
- **`SubagentStop` hook for output validation** — per D5, `SubagentStop` payload lacks subagent identification under parallel fan-out; validation runs on `PostToolUse(Task)` instead. See anthropics/claude-code#7881.
- **CI integration (GitHub Actions, etc.)** — Q10 in HAND_OFF §5. Each agent's JSON output is machine-readable so this is future-additive; not v1.
- **Skills, Python scripts, dynamic-linked binaries** — user principle. Compiled Go binaries with `CGO_ENABLED=0` only.
- **Free-form prose verdicts from any agent** — per D13, all outputs are strict JSON. Prose breaks deduplication and mechanical parsing.

## Context

- **Target runtime:** Claude Code only. No standalone CLI surface in v1.
- **Two Go modules in one repo:** `claude-security-hooks` (lifecycle hook validator binary) and `opengrep-mcp` (scanner MCP server). They develop in parallel and converge at Phase 5 (smoke test).
- **Motivating bug:** an OAuth scope-tampering vulnerability where a previous Claude session trusted the consent-form submission as the source-of-truth for granted scopes instead of the original authorization request. That session conceded a dedicated tracer would have caught it. The whole architecture exists to make that catch reliable.
- **Three failure modes the design must prevent** (HAND_OFF §1):
  1. Instructions in a skill loaded into the main agent silently disappearing when subagents are spawned → solved by per-agent definition files with their own system prompts.
  2. Agents picking cheap heuristic tools (text grep, intuition) over expensive precise tools (Semgrep taint, gopls call hierarchy) → solved by removing cheap tools from allowlists (D12) and embedding numbered protocols that force tool ordering.
  3. Silent drift over time (someone weakens a prompt, no signal surfaces) → solved by hook-based mechanical validation: every output is schema-checked and runtime-invariant-checked, block-on-failure.
- **Pre-pass / fan-out architecture** (CON-arch-three-layers):
  1. Pre-pass (per repo, on dep change): `govulncheck` + `graphify build`
  2. Sequential cartographer (per review): consumes pre-pass artifacts, emits `go-index.json`
  3. Parallel fan-out (per review): tracers run concurrently from `go-index.json`
  4. Synthesis: reads directory of specialist outputs (D18), emits `review-report.{json,md}`
- **51 verifiable per-agent assertions** are lifted from HAND_OFF §3.1–§3.10. Each becomes a Go predicate in `internal/invariants/` (REQ-hooks-H6) plus a unit test in `_test.go`. The unit test suite is Phase 2's exit criterion.
- **Open questions deferred to first-day work inside their owning phase** — Q1 (Graphify MCP tool names), Q4 (OpenGrep build pin), Q9 (model identifier verification). These are integration details to resolve during build, not roadmap blockers.

## Project Principles

These are user-supplied principles that constrain every phase. They come from the user's global `CLAUDE.md` and from HAND_OFF §9 ("Final notes from designer to implementing session").

- **TDD with tests before implementation.** Every plan must have a test section before the implementation section. Every per-agent assertion is a test case before it is a check predicate.
- **Compiled binaries over scripts.** When reaching for Python or shell, prefer Go. The hook binary's cold-start budget (<5ms, CON-nfr-cold-start) is the explicit reason Go was chosen over Python.
- **Declared types over `any`.** Per-agent schemas live in `internal/schema/` as Go structs, not `map[string]any`. Validation is type-driven.
- **gopls over grep for code analysis.** Use the Go language server for symbol/reference/implementation queries. Reach for other tools (`Grep`, ad-hoc Bash) only when gopls cannot help, and tell the user when that happens. This principle applies to the implementing session AND it is the design rationale for D12 (tracer allowlists exclude `Grep`).
- **Ambiguity preferred over false confidence.** Per D14, agent verdicts must return `ambiguous` / `unverifiable` / `unverified` when honest. Applies to the implementer too: surface uncertainty to the user rather than guess. The conversation that produced the handoff was adversarial; the user prefers being told the design is broken to being given a broken implementation.
- **Per-stack first, not parameterized.** Per D17, ship Go end-to-end before adding another language. Don't pre-abstract.

## Constraints

- **Tech stack**: Go (both binaries, CGO_ENABLED=0 static) — CON-nfr-static-binary. No Python in shipped artifacts.
- **Runtime**: Claude Code only (v1). Hook registrations target `PreToolUse`/`PostToolUse` matcher `Task` and `SubagentStart` matcher on the security agent name set — CON-hook-registration.
- **Hook latency**: Cold start <5ms — CON-nfr-cold-start. Drives Go-not-Python.
- **Dependencies**: `claude-security-hooks` core validator path has zero non-stdlib Go imports — CON-nfr-zero-deps / REQ-hooks-H5. Test framework deps allowed only in `_test.go`.
- **Container isolation**: All scanner workspace mounts are read-only (`:ro`); container killed on timeout — CON-nfr-container-isolation.
- **Schema contracts (LOCKED)**: `go-index/v1`, taint verdict, authz tracer, oauth auditor, invariant checker, `review-report/v1` — CON-schema-* family. Every agent output parses to a Go struct in `internal/schema/`.
- **Tool allowlists (LOCKED)**: Per-agent frontmatter `tools:` lists are enumerated in CON-tool-allowlists. Tracer agents MUST NOT have `Grep`, `Bash`, `Edit`, `Write` in their allowlist (D12, T11).
- **OAuth spec set**: OAuth 2.1, OAuth 2.0+RFC9700 BCP, OIDC Core, CIBA Core 1.0. RFC 6819 excluded (D7). CIBA included (D9). Stable check IDs per CON-oauth-checklist-taxonomy.
- **Bug classes**: OAuth audit checklist is owned by `go-oauth-auditor`; OAuth data-flow bugs (scope tampering, redirect_uri tampering, etc.) are owned by `go-taint-tracer` invoked on auditor-emitted `taint_pairs` — D6.

## Locked Decisions (D1–D18)

These 18 entries are LOCKED. Source: HAND_OFF.md §4 (orchestrator-promoted to LOCKED during synthesis). They override any default behavior. Unlocking any one of them requires explicit user request.

<decisions locked="true" source="HAND_OFF.md §4">

| ID | Decision | Scope | Rationale |
|----|----------|-------|-----------|
| **D1** | Per-stack agent definitions, not parameterized | agent definition layout, language coverage | Go source/sink/sanitizer tables are too large to combine across stacks; per-stack files easier to maintain. |
| **D2** | Hooks written in Go, not Python | `claude-security-hooks` binary | Cold-start latency. Python startup compounds across parallel fan-outs. Static Go binary starts in <5ms. |
| **D3** | Semgrep and OpenGrep are containerized | scanner runtime | Reduces host pollution, simplifies CI parity, user preference. |
| **D4** | Single `opengrep-mcp` server fronts both engines via tier parameter | MCP scanner surface | One tool (`scan_with_rule`) with a `tier` arg, not two. Engine selection becomes a runtime decision. |
| **D5** | Task tool's PostToolUse hook validates outputs, NOT SubagentStop | hook registration | `SubagentStop` payload doesn't identify which subagent finished when multiple run in parallel. See anthropics/claude-code#7881. |
| **D6** | OAuth bug classes split across auditor (checklist) and taint tracer (flow analysis) | OAuth analysis pipeline | Scope-tampering is a data-flow bug, not a checklist item. Separation keeps each agent's prompt focused. |
| **D7** | RFC 6819 dropped from audit checklist | OAuth checklist taxonomy | Superseded by RFC 9700. |
| **D8** | OAuth 2.1 draft included as a target profile | `go-oauth-auditor` `target_profile` enum | OAuth 2.1 draft-15 (March 2026) obsoletes 6749/6750 once finalized; auditor must speak both regimes. |
| **D9** | CIBA included as part of OIDC family coverage | OAuth/OIDC checklist coverage | User explicitly requested. CIBA introduces decoupled-flow attack surface. |
| **D10** | `gosec` and `staticcheck` dropped from pre-pass; `govulncheck` kept | cartographer pre-pass tooling | gosec/staticcheck redundant with Semgrep rules; govulncheck is SSA-based reachability against Go vuln DB, irreplaceable. |
| **D11** | Cartographer sequential, tracers parallel | review-time orchestration shape | Cartographer output is small and consumed by every tracer; redundant per-tracer discovery would waste tokens. |
| **D12** | Tool allowlists deliberately omit cheap alternatives | agent frontmatter `tools:` lists | If `Grep` is in the allowlist, the agent uses it instead of LSP `references`. Removing it is the cheapest enforcement mechanism. |
| **D13** | Verdict schemas are strict JSON, prose disallowed | every agent's Output schema | Orchestrator and synthesis must parse mechanically. Prose breaks deduplication. |
| **D14** | "Ambiguous" verdict encouraged over false confidence | verdict policy | False sanitized verdicts create blind spots. Honest uncertainty is the bias. |
| **D15** | Skills NOT used for security review | plugin packaging | Skills load into main agent; instructions silently disappear across subagent boundary. This is the bug the design avoids. |
| **D16** | Invariant checker does NOT generate invariants | `invariant-checker` responsibility | Discovery is a harder problem, out of scope for v1. |
| **D17** | Per-stack expansion is sequential, Go first | project roadmap shape | Cross-stack abstraction before having one stack working produces leaky abstractions. |
| **D18** | Synthesis agent reads from a directory of JSON files | synthesis I/O contract | Decouples synthesis from in-flight Task results. Simpler than streaming state. |

</decisions>

## Key Decisions

Operational / engineering decisions made during ingest. Architectural decisions live in the locked block above.

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Mirror HAND_OFF.md §6 as roadmap phases | Phase plan is already commit-sized and sequenced; faithful mirroring preserves the designer's intent | — Pending |
| Per-agent assertions (51 of them) are Phase 2 REQs, not Phase 3 | Assertions become Go check predicates + unit tests in the hooks binary; that's where they live as code. Agent .md files in Phase 3 deliver the prose that makes the assertions true at runtime; end-to-end verification of agent behavior happens at Phase 5. | — Pending |
| Open questions Q1/Q4/Q9 are first-day work items inside Phases 2–4, not separate phases | They are integration details (model identifier, MCP tool names, OpenGrep build pin) that depend on environment specifics; gating the roadmap on them would inflate phase count. | — Pending |
| Two Go modules in one repo, not two repos | Lower friction for shared schema types, single PR surface during v1, easier smoke-test wiring | — Pending |
| `/security-review` slash command (not orchestrator agent) for v1 | Q7 in HAND_OFF §5. Simpler, more transparent, easier to debug. | — Pending |

---
*Last updated: 2026-05-17 after new-project ingest from HAND_OFF.md*
