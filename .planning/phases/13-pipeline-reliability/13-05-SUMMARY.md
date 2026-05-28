---
phase: 13-pipeline-reliability
plan: 05
subsystem: docs
tags: [json-schema, agent-contracts, oauth, taint-tracer, authz, invariant-checker, synthesis, cartographer, draft-07]

# Dependency graph
requires:
  - phase: 13-04
    provides: preflight.go schema doc constants (embedded schema strings that this plan's files must be consistent with)
provides:
  - "6 JSON Schema draft-07 files under claude-security-hooks/specs/agents/ documenting each agent's Task prompt input contract"
  - "go-oauth-auditor.schema.json explicitly excludes working_directory — documents the root cause of Issue 2"
  - "Canonical schema artifacts for operator debugging when D-09 parse errors occur"
affects: [13-06, phase-14-smoke-test, operators-reading-error-messages]

# Tech tracking
tech-stack:
  added: [JSON Schema draft-07]
  patterns:
    - "Agent input contracts documented as specs/agents/<agent>.schema.json, one per agent type"
    - "additionalProperties: false on all schemas prevents unknown fields at the schema documentation level"
    - "Schema files are documentation artifacts only — runtime enforcement is via DisallowUnknownFields() in preflight.go"

key-files:
  created:
    - claude-security-hooks/specs/agents/go-cartographer.schema.json
    - claude-security-hooks/specs/agents/go-taint-tracer.schema.json
    - claude-security-hooks/specs/agents/go-authz-tracer.schema.json
    - claude-security-hooks/specs/agents/go-oauth-auditor.schema.json
    - claude-security-hooks/specs/agents/invariant-checker.schema.json
    - claude-security-hooks/specs/agents/synthesis.schema.json
  modified: []

key-decisions:
  - "Schema files are documentation artifacts derived from Go struct json tags — they are not loaded at runtime (preflight.go uses DisallowUnknownFields() from the struct itself)"
  - "go-oauth-auditor.schema.json explicitly excludes working_directory to document why the Issue 2 D-09 error occurred"
  - "synthesis.schema.json has only 2 required fields (working_directory + review_id) per D18 — synthesis reads from a directory, not from the other agents' inputs"

patterns-established:
  - "Schema field names mirror Go struct json tags exactly (e.g. oauth_locations, code_ref_dirty, review_session_id)"
  - "All schemas use additionalProperties: false to document the strict contract enforced at runtime"

requirements-completed: [SC-13-6]

# Metrics
duration: 3min
completed: 2026-05-28
---

# Phase 13 Plan 05: W5 Agent Input Schema Documentation Summary

**6 JSON Schema draft-07 files documenting each agent's Task prompt input contract under specs/agents/, with go-oauth-auditor.schema.json explicitly excluding working_directory to document the Issue 2 root cause**

## Performance

- **Duration:** 3 min
- **Started:** 2026-05-28T08:17:43Z
- **Completed:** 2026-05-28T08:20:28Z
- **Tasks:** 1
- **Files modified:** 6 created

## Accomplishments

- Created `claude-security-hooks/specs/agents/` directory with 6 JSON Schema files (one per agent)
- All schemas use JSON Schema draft-07, `additionalProperties: false`, and field names consistent with Go struct json tags in `internal/schema/`
- `go-oauth-auditor.schema.json` documents that `working_directory` is NOT a valid field — this is the canonical reference for why D-09 `oauth input parse: json: unknown field "working_directory"` occurs
- `synthesis.schema.json` requires both `working_directory` and `review_id` per D18 (synthesis reads from a directory of verdict files)
- SC-13-6 requirement satisfied: all 5 specialist agent types + synthesis now have documented input schemas

## Task Commits

1. **Task 1: Create specs/agents/ directory and write all 6 JSON Schema files** - `9bcb14b` (docs)

**Plan metadata:** (see below)

## Files Created/Modified

- `claude-security-hooks/specs/agents/go-cartographer.schema.json` - Cartographer input: working_directory required, code_ref/review_session_id optional
- `claude-security-hooks/specs/agents/go-taint-tracer.schema.json` - TaintInput: source + sink required (each with file/line/expr/kind), semgrep_tier enum, max_depth optional
- `claude-security-hooks/specs/agents/go-authz-tracer.schema.json` - AuthzInput: routes required (with handler/middleware_chain nested schemas)
- `claude-security-hooks/specs/agents/go-oauth-auditor.schema.json` - OAuthInput: oauth_locations required, working_directory explicitly absent (Issue 2 fix documentation)
- `claude-security-hooks/specs/agents/invariant-checker.schema.json` - InvariantCheckerInput: flow_name + invariants required (each invariant has id + statement)
- `claude-security-hooks/specs/agents/synthesis.schema.json` - Synthesis: working_directory + review_id both required per D18

## Decisions Made

- Schema files are documentation-only artifacts; runtime enforcement comes from `DisallowUnknownFields()` in preflight.go (added in 13-04). No binary changes needed in this plan.
- Nested object schemas (TaintEndpoint, AuthzRoute, Handler, InputInvariant) are inlined into the parent schema for operator readability, matching the Go struct hierarchy.
- `go-oauth-auditor.schema.json` includes an explicit `description` at the top-level noting that `working_directory` is NOT valid — this is the primary educational artifact for operators who see D-09 errors.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required. These are documentation files only.

## Next Phase Readiness

- Wave 3 (13-06) can proceed: cartographer route detection + first_param_read fixes
- Schema files are ready to be referenced by operators when D-09 errors occur
- The go-oauth-auditor.schema.json is consistent with the embedded `oauthInputSchemaDoc` constant added to preflight.go in 13-04

## Self-Check

- [x] All 6 schema files exist at `claude-security-hooks/specs/agents/` (verified: `find ... | wc -l` = 6)
- [x] All 6 files are valid JSON (verified via python3 json.load)
- [x] go-oauth-auditor.schema.json has `oauth_locations` in required array
- [x] go-oauth-auditor.schema.json has `additionalProperties: false`
- [x] go-oauth-auditor.schema.json does NOT have `working_directory` in properties
- [x] go-taint-tracer.schema.json has `source` and `sink` in required array
- [x] synthesis.schema.json has `working_directory` and `review_id` in required array
- [x] All 6 files contain `"$schema": "http://json-schema.org/draft-07/schema#"`
- [x] Commit 9bcb14b exists in git log

## Self-Check: PASSED

---
*Phase: 13-pipeline-reliability*
*Completed: 2026-05-28*
