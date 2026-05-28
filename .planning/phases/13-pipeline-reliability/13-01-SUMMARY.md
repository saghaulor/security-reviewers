---
phase: 13-pipeline-reliability
plan: "01"
subsystem: claude-security-hooks
tags:
  - tdd
  - red-phase
  - validate
  - preflight
  - schema
dependency_graph:
  requires: []
  provides:
    - "Failing test: TestValidate_Synthesis_S1_MissingFile_SoftFail (W4 driver)"
    - "Failing test: TestValidate_TaintVerdict_ErrorMessageContainsPreview (W4 driver)"
    - "Failing test: TestPreflight_OAuthUnknownField_SchemaDocInError (W5 driver)"
    - "Compile-error RED: TestHandler_FirstParamReadLine_RoundTrip (W7 driver)"
    - "Compile-error RED: TestHandler_FirstParamReadLine_OmitEmpty (W7 driver)"
  affects:
    - "13-04-PLAN.md (GREEN implementation targets)"
tech_stack:
  added: []
  patterns:
    - "TDD RED phase — failing tests before any implementation"
    - "External test package (schema_test) for schema struct round-trip tests"
    - "compile-error as RED state for struct-addition tests"
key_files:
  created: []
  modified:
    - path: "claude-security-hooks/internal/hooks/validate_test.go"
      change: "Added 2 failing tests for W4: S1 soft-fail and error message preview"
    - path: "claude-security-hooks/internal/hooks/preflight_test.go"
      change: "Added 1 failing test (W5 schema doc) and 1 passing regression guard (cartographer silent-pass)"
    - path: "claude-security-hooks/internal/schema/schema_test.go"
      change: "Added 2 compile-error RED tests for W7 Handler.FirstParamReadLine/FirstParamReadExpr"
decisions:
  - "Used schema.SynthesisReport struct directly in validate_test.go to construct valid inner JSON for S1 soft-fail test — requires schema import in internal test package"
  - "W7 tests use compile-error as RED state (not runtime FAIL) because struct fields are the target of addition; plan acceptance criteria explicitly accepts this"
  - "TestPreflight_CartographerInput_SilentPass added as PASS test to confirm regression safety for cartographer case after W5 changes"
metrics:
  duration: "~15 minutes"
  completed_date: "2026-05-28"
  tasks_completed: 3
  tasks_total: 3
  files_modified: 3
---

# Phase 13 Plan 01: TDD RED Phase for W4/W5/W7 Summary

TDD RED phase writing 5 failing tests (3 runtime FAIL + 2 compile-error) that drive Wave 1c GREEN implementation in 13-04.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | RED — validate.go failing tests (W4: S1 soft-fail + error preview) | 8a36f33 | validate_test.go |
| 2 | RED — preflight.go failing test (W5: schema doc injection) | 3495b7e | preflight_test.go |
| 3 | RED — schema_test.go failing tests (W7: Handler new fields) | 1221024 | schema_test.go |

## RED State Verification

### validate_test.go

- **TestValidate_Synthesis_S1_MissingFile_SoftFail**: FAIL — current code always
  blocks on S1 missing file; W4 soft-fail target requires no block when content is valid JSON.
  Actual error: `"reason":"S1: review-report.json expected 'exists' got 'missing'; ..."`

- **TestValidate_TaintVerdict_ErrorMessageContainsPreview**: FAIL — current code
  produces `"reason":"T1: taint verdict parse"` with no content preview; W4 target
  requires `"content(first 100):"` substring.

### preflight_test.go

- **TestPreflight_OAuthUnknownField_SchemaDocInError**: FAIL — current code emits
  `D-09: oauth input parse: json: unknown field "working_directory"` with no schema doc;
  W5 target requires `"Expected schema:"` + oauth_locations/additionalProperties mention.

- **TestPreflight_CartographerInput_SilentPass**: PASS — cartographer case returns ""
  (silent pass). Included as regression guard.

### schema_test.go

- **TestHandler_FirstParamReadLine_RoundTrip**: COMPILE ERROR — `unknown field
  FirstParamReadLine in struct literal of type schema.Handler`. This IS the correct RED
  state for struct-addition tests.

- **TestHandler_FirstParamReadLine_OmitEmpty**: COMPILE ERROR — same root cause.
  Both will compile and pass once 13-04 Task 2 adds the fields.

## W7 Pre-Existence Check

W7 struct fields (`FirstParamReadLine`, `FirstParamReadExpr`) do NOT pre-exist in
`claude-security-hooks/internal/schema/cartographer.go`. The current Handler struct has
only FQN, File, and Line. 13-04 Task 2 must add the fields (no conditional guard path needed).

## All Previously Passing Tests

All pre-existing tests in `internal/hooks/` continue to pass. The `internal/schema/`
package build fails only because of the new test file's struct literal references to
non-existent fields; production code (`go build ./...`) still builds cleanly.

## Deviations from Plan

### Auto-fixed Issues

None.

### Notes

1. The plan's action for Task 1 specified `schema.SynthesisReport{SchemaVersion: "review-report/v1", ReviewID: "test-session-id"}`. The SynthesisReport struct also has a `Summary` field with `BySeverity map[string]int` and `ByClass map[string]int` — marshaling with nil maps causes S3/S4 invariant violations that would block even with soft-fail. The test initializes these maps to empty maps to ensure the S2–S6 verdict checks pass, so only the S1 filesystem check triggers the current (blocking) behavior. This is correct and necessary for the test to isolate the S1 soft-fail behavior.

2. The validate_test.go required importing the schema package. Added `"github.com/saghaulor/claude-security-hooks/internal/schema"` to the import block. This is allowed (test-only dependency).

## Self-Check: PASSED

Files verified present:
- claude-security-hooks/internal/hooks/validate_test.go: FOUND
- claude-security-hooks/internal/hooks/preflight_test.go: FOUND
- claude-security-hooks/internal/schema/schema_test.go: FOUND

Commits verified:
- 8a36f33: FOUND
- 3495b7e: FOUND
- 1221024: FOUND
