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

Run `Bash: uuidgen` and store the output (trimmed) as SESSION_ID. This UUID will be passed to every agent and embedded in all output files so results can be traced to this exact run.

## Step 2: Delete stale intermediate files

Remove any existing intermediate files from the previous run. These files from any prior run are INVALID for this run because they were produced under a different SESSION_ID.

Run the following (ignore errors if files don't exist):
```
Bash: rm -f TARGET_DIR/go-index.json TARGET_DIR/authz-findings.json TARGET_DIR/oauth-checklist.json TARGET_DIR/invariant-results.json TARGET_DIR/review-report.json TARGET_DIR/review-report.md
Bash: rm -f TARGET_DIR/taint-verdict-*.json
```

Replace TARGET_DIR with the actual resolved path.

## Step 3: Bootstrap opengrep-mcp

Issue `GET http://localhost:8000/health` with a 5-second timeout using Bash curl.

- If the server responds with 2xx, proceed to Step 4.
- If unreachable, start the container:
  ```
  Bash: docker stop opengrep-mcp-e2e-phase8 2>/dev/null; true
  Bash: docker rm opengrep-mcp-e2e-phase8 2>/dev/null; true
  Bash: docker run -d --name opengrep-mcp-e2e-phase8 -p 8000:8000 opengrep-mcp:latest
  ```
  Then poll `GET http://localhost:8000/health` every second for up to 30 seconds. If it never responds 2xx, **stop with error**: `"opengrep-mcp bootstrap failed: health check timeout after 30 seconds"`.

If Docker is unavailable or the image doesn't exist, **stop with a clear error message**.

## Step 4: Run pre-pass artifacts

Run these two commands with TARGET_DIR as the working directory:

1. `Bash (cwd=TARGET_DIR): graphify build .`
   - This produces `TARGET_DIR/graphify-out/graph.json`.
   - If it fails, stop with error.

2. `Bash (cwd=TARGET_DIR): govulncheck -json ./... > govulncheck.json 2>govulncheck.err`
   - If `govulncheck` is not installed, write `{"available":false}` to `govulncheck.json` and continue.
   - If it fails for other reasons, stop with error.

## Step 5: Run cartographer (sequential)

Spawn a Task with agent `go-cartographer` and the following input:

```json
{
  "working_directory": "<TARGET_DIR>",
  "review_session_id": "<SESSION_ID>"
}
```

Wait for it to complete. If it fails or produces an error response, stop with error. The output will be written to `TARGET_DIR/go-index.json`.

## Step 6: Read cartographer output

Read `TARGET_DIR/go-index.json`. Extract the following fields for use in Step 7:
- `entrypoints` — array of route objects
- `authz_primitives` — array of authz primitive objects  
- `sinks_by_kind` — map of sink kind to array of sink locations
- `oauth_locations` — OAuth surface locations

## Step 7: Run tracers (parallel)

Spawn all four tracer agents simultaneously using parallel Task invocations. Do not wait for one before starting the others. Pass SESSION_ID to each.

### Tracer 1: go-authz-tracer

Input:
```json
{
  "routes": <entrypoints array from go-index.json>,
  "authz_primitives": <authz_primitives array from go-index.json>,
  "sensitive_operations": [],
  "review_session_id": "<SESSION_ID>"
}
```

Output file: `TARGET_DIR/authz-findings.json`

### Tracer 2: go-oauth-auditor

Input:
```json
{
  "working_directory": "<TARGET_DIR>",
  "oauth_locations": <oauth_locations object from go-index.json>,
  "review_session_id": "<SESSION_ID>"
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
  "review_session_id": "<SESSION_ID>"
}
```

Output file: `TARGET_DIR/invariant-results.json`

### Tracer 4: go-taint-tracer (one invocation per SQL sink)

For each sink in `sinks_by_kind.sql_exec` from go-index.json, spawn a separate `go-taint-tracer` Task. Construct source/sink pairs by matching each SQL sink's file/line to the nearest handler function from `entrypoints`.

For each (handler entrypoint, SQL sink) pair, the input is:
```json
{
  "source": {
    "file": "<handler file from entrypoints>",
    "line": <handler line from entrypoints>,
    "expr": "r.URL.Query().Get(\"...\") or r.PostForm or r.Body",
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
  "review_session_id": "<SESSION_ID>"
}
```

Name each output file `TARGET_DIR/taint-verdict-<handler-name>-sqli.json` where `<handler-name>` is derived from the handler's FQN (last component, lowercased, with "Handler" stripped).

## Step 8: Run synthesis (sequential)

Wait for ALL four tracer agents from Step 7 to complete. Then spawn a Task with agent `synthesis`:

```json
{
  "working_directory": "<TARGET_DIR>",
  "review_id": "<SESSION_ID>"
}
```

Wait for it to complete. It will produce `TARGET_DIR/review-report.json` and `TARGET_DIR/review-report.md`.

## Step 9: Report completion

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
- **Parallel tracers** — all four tracer Tasks in Step 7 must be started simultaneously, not sequentially.
