---
phase: 13-pipeline-reliability
plan: "06"
subsystem: agent-instructions
tags: [cartographer, semgrep, route-detection, taint-tracer, source-precision]
dependency_graph:
  requires: [13-03, 13-04, 13-05]
  provides: [go-cartographer-step3-metavariables, go-cartographer-step3.5-cross-reference, taint-tracer-source-precision]
  affects: [.claude/agents/go-cartographer.md, .claude/commands/security-review.md]
tech_stack:
  added: []
  patterns: [semgrep-metavariable-patterns, route-registration-gap-warnings, first-param-read-fields]
key_files:
  created: []
  modified:
    - .claude/agents/go-cartographer.md
    - .claude/commands/security-review.md
decisions:
  - "Metavariable Semgrep patterns ($ROUTER.GET) chosen over literal patterns to be variable-name-agnostic across all Go gin codebases"
  - "Step 3.5 uses codegraph_search for FQN resolution of gap routes — unresolvable routes go to warnings (not entrypoints) per A3/A6"
  - "first_param_read_line fields are omitempty in handler output — no regressions for handlers with no gin parameter reads"
metrics:
  duration_minutes: 12
  tasks_completed: 2
  tasks_total: 2
  files_modified: 2
  completed_date: "2026-05-28"
---

# Phase 13 Plan 06: W6+W7 Agent-Side Route Detection and Source Precision Summary

**One-liner:** Metavariable Semgrep patterns + Step 3.5 route cross-reference for variable-name-agnostic gin route detection; first_param_read_line/Expr fields for taint-tracer source precision.

## What Was Built

### Task 1: go-cartographer.md — Step 3 metavariable patterns + Step 3.5 + first_param_read fields

**Problem (W6):** The go-cartographer Step 3 used literal receiver patterns (`r.GET(...)`, `r.POST(...)`) that only matched gin routers where the variable was named `r`. The sample-vulnerable-service uses `router` as the variable name, causing `callChainSQLiHandler` at `main.go:39` (`router.GET("/api/advanced-search", callChainSQLiHandler)`) to be missed — producing 7 entrypoints instead of 8.

**Fix 1 — Metavariable patterns:** All router-specific patterns in Step 3 now use Semgrep metavariables:
- gin: `$ROUTER.GET($PATH, ...)`, `$ROUTER.POST(...)`, `$ROUTER.DELETE(...)`, `$ROUTER.PATCH(...)`, `$ROUTER.PUT(...)`, `$ROUTER.Handle($METHOD, $PATH, ...)`, `$ROUTER.Group($PATH, ...)`
- chi: `$R.Get($PATH, ...)`, `$R.Post(...)`, `$R.Route(...)`, `$R.Group(...)`, `$R.Use(...)`
- gorilla/mux: `$R.HandleFunc(...)`, `$R.Handle(...)`, `$R.PathPrefix(...)`
- echo: `$E.GET(...)`, `$E.POST(...)`, `$E.Group(...)`
- fiber: `$APP.Get(...)`, `$APP.Post(...)`, `$APP.Group(...)`
- httprouter: `$ROUTER.GET(...)`, `$ROUTER.POST(...)`, `$ROUTER.Handle($METHOD, $PATH, ...)`
- net/http: unchanged (already variable-name-agnostic)

**Fix 2 — first_param_read instruction (W7 cartographer side):** Step 3 now instructs the agent to scan each handler body for the first HTTP parameter read expression (gin kinds: `c.Query(...)`, `c.PostForm(...)`, `c.Param(...)`, `c.ShouldBind*(...)`, `c.GetRawData()`). The line number is recorded as `handler.first_param_read_line` and the expression text as `handler.first_param_read_expr` (both fields omitted if no parameter read found).

**Fix 3 — Step 3.5 post-processing cross-reference:** New section between Step 3 and Step 4 performs validation after Semgrep enumeration:
1. Reads `main.go` plus any `router.go`, `routes.go`, `server.go` via Glob
2. Runs `mcp__opengrep__scan_with_rule` with permissive gin pattern to collect all `($METHOD, $PATH, $HANDLER_NAME)` tuples
3. Compares tuples against built entrypoints array — for absent paths: resolves FQN via `mcp__codegraph__codegraph_search`, adds to entrypoints if resolvable, adds `route_registration_gap` warning if not
4. Emits zero warnings if no gaps found

**A5-supplement:** Step 3.5 explicitly states that unresolvable gap routes MUST appear in `warnings` — not silently dropped.

### Task 2: security-review.md — Step 6 source construction uses first_param_read_line

**Problem (W7 orchestration side):** Step 6 Tracer 4 constructed taint-tracer source inputs using `handler.line` (function definition line). For `callChainSQLiHandler`, handler.line=254 but first HTTP read is line 256 (`c.Query("search")`). Tracers received definition line rather than parameter read line.

**Fix:** Added instruction before the taint-tracer JSON template:
> "When constructing the `source` field, use `first_param_read_line` from the entrypoint's `handler` object if it is present and non-zero; otherwise fall back to `handler.line` (the function definition line). Use `first_param_read_expr` as the `expr` value if present; otherwise use the inferred parameter read expression."

Updated the template `source.line` comment from `<handler line from entrypoints>` to `<first_param_read_line from entrypoints if present, otherwise handler.line>` and `source.expr` to reference `first_param_read_expr`.

JSON prompt enforcement instructions from 13-03 are preserved intact.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 1 | 7f26684 | feat(13-06): W6+W7 — cartographer Step 3 metavariable patterns + Step 3.5 + first_param_read fields |
| Task 2 | 07501ea | feat(13-06): W7 — Step 6 taint-tracer source uses first_param_read_line when present |

## Acceptance Criteria Verification

| Criterion | Result |
|-----------|--------|
| `grep -c "\$ROUTER.GET" go-cartographer.md >= 1` | 2 matches |
| `grep -c "route_registration_gap" go-cartographer.md >= 1` | 2 matches |
| `grep -c "Step 3.5" go-cartographer.md >= 1` | 3 matches |
| `grep -c "first_param_read_line" go-cartographer.md >= 1` | 1 match |
| `grep -c "first_param_read_expr" go-cartographer.md >= 1` | 1 match |
| File still has "## 6. Hard Rules" section | PASS |
| File still has Step 4 header | PASS |
| `grep "r\.GET" go-cartographer.md` — old literal in Step 3 removed | PASS (only reference table rows remain) |
| `grep -c "first_param_read_line" security-review.md >= 1` | 2 matches |
| File still has "## Step 6: Run tracers (parallel)" | PASS |
| File still has "### Tracer 4: go-taint-tracer" | PASS |
| JSON prompt enforcement instructions from 13-03 present | PASS |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — both files have complete, wired instructions with no placeholder content.

## Threat Flags

None — no new network endpoints, auth paths, file access patterns, or schema changes. Edits are documentation-only changes to agent instruction markdown files.

## Self-Check: PASSED

- `.claude/agents/go-cartographer.md` — exists and modified
- `.claude/commands/security-review.md` — exists and modified
- Commit 7f26684 — present in git log
- Commit 07501ea — present in git log
