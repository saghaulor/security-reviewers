# Synthesis Summary

Entry point for downstream `gsd-roadmapper`. Produced by `gsd-doc-synthesizer` from the classified `.planning/intel/classifications/` set.

mode: new
precedence: ["ADR", "SPEC", "PRD", "DOC"] (with per-doc override honored)
synthesizer: gsd-doc-synthesizer

---

## Doc inventory

- **Total docs synthesized:** 1
- **By type:**
  - SPEC: 1 (HAND_OFF.md — `Security Review Agent System — Implementation Handoff`, precedence 0 via manifest override)
  - ADR: 0
  - PRD: 0
  - DOC: 0
- **Unresolved classifications (UNKNOWN/low):** 0
- **Cross-reference cycles detected:** 0 (cross_refs point only to artifacts-to-be-authored and external URLs; no incoming refs to other ingest docs)

---

## Decisions

- **Total locked decisions:** 18 (HAND_OFF.md §4, D1–D18)
- **Source paths:** /home/saghaulor/code/security_reviewer/HAND_OFF.md
- **File:** /home/saghaulor/code/security_reviewer/.planning/intel/decisions.md
- **Lock provenance:** orchestrator-promoted (the source doc is SPEC-classified with `locked: false` in classification metadata; orchestrator instructed treating §4 entries as LOCKED for downstream roadmapping).
- **Coverage:** per-stack agent design (D1), Go hooks (D2), containerized scanners (D3), single MCP server with tier param (D4), PostToolUse-on-Task hook strategy (D5), OAuth bug-class split auditor/tracer (D6), RFC 6819 dropped (D7), OAuth 2.1 target profile (D8), CIBA included (D9), gosec/staticcheck dropped, govulncheck kept (D10), sequential cartographer + parallel tracers (D11), tool allowlists exclude cheap alternatives (D12), strict JSON verdicts no prose (D13), ambiguous preferred over false confidence (D14), no skills for security review (D15), invariant-checker verifies not discovers (D16), Go-first sequential per-stack expansion (D17), synthesis reads from directory (D18).

---

## Requirements

- **Total requirements extracted:** 51 verifiable assertions across 9 components.
- **File:** /home/saghaulor/code/security_reviewer/.planning/intel/requirements.md
- **IDs by component:**
  - `go-cartographer`: REQ-cartographer-A1 … A11 (11)
  - `go-taint-tracer`: REQ-taint-T1 … T11 (11)
  - `go-authz-tracer`: REQ-authz-AZ1 … AZ6 (6)
  - `go-oauth-auditor`: REQ-oauth-OA1 … OA7 (7)
  - `invariant-checker`: REQ-invariant-IC1 … IC4 (4)
  - `synthesis`: REQ-synthesis-S1 … S6 (6)
  - `claude-security-hooks`: REQ-hooks-H1 … H7 (7)
  - `opengrep-mcp`: REQ-mcp-O1 … O8 (8)
  - Agent definition format: REQ-agents-prompt-shape (1)
- **Source:** HAND_OFF.md §3.1–§3.10 (numbered assertions A1–A11, T1–T11, AZ1–AZ6, OA1–OA7, IC1–IC4, S1–S6, H1–H7, O1–O8).

---

## Constraints

- **Total constraints captured:** 22
- **File:** /home/saghaulor/code/security_reviewer/.planning/intel/constraints.md
- **By type:**
  - protocol: 9 (CON-arch-three-layers, CON-tool-assignment-matrix, CON-tracer-protocol, CON-tool-allowlists, CON-source-sink-catalog, CON-oauth-source-sink-table, CON-oauth-checklist-taxonomy, CON-engine-tier-abstraction, CON-validate-protocol)
  - schema: 6 (CON-schema-go-index-v1, CON-schema-taint-verdict, CON-schema-authz, CON-schema-oauth-auditor, CON-schema-invariant, CON-schema-review-report)
  - api-contract: 3 (CON-input-taint-tracer, CON-hook-registration, CON-hook-payload-shapes)
  - nfr: 4 (CON-nfr-cold-start, CON-nfr-zero-deps, CON-nfr-static-binary, CON-nfr-container-isolation, CON-stop-conditions-taint)

---

## Context topics

- **Total topics captured:** 7
- **File:** /home/saghaulor/code/security_reviewer/.planning/intel/context.md
- **Topics:**
  1. Problem statement and prior failure modes (HAND_OFF.md §1)
  2. Why two phases of fan-out (§2.2)
  3. How to read the handoff — precedence on conflict (§0)
  4. Implementation plan — 6 phases (§6)
  5. Open questions Q1–Q10 (§5)
  6. References — OAuth/OIDC specs (§7.1)
  7. References — Tooling (§7.2)
  8. References — Go routing libraries (§7.3)
  9. Final notes from designer to implementing session (§9)
  10. Cross-reference index of files-to-be-authored

---

## Conflicts summary

- **BLOCKERS:** 0
- **WARNINGS (competing variants):** 0
- **INFO (auto-resolved + observations):** 4

See full report: /home/saghaulor/code/security_reviewer/.planning/INGEST-CONFLICTS.md

The four INFO entries cover: (1) single-doc trivial-no-conflict surface, (2) manifest precedence override applied, (3) §4 decisions promoted to LOCKED by orchestrator instruction, (4) §5 open questions captured as context not blockers.

---

## Roadmap-shaping signals for downstream gsd-roadmapper

These observations are derived from the intel and are offered to inform PROJECT.md / REQUIREMENTS.md / ROADMAP.md construction. The roadmapper is not bound to follow them.

- The §6 phase plan already provides a sequenced, commit-sized work breakdown (Phase 1 scaffold → 2 hooks binary → 3 agent defs → 4 MCP server → 5 smoke test → 6 docs). A faithful ROADMAP can mirror these phases as milestones with the §3 assertions as exit criteria for Phase 2 (test coverage) and Phase 5 (end-to-end verification).
- Two distinct Go modules will live in the repo: `claude-security-hooks` and `opengrep-mcp`. Each merits its own README, Makefile, and dependency boundary. Roadmap should treat them as parallel deliverables that converge at Phase 5.
- The user's stated preferences (compiled binaries over scripts; ambiguity over false confidence; tests matter; per-stack first not parameterized) should appear as project principles in PROJECT.md.
- §5 Q9 (model identifier verification) and Q1 (Graphify MCP tool names) and Q4 (OpenGrep build pinning) are concrete first-day tasks for Phase 2–4 work items.
- D15 (no skills) plus Q7 (slash command vs orchestrator agent) → a `/security-review` slash command should be planned for after Phase 3 (agent defs exist) but before Phase 5 (smoke test needs an invocation surface).

---

## Files written by this synthesizer

- /home/saghaulor/code/security_reviewer/.planning/intel/decisions.md
- /home/saghaulor/code/security_reviewer/.planning/intel/requirements.md
- /home/saghaulor/code/security_reviewer/.planning/intel/constraints.md
- /home/saghaulor/code/security_reviewer/.planning/intel/context.md
- /home/saghaulor/code/security_reviewer/.planning/intel/SYNTHESIS.md (this file)
- /home/saghaulor/code/security_reviewer/.planning/INGEST-CONFLICTS.md
