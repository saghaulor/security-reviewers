---
phase: 13-pipeline-reliability
reviewed: 2026-05-28T00:00:00Z
depth: standard
files_reviewed: 20
files_reviewed_list:
  - .claude/agents/go-authz-tracer.md
  - .claude/agents/go-cartographer.md
  - .claude/agents/go-oauth-auditor.md
  - .claude/agents/go-taint-tracer.md
  - .claude/agents/invariant-checker.md
  - .claude/commands/security-review.md
  - bootstrap/pre-flight-checks.sh
  - claude-security-hooks/internal/hooks/dispatch.go
  - claude-security-hooks/internal/hooks/preflight.go
  - claude-security-hooks/internal/hooks/preflight_test.go
  - claude-security-hooks/internal/hooks/validate.go
  - claude-security-hooks/internal/hooks/validate_test.go
  - claude-security-hooks/internal/schema/cartographer.go
  - claude-security-hooks/internal/schema/schema_test.go
  - claude-security-hooks/specs/agents/go-authz-tracer.schema.json
  - claude-security-hooks/specs/agents/go-cartographer.schema.json
  - claude-security-hooks/specs/agents/go-oauth-auditor.schema.json
  - claude-security-hooks/specs/agents/go-taint-tracer.schema.json
  - claude-security-hooks/specs/agents/invariant-checker.schema.json
  - claude-security-hooks/specs/agents/synthesis.schema.json
findings:
  critical: 5
  warning: 6
  info: 4
  total: 15
status: issues_found
---

# Phase 13: Code Review Report

**Reviewed:** 2026-05-28T00:00:00Z
**Depth:** standard
**Files Reviewed:** 20
**Status:** issues_found

## Summary

Phase 13 delivers pre-flight checks, the hooks binary validate/dispatch/preflight pipeline, JSON schema spec files, Go schema structs, agent instruction updates, and test coverage. The implementation is substantive and largely correct, but contains five critical defects: a pipeline-breaking contract mismatch that blocks `go-oauth-auditor` on every invocation, two tests explicitly marked RED (failing by design and shipping that way), a JSON field name typo baked into the wire format, and a Go struct/JSON schema mismatch that causes strict-decode failures. Six warnings cover logic errors in the validate fallback, missing `set -e` in shell, schema-to-struct field name mismatches, and a `PaymentSurface` struct that serializes an always-empty required field. Four info items flag dead test code and minor quality issues.

---

## Critical Issues

### CR-01: go-oauth-auditor receives `working_directory` but schema forbids it — preflight blocks every invocation

**File:** `.claude/commands/security-review.md:145`

**Issue:** Step 6, Tracer 2 constructs the `go-oauth-auditor` input with a `working_directory` field:

```json
{
  "working_directory": "<TARGET_DIR>",
  "oauth_locations": ...,
  ...
}
```

The `go-oauth-auditor.schema.json` (line 7) sets `"additionalProperties": false` and explicitly documents that `working_directory` is NOT a valid field. The `OAuthInput` Go struct also has no `working_directory` field. The `preflightPerAgent` function in `preflight.go` (line 69–75) decodes this input with `DisallowUnknownFields()`. Result: the preflight hook will emit a D-09 block and terminate the OAuth auditor Task on every pipeline run, silently disabling the entire OAuth conformance check. The `TestPreflight_OAuthUnknownField_SchemaDocInError` test (preflight_test.go:173) proves this exact field is rejected — but the fix was applied only to the test, not to the command that generates the input.

**Fix:** Remove `"working_directory"` from the Tracer 2 input template in `security-review.md`. The OAuth auditor locates code via the `oauth_locations` fields extracted from `go-index.json` — it does not need a working directory path. The corrected input is:

```json
{
  "oauth_locations": <oauth_locations object from go-index.json>,
  "review_session_id": "<SESSION_ID>",
  "code_ref": "<CODE_REF>",
  "code_ref_dirty": <CODE_REF_DIRTY>
}
```

---

### CR-02: Two tests are explicitly marked RED (intentionally failing) and are shipping in that state

**File:** `claude-security-hooks/internal/hooks/validate_test.go:548` and `validate_test.go:612`

**Issue:** Two tests carry explicit comments saying they MUST FAIL against current code:

- `TestValidate_Synthesis_S1_MissingFile_SoftFail` (line 548): Comment reads "RED: current validate.go always blocks on S1 missing file → this test must FAIL." This test asserts the W4 soft-fail behavior — that a synthesis agent returning valid JSON should not be blocked even when `review-report.json` is absent from disk.
- `TestValidate_TaintVerdict_ErrorMessageContainsPreview` (line 612): Comment reads "RED: current validate.go/dispatch.go does not add a content preview → this test must FAIL."

Shipping known-failing tests means `go test ./...` returns non-zero, breaking any CI gate, and these tests carry no `t.Skip()` — they will actively fail. The described behaviors (W4 soft-fail and W4 error preview) are also correctness requirements for the pipeline: without the soft-fail, a valid synthesis verdict gets blocked; without the preview, operators cannot diagnose which output triggered a block.

Note: `TestPreflight_OAuthUnknownField_SchemaDocInError` (preflight_test.go:173) is also partially RED — its final three assertions about the "Expected schema:" fragment will fail unless `preflight.go` already emits the schema doc. Cross-reference CR-01: this test correctly identifies the block, but the root fix must be in `security-review.md`, not `preflight.go`.

**Fix:** Implement the W4 soft-fail logic in `validate.go` so `TestValidate_Synthesis_S1_MissingFile_SoftFail` passes. The logic described in the comment on line 61–79 of `validate.go` is the implementation; confirm it matches the test expectation. For `TestValidate_TaintVerdict_ErrorMessageContainsPreview`, verify that `dispatch.go`'s `contentPreview` call in `runTaintTracer` (line 71) is reached on parse failure and that the resulting block reason includes `"content(first 100):"`. If the current code path already emits the preview but the test comment is stale, remove the RED annotation.

---

### CR-03: `SynthesisReport.SchemaVersion` serializes as `"scma_version"` — typo is baked into wire format

**File:** `claude-security-hooks/internal/schema/synthesis.go:7`

**Issue:** The JSON struct tag on `SchemaVersion` is `json:"scma_version"` — a clear misspelling of `schema_version`. The comment acknowledges "Note: typo in spec is scma_version", which means the wire format intentionally carries the typo. The `checkS2` invariant in `invariants/synthesis.go:108` validates the path `"scma_version"`, the test fixtures at `testdata/synthesis_out/*/review-report.json` use `"scma_version"`, and the dispatch test at line 143 uses `"scma_version"`. This typo is now load-bearing protocol: any synthesis agent that emits the correct spelling `"schema_version"` will fail the S2 invariant and be blocked.

This is a correctness defect if the synthesis agent's instruction ever diverges from the Go field name, and a maintainability landmine. The spec should canonicalize the correct spelling across all surfaces or document this divergence in a single authoritative place (not just an inline comment).

**Fix:** Either fix the typo consistently across all surfaces (Go struct tag, invariant path string, test fixtures, all agent documentation, and synthesis agent schema) in a coordinated migration, or add a `// PROTOCOL-FROZEN: do not rename` comment to every callsite and add a test that explicitly asserts the wire value equals `"scma_version"` so future refactors fail loudly. The current state — typo acknowledged only in an inline comment — guarantees future breakage.

---

### CR-04: `go-authz-tracer.schema.json` handler object uses `first_param_read` but Go struct uses `first_param_read_line` / `first_param_read_expr`

**File:** `claude-security-hooks/specs/agents/go-authz-tracer.schema.json:46`

**Issue:** The `handler` object in `go-authz-tracer.schema.json` defines a single property named `first_param_read` (type string). The `Handler` Go struct in `cartographer.go` (also used in `authz_tracer.go` via the shared `Handler` type) defines two fields: `FirstParamReadLine int` (JSON `first_param_read_line`) and `FirstParamReadExpr string` (JSON `first_param_read_expr`). These names do not match. When the authz tracer schema is used to validate an input that contains `first_param_read_line` and `first_param_read_expr`, a JSON Schema validator will reject both fields as unknown (since the schema only allows `fqn`, `file`, `line`, and `first_param_read`). Conversely, a prompt containing `first_param_read` would fail the strict Go decoder. Additionally, the authz-tracer schema's `handler` object (line 33) has `"additionalProperties": false`, so any real cartographer output carrying `first_param_read_line` would be schema-invalid.

**Fix:** Update `go-authz-tracer.schema.json` handler properties to match the Go struct:

```json
"first_param_read_line": {
  "type": "integer",
  "description": "1-based line number of the first HTTP parameter read in the handler body."
},
"first_param_read_expr": {
  "type": "string",
  "description": "Expression text of the first HTTP parameter read (e.g. c.Query(\"id\"))."
}
```

Remove the `first_param_read` property.

---

### CR-05: validate.go fallback logic silently drops multi-violation reasons when first reason is a parse error

**File:** `claude-security-hooks/internal/hooks/validate.go:54–59`

**Issue:** The fallback logic at lines 54–59:

```go
reasons, _ := runAgentValidation(ev.ToolInput.SubagentType, stripped, ev.ToolInput.Prompt)
if len(reasons) == 1 && strings.Contains(reasons[0], "verdict parse") {
    fallback, _ := runAgentValidation(ev.ToolInput.SubagentType, verdictText, ev.ToolInput.Prompt)
    if len(fallback) > 0 && !strings.Contains(fallback[0], "verdict parse") {
        reasons = fallback
    }
}
```

The condition `len(reasons) == 1` makes this fallback trigger only when there is exactly one reason. If `runAgentValidation` returns multiple reasons where the first is a parse error (e.g., `reasons = ["T1: verdict parse...", "T6: ..."]`), the fallback is skipped entirely and the stale `reasons` slice (containing parse errors plus other violations from the stripped form) is used unchanged. The additional violations from the fallback on raw text are lost. The condition should be checking whether `reasons[0]` is a parse error, not whether `len(reasons) == 1`. In practice, a parse error from `newStrictDecoder` on a malformed input typically yields a single T1/A1/etc. reason, so this rarely fires — but the condition is logically wrong and will silently swallow violations if the decoder happens to return compound errors.

**Fix:**

```go
reasons, _ := runAgentValidation(ev.ToolInput.SubagentType, stripped, ev.ToolInput.Prompt)
if len(reasons) > 0 && strings.Contains(reasons[0], "verdict parse") {
    fallback, _ := runAgentValidation(ev.ToolInput.SubagentType, verdictText, ev.ToolInput.Prompt)
    if len(fallback) > 0 && !strings.Contains(fallback[0], "verdict parse") {
        reasons = fallback
    }
}
```

---

## Warnings

### WR-01: `pre-flight-checks.sh` missing `set -e` — failures in sub-checks silently continue

**File:** `bootstrap/pre-flight-checks.sh:14`

**Issue:** The script uses `set -uo pipefail` but omits `set -e`. This is intentional for the FAILURES counter pattern — the script must continue after individual failures to report all errors. However, the `rm -rf "$TARGET_DIR/.codegraph/"` on line 140 runs without any error check. If `$TARGET_DIR` resolves to an empty string (which cannot happen here because `set -u` would catch an unset var, but could happen if `realpath` succeeds on a dangling symlink), the command could expand to `rm -rf "/.codegraph/"`. While `set -u` prevents unset variable expansion, the `realpath ... || echo "${1}"` fallback on line 20 means `TARGET_DIR` can be set to a user-supplied path that has not been validated. If the user passes `TARGET_DIR=../../../`, the `rm -rf` on line 140 will delete `../../../.codegraph/` — which could be outside the intended directory.

**Fix:** Add a guard before the `rm -rf`:

```bash
if [ -z "$TARGET_DIR" ] || [ "$TARGET_DIR" = "/" ]; then
  echo "ERROR: TARGET_DIR is empty or root — refusing rm -rf" >&2
  FAILURES=$((FAILURES+1))
else
  rm -rf "$TARGET_DIR/.codegraph/"
fi
```

Also validate that `TARGET_DIR` is a non-root absolute path before any destructive operations.

---

### WR-02: `go-cartographer.md` Step 3 metavariable pattern for `net/http` is inconsistent with other routers

**File:** `.claude/agents/go-cartographer.md:72`

**Issue:** The Step 3 Semgrep patterns for all routers use receiver variable metavariables (e.g., `$ROUTER.GET(...)`, `$R.Get(...)`, `$APP.Get(...)`), but the `net/http` entry lists patterns that are already variable-name-agnostic: `http.HandleFunc(...)`, `http.Handle(...)`, `mux.HandleFunc(...)`, `mux.Handle(...)`. The description says these are "already variable-name-agnostic" but this is incorrect for the `mux.HandleFunc(...)` and `mux.Handle(...)` patterns — `mux` here is a variable name (typically `http.NewServeMux()` receiver), not a package name. A codebase using `myMux.HandleFunc(...)` would not be matched by the literal `mux.HandleFunc(...)` pattern. The agent instructions should use `$MUX.HandleFunc(...)` or note explicitly that `mux` is used only as a conventional name.

**Fix:** Update the `net/http` pattern in Step 3 to:

```
`net/http`: match `http.HandleFunc(...)`, `http.Handle(...)`, `$MUX.HandleFunc($PATH, ...)`, `$MUX.Handle($PATH, ...)`
```

---

### WR-03: `go-cartographer.md` Step 3.5 uses gin-only permissive pattern to validate all detected routers

**File:** `.claude/agents/go-cartographer.md:81`

**Issue:** Step 3.5 instructs the agent to re-scan route-registration files using the permissive gin pattern `$ROUTER.{GET,POST,DELETE,PATCH,PUT,Handle}($PATH, ...)`. This pattern is gin-specific — chi uses `r.Get/r.Post`, echo uses `e.GET/e.POST`, fiber uses `app.Get/app.Post`, etc. For a repository using chi or echo, the Step 3.5 cross-reference scan would find zero matches (since chi methods have different casing), potentially generating false `route_registration_gap` warnings for every route, or silently producing zero gaps even though routes were missed. The net result is Step 3.5 only works correctly for gin applications.

**Fix:** Step 3.5 should run a router-appropriate pattern for each detected router, not a single gin-only pattern. For example:

```
For each router in routers_detected:
  - gin: $ROUTER.{GET,POST,DELETE,PATCH,PUT,Handle}($PATH, ...)
  - chi: $R.{Get,Post,Delete,Patch,Put,Handle}($PATH, ...)
  - echo: $E.{GET,POST,DELETE,PATCH,PUT}($PATH, ...)
  - fiber: $APP.{Get,Post,Delete,Patch,Put}($PATH, ...)
  - net/http: http.HandleFunc(...), $MUX.HandleFunc(...)
```

---

### WR-04: `PaymentSurface.ClusterID` serializes as empty string unconditionally — schema spec says omit it

**File:** `claude-security-hooks/internal/schema/cartographer.go:73`

**Issue:** The `PaymentSurface` struct has `ClusterID string \`json:"cluster_id"\`` without `omitempty`. The agent instruction (`go-cartographer.md` line 110) states `payment_surface.cluster_id` is "omitted — codegraph does not surface community IDs." The go-index schema example shows no `cluster_id` field. However, the Go struct will always serialize to `"cluster_id": ""`. When `dispatch.go` decodes a cartographer verdict with `newStrictDecoder` (which uses `DisallowUnknownFields()`), this empty string will be accepted because the struct includes the field — but any downstream consumer using the JSON schema with `additionalProperties: false` will see an unexpected field. More importantly, the spec says the field should not appear; the implementation contradicts the spec.

**Fix:** Add `omitempty` to the struct tag:

```go
ClusterID string `json:"cluster_id,omitempty"`
```

---

### WR-05: `validate.go` `os.Chdir` side-effect persists across concurrent test runs

**File:** `claude-security-hooks/internal/hooks/validate.go:39–43`

**Issue:** `Validate()` calls `os.Chdir(root)` on line 40 and restores via `defer os.Chdir(prev)`. `os.Chdir` changes the process working directory, which is a global OS-level state — it affects all goroutines. If tests run with `go test -parallel N` or if the hooks binary ever handles concurrent requests, a `Chdir` in one goroutine will corrupt path resolution for all others. The `validate_test.go` tests call `os.Chdir(tmpdir)` directly (lines 28–29 et al.) without protecting against concurrent test execution. This is latent: unit tests that run with `-count=1` and no `-parallel` may pass; tests with higher parallelism will race.

**Fix:** Replace `os.Chdir` with explicit path prefix handling. Pass the workspace root as a parameter to `checkS1Dir` and other invariants that need filesystem access, so they operate on absolute paths rather than CWD-relative paths. Alternatively, use `sync.Mutex` around `Chdir`/invariant check/`Chdir` back — but that only masks the symptom.

---

### WR-06: `preflight_test.go` test `TestPreflight_OAuthUnknownField_SchemaDocInError` will partially fail — "Expected schema:" assertions RED

**File:** `claude-security-hooks/internal/hooks/preflight_test.go:173`

**Issue:** The test at line 173 asserts three properties about the block reason:
1. Decision is "block" (line 207–210) — will pass, since the unknown field triggers `DisallowUnknownFields`.
2. Reason contains "D-09" (line 212–215) — will pass.
3. Reason contains "Expected schema:" (line 219–221) — will FAIL. The `preflightPerAgent` function at `preflight.go:70–75` emits `fmt.Sprintf("D-09: oauth input parse: %v\nExpected schema: %s", err, oauthInputSchemaDoc)` — so this SHOULD pass if `oauthInputSchemaDoc` is included. Cross-reference: the `go-oauth-auditor` case is at line 69–75 and does include the schema doc in the format string, meaning assertions 2–3 will pass. However, assertion 4 (`oauth_locations` in reason, line 223) and assertion 5 (`additionalProperties` in reason, line 227) depend on whether `oauthInputSchemaDoc` contains those strings. Looking at the constant on line 18 of `preflight.go`: it does include `oauth_locations` and `additionalProperties:false`. So this test may actually pass in full. The RED comment in `validate_test.go` may be stale. This needs runtime verification.

This is flagged as WARNING because the test comment says "RED: current preflight.go only emits the parse error, no schema doc → must FAIL" (line 172) but the actual code at line 74 in `preflight.go` does include `oauthInputSchemaDoc`. Either the comment is stale (test was fixed but comment not updated) or there is a code path where the schema doc is not emitted. The discrepancy is a correctness risk.

**Fix:** Run `go test ./... -run TestPreflight_OAuthUnknownField` and verify the result. If the test passes, remove the RED comment. If it fails, the test correctly identifies a bug — document which assertion fails and fix it.

---

## Info

### IN-01: All `schema_test.go` E1 tests are `t.Skip`-ped — zero coverage of CodeRef round-trip

**File:** `claude-security-hooks/internal/schema/schema_test.go:58`

**Issue:** Ten `TestXxx_CodeRefRoundTrip` tests (lines 22–217) are all body-commented-out with `t.Skip(...)`. The skip message says "E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to X (Wave 1)." But the Go schema structs already have `CodeRef` and `CodeRefDirty` fields (e.g., `TaintInput` at `taint_tracer.go:10–11`, `CartographerIndex` at `cartographer.go:7–8`). The fields are present. These tests should be un-skipped and their commented-out bodies activated so actual round-trip behavior is verified.

**Fix:** Un-skip these tests and activate their commented-out bodies. The fields already exist; the E1 compile-error condition has been resolved by Wave 1's implementation.

---

### IN-02: `go-cartographer.md` Step 3 `first_param_read_line` is gin-only despite multi-router context

**File:** `.claude/agents/go-cartographer.md:73–74`

**Issue:** Step 3's instruction to scan handler bodies for the first HTTP parameter read lists only gin-specific expressions: `c.Query(...)`, `c.PostForm(...)`, `c.Param(...)`, `c.ShouldBind*(...)`, `c.GetRawData()`. For chi, net/http, echo, or fiber handlers, the equivalent expressions are different (`r.URL.Query().Get(...)`, `r.FormValue(...)`, `c.QueryParam(...)`, `c.FormValue(...)`, etc.). The step does not instruct the agent to select the appropriate pattern based on `router` type, so for non-gin applications `first_param_read_line` will always be absent from the output, causing the `security-review.md` Tracer 4 logic (line 179) to always fall back to `handler.line`.

**Fix:** Expand the Step 3 instruction to cover per-router parameter read patterns, matching the source kind table in `go-taint-tracer.md` Section 4.

---

### IN-03: `go-authz-tracer.md` AZ3 bucket accounting allows route to appear in multiple buckets — summary can exceed `routes_total`

**File:** `.claude/agents/go-authz-tracer.md:170`

**Issue:** Hard rule AZ3 states: "A route may appear in multiple buckets if it triggers multiple findings; dedup findings per route in the findings array but count in each bucket." This means `protected + missing + weak + idor_risk + public_intentional` can be greater than `routes_total`. The rule says these values "MUST account for all routes" — but that phrase is ambiguous. If an `idor_risk` route is also `protected`, it is counted in both `protected` and `idor_risk`, making the sum exceed `routes_total`. This conflicts with the natural interpretation of an accounting invariant and may confuse consumers of the summary who expect the buckets to sum to `routes_total`. The `AuthzSummary` Go struct and the `AuthzInvariants` check should document this explicitly.

**Fix:** Clarify AZ3 with a concrete example: "A route with IDOR risk that also has authz coverage counts in both `protected` and `idor_risk`; the sum of all buckets may exceed `routes_total`." Add a corresponding note in the `AuthzSummary` struct comment.

---

### IN-04: `dispatch.go` `contentPreview` truncates at byte offset 100, not Unicode-safe

**File:** `claude-security-hooks/internal/hooks/dispatch.go:47`

**Issue:** `contentPreview` uses `s[:100]` for truncation. In Go, slicing a string at a byte index is not guaranteed to align with UTF-8 character boundaries. If the verdict text contains multi-byte Unicode characters at or near byte 100, `s[:100]` will produce a string with an invalid UTF-8 sequence at the cut point. `fmt.Sprintf("%q", ...)` will emit `\x??` escapes for the broken bytes rather than crashing, but the preview will be garbled. This is a diagnostic output quality issue, not a crash.

**Fix:** Use `utf8.RuneCountInString` or `[]rune(s)[:100]` for character-safe truncation:

```go
runes := []rune(s)
if len(runes) > 100 {
    return fmt.Sprintf("%q", string(runes[:100]))
}
return fmt.Sprintf("%q", s)
```

---

_Reviewed: 2026-05-28T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
