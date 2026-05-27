# Roadmap: security-reviewer

## Overview

Six phases that mirror HAND_OFF.md §6 exactly. The journey: scaffold the repo with two parallel Go modules (`claude-security-hooks` and `opengrep-mcp`); build the hooks binary with all 51 per-agent assertion check predicates and their unit-test suite (Phase 2's exit criterion); author the six per-stack agent definition files and wire them to the hooks via `.claude/settings.json`; ship the dual-engine MCP scanner server with its OpenGrep container; converge the two modules at an end-to-end smoke test against a deliberately vulnerable Go service containing SQLi + authz bypass + OAuth scope-tampering bugs (all three flagged with non-ambiguous verdicts is Phase 5's exit criterion); finish with top-level + per-component READMEs and a `CONTRIBUTING.md`. v1 ships Go-only per D17.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Repo scaffold** — Two-module Go monorepo with `.claude/` skeleton, gitignore, license, top-level README stub
- [ ] **Phase 2: claude-security-hooks binary** — Static Go validator binary; per-agent schema structs + invariant predicates + unit tests; 51-assertion test suite passes
- [ ] **Phase 3: Agent definitions + hook registration** — Six `.claude/agents/*.md` files + `.claude/settings.json` hook registrations + `/security-review` slash command
- [ ] **Phase 4: opengrep-mcp server + OpenGrep container** — MCP server with `scan_with_rule`/`scan_directory`/`get_ast`; tier-dispatched containerized scanners
- [ ] **Phase 5: End-to-end smoke test** — `examples/sample-vulnerable-service/` with 3 planted bugs; full review workflow runs; all 3 bugs flagged non-ambiguously
- [ ] **Phase 6: Documentation** — Top-level README + per-component READMEs + `CONTRIBUTING.md`
- [x] **Phase 11: codegraph migration** — Replace graphify with codegraph as the graph/MCP layer in the security-review pipeline *(completed 2026-05-27)*
- [ ] **Phase 12: code review skill evaluation** — Evaluate two third-party code review skills against the sample vulnerable service to determine if they add detection coverage or complementary value to the existing security-review pipeline
- [ ] **Phase 13: pipeline reliability + bootstrap hardening** — Eliminate all 7 failure modes from the first full scan run; implement pre-flight bootstrap script; reach ≥95% first-attempt success rate for fully autonomous pipeline execution

## Phase Details

### Phase 1: Repo scaffold
**Goal**: A two-module Go monorepo exists with the `.claude/` directory skeleton, ready for hook and agent files. Both modules build empty (no functionality yet) under `CGO_ENABLED=0`.
**Depends on**: Nothing (first phase)
**Requirements**: (none — pre-implementation support phase)
**Success Criteria** (what must be TRUE):
  1. Two Go modules exist: `claude-security-hooks/go.mod` and `opengrep-mcp/go.mod`, each with a working `cmd/<name>/main.go` that builds with `CGO_ENABLED=0 go build` and exits 0 when run with no args.
  2. `.claude/agents/` and `.claude/hooks/bin/` and `.claude/security-invariants/` directories exist (may be empty placeholders with `.gitkeep`).
  3. `.gitignore` excludes `bin/`, `*.test`, `.idea/`, `.vscode/`, scanner container output dirs (`graphify-out/`, `review-report.json`, etc.).
  4. `LICENSE` file and stub top-level `README.md` (placeholder title + one-paragraph description) are committed.
  5. First commit is on the trunk branch with message matching the HAND_OFF.md §6 convention (`scaffold: initial project structure`).
**Plans**: 6 plans across 2 waves

Plans:
- [ ] 13-01-PLAN.md — Wave 0: TDD RED — failing tests for W4 (validate.go) + W7 (schema.Handler)
- [ ] 13-02-PLAN.md — Wave 1a: W1 Bootstrap pre-flight script + Makefile preflight target
- [ ] 13-03-PLAN.md — Wave 1b: W2+W3 — JSON serialization enforcement + agent Write permissions
- [ ] 13-04-PLAN.md — Wave 1c: W4+W5+W7 GREEN — validate.go fix + preflight.go schema doc + schema.Handler fields
- [ ] 13-05-PLAN.md — Wave 1d: W5 — Agent input schema JSON files (6 specs/agents/*.schema.json)
- [ ] 13-06-PLAN.md — Wave 2: W6+W7 — Cartographer route cross-reference + security-review.md first_param_read

### Phase 2: claude-security-hooks Go binary
**Goal**: A static `claude-security-hooks` binary exists at `claude-security-hooks/bin/claude-security-hooks` that implements `preflight`, `validate`, and `inject-context` subcommands and validates every specialist agent's JSON verdict against the 51 per-agent assertion check predicates. The Go unit-test suite for those predicates passes.
**Depends on**: Phase 1
**Requirements**: REQ-hooks-H1, REQ-hooks-H2, REQ-hooks-H3, REQ-hooks-H4, REQ-hooks-H5, REQ-hooks-H6, REQ-hooks-H7, REQ-cartographer-A1, REQ-cartographer-A2, REQ-cartographer-A3, REQ-cartographer-A4, REQ-cartographer-A5, REQ-cartographer-A6, REQ-cartographer-A7, REQ-cartographer-A8, REQ-cartographer-A9, REQ-cartographer-A10, REQ-cartographer-A11, REQ-taint-T1, REQ-taint-T2, REQ-taint-T3, REQ-taint-T4, REQ-taint-T5, REQ-taint-T6, REQ-taint-T7, REQ-taint-T8, REQ-taint-T9, REQ-taint-T10, REQ-taint-T11, REQ-authz-AZ1, REQ-authz-AZ2, REQ-authz-AZ3, REQ-authz-AZ4, REQ-authz-AZ5, REQ-authz-AZ6, REQ-oauth-OA1, REQ-oauth-OA2, REQ-oauth-OA3, REQ-oauth-OA4, REQ-oauth-OA5, REQ-oauth-OA6, REQ-oauth-OA7, REQ-invariant-IC1, REQ-invariant-IC2, REQ-invariant-IC3, REQ-invariant-IC4, REQ-synthesis-S1, REQ-synthesis-S2, REQ-synthesis-S3, REQ-synthesis-S4, REQ-synthesis-S5, REQ-synthesis-S6
**First-day work item (Q9 from HAND_OFF §5)**: Verify the current Claude Sonnet model identifier against https://docs.claude.com before encoding it into any agent frontmatter or test fixture. Default is `claude-sonnet-4-6`; confirm or update. (Resolved during planning 2026-05-19 — `claude-sonnet-4-6` is current.)
**Success Criteria** (what must be TRUE):
  1. `CGO_ENABLED=0 go build -o bin/claude-security-hooks ./cmd/claude-security-hooks` from inside `claude-security-hooks/` produces a single static binary with no dynamic library dependencies (verified by `ldd` showing "not a dynamic executable" or equivalent).
  2. `go test ./...` from inside `claude-security-hooks/` passes; every check predicate corresponding to assertions A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6 has at least one passing unit test (the 51-assertion unit-test suite — Phase 2's primary exit criterion).
  3. The compiled binary exits 0 with no stdout when given a `PostToolUse` event JSON whose `tool_input.subagent_type` is NOT in the security agent set (REQ-hooks-H3 verified end-to-end).
  4. The compiled binary emits exactly one `{"decision":"block","reason":"..."}` JSON object and exits 0 when given a `PostToolUse` event whose `tool_response.content` is malformed JSON (REQ-hooks-H7 verified end-to-end).
  5. The core validator path imports only Go stdlib (`go list -deps ./internal/invariants/... ./internal/schema/...` shows no external module paths); test framework deps confined to `_test.go` files (REQ-hooks-H5).
**Plans**: 8 plans across 3 waves

Plans:
- [ ] 02-01-PLAN.md — Wave 0: Test infrastructure, shared types, schema scaffolding, meta-tests
- [ ] 02-02-PLAN.md — Wave 1: Cartographer schema + A1–A11 predicates + tests
- [ ] 02-03-PLAN.md — Wave 1: Taint tracer schema + T1–T11 predicates + tests
- [ ] 02-04-PLAN.md — Wave 1: Authz tracer schema + AZ1–AZ6 predicates + tests
- [ ] 02-05-PLAN.md — Wave 1: OAuth auditor schema + OA1–OA7 predicates + tests
- [ ] 02-06-PLAN.md — Wave 1: Invariant checker schema + IC1–IC4 predicates + tests
- [ ] 02-07-PLAN.md — Wave 1: Synthesis schema + S1–S6 predicates + tests
- [ ] 02-08-PLAN.md — Wave 2: Subcommand wiring (main.go + preflight/validate/inject) + integration tests + Makefile + H1/H2/H3/H4/H7 hardening

### Phase 3: Agent definitions + hook registration
**Goal**: Six agent definition files exist under `.claude/agents/` conforming to the common prompt-shape contract, `.claude/settings.json` registers the three hook events against `claude-security-hooks`, and a `/security-review` slash command is in place to drive the workflow. The hooks binary from Phase 2 validates these agent outputs at runtime without needing additional configuration.
**Depends on**: Phase 2
**Requirements**: REQ-agents-prompt-shape
**First-day work item (Q1 from HAND_OFF §5)**: Verify the exact Graphify MCP tool schemas/names by running `python -m graphify.serve --help` and probing the MCP handshake. Update the `go-cartographer` allowlist if names differ from `{query_graph, get_node, get_neighbors, shortest_path}`.
**Success Criteria** (what must be TRUE):
  1. Six files exist: `.claude/agents/go-cartographer.md`, `.claude/agents/go-taint-tracer.md`, `.claude/agents/go-authz-tracer.md`, `.claude/agents/go-oauth-auditor.md`, `.claude/agents/invariant-checker.md`, `.claude/agents/synthesis.md`. Each has YAML frontmatter declaring `name`, `description`, `model`, `tools` (and `allowed_commands` where applicable) per CON-tool-allowlists, and a body containing the six-section structure required by REQ-agents-prompt-shape.
  2. The `go-taint-tracer.md` body embeds the §8.2 Go source/sink/sanitizer catalog verbatim AND the §3.2.2 OAuth source/sink table; the `go-oauth-auditor.md` body embeds the §8.3 checklist taxonomy with stable IDs.
  3. `.claude/settings.json` registers three hooks per CON-hook-registration: `PreToolUse` matcher `Task` → `claude-security-hooks preflight` (5s timeout); `PostToolUse` matcher `Task` → `claude-security-hooks validate` (15s timeout); `SubagentStart` matcher `go-taint-tracer|go-authz-tracer|go-oauth-auditor|go-cartographer|invariant-checker|synthesis` → `claude-security-hooks inject-context` (5s timeout). `SubagentStop` is NOT registered (D5).
  4. A `/security-review` slash command definition exists (location and shape per Claude Code slash-command convention) that drives the full review workflow: pre-pass artifacts → cartographer → tracer fan-out → synthesis.
  5. No tracer agent frontmatter contains `Grep`, `Bash`, `Edit`, or `Write` in its `tools:` list (D12 negative check, automatable via grep over the agent files at CI time).
**Plans**: 6 plans across 2 waves

Plans:
- [ ] 13-01-PLAN.md — Wave 0: TDD RED — failing tests for W4 (validate.go) + W7 (schema.Handler)
- [ ] 13-02-PLAN.md — Wave 1a: W1 Bootstrap pre-flight script + Makefile preflight target
- [ ] 13-03-PLAN.md — Wave 1b: W2+W3 — JSON serialization enforcement + agent Write permissions
- [ ] 13-04-PLAN.md — Wave 1c: W4+W5+W7 GREEN — validate.go fix + preflight.go schema doc + schema.Handler fields
- [ ] 13-05-PLAN.md — Wave 1d: W5 — Agent input schema JSON files (6 specs/agents/*.schema.json)
- [ ] 13-06-PLAN.md — Wave 2: W6+W7 — Cartographer route cross-reference + security-review.md first_param_read
**UI hint**: no (this phase produces config and prompt files only; no end-user UI surface)

### Phase 4: opengrep-mcp server + OpenGrep container
**Goal**: A static `opengrep-mcp` binary exists that exposes a single MCP server with three tools (`scan_with_rule`, `scan_directory`, `get_ast`), dispatches to the correct containerized engine (Semgrep Pro / OpenGrep intrafile / OpenGrep CE) via a `tier` parameter, mounts workspaces read-only, kills containers on timeout, and never logs Pro tokens. The OpenGrep container image is buildable from the repo Dockerfile.
**Depends on**: Phase 1 (parallel with Phases 2–3; converges into Phase 5)
**Requirements**: REQ-mcp-O1, REQ-mcp-O2, REQ-mcp-O3, REQ-mcp-O4, REQ-mcp-O5, REQ-mcp-O6, REQ-mcp-O7, REQ-mcp-O8
**First-day work item (Q4 from HAND_OFF §5)**: Pin a specific OpenGrep release tag and verify the build commands in the §3.9 Dockerfile sketch against the upstream README at https://github.com/opengrep/opengrep. The §3.9 sketch is incomplete.
**Success Criteria** (what must be TRUE):
  1. `CGO_ENABLED=0 go build -o bin/opengrep-mcp ./cmd/opengrep-mcp` from inside `opengrep-mcp/` produces a working static binary (REQ-mcp-O1) using `github.com/modelcontextprotocol/go-sdk` and `github.com/docker/docker/client`.
  2. An MCP client probing the server's tool list receives exactly three registrations — `scan_with_rule`, `scan_directory`, `get_ast` — each with a JSON Schema validating against the MCP spec (REQ-mcp-O2). Calling `scan_with_rule` with `tier: "bogus"` returns a structured error, not a panic (REQ-mcp-O3).
  3. The OpenGrep Dockerfile in the repo builds successfully against a pinned upstream tag (Q4 resolved); first invocation of `scan_with_rule` with the CE tier pulls and caches the container image locally (REQ-mcp-O4); the container mount inspector (`docker inspect`) shows `/src` mounted with `ro` flag (REQ-mcp-O5).
  4. An integration test that requests `timeout_seconds: 1` against a deliberately slow rule kills the container and returns a partial-results error (REQ-mcp-O6); a separate test confirms server output uses the normalized `internal/schema/findings.go` shape regardless of which engine produced raw findings (REQ-mcp-O7).
  5. Running the server with `SEMGREP_APP_TOKEN=test-secret-do-not-log` in the environment and grepping the server's stdout/stderr output across a full Pro-tier scan invocation produces zero matches for `test-secret-do-not-log` (REQ-mcp-O8).
**Plans**: 6 plans across 2 waves

Plans:
- [ ] 13-01-PLAN.md — Wave 0: TDD RED — failing tests for W4 (validate.go) + W7 (schema.Handler)
- [ ] 13-02-PLAN.md — Wave 1a: W1 Bootstrap pre-flight script + Makefile preflight target
- [ ] 13-03-PLAN.md — Wave 1b: W2+W3 — JSON serialization enforcement + agent Write permissions
- [ ] 13-04-PLAN.md — Wave 1c: W4+W5+W7 GREEN — validate.go fix + preflight.go schema doc + schema.Handler fields
- [ ] 13-05-PLAN.md — Wave 1d: W5 — Agent input schema JSON files (6 specs/agents/*.schema.json)
- [ ] 13-06-PLAN.md — Wave 2: W6+W7 — Cartographer route cross-reference + security-review.md first_param_read

### Phase 5: End-to-end smoke test
**Goal**: `examples/sample-vulnerable-service/` exists as a small Go HTTP service containing exactly three planted security bugs — a SQL injection, an authorization bypass, and an OAuth scope-tampering vulnerability — and running the full `/security-review` workflow against it produces a `review-report.json` whose `findings` array contains all three bugs with non-ambiguous verdicts (`exploitable`, not `ambiguous` or `unverifiable`).
**Depends on**: Phase 2 (hooks binary), Phase 3 (agent definitions + slash command), Phase 4 (opengrep-mcp + OpenGrep container)
**Requirements**: (none from the 61-REQ set — the exit criterion is the 3-bug detection, which is a system-level integration property rather than a per-agent assertion)
**Success Criteria** (what must be TRUE):
  1. `examples/sample-vulnerable-service/` is a buildable Go HTTP service (`go build ./...` succeeds) that includes (a) at least one handler executing a raw SQL query with user-controlled input concatenated into the query string, (b) at least one route whose handler performs a sensitive action without invoking a blocking authz primitive, and (c) at least one OAuth consent flow where the granted scopes are read from the consent-form POST body rather than from the original authorization request (the motivating bug from HAND_OFF §1).
  2. Pre-pass artifacts produced cleanly: `graphify build` against `examples/sample-vulnerable-service/` produces `graphify-out/graph.json` without errors; `govulncheck -json ./...` produces `govulncheck.json` without errors.
  3. Invoking the `/security-review` slash command against `examples/sample-vulnerable-service/` runs to completion: cartographer emits a valid `go-index.json` (Phase 2 hook validates), tracers fan out in parallel (Phase 2 hook validates each verdict), synthesis runs after fan-out (Phase 2 hook validates the report), and no hook returns a `decision:block`.
  4. The resulting `review-report.json` validates against the `review-report/v1` schema and contains exactly three findings whose `(class, file, line)` triples correspond to the three planted bugs (one `class: "injection"` for the SQLi, one `class: "authz"` for the bypass, one `class: "oauth"` for the scope-tampering).
  5. None of the three findings has `confidence: "low"` or a verdict synonym of "ambiguous"/"unverifiable"; each is reported with a concrete data-flow path (for the SQLi and scope-tampering) or a concrete missing-primitive citation (for the authz bypass) per the user's stated success metric.
**Plans**: 6 plans across 2 waves

Plans:
- [ ] 13-01-PLAN.md — Wave 0: TDD RED — failing tests for W4 (validate.go) + W7 (schema.Handler)
- [ ] 13-02-PLAN.md — Wave 1a: W1 Bootstrap pre-flight script + Makefile preflight target
- [ ] 13-03-PLAN.md — Wave 1b: W2+W3 — JSON serialization enforcement + agent Write permissions
- [ ] 13-04-PLAN.md — Wave 1c: W4+W5+W7 GREEN — validate.go fix + preflight.go schema doc + schema.Handler fields
- [ ] 13-05-PLAN.md — Wave 1d: W5 — Agent input schema JSON files (6 specs/agents/*.schema.json)
- [ ] 13-06-PLAN.md — Wave 2: W6+W7 — Cartographer route cross-reference + security-review.md first_param_read

### Phase 6: Documentation
**Goal**: A reader landing on the repo can understand what the system does, how to run a security review, how each component fits together, and how to contribute a new agent or extend the source/sink/checklist catalog — without needing to read HAND_OFF.md.
**Depends on**: Phase 5
**Requirements**: (none — documentation phase)
**Success Criteria** (what must be TRUE):
  1. Top-level `README.md` (replacing the Phase 1 stub) describes what the system is, the three-layer architecture (pre-pass → cartographer → fan-out → synthesis), the slash-command entry point, and how to install and run a review against a target Go repo.
  2. `claude-security-hooks/README.md` documents the three subcommands, the per-agent invariant catalog, the build/install workflow (`make build|test|install`), and the assertion-to-check-predicate mapping (so future contributors can find the test for any A/T/AZ/OA/IC/S assertion).
  3. `opengrep-mcp/README.md` documents the three MCP tools, the `tier` parameter semantics, the `SEMGREP_APP_TOKEN` env contract, and how to rebuild the OpenGrep container against a different upstream tag.
  4. `.claude/agents/README.md` (or equivalent) documents the per-agent file structure and the tool-allowlist policy (especially the negative rule that tracer allowlists must not contain `Grep`/`Bash`/`Edit`/`Write`).
  5. `CONTRIBUTING.md` describes how to add a new specialist agent (e.g., for a future stack), how to extend the source/sink catalog in §8.2, and the TDD-with-tests-before-implementation convention.
**Plans**: 6 plans across 2 waves

Plans:
- [ ] 13-01-PLAN.md — Wave 0: TDD RED — failing tests for W4 (validate.go) + W7 (schema.Handler)
- [ ] 13-02-PLAN.md — Wave 1a: W1 Bootstrap pre-flight script + Makefile preflight target
- [ ] 13-03-PLAN.md — Wave 1b: W2+W3 — JSON serialization enforcement + agent Write permissions
- [ ] 13-04-PLAN.md — Wave 1c: W4+W5+W7 GREEN — validate.go fix + preflight.go schema doc + schema.Handler fields
- [ ] 13-05-PLAN.md — Wave 1d: W5 — Agent input schema JSON files (6 specs/agents/*.schema.json)
- [ ] 13-06-PLAN.md — Wave 2: W6+W7 — Cartographer route cross-reference + security-review.md first_param_read

### Phase 9: opengrep-mcp server + infrastructure hardening
**Goal**: Integrate the standalone `opengrep-mcp` binary into the security-review pipeline. The standalone binary already exists at `/home/saghaulor/code/opengrep-mcp` (Phase 5 complete, tested). Work in this repo is integration-only: top-level Makefile with build/verify targets, `.mcp.json` switched from SSE to stdio, and orchestration command Step 3 (Docker bootstrap) removed.
**Depends on**: Phase 8 (automated E2E testing framework in place; MCP config skeleton committed)
**Requirements**: REQ-mcp-O1, REQ-mcp-O2, REQ-mcp-O3, REQ-mcp-O4, REQ-mcp-O5, REQ-mcp-O6, REQ-mcp-O7, REQ-mcp-O8
**Success Criteria** (what must be TRUE):
  1. `make build-opengrep-mcp` from the security_reviewer root produces a static binary at `.claude/hooks/bin/opengrep-mcp` by delegating to the standalone repo's build system.
  2. `make verify-opengrep-mcp` passes: binary exists, is executable, and is statically linked.
  3. `.mcp.json` opengrep entry uses `type: "stdio"` pointing at `.claude/hooks/bin/opengrep-mcp` with `SAST_ENGINE=opengrep` env var; no SSE/localhost:8000 references remain.
  4. `.claude/commands/security-review.md` Steps are numbered 0–8 with no gaps; no Docker health-check or container bootstrap logic remains in the file.
  5. Claude Code can discover and spawn opengrep-mcp as a stdio subprocess — the three MCP tools (scan_with_rule, scan_directory, get_ast) are accessible to agents via `mcp__opengrep__*` prefix.
**Plans**: 2 plans across 2 waves

Plans:
- [x] 09-01-PLAN.md — Wave 1: TDD verify target (RED) + remove Step 3 from orchestration command
- [x] 09-02-PLAN.md — Wave 2: build-opengrep-mcp implementation (GREEN) + .mcp.json stdio wiring

### Phase 10: Hook compatibility + code-ref schema
**Goal**: Unblock the Phase 5 smoke test by resolving four runtime blockers discovered during the first live `/security-review` run, and add a `code_ref` field to all review artifacts so they can be correlated to the exact code version they analyzed.
**Depends on**: Phase 9
**Blockers addressed**: B1 (Agent tool extra fields rejected by hooks), B2 (T6 blocks honest `ambiguous` verdicts), B3 (taint tracer writes absolute paths that fail T9), B4 (stale hooks binary not rebuilt before reviews)
**Schema enhancement**: E1 — `code_ref` + `code_ref_dirty` on `CartographerIndex` and `SynthesisReport`
**Success Criteria** (what must be TRUE):
  1. `go test ./...` passes in `claude-security-hooks/` including all new Wave 0 tests (B1, B2, E1).
  2. An Agent call with `run_in_background: true` or `model: "sonnet"` is NOT blocked by the preflight hook.
  3. A `TaintVerdict` with `verdict="ambiguous"`, `semgrep.ran=false`, `gopls.references_calls=0` passes T6.
  4. `CartographerIndex` and `SynthesisReport` schema structs contain `code_ref` and `code_ref_dirty` fields and round-trip correctly.
  5. `go-taint-tracer.md` §6 Hard Rules contains an explicit workspace-relative path requirement (T9-PATH).
  6. `security-review.md` contains a binary freshness step (Step 1.5) that runs `make install` before any agent dispatch.
  7. `go-cartographer.md` instructs code_ref computation; `synthesis.md` instructs propagation into `review-report.json`.
  8. A full `/security-review examples/sample-vulnerable-service` run completes without any hook-driven block on B1–B4.
**Plans**: 3 plans across 3 waves

Plans:
- [x] 10-01-PLAN.md — Wave 0: Failing tests (RED) — events_test, taint_tracer_test, schema round-trip
- [x] 10-02-PLAN.md — Wave 1: Go implementation (GREEN) — TaskToolInput, checkT6, schema fields, binary rebuild
- [x] 10-03-PLAN.md — Wave 2: Agent defs + skill — T9-PATH rule, cartographer code_ref, synthesis propagation, freshness check

### Phase 11: codegraph migration
**Goal**: Replace graphify with codegraph as the graph/MCP layer in the security-review pipeline. graphify requires an LLM API key and the `anthropic` pip package, which breaks on Bedrock and adds external dependencies. codegraph is fully local (no API key), has richer security-analysis primitives (`codegraph_trace`, `codegraph_callers`, native route node kind), and is already pulled at `/home/saghaulor/code/codegraph`.
**Depends on**: Phase 10
**Requirements**: (migration — no new REQ entries; replaces graphify integration points end-to-end)
**Success Criteria** (what must be TRUE):
  1. `scripts/graphify-mcp.sh` is deleted or replaced with an equivalent `scripts/codegraph-mcp.sh` that runs `codegraph serve --mcp --path TARGET_DIR`.
  2. `.mcp.json` graphify entry is replaced with a codegraph entry using the stdio transport and `mcp__codegraph__*` tool prefix; no `mcp__graphify__*` references remain.
  3. `go-cartographer.md` tools list is updated: `mcp__graphify__*` tools replaced with `mcp__codegraph__*` equivalents; prompt body updated to use `codegraph_trace` and `codegraph_callers` where applicable.
  4. `security-review.md` Step 3 updated from `graphify build` to `codegraph index` (or equivalent); no graphify invocation remains in the orchestration command.
  5. `go-taint-tracer.md` optionally updated to reference `codegraph_trace` for data-flow confirmation.
  6. `.gitignore` of `examples/sample-vulnerable-service/` excludes `.codegraph/` (the output directory codegraph writes into the target project).
  7. A full `/security-review examples/sample-vulnerable-service` run completes with codegraph as the graph source; `review-report.json` still contains findings with non-ambiguous verdicts.
**Plans**: 2 plans across 2 waves

Plans:
- [ ] 11-01-PLAN.md — Wave 1: Create codegraph-mcp.sh, delete graphify-mcp.sh, update .mcp.json and sample service .gitignore
- [ ] 11-02-PLAN.md — Wave 2: Update go-cartographer.md (tools + body) and security-review.md (Step 2+3)

### Phase 12: code review skill evaluation
**Goal**: Determine whether two third-party code review skills — the project-local `/security-review` (this project's own pipeline) and at least one external skill (e.g. the GSD built-in `/code-review` or `/ultrareview`) — detect the three planted bugs in `examples/sample-vulnerable-service/` and whether either adds coverage, speed, or ergonomic value that the custom pipeline does not already provide. The output is a brief decision record (ADR-style) recommending whether to integrate, defer, or discard each evaluated skill.
**Depends on**: Phase 11
**Requirements**: (evaluation phase — no new REQ entries)
**Success Criteria** (what must be TRUE):
  1. Two code review skills are identified and named in the evaluation plan (one must be an external/third-party skill not authored in this repo).
  2. Each skill is run against `examples/sample-vulnerable-service/` and its output is captured (findings list, confidence levels, runtime).
  3. A comparison table exists mapping each skill's findings against the three planted bugs (SQLi, authz bypass, OAuth scope-tampering) and against each other.
  4. Each evaluated skill receives a verdict: **integrate** (add to pipeline), **complement** (use alongside but separately), or **discard** (no unique value).
  5. A `docs/skill-eval-2026-05-27.md` decision record is committed summarising methodology, findings table, verdicts, and rationale.
**Plans**: TBD

### Phase 13: pipeline reliability + bootstrap hardening
**Goal**: Eliminate all 7 failure modes documented in `SECURITY_REVIEW_SCAN_HANDOFF.md` and all missing pre-flight checks documented in `BOOTSTRAP_REQUIREMENTS.md`. The pipeline reaches ≥95% first-attempt success rate for fully autonomous execution — no manual JSON reformatting, no manual file writes, no false-negative hook blocks.
**Depends on**: Phase 12
**Source documents**: `BOOTSTRAP_REQUIREMENTS.md`, `SECURITY_REVIEW_SCAN_HANDOFF.md`
**Success Criteria** (what must be TRUE):
  1. `bootstrap/pre-flight-checks.sh` exits 0 on a correct environment and non-zero with actionable remediation messages when any required tool (codegraph, make, go, git, docker) is missing or misconfigured; a `make preflight` target invokes it.
  2. All 9 tracer agents (1 authz + 1 oauth + 1 invariant + 6 taint) succeed on first attempt against `examples/sample-vulnerable-service` with no manual prompt reformatting — JSON serialization issue (Handoff Issue 3) is eliminated.
  3. All verdict files (`authz-findings.json`, `oauth-checklist.json`, `invariant-results.json`, `taint-verdict-*.json`, `review-report.json`, `review-report.md`) are written autonomously by agents with no orchestrator Write calls required — agent write permission model (Handoff Issue 4) is resolved.
  4. Post-tool-use validation hooks report 0 false negatives: hooks fire only after all verdict files are present, and error messages include file path, first 100 chars of actual content, and line number of parse failure (Handoff Issue 5).
  5. `go-cartographer.md` post-processing step cross-references all `router.{GET,POST,DELETE,PATCH,PUT}()` calls from `main.go` against detected entrypoints and emits a `warnings` array in `go-index.json` for any gaps — `callChainSQLiHandler` route (Handoff Issue 1) would be caught.
  6. Agent input schemas are documented as `claude-security-hooks/specs/agents/<agent>.schema.json` for all 5 specialist agent types; a schema validation step rejects unknown fields at dispatch with the expected schema in the error message (Handoff Issue 2).
  7. `go-index.json` entrypoints include `first_param_read_line` and `first_param_read_expr` fields so orchestrators use precise source locations rather than function definition lines (Handoff Issue 6).
**Plans**: 6 plans across 2 waves

Plans:
- [ ] 13-01-PLAN.md — Wave 0: TDD RED — failing tests for W4 (validate.go) + W7 (schema.Handler)
- [ ] 13-02-PLAN.md — Wave 1a: W1 Bootstrap pre-flight script + Makefile preflight target
- [ ] 13-03-PLAN.md — Wave 1b: W2+W3 — JSON serialization enforcement + agent Write permissions
- [ ] 13-04-PLAN.md — Wave 1c: W4+W5+W7 GREEN — validate.go fix + preflight.go schema doc + schema.Handler fields
- [ ] 13-05-PLAN.md — Wave 1d: W5 — Agent input schema JSON files (6 specs/agents/*.schema.json)
- [ ] 13-06-PLAN.md — Wave 2: W6+W7 — Cartographer route cross-reference + security-review.md first_param_read

---

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11 → 12 → 13. Phases 2/3/4 develop in parallel from Phase 1; the converge point is Phase 5. Phases 7–13 are post-v1 hardening, migration, evaluation, and reliability passes.

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Repo scaffold | — | ✅ Complete | 2026-05-19 |
| 2. claude-security-hooks binary | 8/8 | ✅ Complete | 2026-05-19 |
| 3. Agent definitions + hook registration | — | ✅ Complete | 2026-05-23 |
| 4. opengrep-mcp server + OpenGrep container | — | ✅ Complete | 2026-05-24 |
| 5. End-to-end smoke test | — | ✅ Complete | 2026-05-24 |
| 6. Documentation | — | ✅ Complete | 2026-05-24 |
| 7. Extended test cases | — | ✅ Complete | 2026-05-24 |
| 8. Automated E2E Testing | — | ✅ Complete | 2026-05-25 |
| 9. opengrep-mcp server + infrastructure hardening | 2/2 | ✅ Complete (SC5 pending human verify) | 2026-05-26 |
| 10. Hook compatibility + code-ref schema | 3/3 | ✅ Complete | 2026-05-26 |
| 11. codegraph migration | 2/2 | ✅ Complete | 2026-05-27 |
| 12. code review skill evaluation | — | 📋 Not planned yet | — |
| 13. pipeline reliability + bootstrap hardening | — | 📋 Not planned yet | — |
