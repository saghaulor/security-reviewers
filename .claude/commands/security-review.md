---
name: security-review
description: Run a complete security review of a Go module using specialist agents and the opengrep-mcp scanner
allowed-tools: Task, Read, Glob, Bash
---

You are running a full security review pipeline. Follow every step below in order. **Do not skip any step. Do not reuse existing intermediate files.** This is a fresh scan.

## Step 0: Parse the target directory

The working directory is `$ARGUMENTS`. If `$ARGUMENTS` is empty, use `.` (current directory). Resolve it to an absolute path using `Bash: realpath $ARGUMENTS`.

Store this as TARGET_DIR. All subsequent steps use this path.

## Step 1: Generate a session ID

Run `Bash: .claude/hooks/bin/claude-security-hooks uuid` from the project root and store the output (trimmed) as SESSION_ID. This UUID will be passed to every agent and embedded in all output files so results can be traced to this exact run.

(`uuidgen` is not used because it may not be installed; `claude-security-hooks uuid` uses `crypto/rand` and is always available.)

## Step 1.5: Ensure hooks binary is current

Rebuild and reinstall the hooks binary before any agent dispatch:

```
Bash: make -C /home/saghaulor/code/security_reviewer/claude-security-hooks install
```

This unconditionally runs `go build` and copies the result to `.claude/hooks/bin/claude-security-hooks`. If `make install` fails, stop — the pipeline cannot proceed without a working hooks binary.

## Step 2: Record target, compute code identity, and delete stale files

First, write TARGET_DIR to `.current-review` in the project root so the codegraph MCP server wrapper knows which graph to serve:
```
Bash: echo "$TARGET_DIR" > /home/saghaulor/code/security_reviewer/.current-review
```

Compute the code identity for this review run. First find the repo root and relative path:

```
Bash: git -C TARGET_DIR rev-parse --show-toplevel
```

Store as REPO_ROOT. If this fails (not a git repo), set CODE_REF="" and CODE_REF_DIRTY=false and skip the remaining git steps.

```
Bash: realpath --relative-to=REPO_ROOT TARGET_DIR
```

Store as RELATIVE_PATH.

```
Bash: git -C REPO_ROOT rev-parse HEAD:RELATIVE_PATH
```

Store as CODE_REF. If this fails, set CODE_REF="".

```
Bash: git -C REPO_ROOT status --porcelain RELATIVE_PATH
```

If output is non-empty, set CODE_REF_DIRTY=true. Otherwise CODE_REF_DIRTY=false.

Then remove any existing intermediate files from the previous run. These files from any prior run are INVALID for this run because they were produced under a different SESSION_ID.

Run the following (ignore errors if files don't exist):
```
Bash: rm -f TARGET_DIR/go-index.json TARGET_DIR/authz-findings.json TARGET_DIR/oauth-checklist.json TARGET_DIR/invariant-results.json TARGET_DIR/review-report.json TARGET_DIR/review-report.md
Bash: rm -f TARGET_DIR/taint-verdict-*.json
```

Replace TARGET_DIR with the actual resolved path.

## Step 3: Run pre-pass artifacts

Run these two commands with TARGET_DIR as the working directory:

1. `Bash: codegraph init TARGET_DIR` (idempotent — safe to re-run if already initialized)
   - This creates `.codegraph/codegraph.db` inside TARGET_DIR on first run; warns and exits 0 if already initialized.
2. `Bash: codegraph index TARGET_DIR`
   - This populates or refreshes `.codegraph/codegraph.db` with the current source index.
   - If it fails, delete `TARGET_DIR/.codegraph/` and re-run both `codegraph init TARGET_DIR` and `codegraph index TARGET_DIR` once before stopping with error. This handles corrupt or version-mismatched databases from prior sessions. (Running index without prior init on a fresh project exits non-zero.)

3. Run govulncheck via Docker (always — do not rely on a local govulncheck install):
   ```
   Bash: docker run --rm -v TARGET_DIR:/workspace -w /workspace golang:latest \
     sh -c "go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck -json ./... > govulncheck.json 2>govulncheck.err"
   ```
   - If Docker or the image is unavailable, write `{"available":false}` to `TARGET_DIR/govulncheck.json` and continue.
   - If govulncheck exits non-zero (vulnerabilities found), that is expected — the output is still valid JSON; continue.
   - If it fails for other reasons (network, permission), write `{"available":false}` and continue.

## Step 4: Run cartographer (sequential)

**JSON prompt enforcement:** The `prompt` value for this Task MUST be the JSON object shown below — no prose, no wrapper text, no markdown code fences. The hooks binary's `preflight` hook validates the prompt as JSON before agent dispatch and will block with D-09 if it receives any non-JSON content.

Spawn a Task with agent `go-cartographer` and the following input:

```json
{
  "working_directory": "<TARGET_DIR>",
  "review_session_id": "<SESSION_ID>",
  "code_ref": "<CODE_REF>",
  "code_ref_dirty": <CODE_REF_DIRTY>
}
```

Wait for it to complete. If it fails or produces an error response, stop with error. The output will be written to `TARGET_DIR/go-index.json`.

## Step 5: Read cartographer output

Read `TARGET_DIR/go-index.json`. Extract the following fields for use in Step 6:
- `entrypoints` — array of route objects
- `authz_primitives` — array of authz primitive objects  
- `sinks_by_kind` — map of sink kind to array of sink locations
- `oauth_locations` — OAuth surface locations

## Step 6: Run tracers (parallel)

**JSON prompt enforcement:** Each tracer Task's `prompt` value MUST be the exact JSON object shown in each sub-section below — no prose, no wrapper text, no markdown. The preflight hook validates all tracer prompts as JSON and will block the entire fan-out if any prompt is not a valid JSON object. Do not paraphrase or annotate the JSON.

Spawn all four tracer agents simultaneously using parallel Task invocations. Do not wait for one before starting the others. Pass SESSION_ID to each.

### Tracer 1: go-authz-tracer

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

Output file: `TARGET_DIR/authz-findings.json`

### Tracer 2: go-oauth-auditor

Input:
```json
{
  "working_directory": "<TARGET_DIR>",
  "oauth_locations": <oauth_locations object from go-index.json>,
  "review_session_id": "<SESSION_ID>",
  "code_ref": "<CODE_REF>",
  "code_ref_dirty": <CODE_REF_DIRTY>
}
```

Output file: `TARGET_DIR/oauth-checklist.json`

### Tracer 3: invariant-checker

Input:
```json
{
  "flow_name": "security-review",
  "invariants": [
    {"id": "I1", "statement": "No SQL query is constructed by string concatenation with user-controlled input without parameterization"},
    {"id": "I2", "statement": "All HTTP handlers that modify state require authentication before processing"},
    {"id": "I3", "statement": "OAuth state parameter is validated on callback before proceeding"},
    {"id": "I4", "statement": "Fund transfer operations validate that source account belongs to authenticated user"}
  ],
  "review_session_id": "<SESSION_ID>",
  "code_ref": "<CODE_REF>",
  "code_ref_dirty": <CODE_REF_DIRTY>
}
```

Output file: `TARGET_DIR/invariant-results.json`

### Tracer 4: go-taint-tracer (one invocation per SQL sink)

For each sink in `sinks_by_kind.sql_exec` from go-index.json, spawn a separate `go-taint-tracer` Task. Construct source/sink pairs by matching each SQL sink's file/line to the nearest handler function from `entrypoints`.

When constructing the `source` field, use `first_param_read_line` from the entrypoint's `handler` object if it is present and non-zero; otherwise fall back to `handler.line` (the function definition line). Use `first_param_read_expr` as the `expr` value if present; otherwise use the inferred parameter read expression.

For each (handler entrypoint, SQL sink) pair, the input is:
```json
{
  "source": {
    "file": "<handler file from entrypoints>",
    "line": <first_param_read_line from entrypoints if present, otherwise handler.line>,
    "expr": "<first_param_read_expr from entrypoints if present, otherwise 'r.URL.Query().Get(...) or r.PostForm or r.Body'>",
    "kind": "http_query"
  },
  "sink": {
    "file": "<sink file>",
    "line": <sink line>,
    "expr": "<sink callee>",
    "kind": "sql_exec"
  },
  "max_depth": 8,
  "semgrep_tier": "intrafile",
  "review_session_id": "<SESSION_ID>",
  "code_ref": "<CODE_REF>",
  "code_ref_dirty": <CODE_REF_DIRTY>
}
```

Name each output file `TARGET_DIR/taint-verdict-<handler-name>-sqli.json` where `<handler-name>` is derived from the handler's FQN (last component, lowercased, with "Handler" stripped).

## Step 7: Run synthesis (sequential)

Wait for ALL four tracer agents from Step 6 to complete. Then spawn a Task with agent `synthesis`:

```json
{
  "working_directory": "<TARGET_DIR>",
  "review_id": "<SESSION_ID>"
}
```

Wait for it to complete. It will produce `TARGET_DIR/review-report.json` and `TARGET_DIR/review-report.md`.

## Step 8: Report completion

Once synthesis completes, report:
- The review ID (SESSION_ID)
- Total findings by severity from review-report.json
- Path to review-report.md

---

## Key constraints (MUST follow)

- **Always run all steps in order** — do not skip any step even if intermediate files exist from a prior run. The stale files were deleted in Step 2.
- **Always generate a fresh SESSION_ID** — never reuse a session ID from a previous run.
- **Never read old intermediate files** — Step 2 deletes them; there should be nothing to read before the agents create them.
- **Pass SESSION_ID to every agent** — it is required in all Task inputs (cartographer, all 4 tracers, synthesis).
- **Parallel tracers** — all four tracer Tasks in Step 6 must be started simultaneously, not sequentially.
- **All Task prompts are pure JSON** — every Task `prompt` field in this skill must be a valid JSON object matching the exact template shown. Never wrap JSON in prose, markdown, or explanation. The preflight hook enforces this at runtime.
