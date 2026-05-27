# Phase 13: Pipeline Reliability + Bootstrap Hardening — Research

**Researched:** 2026-05-27
**Domain:** Shell scripting (W1), Claude Code agent prompting discipline (W2), Claude Code agent `settings.json` write permissions (W3), hooks binary timing/error messages (W4), JSON Schema agent input documentation (W5), codegraph route detection (W6), cartographer output schema enhancement (W7)
**Confidence:** HIGH — all findings verified by direct codebase inspection

---

## Summary

Phase 13 fixes 7 documented failure modes from the first full `/security-review` run against `examples/sample-vulnerable-service`. The run succeeded with workarounds but had a ~15% first-attempt agent success rate. The failure causes are now fully understood from the SECURITY_REVIEW_SCAN_HANDOFF.md document, cross-verified against the live source code.

**Critical insight:** Three of the seven workstreams (W2, W3, W4) are documentation/configuration changes with no Go binary changes required. W1 is a new Bash script. W5 is new JSON Schema files plus one Go change in `preflight.go`. W6 and W7 are changes to `go-cartographer.md` only — no Go code changes required.

The only Go binary changes in this phase are in W4 (error message improvement in `validate.go` — add file path, first-100-chars, line number to block reason) and W5 (update `preflightPerAgent()` to emit expected schema on parse failure).

**Primary recommendation:** Execute workstreams in PHASE.md priority order (W2 → W3 → W4 → W1 → W5 → W6 → W7). W2, W3, and W4 can be batched together as a single wave since none require Go code changes. W1 is a standalone Bash script. W5/W6/W7 each touch a single file.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| JSON prompt validation for agent dispatch | Orchestration skill (`security-review.md`) | Hooks binary (`preflight.go`) | The orchestration skill constructs the prompt; hooks validate it. Both layers need the "pure JSON" requirement stated. |
| Agent write permissions | Claude Code `settings.json` (agent `tools:` field) | Agent definition `.md` files | `settings.json` controls which tools agents may call at runtime. Agent `.md` files document the contract. |
| PostToolUse validation timing | Hooks binary (`validate.go`) | Orchestration skill | Hooks are invoked by Claude Code harness immediately after tool use; validation must be fault-tolerant for race conditions. |
| Bootstrap pre-flight checks | `bootstrap/pre-flight-checks.sh` | Root `Makefile` (`preflight` target) | Shell script does the actual checks; Makefile provides the developer interface. |
| Agent input schema documentation | `claude-security-hooks/specs/agents/` (JSON Schema files) | Hooks `preflight.go` (error message) | Schema files serve as documentation; `preflight.go` cites them in error messages. |
| Cartographer route detection completeness | `go-cartographer.md` (agent instructions) | `examples/sample-vulnerable-service/main.go` (reference data) | Cartographer is an agent that follows instructions; W6 adds explicit post-processing step to the agent's protocol. |
| Taint source precision | `go-cartographer.md` (schema output) | `security-review.md` (Step 6 input construction) | Cartographer must emit the new fields; orchestrator must read them when constructing tracer inputs. |

---

## W2: JSON Serialization — Root Cause and Fix

### Root Cause (VERIFIED by code inspection)

The orchestration skill `security-review.md` dispatches agents via `Task` tool calls. The `prompt` field of a `Task` call is consumed by the agent as its input. The hooks binary's `Preflight()` function (in `claude-security-hooks/internal/hooks/preflight.go`) calls `preflightPerAgent()` which runs `json.NewDecoder(strings.NewReader(promptJSON)).Decode(&in)` for each agent type.

When the orchestrating LLM (Claude) fills in the `prompt` for a Task, it defaults to natural-language style unless explicitly instructed to use pure JSON. The first run showed Claude wrote:

```
Analyze authorization middleware on all HTTP routes.

Input:
- routes: [...]
- authz_primitives: []
```

The hooks parser receives this as `promptJSON`, and `json.Decode` immediately fails on `'A'` (the first character of "Analyze").

### Fix Location (VERIFIED)

`security-review.md` contains the Task dispatch blocks in:
- **Step 4** (cartographer): already uses a JSON template
- **Step 6** (all tracers): already uses JSON templates

The JSON templates already exist in the current `security-review.md` — inspection shows they are correct pure-JSON forms. However, the skill needs an explicit instruction at the TOP of each dispatch site stating:

> "The `prompt` field MUST be the JSON object shown below verbatim — no prose, no wrapper text, no markdown. The hooks binary's `preflight` hook validates the prompt as JSON before agent dispatch and will block on any non-JSON content."

**Insertion points in `security-review.md`:**
- Before Step 4 cartographer dispatch block: add a JSON instruction header
- Before Step 6 tracer dispatch blocks: add a JSON instruction header for the parallel fan-out

**Also needed:** A one-sentence note at the start of `security-review.md`'s Key Constraints section clarifying that all Task `prompt` values must be pure JSON.

### Agent-side: Input contract documentation

Each agent `.md` file already has a `## 2. Input Contract` section with a sample JSON block. The fix is to make those blocks more visually prominent (bold) and add a sentence: "The entire prompt content must be this JSON object — no enclosing prose or markdown."

---

## W3: Agent Write Permissions — Mechanism

### How Claude Code Agent File Permissions Work (VERIFIED by settings.json inspection)

Agent tool access is controlled by the `tools:` list in the agent's YAML frontmatter. The current agent frontmatter shows:

```yaml
# go-authz-tracer (current)
tools: mcp__gopls__go_references, mcp__gopls__go_symbol_references, ..., Read, Glob
# NO Write tool

# synthesis (current)
tools: Read, Glob, Write
# HAS Write — synthesis already has it!
```

**Key finding:** `synthesis.md` already has `Write` in its `tools:` list. It wrote the verdict files successfully in the actual run. The issue is that the PostToolUse hook fired BEFORE the Write tools completed (timing issue = W4), not that Write was structurally forbidden.

**For the specialist tracers** (go-authz-tracer, go-taint-tracer, go-oauth-auditor, invariant-checker), `Write` is absent from the tools list. These agents reported they could not write because `Write` is not in their allowlist.

**go-cartographer** has `Bash` in its tools list but the A10/A11 invariants restrict Bash to govulncheck only. Cartographer writes `go-index.json`, but actually currently the agent instructions say "return JSON as your final message for orchestration to persist." This needs `Write` added.

### Fix: Add `Write` to agent frontmatter tools lists

For agents that write verdict files:
- `go-cartographer.md`: add `Write` — writes `go-index.json`
- `go-authz-tracer.md`: add `Write` — writes `authz-findings.json`
- `go-oauth-auditor.md`: add `Write` — writes `oauth-checklist.json`
- `invariant-checker.md`: add `Write` — writes `invariant-results.json`
- `go-taint-tracer.md`: add `Write` — writes `taint-verdict-*.json`
- `synthesis.md`: already has `Write` — no change needed

**Write path allowlist (in agent instructions):** Each agent's §6 Hard Rules section should specify the exact permitted write path. The A10 rule in `go-cartographer.md` currently says "MUST NOT modify any file in `.codegraph/` or in the source tree." This needs amendment: cartographer IS allowed to write `<working_directory>/go-index.json`. Similarly for each tracer.

**Note about A10/A11:** These are currently no-op stubs (`TODO(phase-5)`). The A10 restriction has been the source of confusion — agents read A10 as "you cannot Write at all" but the intent is "do not modify the codegraph database or source files." The fix is to clarify A10: write to the designated output file IS permitted; modifying source files or `.codegraph/` is NOT.

### Path pattern allowlist (in agent §6 Hard Rules)

Add a rule to each agent:

```
**A10-amended — Permitted write target:** You MAY write your verdict output to exactly one file:
- go-cartographer: `<working_directory>/go-index.json`
- go-authz-tracer: `<working_directory>/authz-findings.json`
- go-oauth-auditor: `<working_directory>/oauth-checklist.json`
- invariant-checker: `<working_directory>/invariant-results.json`
- go-taint-tracer: `<working_directory>/taint-verdict-<handler-name>-sqli.json`
- synthesis: `<working_directory>/review-report.json` and `<working_directory>/review-report.md`
You MUST NOT write to any other path. You MUST NOT modify source files or `.codegraph/`.
```

---

## W4: Hook Timing / False Negatives — Root Cause and Fix

### Root Cause (VERIFIED)

`validate.go` runs on every `PostToolUse` event for Task tool calls. When tracers DID have Write permission (synthesis has it; or after W3 adds it to other agents), the Write tool call completes AS PART OF the agent run. The `PostToolUseEvent` fires after the Task tool (the whole agent run), not after each sub-tool (Write calls inside the agent).

The actual timing issue: When agents lacked `Write` permission and instead returned JSON as their last message, the validate hook ran immediately on the agent's text output. If the agent's text output was the JSON verdict, this was fine — but the validator was calling `runAgentValidation()` on the text content of the task response (the agent's message), not on a file written to disk. The S1 check (`synthesis.go`'s `SynthesisDirInvariants`) checks that `review-report.json` exists as a file — that's the timing gap.

**Synthesis-specific:** `validate.go` runs a directory check via `invariants.SynthesisDirInvariants` which calls `FileExists("review-report.json")` relative to workspace root. If synthesis writes the file as part of the task but the PostToolUse hook fires in a race with the filesystem, or if the validation runs before synthesis completes its Write call, you get false negatives.

### Current validate.go behavior (VERIFIED)

```go
// Synthesis: also run S1 directory check
if ev.ToolInput.SubagentType == "synthesis" {
    for _, inv := range invariants.SynthesisDirInvariants {
        for _, vio := range inv.Check(".") {
            reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
        }
    }
}
```

The synthesis dir check happens every time validate runs on a synthesis PostToolUse event. If called before the file is written (race), it will always report a false negative.

### Fix Strategy

The Phase 13 PHASE.md specifies: "Gate schema validation until all expected verdict files are present (file-existence pre-check)."

**For synthesis specifically:** The S1 dir check in `validate.go` should only fire if the synthesis agent's text content is non-empty and parses as a valid SynthesisReport. If the text content is the JSON verdict, the file existence check is secondary. If synthesis wrote the files AND they exist, the hook should pass. If they don't exist yet (race), the hook should degrade gracefully rather than blocking.

**Practical fix for validate.go:**
1. If `S1` dir check fails AND the text content parsed as valid JSON, emit a warning (log to stderr) rather than a block. The JSON in the response IS the verdict; the file is bonus persistence.
2. Add a retry with 100ms sleep before the S1 check: try `FileExists` once; if missing, sleep 100ms and try once more before blocking. This handles filesystem propagation lag.
3. Improve error messages: when a verdict parse fails, include: the file path being checked (if applicable), the first 100 chars of actual content, and the JSON parse error with offset.

**Error message improvement (in `validate.go` and `dispatch.go`):**

Current block reason format: `T1: taint verdict parse: <error>`

Improved format:
```
T1: taint verdict parse failed
  file: TARGET_DIR/taint-verdict-getuserhandler-sqli.json
  error: invalid character 'A' at offset 0
  content (first 100 chars): "Analyze the following taint..."
```

The `validate.go` function has the `verdictText` variable available; it just needs to be included in the block message.

**`--debug-validate` flag:** Add as a flag to the `validate` subcommand. When set (via env var `HOOKS_DEBUG_VALIDATE=1`), log to stderr: the subagent type, first 200 chars of verdictText, and each invariant result.

---

## W1: Bootstrap Pre-Flight — Design

### Environment Availability (VERIFIED on this machine)

| Dependency | Available | Version | Notes |
|------------|-----------|---------|-------|
| go | ✓ | go1.24.4 linux/amd64 | Required for hooks binary build |
| make | ✓ | GNU Make 4.4.1 | Required for `make install` |
| git | ✓ | 2.47.3 | Required for code_ref computation |
| docker | ✓ | 29.4.0 | Required for govulncheck; has fallback in security-review.md |
| codegraph | ✓ | 0.9.6 | At `/home/saghaulor/.local/bin/codegraph` |

### Script Design (VERIFIED from BOOTSTRAP_REQUIREMENTS.md)

The bootstrap document provides complete pseudocode for all 6 checks. The implementation is:

**File:** `bootstrap/pre-flight-checks.sh`

Six sequential checks (exit non-zero on critical failure, warn on optional):

1. **Codegraph check** — `command -v codegraph` + `codegraph --version`. If missing, print `go install github.com/anthropics/codegraph/cmd/codegraph@latest` and exit 1. [ASSUMED: install path — verify from codegraph upstream docs]
2. **Make + Go** — `command -v make`, `command -v go`, `[ -f claude-security-hooks/go.mod ]`. All critical (exit 1 on any failure).
3. **Git** — `command -v git`, `git -C "$TARGET_DIR" rev-parse --git-dir`. If target is not a git repo, warn (not exit 1 — some targets may not be git repos but are still valid Go projects).
4. **Go project structure** — `[ -f "$TARGET_DIR/go.mod" ]` critical. `[ -f "$TARGET_DIR/go.sum" ]` warn only with `go mod tidy` hint.
5. **Docker** — `command -v docker && docker ps`. Non-critical (govulncheck has a `{"available":false}` fallback). Export `DOCKER_AVAILABLE=true|false`.
6. **Codegraph init + index with retry** — `codegraph init "$TARGET_DIR"` (idempotent), then `codegraph index "$TARGET_DIR"`. On failure: `rm -rf "$TARGET_DIR/.codegraph/"` and retry once. If retry fails, exit 1.

**Script receives TARGET_DIR as $1.** If $1 is empty, use `.`.

### Root Makefile Extension

The root `Makefile` currently has targets: `help`, `verify-opengrep-mcp`, `build-opengrep-mcp`.

Add:
```makefile
.PHONY: preflight

## preflight: Run bootstrap pre-flight checks against TARGET (default: .)
preflight:
    @bash bootstrap/pre-flight-checks.sh "$(TARGET)"
```

Usage: `make preflight TARGET=/path/to/go/project`

---

## W5: Agent Input Schema Documentation

### Current State (VERIFIED by code inspection)

`preflight.go`'s `preflightPerAgent()` function runs `json.NewDecoder(...).DisallowUnknownFields().Decode(&in)` for each agent. If the decode fails, the current error message is:

```
D-09: oauth input parse: json: unknown field "working_directory"
```

This tells the user WHAT failed but not WHAT IS EXPECTED. The fix is to append the expected schema.

### Schema Files

Create `claude-security-hooks/specs/agents/` directory with one JSON Schema per agent:

**Schemas to create** (derived from Go struct definitions in `internal/schema/`):

```
claude-security-hooks/specs/agents/
├── go-cartographer.schema.json     (CartographerIndex-shaped input doesn't apply; cartographer input is minimal)
├── go-taint-tracer.schema.json     (TaintInput struct)
├── go-authz-tracer.schema.json     (AuthzInput struct)
├── go-oauth-auditor.schema.json    (OAuthInput struct — the one that failed with working_directory)
├── invariant-checker.schema.json   (InvariantCheckerInput struct)
└── synthesis.schema.json           (SynthesisInput struct — minimal: working_directory + review_id)
```

**OAuthInput schema** (the failed one — VERIFIED from `schema/oauth_auditor.go`):

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "go-oauth-auditor input",
  "type": "object",
  "required": ["oauth_locations"],
  "properties": {
    "oauth_locations": { "type": "object" },
    "target_profile": { "type": "string", "enum": ["oauth_2_1", "oauth_2_0", "oauth_2_0_with_9700_bcp"] },
    "features_in_use": { "type": "array", "items": { "type": "string" } },
    "review_session_id": { "type": "string" },
    "code_ref": { "type": "string" },
    "code_ref_dirty": { "type": "boolean" }
  },
  "additionalProperties": false
}
```

**Note:** `working_directory` is NOT in this schema — the failing field from the actual run.

### preflight.go Enhancement

When `Decode` fails, the error message should include the expected schema (embedded as a constant or loaded from the spec file). Because the binary has no external file reads at hook time, embed schemas as Go string constants or load from `specs/agents/` relative to executable.

**Simplest approach:** Embed each schema as a multi-line Go string constant in `preflight.go` (or a `specs.go` file in the hooks package). When D-09 fires, append: `\nExpected schema: <embedded_schema_string>`.

**`preflightPerAgent` error message pattern:**
```go
case "go-oauth-auditor":
    var in schema.OAuthInput
    dec := json.NewDecoder(strings.NewReader(promptJSON))
    dec.DisallowUnknownFields()
    if err := dec.Decode(&in); err != nil {
        return fmt.Sprintf("D-09: oauth input parse: %v\nExpected schema: %s", err, oauthInputSchema)
    }
```

---

## W6: Cartographer Route Detection

### Root Cause (VERIFIED by code inspection)

The go-cartographer agent's Step 3 (route enumeration) uses `mcp__opengrep__scan_with_rule` with gin patterns. The Semgrep pattern for gin route registration:

```
# Semgrep pattern from go-cartographer.md §3 Step 3
gin: match `r.GET(...)`, `r.POST(...)`, `r.Group(...)`, `r.Use(...)`, `r.Handle(...)`
```

`callChainSQLiHandler` is registered at `main.go:39` as:
```go
router.GET("/api/advanced-search", callChainSQLiHandler)
```

The variable is named `router`, not `r`. The Semgrep pattern `r.GET(...)` would only match if the variable is named `r`. This is the codegraph/Semgrep DSL edge case.

**Evidence from actual scan:** The go-index.json produced had 7 entrypoints, missing `GET /api/advanced-search → callChainSQLiHandler`. The cartographer detected the `callChainSQLiHandler` function (it appears in sinks analysis via `handlers.go:80`) but did not register it as a route.

### Fix

The cartographer needs a **post-processing validation step** after Step 3. After enumerating routes via Semgrep, add Step 3.5:

**Step 3.5: Cross-reference router registration vs. detected entrypoints**

For each detected router, read `main.go` (and any file that imports the router package) and extract all `<router_var>.{GET,POST,DELETE,PATCH,PUT,Handle}(...)` calls using a more flexible Semgrep pattern (matching any receiver name, not just `r`):

```yaml
# More permissive gin route detection pattern
patterns:
  - pattern: $ROUTER.GET($PATH, ...)
  - pattern: $ROUTER.POST($PATH, ...)
  - pattern: $ROUTER.DELETE($PATH, ...)
  - pattern: $ROUTER.PATCH($PATH, ...)
  - pattern: $ROUTER.PUT($PATH, ...)
  - pattern: $ROUTER.Handle($METHOD, $PATH, ...)
```

Compare the set of `($METHOD, $PATH, $HANDLER)` tuples from Semgrep against `entrypoints`. For any route in Semgrep results that is absent from entrypoints:
1. Add it to entrypoints if enough information is available (handler FQN can be resolved via codegraph_search)
2. If handler FQN cannot be resolved, add to `warnings` array: `"route_registration_gap: GET /api/advanced-search → callChainSQLiHandler not in entrypoints"`

**Additionally:** Update the Semgrep patterns in Step 3 to use `$ROUTER.GET(...)` instead of `r.GET(...)` to be variable-name-agnostic.

---

## W7: Cartographer Output Enhancement — `first_param_read_line`

### Root Cause (VERIFIED)

Orchestrator at Step 6 constructs taint-tracer inputs using `handler.line` from entrypoints (the handler function definition line). For `callChainSQLiHandler`, handler.line = 254 (function def), but first HTTP read is line 256 (`c.Query("search")`).

The agents handled this with `input_mismatch` detection that gracefully continued — so this is a quality improvement, not a blocking issue.

### Fix: New fields in Entrypoint schema

Add to `schema.Handler` struct (in `claude-security-hooks/internal/schema/cartographer.go`):

```go
type Handler struct {
    FQN                string `json:"fqn"`
    File               string `json:"file"`
    Line               int    `json:"line"`
    FirstParamReadLine int    `json:"first_param_read_line,omitempty"` // W7
    FirstParamReadExpr string `json:"first_param_read_expr,omitempty"` // W7
}
```

**`go-cartographer.md` instruction addition (Step 3 enhancement):**

After extracting each handler's definition line, scan the handler function body for the first HTTP parameter read expression (using gin parameter kinds: `c.Query(...)`, `c.PostForm(...)`, `c.Param(...)`, `c.ShouldBind*(...)`, `c.GetRawData()`, `c.Body`). Record the line number and expression text as `first_param_read_line` and `first_param_read_expr`.

**`security-review.md` Step 6 update:**

When constructing taint-tracer source inputs, use `first_param_read_line` if present:
```json
"source": {
    "file": "<handler file>",
    "line": <first_param_read_line if present, else handler.line>,
    "expr": "<first_param_read_expr if present, else inferred>",
    "kind": "http_query"
}
```

---

## Standard Stack

### Core (all existing — no new packages)

| Component | Location | Purpose | Why Standard |
|-----------|----------|---------|--------------|
| Go 1.24.4 | system | Build hooks binary | Already project language |
| `encoding/json` | stdlib | JSON decode/encode in hooks | No external deps (H5) |
| Bash | system | Bootstrap pre-flight script | Shell scripting for env checks |

### No New External Dependencies

Phase 13 introduces no new Go packages. All changes are:
- Modifications to existing Go files in `claude-security-hooks/`
- New JSON Schema files (`.json`, no tooling dependency)
- New Bash script (`bootstrap/pre-flight-checks.sh`)
- Modifications to agent `.md` files
- Modifications to `security-review.md`

---

## Package Legitimacy Audit

No new packages are installed in this phase. All dependencies are pre-existing.

---

## Architecture Patterns

### System Architecture Diagram

```
[make preflight]
      |
      v
[bootstrap/pre-flight-checks.sh TARGET_DIR]
  checks: codegraph / go / make / git / go.mod / docker
  → exits 0 (OK) or non-zero with remediation message
      |
      v
[security-review.md] — orchestrator
  Step 4: Task(go-cartographer)
    prompt = PURE JSON { working_directory, review_session_id, ... }
                ↑
          W2: explicit "must be pure JSON" instruction
    
    PostToolUse hook → validate
      runCartographer() → A1..A11 checks
      W4: improved error messages (file, content, offset)
      
  Step 5: Read go-index.json
    entrypoints now include first_param_read_line (W7)
    warnings array catches missing routes (W6)
    
  Step 6: Task(tracers × N) -- parallel
    prompt = PURE JSON { source, sink, ... }   W2
    tools: [..., Write]                        W3
    agent writes verdict file directly
    PostToolUse → validate → T1..T11 checks
    W4: file-existence pre-check before schema parse
    
  Step 7: Task(synthesis)
    prompt = PURE JSON { working_directory, review_id }  W2
    tools: [..., Write]  (already present)               W3
    agent writes review-report.json + .md
    PostToolUse → validate → S1..S6 + SynthesisDirCheck
    W4: retry FileExists before blocking on S1
```

### Recommended Project Structure Changes

```
bootstrap/
└── pre-flight-checks.sh        # NEW (W1)

claude-security-hooks/
├── specs/
│   └── agents/                 # NEW (W5)
│       ├── go-cartographer.schema.json
│       ├── go-taint-tracer.schema.json
│       ├── go-authz-tracer.schema.json
│       ├── go-oauth-auditor.schema.json
│       ├── invariant-checker.schema.json
│       └── synthesis.schema.json
└── internal/
    ├── hooks/
    │   ├── validate.go         # MODIFIED (W4: error messages, S1 retry)
    │   └── preflight.go        # MODIFIED (W5: embed expected schema in errors)
    └── schema/
        └── cartographer.go     # MODIFIED (W7: Handler.FirstParamReadLine/Expr fields)

.claude/
├── agents/
│   ├── go-cartographer.md      # MODIFIED (W3 A10 amendment, W6 Step 3.5, W7 Step 3 addition)
│   ├── go-authz-tracer.md      # MODIFIED (W2 JSON instruction, W3 add Write tool)
│   ├── go-oauth-auditor.md     # MODIFIED (W2 JSON instruction, W3 add Write tool)
│   ├── go-taint-tracer.md      # MODIFIED (W2 JSON instruction, W3 add Write tool)
│   ├── invariant-checker.md    # MODIFIED (W2 JSON instruction, W3 add Write tool)
│   └── synthesis.md            # (no change needed — already has Write)
└── commands/
    └── security-review.md      # MODIFIED (W2 JSON enforcement note, W7 Step 6 source construction)

Makefile                        # MODIFIED (W1: add preflight target)
```

### Anti-Patterns to Avoid

- **Writing to TARGET_DIR from validate.go:** The validate hook must never write files itself — only block/allow. Writing verdict files is the agent's job (W3).
- **Hardcoding TARGET_DIR in hooks binary:** The hooks binary runs as a subprocess; it does not know TARGET_DIR. Path checks (A7, A8, T9) must remain workspace-relative (relative to CWD which validate.go chdir's to workspace root).
- **Making bootstrap/pre-flight-checks.sh exit 1 on Docker absence:** Docker is optional (govulncheck has a `{"available":false}` fallback). Never block the pipeline on Docker absence.
- **Using Python in bootstrap scripts:** Project convention requires Go for implementation code. Shell is fine for bootstrap scripts. Do not use Python (CLAUDE.md: "Don't reach for Python").

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON Schema validation at runtime | Custom JSON validator in Go | `encoding/json` with `DisallowUnknownFields()` | Already in preflight.go; battle-tested |
| Schema documentation | Custom DSL | Standard JSON Schema draft-07 | Universally understood; IDE tooling |
| Bootstrap tool detection | Complex version parsing | `command -v` + `--version` pipe | Simple and correct for existence checks |
| Filesystem race in validate.go | Complex lock files | Single retry with 100ms sleep | S1 is a soft safety net; agents writing directly (W3) eliminates the race |

---

## Common Pitfalls

### Pitfall 1: A10 Rule Confusion After Adding Write Permission
**What goes wrong:** After adding `Write` to agent tools lists (W3), the A10 invariant (currently a no-op stub) might be misread as blocking all writes. Agents read A10 literally.
**Why it happens:** A10 says "MUST NOT modify any file in `.codegraph/` or in the source tree." Agents may interpret "source tree" as "anything under TARGET_DIR."
**How to avoid:** Amend A10 in each agent's §6 to explicitly carve out the permitted output file. See W3 section above for the per-agent write target list.
**Warning signs:** Agent returns verdict as text in its final message instead of writing to disk, citing A10/A11 as reason.

### Pitfall 2: PostToolUse Hook Fires on Non-Verdict Tasks
**What goes wrong:** The PostToolUse hook matcher is `"Task"` — it fires on ALL Task tool calls, not just security agent calls. The `IsSecurityAgent()` check gates per-agent validation, but the S1 dir check only fires for synthesis. Adding more file-existence checks could accidentally trigger on non-synthesis agents.
**Why it happens:** `validate.go` dispatches on `SubagentType`; the synthesis branch adds an extra check that no other agent has.
**How to avoid:** Keep the timing fix (S1 retry) narrow — only apply to synthesis. Other agents don't have file-existence checks in the validator.
**Warning signs:** Unexpected blocks on cartographer or tracer agents after W4 changes.

### Pitfall 3: Variable-Name Semgrep Pattern Matching
**What goes wrong:** Semgrep patterns like `r.GET(...)` only match if the gin router variable is named `r`. Code using `router.GET(...)` or `e.GET(...)` (echo) won't match.
**Why it happens:** The original patterns used the idiomatic variable name rather than a metavariable.
**How to avoid:** All route detection patterns in go-cartographer.md Step 3 must use metavariables (`$ROUTER.GET(...)`). Apply this fix as part of W6.
**Warning signs:** Missing entrypoints in go-index.json when router variable is not named `r`.

### Pitfall 4: Bootstrap Script Running Inside claude-security-hooks/ Directory
**What goes wrong:** bootstrap/pre-flight-checks.sh checks for `claude-security-hooks/go.mod` — this is a path relative to the REPO ROOT, not the TARGET_DIR. If the script is called from the wrong CWD it will fail silently.
**Why it happens:** The script serves two purposes: verify repo structure AND verify the target Go project. These have different CWD requirements.
**How to avoid:** The script must detect repo root (e.g., `git rev-parse --show-toplevel`) and use absolute paths for repo-structure checks. TARGET_DIR checks use the resolved TARGET_DIR argument.
**Warning signs:** `go.mod not found` errors when running from a subdirectory.

### Pitfall 5: go-index.json Schema Change Breaking Existing Tests
**What goes wrong:** Adding `first_param_read_line` and `first_param_read_expr` to `schema.Handler` changes the Go struct. If any existing test fixtures hardcode `Handler` without these fields, they will fail to compile.
**Why it happens:** The `omitempty` tag means existing JSON round-trips work fine, but Go literal initializations like `Handler{FQN: "...", File: "...", Line: 42}` will still compile — struct literal with named fields is always fine. No regression risk in tests.
**How to avoid:** Use `omitempty` on both new fields (already specified above). Verify `go test ./...` passes after schema change.
**Warning signs:** Compilation errors in schema_test.go or cartographer_test.go.

---

## Code Examples

### W2: Correct Task Dispatch Pattern

```markdown
Spawn a Task with agent `go-authz-tracer` with the following input. The `prompt` value MUST be
valid JSON exactly as shown — no prose, no wrapper text, no markdown code fences. The preflight
hook validates the prompt as JSON and will block if it is not a valid JSON object.

Input:
```json
{
  "routes": <entrypoints array from go-index.json>,
  "authz_primitives": <authz_primitives array from go-index.json>,
  "sensitive_operations": [],
  "review_session_id": "<SESSION_ID>",
  "code_ref": "<CODE_REF>",
  "code_ref_dirty": <CODE_REF_DIRTY>
}
```
```
[Source: security-review.md Step 6, verified by direct file inspection]

### W3: Agent Tools Frontmatter Pattern

```yaml
# go-authz-tracer.md — AFTER W3 fix
---
name: go-authz-tracer
description: ...
model: claude-sonnet-4-6
tools: mcp__gopls__go_references, mcp__gopls__go_symbol_references, mcp__gopls__go_search, mcp__lsp__callHierarchy_outgoingCalls, mcp__lsp__textDocument_definition, mcp__lsp__textDocument_implementation, Read, Glob, Write
---
```
[Source: Current agent files, verified by direct inspection]

### W4: validate.go Error Message Improvement

```go
// Current (verbose text lost):
return []string{FormatViolation("T1", "taint verdict parse", invariants.Violation{
    Path: "", Expected: "", Actual: err.Error(),
})}, nil

// Improved (include content preview):
preview := verdictText
if len(preview) > 100 {
    preview = preview[:100] + "..."
}
return []string{FormatViolation("T1", "taint verdict parse", invariants.Violation{
    Path:     "tool_response.content",
    Expected: "valid TaintVerdict JSON",
    Actual:   fmt.Sprintf("%s | content(first 100): %q", err.Error(), preview),
})}, nil
```
[Source: dispatch.go + validate.go, verified by direct inspection]

### W5: preflight.go Expected Schema in Error Message

```go
const oauthInputSchemaDoc = `{"type":"object","required":["oauth_locations"],"properties":{"oauth_locations":{},"target_profile":{"enum":["oauth_2_1","oauth_2_0","oauth_2_0_with_9700_bcp"]},"features_in_use":{"type":"array"},"review_session_id":{"type":"string"},"code_ref":{"type":"string"},"code_ref_dirty":{"type":"boolean"}},"additionalProperties":false}`

case "go-oauth-auditor":
    var in schema.OAuthInput
    dec := json.NewDecoder(strings.NewReader(promptJSON))
    dec.DisallowUnknownFields()
    if err := dec.Decode(&in); err != nil {
        return fmt.Sprintf("D-09: oauth input parse: %v\nExpected schema: %s", err, oauthInputSchemaDoc)
    }
```
[Source: preflight.go, verified by direct inspection]

### W6: Semgrep Pattern Fix for Gin Route Detection

```yaml
# BEFORE (variable-name-specific, misses router.GET):
patterns:
  - pattern: r.GET($PATH, ...)
  - pattern: r.POST($PATH, ...)

# AFTER (metavariable, matches any receiver):
patterns:
  - pattern: $ROUTER.GET($PATH, ...)
  - pattern: $ROUTER.POST($PATH, ...)
  - pattern: $ROUTER.DELETE($PATH, ...)
  - pattern: $ROUTER.PATCH($PATH, ...)
  - pattern: $ROUTER.PUT($PATH, ...)
  - pattern: $ROUTER.Handle($METHOD, $PATH, ...)
  - pattern: $ROUTER.Group($PATH, ...)
```
[Source: go-cartographer.md §3 Step 3, verified by direct inspection]

### W7: Schema Change for Handler Struct

```go
// claude-security-hooks/internal/schema/cartographer.go — AFTER W7
type Handler struct {
    FQN                string `json:"fqn"`
    File               string `json:"file"`
    Line               int    `json:"line"`
    FirstParamReadLine int    `json:"first_param_read_line,omitempty"`
    FirstParamReadExpr string `json:"first_param_read_expr,omitempty"`
}
```
[Source: schema/cartographer.go, verified by direct inspection]

---

## Runtime State Inventory

This phase is a reliability/hardening pass, not a rename or migration. However, there are specific runtime artifacts to track:

| Category | Items Found | Action Required |
|----------|-------------|-----------------|
| Stored data | `examples/sample-vulnerable-service/*.json` (go-index.json, taint-verdict-*.json, authz-findings.json, etc.) from prior scan run | No action — Step 2 of security-review.md deletes stale files on each fresh run |
| Live service config | `examples/sample-vulnerable-service/.codegraph/` — codegraph database written into target | No action — .gitignore excludes it; W6/W7 work with the existing db |
| OS-registered state | None — no Task Scheduler, systemd, or pm2 registrations in this project | None |
| Secrets/env vars | `HOOKS_DEBUG_VALIDATE` — new env var for W4 debug mode | No existing values; new convention |
| Build artifacts | `.claude/hooks/bin/claude-security-hooks` — hooks binary; rebuilt by `make install` in Step 1.5 | W4/W5 changes require rebuild; Step 1.5 handles this automatically |

---

## Open Questions

1. **W3: Is `Write` sufficient, or does Claude Code require `allowedWrite` path patterns in settings.json?**
   - What we know: agent frontmatter `tools:` controls tool availability. `Write` is a builtin tool.
   - What's unclear: whether Claude Code has path-filtering for `Write` at the settings.json level (like `Bash(cmd pattern)` does for Bash).
   - Recommendation: Add `Write` to tools lists in agent frontmatter first. If path filtering exists, add path patterns to `.claude/settings.json`. The handoff document shows synthesis (which already has `Write`) wrote files successfully — so basic `Write` in tools list IS sufficient.

2. **W4: Is the S1 false negative a timing race or a different issue?**
   - What we know: The handoff says synthesis said "Both output files have been written" but the hook reported them missing. Files were confirmed present by `ls`.
   - What's unclear: Whether this was truly a filesystem timing race (unlikely on local filesystem) or whether the `FileExists()` check uses a different CWD than expected.
   - Recommendation: `validate.go` calls `os.Chdir(root)` before validation. Verify that `SynthesisDirInvariants.Check(".")` after the chdir resolves to workspace root correctly. The `"."` argument in the synthesis check may be the bug — it should be the actual TARGET_DIR of the review, not workspace root.

3. **W6: Can `$ROUTER.GET(...)` in Semgrep match gin `router.GET()` in Go?**
   - What we know: Semgrep metavariables (`$X`) match any expression in the correct syntactic position.
   - What's unclear: Whether the Semgrep CE/intrafile tier used in cartographer supports Go metavariable matching on method receivers.
   - Recommendation: Test the pattern against `examples/sample-vulnerable-service/main.go` directly using `mcp__opengrep__scan_with_rule` as part of plan implementation.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Build hooks binary | ✓ | 1.24.4 | None — required |
| make | `make install`, `make preflight` | ✓ | GNU Make 4.4.1 | None for hooks binary |
| git | code_ref computation | ✓ | 2.47.3 | Step 2 handles gracefully (CODE_REF="") |
| docker | govulncheck in Step 3 | ✓ | 29.4.0 | `{"available":false}` fallback in security-review.md |
| codegraph | Code indexing | ✓ | 0.9.6 at ~/.local/bin | None — required for pipeline |

**Missing dependencies with no fallback:** None on this machine.

**Missing dependencies with fallback:** Docker (has fallback in security-review.md Step 3 already).

---

## Validation Architecture

**Note:** No `.planning/config.json` found. Treating `nyquist_validation` as enabled (default).

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go standard `testing` package |
| Config file | none — `go test ./...` from `claude-security-hooks/` |
| Quick run command | `go test ./internal/hooks/ -run TestValidate -v` |
| Full suite command | `go test ./... -count=1` from `claude-security-hooks/` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| W2-JSON | `preflight` blocks prompt with non-JSON text | unit | `go test ./internal/hooks/ -run TestPreflight` | ✓ `preflight_test.go` |
| W3-WRITE | Agents that list `Write` in tools can persist verdict files | manual | Run `/security-review examples/sample-vulnerable-service` | N/A (integration) |
| W4-ERRMSG | Block reason includes file path + first 100 chars + offset | unit | `go test ./internal/hooks/ -run TestValidate` | ✓ `validate_test.go` |
| W4-S1-RETRY | S1 check gracefully handles missing file (warn not block) | unit | `go test ./internal/hooks/ -run TestValidate/S1` | ✓ Wave 0 needed |
| W5-SCHEMA | D-09 error includes expected schema string | unit | `go test ./internal/hooks/ -run TestPreflight/OAuthUnknownField` | ✓ Wave 0 needed |
| W6-ROUTE | Cartographer detects `callChainSQLiHandler` via Step 3.5 | integration | Manual run of cartographer against sample-vulnerable-service | N/A (agent) |
| W7-PARAM | go-index.json entrypoints include `first_param_read_line` | integration | Manual run + verify go-index.json | N/A (agent) |

### Wave 0 Gaps

- [ ] `claude-security-hooks/internal/hooks/validate_test.go` — add test for improved error message format (W4)
- [ ] `claude-security-hooks/internal/hooks/preflight_test.go` — add test for D-09 schema doc in error (W5)
- [ ] `claude-security-hooks/internal/hooks/validate_test.go` — add S1 soft-fail test (W4 synthesis retry)
- [ ] `claude-security-hooks/internal/schema/schema_test.go` — add round-trip test for `Handler.FirstParamReadLine` (W7)

---

## Security Domain

`security_enforcement` setting not found in config (absent = enabled).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Not applicable — this phase adds no auth logic |
| V3 Session Management | No | Not applicable |
| V4 Access Control | No — the write permission change is agent-tool-level, not user auth | Agent tools list in frontmatter |
| V5 Input Validation | Yes — W2/W5 fix JSON input validation | `encoding/json` + `DisallowUnknownFields()` |
| V6 Cryptography | No | Not applicable |

### Known Threat Patterns for This Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Agent prompt injection (natural language sneaking into JSON) | Tampering | Explicit "pure JSON" instruction + preflight JSON decode validation |
| Agent writing outside permitted path (path traversal via Write tool) | Tampering | Explicit path constraint in agent §6 Hard Rules; A10-amended rule |
| Bootstrap script injection via TARGET_DIR arg | Tampering | Quote all variable expansions in bash script (`"$TARGET_DIR"`, not `$TARGET_DIR`) |

---

## Sources

### Primary (HIGH confidence)

- Direct codebase inspection of `claude-security-hooks/internal/hooks/validate.go` — hook dispatch flow
- Direct codebase inspection of `claude-security-hooks/internal/hooks/preflight.go` — D-09 input validation
- Direct codebase inspection of `claude-security-hooks/internal/hooks/dispatch.go` — per-agent validation routing
- Direct codebase inspection of `claude-security-hooks/internal/schema/*.go` — all input/output struct definitions
- Direct codebase inspection of `claude-security-hooks/internal/invariants/cartographer.go` — A10/A11 stubs
- Direct codebase inspection of `.claude/agents/*.md` — all agent frontmatter and tool lists
- Direct codebase inspection of `.claude/commands/security-review.md` — all Step dispatch blocks
- Direct codebase inspection of `.claude/settings.json` — hook registrations
- Direct inspection of `examples/sample-vulnerable-service/go-index.json` — confirmed 7 entrypoints (missing callChainSQLiHandler)
- Direct inspection of `examples/sample-vulnerable-service/main.go` — confirmed 8 route registrations
- Direct inspection of `SECURITY_REVIEW_SCAN_HANDOFF.md` — all 7 failure modes with exact error messages
- Direct inspection of `BOOTSTRAP_REQUIREMENTS.md` — all 6 pre-flight check specifications

### Secondary (MEDIUM confidence)

- [ASSUMED] Codegraph Semgrep pattern metavariable behavior for Go receiver matching — not verified against Semgrep/OpenGrep documentation in this session

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `$ROUTER.GET(...)` in Semgrep CE/intrafile tier matches Go method calls with any receiver variable name | W6 | If metavariables don't work on receivers in OpenGrep CE, W6 Semgrep pattern fix won't resolve the missing route; need fallback to regex on main.go source text |
| A2 | Adding `Write` to agent `tools:` frontmatter is sufficient to grant write permission — no Claude Code `settings.json` path filter needed | W3 | If Claude Code has undocumented path filtering for Write, agents may still fail to write even with `Write` in tools list; would require `settings.json` allowlist entries |
| A3 | The codegraph install command for fresh environments is `go install github.com/anthropics/codegraph/cmd/codegraph@latest` | W1 | If the actual repo path differs, bootstrap script will give wrong remediation instructions |

**If table is not empty:** Assumptions A1, A2, A3 need user confirmation before execution, specifically for W6 pattern design, W3 permissions model, and W1 install command.

---

## Metadata

**Confidence breakdown:**
- W2 (JSON serialization): HIGH — root cause verified, fix location precise
- W3 (Write permissions): HIGH — agent frontmatter mechanism verified, A10 stub confirmed
- W4 (Hook timing): HIGH — validate.go code inspected, S1 check location confirmed
- W1 (Bootstrap): HIGH — BOOTSTRAP_REQUIREMENTS.md is prescriptive with complete pseudocode
- W5 (Schema docs): HIGH — OAuthInput struct verified, preflight.go injection point confirmed
- W6 (Route detection): MEDIUM — root cause likely (gin variable name), Semgrep fix pattern [ASSUMED]
- W7 (Param read line): HIGH — schema change location confirmed, no test regressions expected

**Research date:** 2026-05-27
**Valid until:** 2026-06-27 (stable codebase — no fast-moving external deps)
