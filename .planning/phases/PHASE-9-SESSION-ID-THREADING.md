# Phase 9: Session ID Threading

**Goal:** Add session ID tracking to all intermediate security review outputs to enable cross-agent deduplication and tracing. All specialist agents echo their session ID in their output, and synthesis validates matching session IDs during result merging.

**Phase type:** Infrastructure/Hardening

**Dependencies:** Phases 1–8 complete (all agent implementations and hooks binary finalized)

## What the Codebase Looks Like (Before)

After Phase 8, the security review workflow has:

1. **Agent schemas** (`internal/schema/*.go`):
   - `TaintInput` / `TaintVerdict` with optional `ReviewSessionID` field
   - `AuthzInput` / `AuthzVerdict` with optional `ReviewSessionID` field
   - `OAuthInput` / `OAuthVerdict` with optional `ReviewSessionID` field
   - `InvariantCheckerInput` / `InvariantCheckerVerdict` with optional `ReviewSessionID` field
   - `SynthesisReport` (no session ID — uses `review_id` from orchestration)

2. **Orchestration command** (`.claude/commands/security-review.md`):
   - Generates a `SESSION_ID` via `uuidgen` at command start
   - Passes `SESSION_ID` to cartographer as `review_session_id`
   - Passes `SESSION_ID` to synthesis as `review_id`
   - Does NOT pass `SESSION_ID` to the four fan-out tracers (go-taint-tracer, go-authz-tracer, go-oauth-auditor, invariant-checker)

3. **Agent definitions** (`.claude/agents/*.md`):
   - Input/output schemas document `review_session_id` field (optional)
   - Synthesis protocol does NOT validate matching session IDs during deduplication

## What Needs to Change

### 1. Update Agent Definition Files (5 files)

**Affected files:**
- `.claude/agents/go-taint-tracer.md`
- `.claude/agents/go-authz-tracer.md`
- `.claude/agents/go-oauth-auditor.md`
- `.claude/agents/invariant-checker.md`
- `.claude/agents/synthesis.md`

**Changes:**

For each of the first four tracer agents, update the "Input Contract" section to include:
```
"review_session_id": "<uuid>" (optional, string) — Session identifier passed from orchestration command
```

For each of the first four tracer agents, update the "Output Schema" section to include:
```
"review_session_id": <uuid> — Echo of input review_session_id if provided
```

For synthesis.md, update "Step 1: Read all specialist outputs" to add a new sub-step:
> Check each file's `review_session_id` field. Skip any file where `review_session_id` is present but does not match the `review_id` passed in the Task input. Log skipped files as informational warnings in `deduplication_notes`.

### 2. Update Orchestration Command (1 file)

**Affected file:** `.claude/commands/security-review.md`

**Changes:**

In Stage 3 (Tracer Fan-Out), for each of the four tracer Task invocations, add the session ID to the input JSON:
- `go-taint-tracer` input: add `"review_session_id": "$SESSION_ID"`
- `go-authz-tracer` input: add `"review_session_id": "$SESSION_ID"`
- `go-oauth-auditor` input: add `"review_session_id": "$SESSION_ID"`
- `invariant-checker` input: add `"review_session_id": "$SESSION_ID"`

Update Implementation Notes to clarify:
> The `review_id` is generated once at command start and stored in `$SESSION_ID`. This UUID is passed to cartographer as `review_session_id`, to all four tracers as `review_session_id`, and to synthesis as `review_id`.

### 3. Build and Install Hooks Binary (1 step)

**Affected files:** `claude-security-hooks/bin/claude-security-hooks`

The hooks binary is already built with all schema fields in place. Copy it to `.claude/hooks/bin/claude-security-hooks`.

## Success Criteria

1. **Agent documentation updated**: All 5 agent .md files include session ID in input and output schema sections.

2. **Orchestration updated**: All four tracer Task invocations in `.claude/commands/security-review.md` Stage 3 include `"review_session_id": "$SESSION_ID"`.

3. **Binary installed**: Hooks binary exists at `.claude/hooks/bin/claude-security-hooks` (2.3 MB static binary, no dynamic lib dependencies).

4. **Session ID echoing**: All tracer agents will echo `review_session_id` in their JSON output when provided in input.

5. **Synthesis deduplication**: Synthesis validates that intermediate file session IDs match the `review_id` passed from orchestration; skipped files are logged in deduplication_notes.

6. **Cross-review isolation**: If multiple review runs happen in the same directory, their intermediate outputs are isolated by session ID; synthesis from run A will not merge findings from run B.

## Testing Notes

- **Unit tests**: The 51-assertion test suite in `claude-security-hooks/internal/invariants/` validates that session ID fields are properly echoed in all verdicts.
- **Integration testing**: The e2e workflow in Phase 5 can be extended to verify that session IDs are threaded through all five agents.
- **Manual verification**: After this phase, running `/security-review` twice in the same directory should produce two independent review-report.json files (by separate naming or session-based filtering).

## Deliverables

- [ ] 5 agent .md files updated with session ID documentation
- [ ] `.claude/commands/security-review.md` updated with session ID injection in Stage 3
- [ ] `.claude/hooks/bin/claude-security-hooks` binary installed (2.3 MB)
- [ ] All changes committed with phase-tagged message

## Files Modified

1. `.claude/agents/go-taint-tracer.md` — Input Contract + Output Schema
2. `.claude/agents/go-authz-tracer.md` — Input Contract + Output Schema
3. `.claude/agents/go-oauth-auditor.md` — Input Contract + Output Schema
4. `.claude/agents/invariant-checker.md` — Input Contract + Output Schema
5. `.claude/agents/synthesis.md` — Step 1 protocol (deduplication logic)
6. `.claude/commands/security-review.md` — Stage 3 + Implementation Notes

**Total files modified:** 6

**Summary:** Session ID threading adds cross-agent tracing and deduplication support to the security review workflow. All intermediate outputs now carry a session identifier that ties them to a single orchestrated review run, enabling synthesis to validate and deduplicate findings with confidence that they originated from the same session.
