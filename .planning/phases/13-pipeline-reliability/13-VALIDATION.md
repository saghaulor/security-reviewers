---
phase: 13
slug: pipeline-reliability
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-05-27
---

# Phase 13 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` package |
| **Config file** | none — `go test ./...` from `claude-security-hooks/` |
| **Quick run command** | `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./internal/hooks/ -run TestValidate -v` |
| **Full suite command** | `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./... -count=1` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./internal/hooks/ -run TestValidate -v`
- **After every plan wave:** Run `go test -C /home/saghaulor/code/security_reviewer/claude-security-hooks ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------------|-----------|-------------------|-------------|--------|
| 13-01-T1 | 13-01 | 0 | W4-ERRMSG | Block reason includes `content(first 100)` preview | unit (RED) | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestValidate_TaintVerdict_ErrorMessageContainsPreview -v` | ❌ W0 | ⬜ pending |
| 13-01-T1-AZ | 13-01 | 0 | W4-ERRMSG | AZ block reason includes preview | unit (RED) | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestValidate_AuthzVerdict_ErrorMessageContainsPreview -v` | ❌ W0 | ⬜ pending |
| 13-01-T1-OA | 13-01 | 0 | W4-ERRMSG | OA block reason includes preview | unit (RED) | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestValidate_OAuthVerdict_ErrorMessageContainsPreview -v` | ❌ W0 | ⬜ pending |
| 13-01-T1-IC | 13-01 | 0 | W4-ERRMSG | IC block reason includes preview | unit (RED) | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestValidate_InvariantVerdict_ErrorMessageContainsPreview -v` | ❌ W0 | ⬜ pending |
| 13-01-T2 | 13-01 | 0 | W4-S1-RETRY | S1 check warns-not-blocks on missing synthesis file | unit (RED) | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestValidate_Synthesis_S1_MissingFile_SoftFail -v` | ❌ W0 | ⬜ pending |
| 13-01-T2b | 13-01 | 0 | W5-SCHEMA | D-09 error includes expected schema string | unit (RED) | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestPreflight_OAuthUnknownField_ErrorContainsSchema -v` | ❌ W0 | ⬜ pending |
| 13-01-T3 | 13-01 | 0 | W7-PARAM | Handler.FirstParamReadLine field round-trips | unit (RED) | `go test -C .../claude-security-hooks ./internal/schema/ -run TestHandler_FirstParamReadLine -v` | ❌ W0 | ⬜ pending |
| 13-02-T1 | 13-02 | 1 | SC-13-1 | bootstrap script exits 0 on correct env | integration | `bash /home/saghaulor/code/security_reviewer/bootstrap/pre-flight-checks.sh /home/saghaulor/code/security_reviewer/examples/sample-vulnerable-service` | ❌ W1 | ⬜ pending |
| 13-02-T2 | 13-02 | 1 | SC-13-1 | make preflight target invokes script | integration | `make -C /home/saghaulor/code/security_reviewer preflight TARGET=/home/saghaulor/code/security_reviewer/examples/sample-vulnerable-service` | ❌ W1 | ⬜ pending |
| 13-03-T1 | 13-03 | 1 | SC-13-2 | security-review.md has JSON enforcement instruction | structural | `grep -c "JSON prompt enforcement" /home/saghaulor/code/security_reviewer/.claude/commands/security-review.md` | ✅ | ⬜ pending |
| 13-03-T2 | 13-03 | 1 | SC-13-3 | All 5 tracer agents have Write in tools frontmatter | structural | `grep -l "Write" /home/saghaulor/code/security_reviewer/.claude/agents/go-cartographer.md ...` | ✅ | ⬜ pending |
| 13-04-T1 | 13-04 | 1 | W4-ERRMSG | All 5 agent parse-failure paths include content preview (GREEN) | unit | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestValidate.*ErrorMessageContainsPreview -v` | ❌ W0 deps | ⬜ pending |
| 13-04-T2 | 13-04 | 1 | W5-SCHEMA | D-09 error includes expected schema string (GREEN) | unit | `go test -C .../claude-security-hooks ./internal/hooks/ -run TestPreflight_OAuthUnknownField -v` | ❌ W0 deps | ⬜ pending |
| 13-04-T3 | 13-04 | 1 | W7-PARAM | Handler.FirstParamReadLine round-trips (GREEN) + full suite | unit | `go test -C .../claude-security-hooks ./... -count=1` | ❌ W0 deps | ⬜ pending |
| 13-05-T1 | 13-05 | 2 | SC-13-6 | 6 JSON Schema files exist and are valid JSON | structural | `find .../claude-security-hooks/specs/agents/ -name "*.schema.json" -type f \| wc -l` | ❌ W1 deps | ⬜ pending |
| 13-06-T1 | 13-06 | 3 | SC-13-5 | go-cartographer.md contains Step 3.5 cross-reference + $ROUTER pattern | structural | `grep -c "Step 3.5" /home/saghaulor/code/security_reviewer/.claude/agents/go-cartographer.md` | ✅ | ⬜ pending |
| 13-06-T2 | 13-06 | 3 | SC-13-7 | security-review.md uses first_param_read_line for taint source | structural | `grep -c "first_param_read_line" /home/saghaulor/code/security_reviewer/.claude/commands/security-review.md` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `claude-security-hooks/internal/hooks/validate_test.go` — add RED tests for W4 error message format (taint + authz + oauth + invariant) and W4 S1 soft-fail
- [ ] `claude-security-hooks/internal/hooks/preflight_test.go` — add RED test for W5 schema doc in D-09 error
- [ ] `claude-security-hooks/internal/schema/schema_test.go` — add RED test for W7 Handler.FirstParamReadLine field round-trip

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| All 9 tracers succeed first attempt (JSON enforcement) | SC-13-2 | Agent execution is not unit-testable; requires live Claude Code session | Run `/security-review examples/sample-vulnerable-service` after Phase 13 execution; confirm 0 manual prompt reformats needed |
| All verdict files written autonomously (Write perms) | SC-13-3 | Write permission requires a live agent session | Run full `/security-review`; confirm `authz-findings.json`, `oauth-checklist.json`, `invariant-results.json`, `taint-verdict-*.json`, `review-report.{json,md}` are all written without orchestrator Write calls |
| Cartographer detects callChainSQLiHandler (route cross-ref) | SC-13-5 | Agent output can only be verified by running the cartographer agent | Run `/security-review examples/sample-vulnerable-service`; verify `go-index.json` entrypoints include `callChainSQLiHandler` or that a `warnings` array flags the gap |
| go-index.json includes first_param_read_line (W7) | SC-13-7 | Agent output requires live run | After execution, inspect `examples/sample-vulnerable-service/go-index.json` entrypoints for `first_param_read_line` fields |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (13-01 Task 1/2/3)
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** 2026-05-27
