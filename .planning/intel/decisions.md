# Decisions Intel

Locked design decisions lifted from HAND_OFF.md §4 (Decisions log). Per ingest instruction, every entry below is treated as **LOCKED** for downstream roadmapping. The source document is `SPEC`-classified but its §4 carries ADR-grade decisions that the project must not silently override.

source-doc: /home/saghaulor/code/security_reviewer/HAND_OFF.md
section: §4 Decisions log (D1–D18)
status: locked
precedence: 0 (manifest override)

---

## D1 — Per-stack agent definitions, not parameterized

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D1)
- status: locked
- scope: agent definition layout, language coverage
- decision: Agent definitions are per-stack (Go-first slice). No cross-stack parameterized agent files.
- rationale: Go source/sink/sanitizer tables are large enough that combining stacks blows up prompt size and degrades accuracy. Per-stack files are easier to maintain.

## D2 — Hooks written in Go, not Python

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D2)
- status: locked
- scope: `claude-security-hooks` binary
- decision: Lifecycle hooks are implemented as a static Go binary, not Python scripts.
- rationale: Cold-start latency. Python interpreter startup compounds across parallel fan-outs. Static Go binary starts in <5ms. Aligns with user preference for compiled binaries.

## D3 — Semgrep and OpenGrep are containerized

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D3)
- status: locked
- scope: scanner runtime
- decision: Both Semgrep (Pro/CE) and OpenGrep run in Docker containers, never on the host.
- rationale: Reduces host pollution, simplifies CI parity, user preference.

## D4 — Single `opengrep-mcp` server fronts both engines via tier parameter

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D4)
- status: locked
- scope: MCP scanner surface
- decision: One MCP server exposes a single `scan_with_rule` tool with a `tier` parameter (`pro|intrafile|ce`). Agents do not select engines directly.
- rationale: Simpler agent prompt (one tool, not two). Engine selection becomes a runtime decision.

## D5 — Task tool's PostToolUse hook validates outputs, NOT SubagentStop

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D5)
- status: locked
- scope: hook registration in `.claude/settings.json`
- decision: Output validation hooks register against `PostToolUse` on the `Task` tool, not against `SubagentStop`.
- rationale: `SubagentStop` payload doesn't identify which subagent finished when multiple run in parallel. `Task` tool hooks give per-invocation routing. See anthropics/claude-code#7881.

## D6 — OAuth bug classes split across auditor (checklist) and taint tracer (flow analysis)

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D6)
- status: locked
- scope: OAuth analysis pipeline
- decision: `go-oauth-auditor` runs the RFC conformance checklist and emits taint-pair candidates; `go-taint-tracer` verifies the actual data-flow exploitability of those pairs.
- rationale: Scope-tampering bugs are data-flow bugs, not checklist items. Separation of concerns keeps each agent's prompt focused.

## D7 — RFC 6819 dropped from audit checklist

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D7)
- status: locked
- scope: OAuth checklist taxonomy
- decision: RFC 6819 is not part of the auditor's checklist.
- rationale: Superseded by RFC 9700.

## D8 — OAuth 2.1 draft included as a target profile

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D8)
- status: locked
- scope: `go-oauth-auditor` `target_profile` enum
- decision: The auditor accepts `target_profile ∈ {oauth_2_1, oauth_2_0, oauth_2_0_with_9700_bcp}` and handles each regime.
- rationale: OAuth 2.1 draft-15 (March 2026) obsoletes 6749/6750 once finalized; the auditor must speak both regimes during transition.

## D9 — CIBA included as part of OIDC family coverage

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D9)
- status: locked
- scope: OAuth/OIDC checklist coverage
- decision: CIBA Core 1.0 checks are first-class items in the auditor's checklist.
- rationale: User explicitly requested. CIBA introduces decoupled-flow attack surface not covered by baseline OAuth.

## D10 — gosec and staticcheck dropped from pre-pass; govulncheck kept

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D10)
- status: locked
- scope: cartographer pre-pass tooling
- decision: `gosec` and `staticcheck` are NOT invoked. `govulncheck` is REQUIRED in the cartographer pre-pass.
- rationale: gosec/staticcheck are redundant with custom Semgrep rules. `govulncheck` is SSA-based reachability against the Go vuln DB and cannot be replicated by Semgrep patterns.

## D11 — Cartographer sequential, tracers parallel

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D11)
- status: locked
- scope: review-time orchestration shape
- decision: `go-cartographer` runs once sequentially. All tracers fan out in parallel from its output.
- rationale: Cartographer output is small and consumed by every tracer; redundant per-tracer discovery would waste tokens. Tracers are independent and parallelizable.

## D12 — Tool allowlists deliberately omit cheap alternatives

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D12)
- status: locked
- scope: agent frontmatter `tools:` lists
- decision: Tracer agent allowlists exclude `Grep`, `Bash` (except where strictly required), `Edit`, `Write`. Cheap text-search alternatives are removed so agents are forced to LSP/Semgrep paths.
- rationale: If `Grep` is in the allowlist, the agent will use it instead of LSP `references`. Removing it is the cheapest enforcement mechanism.

## D13 — Verdict schemas are strict JSON, prose disallowed

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D13)
- status: locked
- scope: every agent's `Output schema`
- decision: All agent outputs conform to a strict JSON schema. Free-form prose verdicts are forbidden.
- rationale: Orchestrator and synthesis agent must parse outputs mechanically. Prose verdicts force English parsing, which scales badly and breaks deduplication.

## D14 — "Ambiguous" verdict encouraged over false confidence

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D14)
- status: locked
- scope: verdict policy
- decision: Agents must return `ambiguous` (or `unverifiable`, `unverified`) when honest. Synthetic confidence is forbidden.
- rationale: False sanitized verdicts create blind spots. Honest uncertainty is the bias.

## D15 — Skills NOT used for security review

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D15)
- status: locked
- scope: plugin packaging strategy
- decision: The security-review entrypoint is a slash command (or manually-invoked workflow), not a Claude Code skill.
- rationale: Skills load into the main agent. Subagents start with empty context; instructions in a skill silently disappear across the boundary. This is exactly the bug the design avoids.

## D16 — Invariant checker does NOT generate invariants

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D16)
- status: locked
- scope: `invariant-checker` agent responsibility
- decision: The invariant-checker verifies caller-supplied invariants only. It does not discover invariants.
- rationale: Discovery is a harder problem and out of scope for v1.

## D17 — Per-stack expansion is sequential, Go first

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D17)
- status: locked
- scope: project roadmap shape
- decision: Ship the Go slice end-to-end before adding another language. The shared layer is JSON schemas and hook validators; agents themselves remain per-stack.
- rationale: User confirmed. Building cross-stack abstraction before having one stack working produces leaky abstractions.

## D18 — Synthesis agent reads from a directory of JSON files

- source: /home/saghaulor/code/security_reviewer/HAND_OFF.md §4 (row D18)
- status: locked
- scope: synthesis I/O contract
- decision: The orchestrator dumps each specialist's verdict JSON to disk in a single directory. The synthesis agent reads the whole directory at completion.
- rationale: Decouples synthesis from in-flight Task results. Simpler than streaming state.
