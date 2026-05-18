# Phase 2: claude-security-hooks Go binary - Context

**Gathered:** 2026-05-18
**Status:** Ready for planning

<domain>
## Phase Boundary

A static `claude-security-hooks` Go binary built with `CGO_ENABLED=0` at `claude-security-hooks/bin/claude-security-hooks` that implements three subcommands — `preflight`, `validate`, `inject-context` — and validates every specialist agent's JSON verdict against the 51 per-agent assertion check predicates (A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6) plus the H1–H7 build/runtime invariants. The Go unit-test suite for those 51 predicates passes (`go test ./...`). The core validator path has zero non-stdlib Go dependencies; test framework deps confined to `_test.go` files.

**In scope:**
- `cmd/claude-security-hooks/main.go` subcommand dispatch.
- `internal/schema/` Go structs mirroring every agent's verdict schema (taint_tracer, cartographer, authz_tracer, oauth_auditor, invariant_checker, synthesis).
- `internal/invariants/` 51 assertion predicates + H1–H7 build-time / runtime invariants.
- `internal/hooks/preflight.go`, `validate.go`, `inject.go` subcommand bodies.
- 51 unit tests (one per assertion predicate, table-driven).
- `Makefile` with `build`, `test`, `install` targets (`install` copies binary to `.claude/hooks/bin/`).

**Out of scope:**
- Agent definition files in `.claude/agents/*.md` (Phase 3).
- `.claude/settings.json` hook registration (Phase 3).
- `/security-review` slash command (Phase 3).
- `opengrep-mcp` server (Phase 4).
- E2E smoke test against `examples/sample-vulnerable-service/` (Phase 5).
- Real semantic context injection from `inject-context` (deferred — see D-08).

</domain>

<decisions>
## Implementation Decisions

### Invariant registry shape
- **D-01:** Per-agent registries as structured slices of typed `Invariant` records, NOT a flat `map[string]func`. Each entry carries `ID`, `Description`, `Severity`, and a typed `Check` function. The slice is iterated in registration order for deterministic block-reason output.
- **D-02:** Per the user's global CLAUDE.md no-`any` rule, each agent has its OWN registry type with its OWN typed `Check` signature — no `func(v any)` shared signature. Example shape:
  ```go
  type Severity string  // "critical" | "high" | "medium" | "low" | "info"
  type Violation struct {
      Path     string  // e.g. "path[0].step"
      Expected string  // e.g. "source"
      Actual   string  // e.g. "sink"
  }
  type TaintInvariant struct {
      ID          string  // "T3"
      Description string  // "sanitized/exploitable requires non-empty path with source first and sink last"
      Severity    Severity
      Check       func(*schema.TaintVerdict) []Violation
  }
  var TaintTracerInvariants = []TaintInvariant{
      {ID: "T1", Description: "...", Severity: SeverityCritical, Check: checkT1},
      // ... 10 more for T2–T11
  }
  ```
  Repeat per agent: `CartographerInvariants []CartographerInvariant`, `AuthzInvariants []AuthzInvariant`, `OAuthInvariants []OAuthInvariant`, `InvariantCheckerInvariants []InvariantCheckerInvariant`, `SynthesisInvariants []SynthesisInvariant`.
- **D-03:** Block-reason string format is stable and parseable: `"<ID>: <Path> expected '<Expected>' got '<Actual>'"` for path-style violations and `"<ID>: <Description>"` for boolean-style violations. Multiple violations in a single verdict concatenate with `; `.
- **D-04:** H6 enforcement is mechanical: a dedicated unit test in `internal/invariants/registry_test.go` enumerates each per-agent registry, asserts the set of IDs equals the expected set (`{T1..T11}`, `{A1..A11}`, etc.), and asserts every ID has at least one corresponding `Test*` function in the agent's `_test.go` file (discovered via the `testing.InternalTest` registry pattern). Drift between spec and code fails CI.

### Test framework
- **D-05:** stdlib `testing` for runners, `github.com/google/go-cmp` v0.6.0+ for `cmp.Diff` on `[]Violation` comparisons in `_test.go` files only. `go-cmp` is the ONLY non-stdlib dependency permitted anywhere in the module, and it must appear only in `_test.go` files (enforced by H5 unit test: parse `go list -deps -test ./...` and assert non-test deps are stdlib-only).
- **D-06:** Tests are table-driven. One `*_test.go` file per agent (`taint_tracer_test.go`, `cartographer_test.go`, etc.) co-located in `internal/invariants/`. Each test function is `TestT3_PathNonEmpty(t *testing.T)` style (one test func per assertion ID, named `Test<ID>_<short-description>`) so `go test -run TestT3_` selects a single assertion.

### Subagent-set source of truth
- **D-07:** Hardcoded const slice `var SecurityAgentSet = []string{"go-cartographer", "go-taint-tracer", "go-authz-tracer", "go-oauth-auditor", "invariant-checker", "synthesis"}` lives in `internal/hooks/agents.go`. Zero coupling to Phase 3 agent files, zero I/O at hook startup (preserves CON-nfr-cold-start <5ms budget).
  Drift mitigation: a unit test `TestSecurityAgentSet_MatchesSpec` in `internal/hooks/agents_test.go` lists the expected set verbatim from PROJECT.md and fails if `SecurityAgentSet` diverges. Updating the set after Phase 3 adds a 7th agent requires editing BOTH the const AND the test — forced visibility.

### `inject-context` v1 scope
- **D-08:** `inject-context` is a no-op stub in Phase 2. It MUST be a real subcommand (REQ-hooks-H2: all three subcommands accept Claude Code hook event JSON on stdin matching the documented schemas), but its body reads the `SubagentStart` event from stdin, validates it parses against the §8.1 schema, and exits 0 silently. No stdout output.
  Rationale: spec calls it "optional"; no IC* / H* assertion targets `inject-context`; what to inject and how Claude Code consumes injected context is not fully specified in HAND_OFF.md. Defer to Phase 3 (when agent files exist and we know what context they actually need) or Phase 5 (when E2E smoke surfaces real injection requirements). Phase 2 satisfies H2 by accepting+validating the payload but contributes nothing further.

### Preflight strictness
- **D-09:** `preflight` rejects (emits `{"decision":"block","reason":"..."}` and exits 0) when `tool_input.subagent_type` IS in `SecurityAgentSet` but `tool_input.prompt` does not parse as JSON matching the per-agent input contract from §3.x (e.g., CON-input-taint-tracer for `go-taint-tracer`). This is symmetric with `validate`'s H7 behavior on malformed output. Without this gate, malformed dispatches silently reach the subagent and waste tokens before `validate` catches the bad output downstream.
  When `tool_input.subagent_type` is NOT in `SecurityAgentSet`: exit 0 silently (no validation duty).

### JSON parsing strictness
- **D-10:** All `internal/schema/*` Go-struct unmarshalling uses `json.NewDecoder(input)` with `dec.DisallowUnknownFields()` enabled. New unknown fields in an agent's verdict trigger a block decision at hook time. Aligned with D13 (strict JSON) and D14 (prefer ambiguous over false confidence): a verdict the validator cannot fully understand is BLOCKED, not silently accepted.
  Schema-bump protocol: when a verdict schema gains a field, both the agent prompt and the hook struct must be updated in the same commit. The Phase 2 binary is the gatekeeper enforcing the schema contract.

### File-system access for A7/A8
- **D-11:** Assertions A7 ("every file path cited in cartographer output exists") and A8 ("every line number cited is within the corresponding file's line count") require the validator to read source files. Phase 2 ASSUMES Claude Code launches hooks with CWD set to the workspace root and uses `os.Stat(path)` for A7 and `bufio.Scanner` over `os.Open(path)` counting `\n` for A8. The same CWD assumption applies wherever a tracer cites `(file, line)` (T9, AZ4, OA4, IC3).
  If Claude Code's CWD semantics differ from this assumption, the planner / Q-style verification will catch it; for now Phase 2 encodes the assumption and adds a guard in `validate` that emits a `block` reason `"workspace_root_not_found"` if CWD doesn't contain expected workspace markers (`go.mod`, `.claude/`, or `.planning/`).

### Build & module layout
- **D-12:** Per HAND_OFF §3.7 recommended layout already in place from Phase 1 scaffold (`claude-security-hooks/cmd/claude-security-hooks/`, `claude-security-hooks/internal/{hooks,invariants,schema}/`, `go.mod` declares `github.com/saghaulor/claude-security-hooks` and `go 1.22`). Phase 2 fills files into this layout; no directory restructuring.
- **D-13:** `Makefile` lives at `claude-security-hooks/Makefile`. Targets: `build` (`CGO_ENABLED=0 go build -o bin/claude-security-hooks ./cmd/claude-security-hooks`), `test` (`go test ./...`), `install` (`build` + `cp bin/claude-security-hooks ../.claude/hooks/bin/`), `clean`, `lint` (`go vet ./...`). No `go generate` step (would force Phase 3 dependency).

### Error / exit-code policy
- **D-14:** Two-class error policy:
  - **Class A — validation block decisions:** anything that says "the security agent did something I cannot accept" → emit `{"decision":"block","reason":"..."}` to stdout, exit 0. Covers all 51 assertion violations, schema parse failures (H7), preflight input mismatch (D-09), unknown-field rejections (D-10), workspace CWD missing (D-11).
  - **Class B — infrastructure failures:** stdin read errors, broken pipes, OS-level failures (e.g., cannot stat workspace at all when needed) → write a human-readable line to stderr, exit 1. These signal "the hook itself is broken," not "the agent did something bad."
  Documented in `internal/hooks/doc.go` as the binary's exit-code contract.

### First-day work item (Q9 from HAND_OFF §5)
- **D-15:** Before writing any test fixture or schema doc that hardcodes a model identifier, fetch https://docs.claude.com to verify `claude-sonnet-4-6` is current as of 2026-05-18. The model identifier does NOT appear in the hook binary itself (the binary is model-agnostic), but it DOES appear in §8.1 reference fixtures used by `_test.go`. If a newer Sonnet identifier exists, update fixtures and note the new ID in CONTEXT.md as a follow-up for Phase 3 (where it appears in agent frontmatter).

### Claude's Discretion
- Internal package import paths within `claude-security-hooks` module.
- Specific Go struct field names within `internal/schema/*` (must mirror JSON keys per CON-schema-*).
- Test fixture data shapes inside `_test.go` files.
- Specific error message wording for block reasons beyond the D-03 format pattern.
- Whether to add a `--version` flag (low priority, defer to Phase 6 docs phase).
- Code-fence stripping algorithm in `validate` (per HAND_OFF §3.7 step 4): default to "strip leading/trailing ```/```json fences and surrounding whitespace, attempt JSON parse, if that fails try entire stdin verbatim." Recoverable; not worth bikeshedding.
- Whether each `Check` function lives in `internal/invariants/<agent>_check_<id>.go` (one file per ID) or `internal/invariants/<agent>.go` (one file per agent). Planner picks based on resulting file size (one-per-agent unless any agent's file exceeds ~400 lines).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Source-of-truth handoff document
- `HAND_OFF.md` §0 — How to read this document (precedence on conflict).
- `HAND_OFF.md` §3.7 — `claude-security-hooks` Go binary component spec (subcommands, main.go shape, validate logic, assertions H1–H7).
- `HAND_OFF.md` §8.1 — Hook event payload shapes (PreToolUse, PostToolUse, SubagentStart). MUST verify against https://code.claude.com/docs/en/hooks before encoding fixtures.
- `HAND_OFF.md` §3.1 — `go-cartographer` agent contract (output schema, assertions A1–A11).
- `HAND_OFF.md` §3.2 — `go-taint-tracer` agent contract (verdict schema, behavior protocol, assertions T1–T11).
- `HAND_OFF.md` §3.3 — `go-authz-tracer` agent contract (schema, assertions AZ1–AZ6).
- `HAND_OFF.md` §3.4 — `go-oauth-auditor` agent contract (schema, assertions OA1–OA7).
- `HAND_OFF.md` §3.5 — `invariant-checker` agent contract (schema, assertions IC1–IC4).
- `HAND_OFF.md` §3.6 — `synthesis` agent contract (review-report/v1 schema, assertions S1–S6).
- `HAND_OFF.md` §4 — Decisions log D1–D18 (LOCKED architectural decisions).
- `HAND_OFF.md` §6 — Implementation plan Phase 2 line items.

### Intel layer (synthesized from HAND_OFF.md)
- `.planning/PROJECT.md` — Project principles + LOCKED D1–D18 table (authoritative for downstream agents).
- `.planning/REQUIREMENTS.md` — 61 REQ-IDs with traceability to phases (Phase 2 owns 52 of them).
- `.planning/STATE.md` — Project state, Q9 deferred-resolution location.
- `.planning/intel/constraints.md` — CON-* constraint table: CON-arch-three-layers, CON-engine-tier-abstraction, CON-tool-assignment-matrix, CON-schema-* (6 schemas), CON-input-taint-tracer, CON-tracer-protocol, CON-tool-allowlists, CON-hook-registration, CON-validate-protocol, CON-nfr-cold-start, CON-nfr-zero-deps, CON-nfr-static-binary, CON-nfr-container-isolation, CON-source-sink-catalog, CON-oauth-source-sink-table, CON-oauth-checklist-taxonomy, CON-stop-conditions-taint, CON-hook-payload-shapes.
- `.planning/intel/decisions.md` — D1–D18 expanded with rationale and source citations.
- `.planning/intel/requirements.md` — Requirements traceability companion to REQUIREMENTS.md.
- `.planning/intel/context.md` — Project context topics (problem statement, fan-out rationale, etc.).
- `.planning/intel/SYNTHESIS.md` — Synthesizer summary; entry-point.
- `.planning/INGEST-CONFLICTS.md` — Conflict report from initial ingest (0 BLOCKERS, 0 WARNINGS).

### Roadmap
- `.planning/ROADMAP.md` §"Phase 2: claude-security-hooks Go binary" — phase goal, requirements list, success criteria, first-day work item (Q9).

### External (must verify at first-day work)
- https://docs.claude.com — verify current Claude Sonnet model identifier (Q9, D-15).
- https://code.claude.com/docs/en/hooks — verify hook event payload field names (CON-hook-payload-shapes, D-15 by extension).
- https://github.com/anthropics/claude-code/issues/7881 — `SubagentStop` payload deficiency explanation (D5 rationale).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `claude-security-hooks/go.mod` exists from Phase 1 scaffold: `module github.com/saghaulor/claude-security-hooks` / `go 1.22`. Phase 2 builds on this; no module rename.
- Directory skeleton from Phase 1: `claude-security-hooks/cmd/claude-security-hooks/`, `claude-security-hooks/internal/{hooks,invariants,schema}/` exist as empty dirs. Phase 2 fills them.
- `.claude/hooks/bin/` exists (empty, `.gitkeep` placeholder). Makefile `install` target writes the binary here.
- `.claude/security-invariants/` exists (empty). Reserved for future `inject-context` per-agent context files (D-08 deferred); Phase 2 does NOT add files here.

### Established Patterns
- Go module per binary (this project has two: `claude-security-hooks` and `opengrep-mcp`). Each has its own `go.mod`, `go.sum` (when deps appear), and its own `Makefile`.
- HAND_OFF §3.7 layout recommendation is the established pattern: `cmd/<name>/main.go` for entrypoint dispatch + `internal/<concern>/` for behavior. Phase 2 follows it verbatim.
- Test files co-located with implementation under `internal/` (Go convention, reinforced by H6 requiring per-assertion tests next to per-assertion checks).

### Integration Points
- The compiled binary `claude-security-hooks/bin/claude-security-hooks` (built via `make build`) is copied to `.claude/hooks/bin/claude-security-hooks` (via `make install`). Phase 3 references this path in `.claude/settings.json` hook registrations.
- `internal/schema/*` Go structs are the type-level contract that downstream phases must respect: agent prompts in Phase 3 must produce JSON parsable by these structs.
- `SecurityAgentSet` in `internal/hooks/agents.go` is the name-level contract: Phase 3 agent files must use one of those 6 frontmatter names.

</code_context>

<specifics>
## Specific Ideas

- **Block-reason format** (D-03): `"<ID>: <Path> expected '<Expected>' got '<Actual>'"` for path-style; `"<ID>: <Description>"` for boolean-style. Multiple violations join with `; `. Example block-reason JSON output: `{"decision":"block","reason":"T3: path[0].step expected 'source' got 'sink'; T6: neither semgrep.ran nor gopls.references_calls > 0"}`. Stable enough that agents and humans can grep for assertion IDs in CI logs.
- **Drift-detection tests**: `TestSecurityAgentSet_MatchesSpec` (D-07) and `TestInvariantRegistry_AllIDsCovered` (D-04) are the meta-tests that keep Phase 2 honest as later phases evolve. Both fail loudly when reality drifts from spec.
- **Two-class exit policy** (D-14) documented in `internal/hooks/doc.go` package comment so it's discoverable via `go doc internal/hooks`.

</specifics>

<deferred>
## Deferred Ideas

- **`--version` flag on the binary**: useful for diagnostics, defer to Phase 6 (Documentation) when version semantics are settled.
- **Real `inject-context` semantic injection**: defer to Phase 3 or Phase 5 once Claude Code's context-injection protocol is verified and we know what each agent actually needs.
- **Generate `SecurityAgentSet` from `.claude/agents/*.md` at build time**: rejected for Phase 2 (would force Phase 3 dependency); could be revisited post-v1 if drift becomes painful.
- **YAML support for `.claude/security-invariants/*.yaml`**: deferred indefinitely. If Phase 5 needs per-agent context, use `.json` to keep CON-nfr-zero-deps intact.
- **CI integration** (Q10 in HAND_OFF §5): out of scope for v1 per PROJECT.md.

</deferred>

---

*Phase: 02-claude-security-hooks-go-binary*
*Context gathered: 2026-05-18 via /gsd-discuss-phase*
