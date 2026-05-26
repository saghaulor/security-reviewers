---
phase: 10-hook-compat-code-ref
plan: 03
subsystem: configuration-wiring
tags:
  - documentation
  - configuration
  - code-reference
  - agent-definitions
duration: 45m
completed_date: "2026-05-26"
---

# Phase 10 Plan 03: Hook Compatibility Code Reference Configuration — Summary

**Objective:** Apply all text and configuration changes that wire the Wave 1 Go implementation into the running system. No Go code changes in this plan.

**One-liner:** Added T9-PATH workspace-relative path rule to taint tracer; wired code_ref through cartographer, all four tracers, and synthesis; implemented binary freshness check and CODE_REF computation in security-review command.

## Overview

This plan completes the configuration and documentation wiring for Wave 1's code reference support (requirement E1) and the binary freshness gate (requirement B4). Six agent definition files updated with code_ref echo/propagation instructions, the T9-PATH rule added to enforce workspace-relative paths (B3), and the security-review orchestration command extended with Step 1.5 (unconditional binary rebuild) and Step 2 CODE_REF computation (git tree hash + dirty flag).

## Implementation Details

### Task 1: go-taint-tracer.md — T9-PATH Hard Rule
**File:** `.claude/agents/go-taint-tracer.md`
**Commit:** 7f4a7d4

Added Section 6 (Hard Rules) with new **T9-PATH** rule:
- Requires all `file` fields in `path[]` entries be workspace-relative
- Examples of correct paths: `examples/service/handlers.go`, `internal/auth/middleware.go`
- Examples of incorrect paths: absolute paths, relative-to-CWD paths
- Rule enforces verification via Read/Glob before recording file paths in verdict

### Task 2: go-cartographer.md — Code Ref Computation
**File:** `.claude/agents/go-cartographer.md`
**Commit:** 5756313

**Input Contract:** Added `code_ref` and `code_ref_dirty` fields to minimal valid input with description.

**Protocol:** Added **Step 1.5: Write code identity to output**
- Copy `code_ref` and `code_ref_dirty` verbatim from input to go-index.json output
- Do not recompute; cartographer is passive receiver
- Omit both fields if input code_ref is absent or empty

**Output Schema:** Added `code_ref` and `code_ref_dirty` fields to go-index/v1 JSON example.

### Task 3: Four Tracer Updates — Code Ref Echo Instructions
**Files:** `.claude/agents/go-authz-tracer.md`, `go-oauth-auditor.md`, `invariant-checker.md`, `go-taint-tracer.md`
**Commits:** 3bd10eb, 4cc9417, 62af270, 5796725

For each tracer:

**Input Contract:** Added `code_ref` and `code_ref_dirty` with description:
- Copy verbatim from input into verdict output
- Enables joint invariant (AZ-CodeRef, OA-CodeRef, IC-CodeRef, T-CodeRef) to verify correctness

**Output Schema:** Added `code_ref` and `code_ref_dirty` fields to verdict JSON example.

**Field Notes:** Added explanation that both fields echo input verbatim; omit if input code_ref is empty (omitempty).

### Task 4: synthesis.md — Code Ref Propagation
**File:** `.claude/agents/synthesis.md`
**Commit:** ec31e87

**Step 1 (Read all specialist outputs):** Added code_ref extraction block:
- Extract `code_ref` from go-index.json (default "" if absent)
- Extract `code_ref_dirty` from go-index.json (default false if absent)
- Propagate to review-report.json in Step 5
- Cartographer is single source of truth; synthesis does NOT recompute

**Step 5 (Write output files):** Added `code_ref` and `code_ref_dirty` to review-report.json schema.

**Field Notes:** Added documentation that both fields propagate from go-index.json; omit if empty (omitempty).

### Task 5: security-review.md — Step 1.5, Step 2 CODE_REF, Steps 4+6
**File:** `.claude/commands/security-review.md`
**Commit:** c35bde8

**Step 1.5 (Ensure hooks binary is current):** Inserted between Step 1 and Step 2:
- Run `make -C /home/saghaulor/code/security_reviewer/claude-security-hooks install`
- Unconditional rebuild + reinstall before any agent dispatch
- ~2s cost per review run acceptable vs. risk of stale binary

**Step 2 (Extended to compute CODE_REF):**
- Find repo root: `git -C TARGET_DIR rev-parse --show-toplevel` → REPO_ROOT
- If fails, set CODE_REF="" and CODE_REF_DIRTY=false; skip remaining git steps
- Compute relative path: `realpath --relative-to=REPO_ROOT TARGET_DIR` → RELATIVE_PATH
- Get tree hash: `git -C REPO_ROOT rev-parse HEAD:RELATIVE_PATH` → CODE_REF
- Check dirty: `git -C REPO_ROOT status --porcelain RELATIVE_PATH`
  - Non-empty output → CODE_REF_DIRTY=true
  - Empty output → CODE_REF_DIRTY=false

**Step 4 (Cartographer input):** Added code_ref and code_ref_dirty fields to Task input JSON.

**Step 6 (All four tracers):** Added code_ref and code_ref_dirty to each tracer's Task input JSON:
- go-authz-tracer
- go-oauth-auditor
- invariant-checker
- go-taint-tracer (all four instances)

## Success Criteria Met

- ✅ go-taint-tracer.md §6 contains T9-PATH hard rule requiring workspace-relative paths
- ✅ go-cartographer.md updated: Input Contract + Protocol (Step 1.5) + Output Schema with code_ref
- ✅ 4 tracers updated: Input Contract + Output Schema with code_ref echo instructions
- ✅ synthesis.md updated: Step 1 extraction + Step 5 propagation of code_ref
- ✅ security-review.md Step 1.5 binary rebuild added
- ✅ security-review.md Step 2 CODE_REF computation added (git commands)
- ✅ security-review.md Step 4 cartographer + Step 6 tracers include code_ref/code_ref_dirty
- ✅ Verification checklist: all mandatory strings present in all files
- ✅ Each task committed individually with proper formatting
- ✅ SUMMARY.md created with complete implementation record

## Verification Checklist

| File | Search term | Found | Location |
|------|-------------|-------|----------|
| go-taint-tracer.md | `workspace-relative` | ✅ | §6 Hard Rules |
| go-cartographer.md | `code_ref` | ✅ | Input Contract, Step 1.5, Output Schema |
| go-authz-tracer.md | `code_ref` | ✅ | Input Contract, Output Schema |
| go-oauth-auditor.md | `code_ref` | ✅ | Input Contract, Output Schema |
| invariant-checker.md | `code_ref` | ✅ | Input Contract, Output Schema |
| go-taint-tracer.md | `code_ref` | ✅ | Input Contract, Output Schema |
| synthesis.md | `code_ref` | ✅ | Step 1, Step 5 |
| security-review.md | `make -C claude-security-hooks install` | ✅ | Step 1.5 |
| security-review.md | `CODE_REF` | ✅ | Steps 2, 4, 6 (13 occurrences) |

## Deviations from Plan

None — plan executed exactly as written. All requirements met, all configuration changes applied, all agent definitions and command specs updated correctly.

## Files Modified

| File | Changes | Purpose |
|------|---------|---------|
| go-taint-tracer.md | +10 lines | T9-PATH rule (B3) + code_ref echo |
| go-cartographer.md | +9 lines | code_ref input/output wiring |
| go-authz-tracer.md | +10 lines | code_ref echo instructions |
| go-oauth-auditor.md | +10 lines | code_ref echo instructions |
| invariant-checker.md | +10 lines | code_ref echo instructions |
| synthesis.md | +10 lines | code_ref propagation |
| security-review.md | +52 lines | Step 1.5, Step 2 CODE_REF, Steps 4+6 injection |

## Key Decisions Documented in Plan

- **B4: Unconditional rebuild:** Step 1.5 makes ~2s cost acceptable vs. risk of stale binary silently corrupting the pipeline
- **E1: CODE_REF computation:** git rev-parse HEAD:RELATIVE_PATH for tree hash; changes only when reviewed directory changes
- **E1: CODE_REF_DIRTY:** git status --porcelain detects uncommitted changes; no mtime/hash comparison overhead
- **E1: Cartographer as single source:** Cartographer receives code_ref from orchestration, writes to go-index.json; does not recompute to avoid duplicating git logic and ensure consistency
- **E1: Tracer echo pattern:** Each tracer receives code_ref in input, echoes verbatim in output; T-CodeRef/AZ-CodeRef/OA-CodeRef/IC-CodeRef joint invariants (Wave 1) enforce correctness
- **B3: T9-PATH enforcement:** Workspace-relative paths only; T9 FileExists check after os.Chdir(workspace_root) rejects absolute paths; rule numbered in §6 for same weight as T1–T11

## Commits

1. **7f4a7d4** docs(10-03): add T9-PATH hard rule to go-taint-tracer.md
2. **5756313** docs(10-03): add code_ref computation step to go-cartographer.md
3. **3bd10eb** docs(10-03): add code_ref echo instructions to go-authz-tracer.md
4. **4cc9417** docs(10-03): add code_ref echo instructions to go-oauth-auditor.md
5. **62af270** docs(10-03): add code_ref echo instructions to invariant-checker.md
6. **5796725** docs(10-03): add code_ref echo instructions to go-taint-tracer.md
7. **ec31e87** docs(10-03): add code_ref propagation to synthesis.md
8. **c35bde8** docs(10-03): update security-review.md with Step 1.5 and code_ref wiring

## Self-Check Results

✅ PASSED

- All task files modified as specified (8 files, 7 commits)
- All commits created with proper messaging
- T9-PATH rule present in go-taint-tracer.md §6
- code_ref fields present in all agent definitions (input + output contracts)
- Step 1.5 binary freshness check added to security-review.md
- Step 2 CODE_REF computation with git commands added
- Steps 4 and 6 tracer inputs include code_ref and code_ref_dirty fields
- Verification checklist: all required strings found in expected locations
- No pre-existing files broken
- Plan objective complete: configuration and documentation wiring finished for Wave 1
