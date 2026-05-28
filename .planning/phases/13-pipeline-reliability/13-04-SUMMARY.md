---
phase: 13-pipeline-reliability
plan: "04"
subsystem: claude-security-hooks
tags:
  - tdd
  - green-phase
  - validate
  - dispatch
  - preflight
  - schema
dependency_graph:
  requires:
    - "13-01: RED tests for W4/W5/W7"
  provides:
    - "GREEN: TestValidate_Synthesis_S1_MissingFile_SoftFail PASS"
    - "GREEN: TestValidate_TaintVerdict_ErrorMessageContainsPreview PASS"
    - "GREEN: TestPreflight_OAuthUnknownField_SchemaDocInError PASS"
    - "GREEN: TestHandler_FirstParamReadLine_RoundTrip PASS"
    - "GREEN: TestHandler_FirstParamReadLine_OmitEmpty PASS"
  affects:
    - "claude-security-hooks/internal/hooks/validate.go"
    - "claude-security-hooks/internal/hooks/dispatch.go"
    - "claude-security-hooks/internal/hooks/preflight.go"
    - "claude-security-hooks/internal/schema/cartographer.go"
tech_stack:
  added: []
  patterns:
    - "TDD GREEN phase — implementation to make RED tests pass"
    - "S1 soft-fail conditional logic (only block on file absence when parse also failed)"
    - "Content preview in error messages (first 100 chars)"
    - "Embedded compact JSON Schema constants for D-09 diagnostics"
key_files:
  created: []
  modified:
    - path: "claude-security-hooks/internal/hooks/dispatch.go"
      change: "All verdict parse failures now include 'content(first 100):' preview in description; added contentPreview() helper"
    - path: "claude-security-hooks/internal/hooks/validate.go"
      change: "S1 SynthesisDirInvariants.Check skipped when synthesis content parsed successfully (synthContentValid guard)"
    - path: "claude-security-hooks/internal/hooks/preflight.go"
      change: "D-09 error messages include 'Expected schema:' with compact JSON Schema doc constants per agent"
    - path: "claude-security-hooks/internal/schema/cartographer.go"
      change: "Handler struct gains FirstParamReadLine int and FirstParamReadExpr string fields with omitempty json tags"
decisions:
  - "contentPreview() helper uses fmt.Sprintf(%q) to produce a quoted, safe-to-print string for content previews in error messages"
  - "S1 soft-fail uses synthContentValid bool: checks for 'verdict parse' or 'content parse' substring in reasons slice from runAgentValidation"
  - "Schema doc constants placed at package level as Go const (not var) per plan requirement; not embedded from external files"
  - "dispatch.go parse failures use description string (not Actual field) to carry the preview, since FormatViolation ignores Actual when Path is empty"
  - "go.sum was absent from worktree (gitignored per .gitignore); copied from main repo on-disk — not tracked"
metrics:
  duration: "~25 minutes"
  completed_date: "2026-05-28"
  tasks_completed: 3
  tasks_total: 3
  files_modified: 4
---

# Phase 13 Plan 04: TDD GREEN Phase for W4/W5/W7 Summary

TDD GREEN phase implementing three Go changes that make all five Wave 0 RED tests pass: S1 soft-fail (W4), content preview in error messages (W4), schema doc injection in D-09 errors (W5), and Handler struct extension (W7).

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | GREEN W4 — validate.go S1 soft-fail + dispatch.go error preview | f78919e | validate.go, dispatch.go |
| 2 | GREEN W5+W7 — preflight schema doc + Handler new fields | a1ac46a | preflight.go, cartographer.go |
| 3 | Rebuild hooks binary | (no commit — binary gitignored) | .claude/hooks/bin/claude-security-hooks |

## GREEN State Verification

### Wave 0 RED Tests — All GREEN

- **TestValidate_Synthesis_S1_MissingFile_SoftFail**: PASS — S1 file-existence check now skipped when synthesis content is valid JSON
- **TestValidate_TaintVerdict_ErrorMessageContainsPreview**: PASS — block reason now includes `content(first 100):` preview
- **TestPreflight_OAuthUnknownField_SchemaDocInError**: PASS — D-09 block reason includes `Expected schema:` with oauth_locations/additionalProperties
- **TestHandler_FirstParamReadLine_RoundTrip**: PASS — Handler has FirstParamReadLine/FirstParamReadExpr fields with correct json tags
- **TestHandler_FirstParamReadLine_OmitEmpty**: PASS — zero-value fields omitted from JSON

### Full Suite Result

```
ok  github.com/saghaulor/claude-security-hooks/internal/agentcheck
ok  github.com/saghaulor/claude-security-hooks/internal/hooks
ok  github.com/saghaulor/claude-security-hooks/internal/invariants
ok  github.com/saghaulor/claude-security-hooks/internal/schema
ok  github.com/saghaulor/claude-security-hooks/internal/uuidgen
```

go test ./... exits 0. No FAIL lines.

## Implementation Details

### validate.go — S1 Soft-Fail (W4)

Changed the synthesis branch to track `synthContentValid` (bool): if no reason from `runAgentValidation` contains "verdict parse" or "content parse", content is considered valid and `SynthesisDirInvariants.Check` is skipped. If parse failed, S1 check still runs as secondary diagnostic.

### dispatch.go — Content Preview (W4)

Added `contentPreview(s string) string` helper that returns `fmt.Sprintf("%q", s[:100])` for strings >100 chars, `fmt.Sprintf("%q", s)` otherwise. All six per-agent parse-failure violations now pass a description string like `"taint verdict parse | content(first 100): <preview>"` to `FormatViolation`. Since `FormatViolation` ignores `Actual` when `Path` is empty, the description carries the preview.

### preflight.go — Schema Doc (W5)

Added four `const` string constants (`taintInputSchemaDoc`, `authzInputSchemaDoc`, `oauthInputSchemaDoc`, `invariantInputSchemaDoc`) containing compact JSON Schema summaries. Each `preflightPerAgent` error return appends `\nExpected schema: <constant>` via `fmt.Sprintf`.

### cartographer.go — Handler Fields (W7)

Added two fields to `Handler` struct after `Line int`:
- `FirstParamReadLine int    json:"first_param_read_line,omitempty"`
- `FirstParamReadExpr string json:"first_param_read_expr,omitempty"`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] go.sum absent from worktree**
- **Found during:** Task 1 test run setup
- **Issue:** go.sum is gitignored per `.gitignore`; the worktree was created without the file. `go test` failed with "missing go.sum entry" for `github.com/google/go-cmp`.
- **Fix:** Copied go.sum from main repo on-disk (not tracked). Not staged for commit — gitignore applies correctly.
- **Files modified:** claude-security-hooks/go.sum (untracked, not committed)

### Notes

1. `FormatViolation` in decision.go ignores the `Actual` field when `Path` is empty — it only uses `Actual` in the path-style format `"ID: path expected 'X' got 'Y'"`. The plan said "Actual field" but since `Path` is always empty for parse errors, the description string is the correct carrier for the content preview. All tests pass with this approach.

2. Task 3 binary rebuild: `make install` succeeded, binary deployed to `.claude/hooks/bin/claude-security-hooks`. No git commit because both `claude-security-hooks/bin/` and `.claude/hooks/bin/*` are gitignored.

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes at trust boundaries introduced. Schema doc constants are read-only strings with no secrets. S1 soft-fail only activates on successful JSON parse — invalid JSON still triggers blocks.

## Known Stubs

None.

## Self-Check: PASSED

Files verified present:
- claude-security-hooks/internal/hooks/dispatch.go: FOUND
- claude-security-hooks/internal/hooks/validate.go: FOUND
- claude-security-hooks/internal/hooks/preflight.go: FOUND
- claude-security-hooks/internal/schema/cartographer.go: FOUND

Commits verified:
- f78919e: FOUND
- a1ac46a: FOUND
