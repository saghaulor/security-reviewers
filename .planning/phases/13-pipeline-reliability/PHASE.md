# Phase 13: Pipeline Reliability + Bootstrap Hardening

**Status:** Not planned yet  
**Source documents:** `BOOTSTRAP_REQUIREMENTS.md`, `SECURITY_REVIEW_SCAN_HANDOFF.md`  
**Depends on:** Phase 12

---

## Goal

Eliminate all 7 failure modes documented in `SECURITY_REVIEW_SCAN_HANDOFF.md` and all missing pre-flight checks documented in `BOOTSTRAP_REQUIREMENTS.md`. The pipeline reaches ≥95% first-attempt success rate for fully autonomous execution — no manual JSON reformatting, no manual file writes, no false-negative hook blocks.

---

## Failure Modes to Fix (from SECURITY_REVIEW_SCAN_HANDOFF.md)

| # | Issue | Severity | Root Cause | Time Lost |
|---|-------|----------|-----------|-----------|
| 1 | Cartographer missing route detection (`callChainSQLiHandler`) | Medium | gin router DSL edge case in codegraph static parse | 3 min |
| 2 | OAuth-auditor input schema mismatch (`working_directory` rejected) | High | No schema docs; field inferred incorrectly | 75 sec |
| 3 | **All 9 tracers failed — JSON serialization** | **Critical** | Prompt was natural language + JSON, not pure JSON | 105 sec |
| 4 | **Agents cannot write verdict files** (A10 read-only enforcement) | **High** | Permission model too restrictive; no scoped write allowlist | 150 sec |
| 5 | Validation hooks reporting false negatives | Medium | Hooks fired before files were written (timing race) | 105 sec |
| 6 | Source line number mismatches (func def line vs param read line) | Low | Orchestrator used handler def lines, agents handled gracefully | 0 sec |
| 7 | Synthesis agent file write failure (same as Issue 4) | High | Same permission model issue | 30 sec |

---

## Workstreams

### W1 – Bootstrap pre-flight checks (from BOOTSTRAP_REQUIREMENTS.md)
- Implement `bootstrap/pre-flight-checks.sh` with 6 sequential checks:
  1. Codegraph installation (install if missing via `go install`)
  2. Make + Go build-tool verification with fast-fail and clear remediation messages
  3. Git + repository verification with `git rev-parse` guard
  4. Go project structure (go.mod required, go.sum warning with `go mod tidy` hint)
  5. Docker availability (export `DOCKER_AVAILABLE` env for downstream steps)
  6. Codegraph init + index with retry-on-corruption logic (remove `.codegraph/` and retry)
- Wire into `Makefile` as a `preflight` target
- **Estimated effort:** 4–6 hrs

### W2 – Fix JSON serialization for agent dispatch (Critical — Issue 3)
All 9 tracer agents failed on first attempt because the orchestrator passed natural-language text instead of pure JSON as the `prompt` field.
- Add explicit "prompt must be valid JSON" instruction in `security-review.md` skill at each agent dispatch site
- Add per-agent JSON example templates in each agent's SKILL.md or spec
- Add a pre-dispatch validation step in the orchestration command that fails fast with a helpful error if the prompt is not JSON
- **Estimated effort:** 2–4 hrs

### W3 – Grant scoped write permissions to tracer agents (High — Issues 4, 7)
Agents produced correct output but couldn't persist verdict files (A10 read-only enforcement).
- Define per-agent write path allowlists:
  - `go-cartographer` → `<TARGET_DIR>/go-index.json`
  - `go-authz-tracer` → `<TARGET_DIR>/authz-findings.json`
  - `go-oauth-auditor` → `<TARGET_DIR>/oauth-checklist.json`
  - `go-taint-tracer` → `<TARGET_DIR>/taint-verdict-*.json`
  - `invariant-checker` → `<TARGET_DIR>/invariant-results.json`
  - `synthesis` → `<TARGET_DIR>/review-report.{json,md}`
- Implement path-pattern matching (allowlist only; no parent-directory escalation)
- Update A10/A11 policy documentation to clarify which agents may write
- **Estimated effort:** 4–6 hrs

### W4 – Fix validation hook timing / false negatives (Medium — Issues 5, 7)
Post-tool-use hooks fired before agent verdict files were written, producing false parse errors even though files were valid.
- Gate schema validation until all expected verdict files are present (file-existence pre-check)
- Improve error messages: include file path, expected vs. actual first 100 chars, line number of parse failure
- Add `--debug-validate` flag: logs schema being checked + file content before parse attempt
- **Estimated effort:** 2–3 hrs

### W5 – Document and validate agent input schemas (High — Issue 2)
`go-oauth-auditor` rejected `working_directory` because there was no schema documentation.
- Create `claude-security-hooks/specs/agents/<agent>.schema.json` for all 5 specialist agent types
- Add pre-dispatch schema validation: reject with the expected schema in the error message
- Include schema in D-09 error messages (when input JSON parse fails, show expected fields)
- **Estimated effort:** 4–6 hrs

### W6 – Fix cartographer missing route detection (Medium — Issue 1)
`GET /api/advanced-search → callChainSQLiHandler` was absent from cartographer entrypoints despite being registered at `main.go:39`.
- Add post-processing step in `go-cartographer.md`: parse all `router.{GET,POST,DELETE,PATCH,PUT}()` calls
- Cross-reference against detected entrypoints; emit `warnings` array in `go-index.json` for any gaps
- Emit explicit warning message when a registered route is not in the detected entrypoints
- **Estimated effort:** 6–8 hrs

### W7 – Enhance cartographer output with HTTP parameter read locations (Low — Issue 6)
Orchestrator used handler function-definition lines as taint-tracer sources; agents handled the mismatch gracefully but with `input_mismatch` warnings.
- Add `first_param_read_line` and `first_param_read_expr` fields to each entrypoint entry in `go-index.json`
- `go-cartographer.md` instructions: report all HTTP parameter read lines (Query, PostForm, Param, ShouldBind)
- Update orchestration skill to use `first_param_read_line` when constructing taint-tracer source inputs
- **Estimated effort:** 4–6 hrs

---

## Priority Order (by time saved / impact)

| Priority | Workstream | Rationale | Time Saved |
|----------|-----------|-----------|-----------|
| 1 | W2 — JSON serialization | Critical; unblocks all 9 tracers on first attempt | 105 sec |
| 2 | W3 — Agent write permissions | Enables full autonomy; eliminates all manual Write calls | 150 sec |
| 3 | W4 — Hook timing / false negatives | Reliability + clear error reporting | 135 sec |
| 4 | W1 — Bootstrap pre-flight | Prevents fresh-environment failures; CI readiness | — |
| 5 | W5 — Schema docs + dispatch validation | Prevents schema mismatch errors like Issue 2 | 75 sec |
| 6 | W6 — Cartographer route detection | Correctness gap; prevents silent missing-route issues | 3 min |
| 7 | W7 — Param read line locations | Quality improvement; agents already handle gracefully | 0 sec |

---

## Estimated Total Effort

| Workstream | Estimate |
|-----------|---------|
| W1 Bootstrap | 4–6 hrs |
| W2 JSON serialization | 2–4 hrs |
| W3 Agent write permissions | 4–6 hrs |
| W4 Hook timing | 2–3 hrs |
| W5 Schema docs | 4–6 hrs |
| W6 Cartographer routes | 6–8 hrs |
| W7 Param read locations | 4–6 hrs |
| **Total** | **26–39 hrs** |

---

## Success Criteria

1. `bootstrap/pre-flight-checks.sh` exits 0 on a correct environment and non-zero with actionable remediation messages when any required tool is missing.
2. All 9 tracer agents succeed on **first attempt** against `examples/sample-vulnerable-service` with no manual intervention.
3. All verdict files are written **autonomously by agents** — zero orchestrator Write calls required.
4. Validation hooks report **0 false negatives** after all agents complete.
5. Cartographer detects **all 8 routes** from `examples/sample-vulnerable-service/main.go`.
6. Agent input schema docs exist for **all 5 specialist agent types**.
7. `go-index.json` entrypoints include `first_param_read_line` and `first_param_read_expr`.
