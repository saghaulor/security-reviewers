# security-reviewer

A Claude Code plugin for systematic security review of Go services. Combines multi-agent analysis (cartography, taint tracing, authorization verification, OAuth auditing, invariant checking, synthesis) with mechanical output validation.

**Status:** Pre-release. See `.planning/ROADMAP.md` for the 6-phase implementation plan.

## Quick Start

Once the plugin is ready (Phase 3+):

```bash
/security-review
```

This runs the full workflow:
1. Build a structural index of your Go service (`go-cartographer`)
2. Dispatch parallel tracers for taint-based flows, authorization, OAuth conformance
3. Synthesize findings into a machine-readable report

## Components

- **`claude-security-hooks`** — Go binary that validates agent outputs via lifecycle hooks
- **`opengrep-mcp`** — MCP server fronting Semgrep Pro and OpenGrep for policy-driven scanning
- **`.claude/agents/`** — Six Claude subagent definitions (cartographer, tracers, auditors, synthesis)

## Documentation

- Architecture and design: see [`HAND_OFF.md`](HAND_OFF.md) (the design document bootstrapping this project)
- Planning: see [`.planning/`](.planning/) directory (PROJECT.md, ROADMAP.md, REQUIREMENTS.md)
- Implementation progress: tracked in `.planning/STATE.md`

## Development

This project follows Go-first, TDD-with-tests-before-implementation discipline per the design handoff (HAND_OFF.md). See `CONTRIBUTING.md` (coming Phase 6) for contributor guidelines.

---

**Copyright:** 2026 Stephen Aghaulor. Licensed under MIT. See LICENSE for details.
