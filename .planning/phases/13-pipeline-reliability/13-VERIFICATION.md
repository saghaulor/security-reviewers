---
phase: 13-pipeline-reliability
verified: 2026-05-28T00:00:00Z
status: gaps_found
score: 6/7 must-haves verified
overrides_applied: 0
gaps:
  - truth: "go test ./... passes GREEN — all tests pass including pre-existing D-12 enforcement"
    status: failed
    reason: "TestTracerAgentsExcludeForbiddenTools fails in internal/agentcheck because W3 (Plan 13-03) added Write to go-taint-tracer, go-authz-tracer, go-oauth-auditor, and invariant-checker, but the D-12 enforcement test (which explicitly lists Write as forbidden for tracer agents) was never updated to reflect this intentional policy change. go test ./... exits non-zero."
    artifacts:
      - path: "claude-security-hooks/internal/agentcheck/agentcheck_test.go"
        issue: "Line 29 defines forbiddenTools = [Grep Bash Edit Write]. The test enforces D-12 which forbids Write from tracer agents. Plan 13-03 added Write to tracer agents without updating this test."
      - path: ".claude/agents/go-taint-tracer.md"
        issue: "tools frontmatter now includes Write — violates the D-12 check in agentcheck_test.go"
      - path: ".claude/agents/go-authz-tracer.md"
        issue: "tools frontmatter now includes Write — violates the D-12 check in agentcheck_test.go"
      - path: ".claude/agents/go-oauth-auditor.md"
        issue: "tools frontmatter now includes Write — violates the D-12 check in agentcheck_test.go"
      - path: ".claude/agents/invariant-checker.md"
        issue: "tools frontmatter now includes Write — violates the D-12 check in agentcheck_test.go"
    missing:
      - "Update agentcheck_test.go to reflect the D-12 policy amendment: tracer agents now have scoped Write permission to exactly one output file (A10-amended). Remove Write from forbiddenTools or add a per-agent allowlist that permits Write if an A10-amended rule is present in the agent file."
      - "Optionally update REQUIREMENTS.md REQ-taint-T11 comment to note that Write is now permitted for verdict file output (A10-amended)"
human_verification:
  - test: "Run /security-review against examples/sample-vulnerable-service and observe first-attempt success rate"
    expected: "All 9 tracer agents complete on first attempt with no D-09 preflight blocks, no manual JSON reformatting, and all verdict files written autonomously. Pipeline reaches >=95% first-attempt success rate."
    why_human: "Pipeline behavior under real LLM execution cannot be verified with grep/file checks. The fix correctness (JSON enforcement instructions, Write permissions, schema docs) is verified structurally, but runtime behavior requires an actual /security-review invocation."
---

# Phase 13: Pipeline Reliability + Bootstrap Hardening Verification Report

**Phase Goal:** Eliminate all 7 failure modes documented in SECURITY_REVIEW_SCAN_HANDOFF.md and all missing pre-flight checks documented in BOOTSTRAP_REQUIREMENTS.md. The pipeline reaches >=95% first-attempt success rate for fully autonomous execution — no manual JSON reformatting, no manual file writes, no false-negative hook blocks.
**Verified:** 2026-05-28
**Status:** GAPS FOUND
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|---------|
| 1 | bootstrap/pre-flight-checks.sh exits 0 on correct environment and non-zero with actionable remediation messages on failures; make preflight target invokes it | VERIFIED | Script exists at bootstrap/pre-flight-checks.sh, implements all 6 checks with FAILURES counter, exits 0 on sample-vulnerable-service (confirmed via behavioral spot-check). make preflight TARGET=... exits 0 and propagates exit code. |
| 2 | All 9 tracer agents succeed on first attempt — JSON serialization issue eliminated | VERIFIED (structural) | security-review.md has 2 "JSON prompt enforcement" blocks (Steps 4 and 6) and 1 "All Task prompts are pure JSON" in Key constraints. Runtime behavior needs human verification. |
| 3 | All verdict files written autonomously by agents with no orchestrator Write calls — agent write permission model resolved | VERIFIED (structural) | All 5 tracer agents (cartographer, authz, oauth, taint, invariant) have Write in tools frontmatter. A10-amended write-target rules present in all 5. Runtime behavior needs human verification. |
| 4 | Post-tool-use validation hooks report 0 false negatives: S1 check skipped when content parsed successfully; error messages include first 100 chars of content | VERIFIED | validate.go has synthContentValid guard (lines 66-74). dispatch.go has contentPreview() helper applied to all 6 parse-failure paths. Wave 0 tests TestValidate_Synthesis_S1_MissingFile_SoftFail and TestValidate_TaintVerdict_ErrorMessageContainsPreview both PASS. |
| 5 | go-cartographer.md post-processing cross-references all router registration calls and emits warnings array for gaps | VERIFIED | Step 3 uses metavariable patterns ($ROUTER.GET). Step 3.5 is present with route_registration_gap warning logic. Confirmed by grep. |
| 6 | Agent input schemas documented as specs/agents/<agent>.schema.json for all 5 specialist agent types | VERIFIED | 6 JSON Schema files exist at claude-security-hooks/specs/agents/. go-oauth-auditor.schema.json correctly requires oauth_locations, excludes working_directory, has additionalProperties: false. All have $schema draft-07. |
| 7 | go-index.json entrypoints include first_param_read_line and first_param_read_expr fields | VERIFIED | schema.Handler struct has FirstParamReadLine int (json:"first_param_read_line,omitempty") and FirstParamReadExpr string (json:"first_param_read_expr,omitempty"). go-cartographer.md Step 3 instructs their reporting. security-review.md Step 6 references first_param_read_line. |

**Score:** 6/7 truths verified (truth #3 is verified structurally but blocked by test suite failure — see gaps)

### Critical Test Suite Failure

`go test ./...` from `claude-security-hooks/` exits non-zero:

```
--- FAIL: TestTracerAgentsExcludeForbiddenTools (0.00s)
    --- FAIL: TestTracerAgentsExcludeForbiddenTools/go-taint-tracer (0.00s)
        agentcheck_test.go:50: agent "go-taint-tracer" contains D-12 forbidden tools: [Write]
    --- FAIL: TestTracerAgentsExcludeForbiddenTools/go-authz-tracer (0.00s)
        agentcheck_test.go:50: agent "go-authz-tracer" contains D-12 forbidden tools: [Write]
    --- FAIL: TestTracerAgentsExcludeForbiddenTools/go-oauth-auditor (0.00s)
        agentcheck_test.go:50: agent "go-oauth-auditor" contains D-12 forbidden tools: [Write]
    --- FAIL: TestTracerAgentsExcludeForbiddenTools/invariant-checker (0.00s)
        agentcheck_test.go:50: agent "invariant-checker" contains D-12 forbidden tools: [Write]
FAIL    github.com/saghaulor/claude-security-hooks/internal/agentcheck    0.005s
FAIL
```

Plan 13-03 intentionally added Write to tracer agents (SC-13-3 requirement). However, the existing D-12 enforcement test in `internal/agentcheck/agentcheck_test.go` was not updated to reflect this policy amendment. This test was established in Phase 3 and encodes the rule that `Write` is forbidden for tracer agents. The Phase 13 plans never referenced `agentcheck_test.go` or planned to update D-12 enforcement.

**This is a blocker.** The top-level exit criterion for Phase 2 (SC-2-2) requires `go test ./...` to pass. Phase 13 breaks that invariant.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `bootstrap/pre-flight-checks.sh` | 6-check pre-flight script; executable | VERIFIED | 172 lines, all 6 checks with FAILURES counter, exits 0 on valid env |
| `Makefile` | preflight target wired to bootstrap script | VERIFIED | preflight: @bash bootstrap/pre-flight-checks.sh "$(TARGET)" present |
| `.claude/commands/security-review.md` | JSON enforcement at Steps 4+6 + Key constraints | VERIFIED | 2 occurrences of "JSON prompt enforcement", 1 of "All Task prompts are pure JSON" |
| `.claude/agents/go-cartographer.md` | Write in tools + A10 amendment + Step 3 metavariables + Step 3.5 | VERIFIED | Write in tools, A10 text includes go-index.json, $ROUTER.GET patterns, Step 3.5 section |
| `.claude/agents/go-authz-tracer.md` | Write in tools + A10-amended rule | VERIFIED | Write in tools, A10-amended → authz-findings.json |
| `.claude/agents/go-oauth-auditor.md` | Write in tools + A10-amended rule | VERIFIED | Write in tools, A10-amended → oauth-checklist.json |
| `.claude/agents/go-taint-tracer.md` | Write in tools + A10-amended rule | VERIFIED | Write in tools, A10-amended → taint-verdict-*.json |
| `.claude/agents/invariant-checker.md` | Write in tools + A10-amended rule | VERIFIED | Write in tools, A10-amended → invariant-results.json |
| `claude-security-hooks/internal/hooks/validate.go` | S1 soft-fail guard | VERIFIED | synthContentValid bool at lines 66-74; SynthesisDirInvariants.Check only runs if parse failed |
| `claude-security-hooks/internal/hooks/dispatch.go` | content preview in error messages | VERIFIED | contentPreview() helper, applied to all 6 per-agent parse failure paths |
| `claude-security-hooks/internal/hooks/preflight.go` | Expected schema in D-09 error | VERIFIED | 4 schema doc constants, each injected with "Expected schema:" suffix in error message |
| `claude-security-hooks/internal/schema/cartographer.go` | Handler.FirstParamReadLine and FirstParamReadExpr fields | VERIFIED | Both fields present with omitempty json tags |
| `claude-security-hooks/specs/agents/go-cartographer.schema.json` | Cartographer input schema | VERIFIED | working_directory required, $schema draft-07 |
| `claude-security-hooks/specs/agents/go-taint-tracer.schema.json` | Taint tracer input schema | VERIFIED | source + sink required |
| `claude-security-hooks/specs/agents/go-authz-tracer.schema.json` | Authz tracer input schema | VERIFIED | routes required |
| `claude-security-hooks/specs/agents/go-oauth-auditor.schema.json` | OAuth auditor input schema | VERIFIED | oauth_locations required, working_directory absent, additionalProperties: false |
| `claude-security-hooks/specs/agents/invariant-checker.schema.json` | Invariant checker input schema | VERIFIED | flow_name + invariants required |
| `claude-security-hooks/specs/agents/synthesis.schema.json` | Synthesis input schema | VERIFIED | working_directory + review_id required |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| Makefile preflight target | bootstrap/pre-flight-checks.sh | bash bootstrap/pre-flight-checks.sh | VERIFIED | @bash bootstrap/pre-flight-checks.sh "$(TARGET)" in Makefile |
| security-review.md Step 4 | JSON enforcement instruction | "prompt MUST be pure JSON" | VERIFIED | "JSON prompt enforcement" block before Step 4 cartographer dispatch |
| security-review.md Step 6 | JSON enforcement instruction | "pure JSON" | VERIFIED | "JSON prompt enforcement" block before Step 6 tracer fan-out |
| go-authz-tracer.md tools frontmatter | Claude Code runtime Write permission | Write in tools list | VERIFIED | tools line contains Write |
| preflight.go D-09 error | Expected schema string | fmt.Sprintf with "Expected schema:" | VERIFIED | All 4 agent error paths include Expected schema: |
| validate.go synthesis branch | S1 soft-fail | synthContentValid guard | VERIFIED | SynthesisDirInvariants.Check skipped when synthContentValid=true |
| dispatch.go parse failures | content preview | contentPreview() helper | VERIFIED | All 6 agent parse paths use contentPreview() |
| go-cartographer.md Step 3.5 | CartographerIndex.Warnings | route_registration_gap | VERIFIED | Step 3.5 text present with route_registration_gap warning format |
| security-review.md Step 6 | go-index.json first_param_read_line | first_param_read_line instruction | VERIFIED | "use first_param_read_line from entrypoint handler if present and non-zero" |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Pre-flight exits 0 on valid environment | bash bootstrap/pre-flight-checks.sh examples/sample-vulnerable-service | pre-flight-checks: PASSED | PASS |
| Pre-flight exits 1 on missing go.mod | bash bootstrap/pre-flight-checks.sh /nonexistent/path | ERROR: /nonexistent/path/go.mod not found. EXIT 1 | PASS |
| make preflight propagates exit code | make preflight TARGET=examples/sample-vulnerable-service | exits 0, PASSED | PASS |
| Wave 0 RED tests now GREEN | go -C claude-security-hooks test ./internal/hooks/ -run S1_SoftFail,TaintPreview,PrefPreview | all PASS | PASS |
| Handler fields round-trip | go -C claude-security-hooks test ./internal/schema/ -run TestHandler_FirstParamReadLine | all PASS | PASS |
| Full test suite | go -C claude-security-hooks test ./... | FAIL (agentcheck D-12) | FAIL |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|---------|
| SC-13-1 | 13-02-PLAN.md | Bootstrap pre-flight script exits 0 on correct env, non-zero with remediation on failures | SATISFIED | bootstrap/pre-flight-checks.sh verified end-to-end |
| SC-13-2 | 13-03-PLAN.md | All 9 tracers succeed on first attempt — JSON serialization eliminated | SATISFIED (structural) | JSON enforcement instructions at Steps 4 and 6 of security-review.md present; runtime requires human verification |
| SC-13-3 | 13-03-PLAN.md | All verdict files written autonomously by agents | SATISFIED (structural; test broken) | Write in all 5 tracer agent tools frontmatter; A10-amended rules present; TestTracerAgentsExcludeForbiddenTools FAILS |
| SC-13-4 | 13-01, 13-04-PLAN.md | Hook false negatives 0: S1 soft-fail + 100-char content preview + Expected schema in D-09 | SATISFIED | validate.go synthContentValid, dispatch.go contentPreview(), preflight.go Expected schema — all wave 0 tests pass |
| SC-13-5 | 13-06-PLAN.md | Cartographer cross-references all router.{GET,POST,...} calls, emits warnings for gaps | SATISFIED | go-cartographer.md Step 3.5 with route_registration_gap warnings present |
| SC-13-6 | 13-01, 13-04, 13-05-PLAN.md | Agent input schemas documented as specs/agents/*.schema.json for all 5 types | SATISFIED | 6 schema files present, all valid JSON, oauth_locations required, additionalProperties: false |
| SC-13-7 | 13-01, 13-04, 13-06-PLAN.md | go-index.json entrypoints include first_param_read_line and first_param_read_expr | SATISFIED | Handler struct fields present with omitempty; go-cartographer.md instructs their reporting; security-review.md uses them |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| claude-security-hooks/internal/agentcheck/agentcheck_test.go | 29 | forbiddenTools includes Write — policy now conflicts with W3 changes | BLOCKER | `go test ./...` exits non-zero; 4 sub-tests fail. The D-12 enforcement test was not updated when Phase 13 intentionally changed the Write permission model for tracer agents. |

No unreferenced TBD/FIXME/XXX markers found in any Phase 13-modified files.

### Human Verification Required

#### 1. Pipeline First-Attempt Success Rate

**Test:** Run `/security-review examples/sample-vulnerable-service` once (no retries) after rebuilding the hooks binary.
**Expected:** All 9 tracer agents complete on first attempt. No D-09 preflight blocks from non-JSON prompts. All verdict files (authz-findings.json, oauth-checklist.json, invariant-results.json, taint-verdict-*.json, review-report.json) appear in the target directory written by agents. No orchestrator Write calls required. Pipeline ends with a valid review-report.json.
**Why human:** Runtime LLM behavior cannot be verified with file checks. JSON enforcement instructions and Write permission grants are structural — whether Claude actually follows them on first attempt requires execution.

### Gaps Summary

**One blocker** prevents marking this phase as complete.

**Blocker: D-12 test conflict (SC-13-3)**

Plan 13-03 added `Write` to the tools frontmatter of four tracer agents (go-taint-tracer, go-authz-tracer, go-oauth-auditor, invariant-checker) to address Handoff Issue 4 (SC-13-3). This was the correct change per the phase goal.

However, the existing test `TestTracerAgentsExcludeForbiddenTools` in `claude-security-hooks/internal/agentcheck/agentcheck_test.go` enforces D-12 (HAND_OFF §3 decision) which explicitly lists `Write` as forbidden for tracer agents. This test was written in Phase 3 as a permanent CI enforcement gate. Phase 13 plans never referenced this test, never planned to update it, and the executor did not notice it.

The Phase 13 W3 fix (granting Write to tracer agents) is the correct resolution for the real-world runtime failure (Issue 4 — 150 seconds lost), but the D-12 enforcement test needs to be updated to reflect the amended policy: tracer agents now have scoped Write permission to a single designated output file (enforced by A10-amended rules in each agent file), which is a narrower and still-safe contract different from the blanket "no Write at all" rule.

**Resolution required:** Update `TestTracerAgentsExcludeForbiddenTools` or create an amended version that either (a) removes `Write` from `forbiddenTools` and adds a separate test verifying the A10-amended rule is present when Write is in the tools list, or (b) checks that the agent file contains an A10-amended constraint when Write is present. Then `go test ./...` will pass and the phase goal is achieved.

---

_Verified: 2026-05-28_
_Verifier: Claude (gsd-verifier)_
