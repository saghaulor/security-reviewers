# Phase 9 Diagnosis: Orchestration Not Running Agents

**Date:** 2026-05-25  
**Status:** Phase 9 code complete, but orchestration command not executing agents  
**Token Budget:** Exceeded - needs fresh session

## Problem Observed

Running `/security-review ./examples/sample-vulnerable-service`:
1. ✗ No agents spawned (go-cartographer, tracers, synthesis)
2. ✗ Old Phase 8 intermediate files reused silently
3. ✗ OpenGrep MCP server not started (suggests early exit)
4. ✗ No error message - just returned with stale results

## Root Cause (Hypothesis)

The `.claude/commands/security-review.md` file likely has:
- Syntax errors in the orchestration Task invocations
- Incorrect JSON structure for agent input contracts
- Malformed `$SESSION_ID` variable substitution
- Stage 0 (bootstrap) failing silently without preventing continuation

## Investigation Needed

1. **Read `.claude/commands/security-review.md` completely** - Check:
   - Stage 0 (opengrep-mcp bootstrap) - does it fail gracefully or exit?
   - Stage 1 (Pre-Pass) - are govulncheck and graphify invoked?
   - Stage 2 (Cartographer) - is the JSON input valid?
   - Stage 3 (Tracer Fan-Out) - are all 4 agents' JSON inputs valid?
   - Stage 4 (Synthesis) - is review_id passed correctly?

2. **Verify agent .md files** - Check:
   - Each agent's input JSON examples are valid
   - review_session_id is properly documented
   - Output schema includes review_session_id

3. **Delete stale intermediate files** - Before re-testing:
   ```bash
   rm examples/sample-vulnerable-service/{go-index.json,authz-findings.json,oauth-checklist.json,taint-verdict-*.json,invariant-results.json}
   ```

4. **Test orchestration in isolation** - Run just Stage 1-2 to verify:
   - govulncheck produces output
   - graphify builds the graph
   - cartographer is invoked

## Phase 9 Code Status

✅ Schema changes complete (ReviewSessionID in all 5 types)
✅ Invariants T12/AZ7/OA8/IC5 implemented
✅ Tests written (5 functions, 20 cases)
✅ Agent .md files updated
✅ Hooks binary rebuilt with Status field fix
✅ PHASE-9 spec created

❌ **Orchestration command not executing** - needs debugging

## Files That Need Review

Priority order:
1. `.claude/commands/security-review.md` - **CRITICAL** - the orchestration definition
2. `examples/sample-vulnerable-service/` - delete stale files before retesting
3. Agent .md files - verify JSON examples are syntactically valid

## Next Steps (Fresh Session)

1. Read and validate `.claude/commands/security-review.md` syntax
2. Check each agent's input/output JSON in their .md files
3. Delete stale intermediate files from sample service
4. Re-run `/security-review ./examples/sample-vulnerable-service` with diagnostics
5. If agents still don't spawn, trace the command execution step-by-step

## Commands to Run

```bash
# Delete stale files
rm examples/sample-vulnerable-service/{go-index.json,authz-findings.json,oauth-checklist.json,taint-verdict-*.json,invariant-results.json,review-report.json,review-report.md} 2>/dev/null

# Verify binary is installed
ls -lh .claude/hooks/bin/claude-security-hooks

# Check orchestration syntax (read the full file)
cat .claude/commands/security-review.md | head -100
```

---

**Critical Files:**
- `/home/saghaulor/code/security_reviewer/.claude/commands/security-review.md` - the orchestration definition
- `/home/saghaulor/code/security_reviewer/.planning/phases/PHASE-9-SESSION-ID-THREADING.md` - completed spec
- `/home/saghaulor/code/security_reviewer/examples/sample-vulnerable-service/` - test target (with stale files)
