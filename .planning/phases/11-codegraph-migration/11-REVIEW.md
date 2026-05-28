---
phase: 11-codegraph-migration
reviewed: 2026-05-27T14:30:00Z
depth: standard
files_reviewed: 4
files_reviewed_list:
  - scripts/codegraph-mcp.sh
  - .mcp.json
  - .claude/agents/go-cartographer.md
  - .claude/commands/security-review.md
findings:
  critical: 3
  warning: 3
  info: 1
  total: 7
status: issues_found
---

# Phase 11: Code Review Report

**Reviewed:** 2026-05-27T14:30:00Z
**Depth:** standard
**Files Reviewed:** 4
**Status:** issues_found

## Summary

This review covers the codegraph migration — replacing graphify with codegraph across the MCP wrapper script, MCP server config, cartographer agent, and orchestration command. The migration is largely coherent: tool references in go-cartographer.md are fully updated, the MCP config is correctly updated, and the shell wrapper is well-structured.

Three critical issues found: a literal-string bug that writes the placeholder text "TARGET_DIR" instead of the actual path to `.current-review` (breaking the MCP server's target-directory resolution), a schema contradiction between prose instructions and the output example for `payment_surface.cluster_id`, and an unquoted variable in the shell wrapper that creates a path-injection/word-splitting risk. Three warnings cover the stale SKILL.md (out-of-scope but in the migration's kill-list), a missing tool declaration for `mcp__codegraph__codegraph_callees` used by the agent, and a `codegraph index` failure-handling gap.

---

## Critical Issues

### CR-01: Literal string "TARGET_DIR" written to `.current-review` instead of the resolved path

**File:** `.claude/commands/security-review.md:35`

**Issue:** Step 2 instructs the orchestrator to run:

```
Bash: echo "TARGET_DIR" > /home/saghaulor/code/security_reviewer/.current-review
```

The double-quoted string `"TARGET_DIR"` is a shell literal, not a variable expansion. This writes the four characters `TARGET_DIR` to `.current-review` rather than the resolved absolute path. The codegraph MCP wrapper (`scripts/codegraph-mcp.sh`) reads this file at server startup to determine which directory to index. If the file contains the literal string `TARGET_DIR`, the server will attempt `codegraph serve --mcp --path TARGET_DIR`, which is either a relative path that resolves to the wrong location or a path that does not exist — both result in the codegraph server failing silently or indexing the wrong codebase. This means every `mcp__codegraph__*` call in the cartographer agent operates against stale or wrong data, which can cause missed vulnerabilities to be reported as clean.

This identical bug also exists in `.claude/skills/security-review/SKILL.md:50`, which is the skill copy of the same command and is treated as authoritative by agents that load it.

**Fix:** Use `$TARGET_DIR` (without quotes, or with double-quotes that allow variable expansion) in the Bash invocation. In an LLM-orchestrated context where the variable is held as a string in the session, the command block must make clear that the resolved value must be substituted:

```bash
# security-review.md Step 2 — replace the literal placeholder:
echo "$TARGET_DIR" > /home/saghaulor/code/security_reviewer/.current-review
```

The instruction prose should read: `Bash: echo "$TARGET_DIR" > ...` or use the actual resolved path value (e.g., `/absolute/path/to/service`). The template marker must NOT be a bare word inside double-quotes if the intent is shell-variable expansion.

---

### CR-02: `payment_surface.cluster_id` schema contradiction — prose says omit, example says include

**File:** `.claude/agents/go-cartographer.md:93` and `:176`

**Issue:** Step 7 (line 93) instructs the agent:

> Record matching files as `payment_surface.files`, **omit cluster_id** (codegraph does not surface community IDs), and a confidence level

Yet the output schema example at line 176 includes `cluster_id` in `payment_surface`:

```json
"payment_surface": {"files": ["billing/charge.go"], "cluster_id": "42", "confidence": "extracted"},
```

These two directives contradict each other. An agent following the prose instruction will omit `cluster_id`. An agent pattern-matching on the example will include it with a fabricated value (`"42"` is a magic number with no derivation path described). Because codegraph does not surface community IDs, any agent that follows the example schema and populates `cluster_id` will invent a value, which violates the agent's own no-invent rule (Hard Rules section, last paragraph). Downstream consumers that validate the schema against the example will break on outputs that omit the field; consumers that trust the prose will fail on outputs that include it.

**Fix:** Remove `cluster_id` from the schema example at line 176 to match the prose instruction:

```json
"payment_surface": {"files": ["billing/charge.go"], "confidence": "extracted"},
```

Add a field note under Section 5 Field Notes: "`payment_surface.cluster_id` is omitted — codegraph does not surface community IDs."

---

### CR-03: Unquoted `$TARGET_DIR` variable in shell wrapper exposes word-splitting / path-injection

**File:** `scripts/codegraph-mcp.sh:21`

**Issue:** The final `exec` line is:

```bash
exec codegraph serve --mcp --path "$TARGET_DIR"
```

Wait — the variable IS double-quoted on line 21. However, the content written to `TARGET_FILE` (`.current-review`) is produced by Step 2 of the orchestration command. As documented in CR-01, this file can contain an unintended value. More critically: the file is not validated for content before use. The script reads arbitrary content from `.current-review` via `cat "$TARGET_FILE"` (line 16) and passes it unsanitized to `exec codegraph serve --mcp --path`.

If `.current-review` contains a path with spaces (e.g., `/home/user/my projects/service`), the double-quoting on line 21 handles that correctly. However, if the file contains newlines or embedded shell metacharacters (e.g., injected via a malicious `TARGET_DIR` value such as `/tmp/evil; rm -rf /`), the `exec` form with double-quoting will pass the entire string as a single argument to `codegraph` — which is correct and safe. The real risk is that `cat "$TARGET_FILE"` without stripping trailing newlines can silently embed a `\n` in `TARGET_DIR`, causing `codegraph` to receive a path ending with a newline character, which will fail to match any real directory.

**Fix:** Strip trailing whitespace/newlines when reading the file:

```bash
TARGET_DIR="$(tr -d '[:space:]' < "$TARGET_FILE")"
```

Or more conservatively, strip only trailing newlines:

```bash
TARGET_DIR="$(cat "$TARGET_FILE" | tr -d '\n')"
```

Additionally, add a basic sanity check before exec:

```bash
if [ -z "$TARGET_DIR" ]; then
  echo "codegraph-mcp.sh: TARGET_DIR is empty after reading $TARGET_FILE" >&2
  exit 1
fi
if [ ! -d "$TARGET_DIR" ]; then
  echo "codegraph-mcp.sh: TARGET_DIR '$TARGET_DIR' is not a directory" >&2
  exit 1
fi
```

---

## Warnings

### WR-01: SKILL.md not updated — still references `graphify` in three places

**File:** `.claude/skills/security-review/SKILL.md:19`, `:48`, `:93-94`

**Issue:** The skill file (which is the authoritative copy loaded by agents that invoke the `/security-review` skill) was not updated as part of this migration. It still contains three stale `graphify` references:

- Line 19: Flow summary still says "graphify + govulncheck"
- Line 48: Step 2 comment still says "so the graphify MCP server wrapper knows which graph to serve"
- Lines 93-94: Step 3 still instructs `graphify update TARGET_DIR` and says it produces `TARGET_DIR/graphify-out/graph.json`

The orchestration command (`.claude/commands/security-review.md`) was migrated to codegraph. The SKILL.md copy was not. Any agent that loads SKILL.md (rather than the command file directly) will attempt to run `graphify update TARGET_DIR`, which will fail — because graphify has been removed from the MCP config. This is a silent pipeline break: the agent will encounter an unknown command error at Step 3 and halt the entire review.

This is also where the SKILL.md version of the CR-01 `echo "TARGET_DIR"` bug lives (line 50).

**Fix:** Apply the same Step 3 migration to SKILL.md that was applied to security-review.md: replace `graphify update TARGET_DIR` with `codegraph init TARGET_DIR` + `codegraph index TARGET_DIR`, update the flow summary line, and update the Step 2 comment. Also fix the `echo "TARGET_DIR"` literal on line 50 to `echo "$TARGET_DIR"`.

---

### WR-02: `mcp__codegraph__codegraph_callees` used implicitly but not declared in agent tools list

**File:** `.claude/agents/go-cartographer.md:5`

**Issue:** The agent's declared tool list (line 5) includes:

```
mcp__codegraph__codegraph_search, mcp__codegraph__codegraph_callers, mcp__codegraph__codegraph_trace, mcp__codegraph__codegraph_node, mcp__codegraph__codegraph_status
```

`mcp__codegraph__codegraph_callees` is not declared. The CLAUDE.md system instructions document `codegraph_callees` as a standard codegraph tool ("What does this call?" → `codegraph_callees`). Step 5 (Detect authz primitives) uses `mcp__codegraph__codegraph_callers` to find callers but then needs to explore what those callers call to trace middleware chains. The absence of `codegraph_callees` from the tool declaration means the agent cannot use it even if its reasoning requires it — the tool will be blocked by the runtime tool allowlist enforcement.

The migration replaced `mcp__graphify__*` tools but may not have evaluated whether the replacement codegraph tool set is complete for the agent's task.

**Fix:** Evaluate whether Steps 5 and 9 require callees traversal. If so, add `mcp__codegraph__codegraph_callees` to the tools declaration on line 5:

```yaml
tools: mcp__codegraph__codegraph_search, mcp__codegraph__codegraph_callers, mcp__codegraph__codegraph_callees, mcp__codegraph__codegraph_trace, mcp__codegraph__codegraph_node, mcp__codegraph__codegraph_status, ...
```

---

### WR-03: `codegraph index` failure after a prior partial init leaves `.codegraph/codegraph.db` in an unknown state

**File:** `.claude/commands/security-review.md:80-82`

**Issue:** Step 3 runs `codegraph init TARGET_DIR` (described as idempotent) and then `codegraph index TARGET_DIR`. The instructions state: "If [index] fails, stop with error. (Running index without prior init on a fresh project exits non-zero.)"

The parenthetical only calls out the case where `init` was never run. It does not account for the case where `init` ran successfully in a prior review session but the database is corrupt, locked, or contains stale schema from an older codegraph version. In that case `codegraph init` will exit 0 (already initialized, idempotent), `codegraph index` may fail with an obscure error, and the orchestrator stops — but there is no recovery path instructed. The operator must manually delete `.codegraph/` from TARGET_DIR to reset state.

This is especially likely when reviewing the same service repeatedly (the normal use case) and a prior session was interrupted mid-index.

**Fix:** Add a recovery instruction after the `codegraph index` failure branch:

> If `codegraph index TARGET_DIR` fails, delete `TARGET_DIR/.codegraph/` and re-run both `codegraph init TARGET_DIR` and `codegraph index TARGET_DIR` once before stopping with error. This handles corrupt or version-mismatched databases from prior sessions.

---

## Info

### IN-01: Hardcoded absolute path in `.mcp.json` is non-portable

**File:** `.mcp.json:6`

**Issue:** The codegraph MCP server command is an absolute path:

```json
"command": "/home/saghaulor/code/security_reviewer/scripts/codegraph-mcp.sh"
```

This path works only on the machine where the file was authored. Anyone cloning this repo to a different path or user account will get a "command not found" MCP server failure with no diagnostic hint. The opengrep entry on line 9 has the same pattern.

**Fix:** Use a relative path from the repo root, or document in the README that `.mcp.json` must be regenerated on first clone. For portability, a `scripts/install.sh` post-clone hook that rewrites `.mcp.json` with `$(pwd)` substitution would eliminate the manual step.

---

_Reviewed: 2026-05-27T14:30:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
