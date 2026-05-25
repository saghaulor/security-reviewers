# claude-security-hooks

The hooks binary (`claude-security-hooks`) is a lifecycle validator that runs between Claude Code subagents during security review. It ensures that specialist agents (taint-tracer, authz-tracer, oauth-auditor, invariant-checker, synthesis) output valid, well-formed verdicts that conform to security invariants before those verdicts are accepted by the main orchestrator.

## What It Does

When a specialist agent completes its analysis (e.g., go-taint-tracer finishes tracing a source/sink pair), the hooks binary intercepts the output and answers three questions:

1. **Preconditions met?** (PreToolUse hook) — Does the agent have the right tools to do its job? Is the input structurally sound?
2. **Output valid?** (PostToolUse hook) — Is the verdict JSON well-formed and does it conform to the agent's schema?
3. **Invariants hold?** (PostToolUse hook) — Does the verdict satisfy all 51 security and structural assertions?

If any check fails, the hooks binary blocks the verdict and forces retry with explicit feedback. This prevents silent drift: if an agent's instructions are weakened or a tool allowlist is loosened, the mechanical validation catches it immediately.

## Three Subcommands

The hooks binary exposes three subcommands, invoked via Claude Code's PreToolUse and PostToolUse hooks:

### `preflight`

Runs before a tracer starts analysis. Validates preconditions.

```bash
claude-security-hooks preflight --agent go-taint-tracer --input input.json
```

**Checks:**
- Input JSON is syntactically valid
- All required fields are present (source, sink, max_depth, etc.)
- No prohibited tools in the agent's allowlist (e.g., go-taint-tracer must NOT have Bash or Grep)
- Agent definition file exists and is valid YAML

**Exit codes:**
- `0` — All checks passed; agent may start
- `1` — Input malformed or precondition failed; shows structured error message with remediation

### `validate`

Runs after a tracer finishes analysis. Validates output verdict.

```bash
claude-security-hooks validate --agent go-taint-tracer --verdict verdict.json
```

**Checks:**
- Verdict JSON is syntactically valid
- Schema matches expected output for this agent (e.g., taint-tracer must have `{confidence, data_flow_path, ...}`)
- All 51 assertions pass (see "The 51-Assertion Concept" below)

**Exit codes:**
- `0` — Verdict is valid; accept it
- `1` — Verdict is invalid or assertions failed; show details on which assertions failed, suggest retry

### `inject-context`

Runs when a tracer is started (SubagentStart hook). Pre-loads cartographer output and prior findings into the tracer's context.

```bash
claude-security-hooks inject-context --agent go-taint-tracer --cartographer-output go-index.json
```

**Purpose:** Reduce token usage by staging cartographer output once instead of requiring each tracer to re-discover it.

**No validation:** This subcommand succeeds unless the cartographer output file is missing or malformed. It does not validate the context itself; that's the tracer's job when it processes the context.

## Build, Test, Install

All operations go through the Makefile:

```bash
# Build the binary (static, CGO disabled)
make build

# Output: ./bin/claude-security-hooks (statically linked)
# Flags: CGO_ENABLED=0 -ldflags='-s -w' (optimize for size)
```

**Test:**

```bash
make test

# Runs: go test ./...
# Must pass before deployment
```

**Install:**

```bash
make install

# Copies ./bin/claude-security-hooks to ~/.claude/hooks/bin/
# Makes it available to Claude Code's hook system
```

**Verification (optional):**

```bash
make verify-static

# Confirms binary is statically linked (no CGO dependencies)
# Fails if binary depends on dynamic libraries
```

## The 51-Assertion Concept

The hooks binary enforces 51 security and structural assertions across six domains:

| Domain | Count | Purpose |
|--------|-------|---------|
| **Cartographer (A)** | A1–A11 | Structural index soundness (no cycles, all entrypoints indexed, etc.) |
| **Taint Analysis (T)** | T1–T11 | Taint tracing correctness (data-flow path valid, sanitizers recognized, etc.) |
| **Authorization (AZ)** | AZ1–AZ6 | Middleware chain analysis soundness (no missing guards, scope correct) |
| **OAuth (OA)** | OA1–OA7 | OAuth conformance (scope validation, redirect_uri checks, token endpoint hardening) |
| **Invariants (IC)** | IC1–IC4 | Business logic assertions (no implicit trust boundaries, etc.) |
| **Synthesis (S)** | S1–S6 | Report generation (deduplication, evidence completeness, confidence calibration) |

Each assertion is a test case in the codebase (`./*_test.go` files in `internal/invariants/`) plus a check predicate that runs on every verdict.

**Example:** Assertion T5 ("Sanitizer exclusion") tests that when a taint path passes through a recognized sanitizer (e.g., `sql.EscapeString`), the tracer correctly terminates the flow and does not report a false positive. The test case verifies this; the check predicate validates that the verdict respects it.

**Full specification:** See `HAND_OFF.md` §3 "Assertion Catalog" for the complete list, test names, and detailed semantics.

## Development: How to Add a New Assertion

Follow TDD (test-first) discipline. This is the pattern Phase 6 uses to extend the system:

1. **Write the test case** (RED phase):
   - Create a new test in `internal/invariants/{domain}_test.go`
   - Describe the behavior you want to assert: `TestT12_CustomSanitizerRecognized(t *testing.T)`
   - Write test expectations (what should be true of a valid verdict)
   - Run `go test ./...` — test should fail (RED)

2. **Implement the check predicate** (GREEN phase):
   - Add the check function to `internal/invariants/{domain}.go`
   - Implement the predicate that validates the assertion on verdicts
   - Run `go test ./...` — test should pass (GREEN)

3. **Document the assertion**:
   - Add a comment in the code: `// T12: Custom sanitizers are recognized and terminate taint flow`
   - Update `HAND_OFF.md` §3 to include the new assertion

4. **Version the catalog**:
   - If this is a minor enhancement, increment a patch version in the hooks binary version
   - If this enables new stack support (e.g., Python), increment minor version

**Why test-first?** Writing the test before the predicate forces you to define expected behavior explicitly. This prevents silent drift: if someone later weakens the assertion, the test will catch it.

## See Also

- `HAND_OFF.md` §3 — Full assertion catalog and test names
- `internal/schema/` — Go struct definitions for each agent's verdict type
- `.claude/agents/README.md` — How agents interact with these hooks
