---
phase: 11
slug: codegraph-migration
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-05-27
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> This phase is a pure config/script/markdown migration — no Go code changes.
> All automated verification is via shell assertions embedded in task `<verify>` blocks.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | None (no Go/TS code changes — shell `test`/`grep`/`jq` checks only) |
| **Config file** | N/A |
| **Quick run command** | `which codegraph && jq -e '.mcpServers.codegraph != null and .mcpServers.graphify == null' .mcp.json` |
| **Full suite command** | manual `/security-review examples/sample-vulnerable-service` |
| **Estimated runtime** | < 1 second (automated); ~5 min (manual integration) |

---

## Sampling Rate

- **After every task commit:** Run quick `which codegraph` + `jq` + `grep` checks from task `<verify>` block
- **After every plan wave:** Run full `grep -r mcp__graphify__ .claude/` (should return nothing)
- **Before `/gsd-verify-work`:** Manual smoke test must complete without `command not found: codegraph`
- **Max feedback latency:** < 1 second (automated assertions)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | SC | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|----|-----------------|-----------|-------------------|-------------|--------|
| 11-01-T0 | 01 | 1 | prerequisite | `which codegraph` exits 0 | smoke | `which codegraph` | ❌ prereq | ⬜ pending |
| 11-01-T1 | 01 | 1 | SC-1, SC-2, SC-3 | codegraph-mcp.sh exists+executable; graphify-mcp.sh absent; .mcp.json has codegraph key, no graphify key | smoke | `test -x scripts/codegraph-mcp.sh && test ! -f scripts/graphify-mcp.sh && jq -e '.mcpServers.codegraph != null and .mcpServers.graphify == null' .mcp.json` | ❌ W1 | ⬜ pending |
| 11-01-T2 | 01 | 1 | SC-6 | `.gitignore` gains `.codegraph/` | smoke | `grep -c "^\.codegraph/$" examples/sample-vulnerable-service/.gitignore` | ✅ existing | ⬜ pending |
| 11-02-T1 | 02 | 2 | SC-4 | No `mcp__graphify__` in go-cartographer.md | smoke | `grep -c "mcp__graphify__" .claude/agents/go-cartographer.md` (must be 0) | ✅ existing | ⬜ pending |
| 11-02-T2 | 02 | 2 | SC-4 | `security-review.md` Step 3 uses `codegraph init` + `codegraph index`; no `graphify` | smoke | `grep -c "graphify" .claude/commands/security-review.md` (must be 0) | ✅ existing | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

No Wave 0 test scaffolding required — this phase has no Go code changes.
All test infrastructure (shell `test`, `grep`, `jq`) is available in the environment.

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

| Behavior | SC | Why Manual | Test Instructions |
|----------|----|------------|-------------------|
| Full `/security-review` run completes with codegraph as graph source | SC-7 | Requires `codegraph` binary on PATH, live MCP server, full pipeline | Run: `/security-review examples/sample-vulnerable-service`; confirm `review-report.json` exists with non-ambiguous findings |
| `codegraph_status` MCP reachability confirmed | SC-4 | Requires MCP session active | Open a session with codegraph MCP server; call `mcp__codegraph__codegraph_status`; confirm `fileCount > 0` |

---

## Phase Success Criteria (from ROADMAP.md)

| SC | Criterion | Automated | Plan Task |
|----|-----------|-----------|-----------|
| SC-1 | `scripts/codegraph-mcp.sh` exists, executable, runs `codegraph serve --mcp --path TARGET_DIR` | Yes | 11-01 T1 |
| SC-2 | `.mcp.json` has `codegraph` stdio entry; no `graphify` key | Yes | 11-01 T1 |
| SC-3 | `go-cartographer.md` `tools:` has `mcp__codegraph__*`; no `mcp__graphify__*` | Yes | 11-02 T1 |
| SC-4 | `security-review.md` Step 3 has `codegraph init` + `codegraph index`; no `graphify` | Yes | 11-02 T2 |
| SC-5 | `go-taint-tracer.md` updated (stretch/optional per D-10) | n/a | Deferred |
| SC-6 | `examples/sample-vulnerable-service/.gitignore` contains `.codegraph/` | Yes | 11-01 T2 |
| SC-7 | Full `/security-review` run completes with codegraph; `review-report.json` has findings | Manual | — |
