---
phase: 11-codegraph-migration
fixed_at: 2026-05-27T15:00:00Z
review_path: .planning/phases/11-codegraph-migration/11-REVIEW.md
iteration: 1
findings_in_scope: 6
fixed: 6
skipped: 0
status: all_fixed
---

# Phase 11: Code Review Fix Report

**Fixed at:** 2026-05-27T15:00:00Z
**Source review:** .planning/phases/11-codegraph-migration/11-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 6 (CR-01, CR-02, CR-03, WR-01, WR-02, WR-03)
- Fixed: 6
- Skipped: 0

## Fixed Issues

### CR-01: Literal string "TARGET_DIR" written to `.current-review` instead of the resolved path

**Files modified:** `.claude/commands/security-review.md`, `.claude/skills/security-review/SKILL.md`
**Commit:** 1ac0dcd (commands), 3fe90a0 (SKILL.md)
**Applied fix:** Changed `echo "TARGET_DIR"` to `echo "$TARGET_DIR"` in Step 2 of both files so the shell variable expands to the resolved absolute path rather than writing the literal four characters.

---

### CR-02: `payment_surface.cluster_id` schema contradiction — prose says omit, example says include

**Files modified:** `.claude/agents/go-cartographer.md`
**Commit:** 02e9381
**Applied fix:** Removed `"cluster_id": "42"` from the `payment_surface` schema example at line 176, leaving `{"files": ["billing/charge.go"], "confidence": "extracted"}`. Added a field note under Section 5 Field Notes: "`payment_surface.cluster_id` is omitted — codegraph does not surface community IDs."

---

### CR-03: Unquoted `$TARGET_DIR` / trailing-newline risk in shell wrapper

**Files modified:** `scripts/codegraph-mcp.sh`
**Commit:** 31a1717
**Applied fix:** Replaced `TARGET_DIR="$(cat "$TARGET_FILE")"` with `TARGET_DIR="$(tr -d '\n' < "$TARGET_FILE")"` to strip trailing newlines. Added a two-condition guard before `exec`: checks that TARGET_DIR is non-empty and that the path is an existing directory, printing a diagnostic message to stderr and exiting 1 on failure.

---

### WR-01: SKILL.md not updated — still references `graphify` in three places

**Files modified:** `.claude/skills/security-review/SKILL.md`
**Commit:** 3fe90a0
**Applied fix:**
- Objective flow summary: `Pre-pass (graphify + govulncheck)` → `Pre-pass (codegraph + govulncheck)`
- Step 2 comment: `the graphify MCP server wrapper` → `the codegraph MCP server wrapper`
- Step 2 echo literal: `echo "TARGET_DIR"` → `echo "$TARGET_DIR"` (also fixes CR-01 in SKILL.md)
- Step 3: Replaced `graphify update TARGET_DIR` (producing `graphify-out/graph.json`) with two-step `codegraph init TARGET_DIR` + `codegraph index TARGET_DIR` with recovery instruction, matching the migrated commands file. Old step numbered as "1"; new steps renumbered to match the security-review.md structure (govulncheck became step 3).

---

### WR-02: `mcp__codegraph__codegraph_callees` used implicitly but not declared in agent tools list

**Files modified:** `.claude/agents/go-cartographer.md`
**Commit:** 02e9381
**Applied fix:** Added `mcp__codegraph__codegraph_callees` to the tools declaration on line 5, inserted after `mcp__codegraph__codegraph_callers`. The tool is needed for Steps 5 and 9 to traverse what callers call (middleware chain exploration).

---

### WR-03: `codegraph index` failure leaves no recovery path

**Files modified:** `.claude/commands/security-review.md`
**Commit:** 1ac0dcd
**Applied fix:** Extended the Step 3 failure instruction for `codegraph index`: instead of "stop with error", the instruction now reads "delete `TARGET_DIR/.codegraph/` and re-run both `codegraph init TARGET_DIR` and `codegraph index TARGET_DIR` once before stopping with error. This handles corrupt or version-mismatched databases from prior sessions." The same recovery instruction was mirrored in SKILL.md (commit 3fe90a0).

---

## Skipped Issues

None — all findings were fixed.

---

_Fixed: 2026-05-27T15:00:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
