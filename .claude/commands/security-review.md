---
name: security-review
description: Run a complete security review of a Go module using specialist agents and the opengrep-mcp scanner
allowed-tools: Task, Read, Glob, Bash
---

# /security-review Command

Run a complete security review of a Go module, orchestrating the full pipeline: pre-pass artifacts → cartographer structural analysis → fan-out tracers → synthesis.

## Usage

```
/security-review [directory]
```

**Parameters:**
- `directory` (optional): Path to the Go module root. Defaults to current working directory.

## Workflow Stages

### Stage 0: Bootstrap opengrep-mcp (NEW)

Before running any security analysis, ensure the opengrep-mcp server is available. This stage checks if the server is running and starts it via Docker if needed.

**Procedure:**

1. **Health Check:** Issue `GET http://localhost:8000/health` with a 5-second timeout.
   - If the server responds with a 2xx status code, proceed to Stage 1.
   - If the server is unreachable (connection refused, timeout), proceed to step 2.

2. **Start Docker Container (if needed):**
   - Execute: `docker run -d --name opengrep-mcp-e2e-phase8 -p 8000:8000 opengrep-mcp:latest`
   - If the container name already exists (from a prior run), stop and remove it first:
     ```bash
     docker stop opengrep-mcp-e2e-phase8 2>/dev/null || true
     docker rm opengrep-mcp-e2e-phase8 2>/dev/null || true
     docker run -d --name opengrep-mcp-e2e-phase8 -p 8000:8000 opengrep-mcp:latest
     ```

3. **Wait for Health Check to Succeed:**
   - Poll `GET http://localhost:8000/health` every 1 second, up to 30 seconds.
   - Once the endpoint responds 2xx, proceed to Stage 1.
   - If the health check never succeeds within 30 seconds, fail hard with error message.

4. **Error Handling:**
   - If Docker is not available (`docker: command not found`), return error: `"opengrep-mcp bootstrap failed: docker not available"`
   - If the image `opengrep-mcp:latest` does not exist, return error: `"opengrep-mcp bootstrap failed: image not found. Build with: cd /path/to/opengrep-mcp && make build && docker build -t opengrep-mcp:latest ."`
   - If the container fails to start (non-zero exit), return error: `"opengrep-mcp bootstrap failed: container startup error. Check 'docker logs opengrep-mcp-e2e-phase8'"`
   - If the health check times out after 30 seconds, return error: `"opengrep-mcp bootstrap failed: health check timeout after 30 seconds"`
   - **No graceful degradation:** If bootstrap fails, `/security-review` stops immediately. Do not attempt to run the workflow without the server.

**Rationale:** The opengrep-mcp scanner is essential to the analysis pipeline. Bootstrapping it automatically ensures users can run `/security-review` without manual Docker setup. Failing hard on bootstrap errors provides clear feedback rather than producing partial or incorrect results.

### Stage 1: Pre-Pass Artifacts

Generate prerequisite analysis files:
1. **Graphify graph:** `graphify build [directory]` produces `graphify-out/graph.json`
2. **Vulnerability check:** `govulncheck -json ./...` produces `govulncheck.json`

If either command fails, return structured error and stop.

### Stage 2: Cartographer (Sequential)

Invoke the `go-cartographer` agent once. It reads the graphify output and produces `go-index.json`, a structural index of:
- HTTP routers and entrypoints
- Data sinks by kind (SQL, cmd, path, etc.)
- OAuth surfaces and redirect URIs
- Payment processing surfaces
- Authorization primitives (canaries for missing authz)
- Vulnerability findings from govulncheck

**Input to cartographer:**
```json
{
  "working_directory": "[directory]",
  "review_session_id": "[generated_uuid]"
}
```

Wait for completion. On error, return structured error and stop.

### Stage 3: Tracer Fan-Out (Parallel)

Once cartographer completes, dispatch four tracer agents in parallel against the cartographer output:

1. **go-taint-tracer:** For each (source, sink) pair from the cartographer index, verify exploitability of data flow from source to sink. Produces `taint-verdict-*.json` files.

2. **go-authz-tracer:** For each authorization-missing endpoint, trace whether the missing authorization is retrievable via a call to a known authz primitive. Produces `authz-findings.json`.

3. **go-oauth-auditor:** For OAuth surfaces, apply the checklist taxonomy to all OAuth code paths. Dispatches taint pairs to go-taint-tracer for scope-tampering verification. Produces `oauth-checklist.json`.

4. **invariant-checker:** Verify language-level invariants (type safety, nil receiver, race conditions). Produces `invariant-results.json`.

All four run in parallel; synthesis must wait for all four to complete.

### Stage 4: Synthesis (Sequential)

Once all tracers complete, invoke the `synthesis` agent once. It reads all tracer outputs from the working directory and produces:
- `review-report.json` (machine-readable, conforming to review-report/v1 schema)
- `review-report.md` (human-readable markdown)

**Input to synthesis:**
```json
{
  "working_directory": "[directory]",
  "review_id": "[same_uuid_from_stage_1]"
}
```

The `review_id` is generated once at command start and injected into both cartographer and synthesis for consistency.

## Implementation Notes

- **review_id generation:** Use Bash `uuidgen` to generate a UUID at command start. This UUID is passed to cartographer as `review_session_id` and to synthesis as `review_id`.
- **Pre-pass failures:** If graphify or govulncheck fails, return error immediately without invoking agents.
- **Agent failures:** If any agent (cartographer, tracer, synthesis) fails, return structured error and stop. Partial results are not acceptable.
- **Output directory:** All JSON artifacts are written to the working directory (where graphify-out/ is located). The synthesis agent produces `review-report.json` and `review-report.md` in the same directory.

## Example

```bash
/security-review /home/user/vulnerable-app
```

This command will:
1. Build the graphify graph and run govulncheck against `/home/user/vulnerable-app`
2. Run cartographer to produce `go-index.json`
3. Fan out four tracers in parallel
4. Wait for all tracers, then run synthesis
5. Produce `review-report.json` and `review-report.md`
