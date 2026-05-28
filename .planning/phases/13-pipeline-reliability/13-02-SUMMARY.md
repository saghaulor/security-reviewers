---
phase: 13-pipeline-reliability
plan: "02"
subsystem: bootstrap
tags: [bootstrap, preflight, shell, makefile]
dependency_graph:
  requires: []
  provides: [bootstrap/pre-flight-checks.sh, Makefile:preflight]
  affects: [Makefile]
tech_stack:
  added: []
  patterns: [bash-preflight-checks, makefile-target-delegation]
key_files:
  created:
    - bootstrap/pre-flight-checks.sh
  modified:
    - Makefile
    - .gitignore
decisions:
  - "Use FAILURES counter (not set -e) so all 6 checks run and user gets a complete picture of missing deps in one pass"
  - "Docker absence is warning-only with DOCKER_AVAILABLE export so govulncheck fallback works transparently"
  - "codegraph index retry clears .codegraph/ on first failure to recover from corrupted partial index"
  - "realpath with 2>/dev/null fallback to raw arg handles nonexistent TARGET_DIR gracefully"
metrics:
  duration: "~3 minutes"
  completed: "2026-05-28"
  tasks_completed: 2
  files_changed: 3
---

# Phase 13 Plan 02: Bootstrap Pre-flight Checks Summary

Implemented `bootstrap/pre-flight-checks.sh` with 6 sequential environment checks and wired it to the root Makefile as a `preflight` target. Eliminates fresh-environment failures — users run `make preflight TARGET=/path/to/project` and get actionable remediation messages for any missing dependency.

## What Was Built

### bootstrap/pre-flight-checks.sh

A bash script that runs 6 checks sequentially. Uses a `FAILURES` counter rather than `set -e` so all checks run even when earlier ones fail — the user gets a complete picture in one pass.

**Checks implemented:**

1. **Codegraph** — `command -v codegraph`; hard failure with `go install` remediation on miss. Sets `CODEGRAPH_OK` flag to gate Check 6.
2. **Make + Go + hooks module** — three individual hard failures: `command -v make`, `command -v go`, presence of `claude-security-hooks/go.mod` relative to `REPO_ROOT`.
3. **Git + target repo** — git binary absence is hard failure; TARGET_DIR not being a git repo is a warning (non-git Go projects are valid targets).
4. **Go project structure** — `TARGET_DIR/go.mod` absence is hard failure; `go.sum` absence is a warning with `go mod tidy` hint.
5. **Docker** — all outcomes are warnings. Exports `DOCKER_AVAILABLE=true|false`; govulncheck has a `{"available":false}` fallback mode.
6. **Codegraph init + index with retry** — runs only if Check 1 passed. On first index failure, removes `.codegraph/` and retries; retry failure is a hard failure.

**Security:** All variable expansions are double-quoted (`"$TARGET_DIR"`, `"$REPO_ROOT"`, `"$output"`) per T-13-02 threat mitigation.

### Makefile changes

- Added `preflight` to `.PHONY` list
- Added `preflight` to `help` target echo block
- Added `preflight` target delegating to `bash bootstrap/pre-flight-checks.sh "$(TARGET)"`

### .gitignore update (deviation)

Added `.codegraph/` to `.gitignore`. The codegraph init/index step during verification generated a `.codegraph/` directory in the worktree root. This is a per-machine generated index and must not be committed (Rule 2 auto-fix — missing critical configuration).

## Verification Results

```
# Against examples/sample-vulnerable-service (exits 0):
OK: codegraph found at /home/saghaulor/.local/bin/codegraph
OK: make/go/claude-security-hooks/go.mod found
OK: git found; TARGET_DIR is a git repository
OK: go.mod found; WARNING: go.sum not found (expected — sample service has no go.sum)
OK: docker available.
OK: codegraph index complete.
pre-flight-checks: PASSED

# Against /nonexistent/path (exits 1):
ERROR: /nonexistent/path/go.mod not found. TARGET_DIR must be a Go module.
pre-flight-checks: FAILED (1+ hard failure(s))

# make preflight TARGET=.../sample-vulnerable-service: exits 0
# make help: shows "preflight" entry
```

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical Config] Added .codegraph/ to .gitignore**
- **Found during:** Post-Task-1 verification
- **Issue:** Running `codegraph index` during verification created `.codegraph/` in the worktree root, leaving an untracked generated directory
- **Fix:** Added `.codegraph/` entry to `.gitignore` with explanatory comment
- **Files modified:** `.gitignore`
- **Commit:** 7b1a7a7

**2. [Rule 1 - Bug] Fixed realpath failure for nonexistent TARGET_DIR**
- **Found during:** Task 1 verification (nonexistent path test)
- **Issue:** `realpath "${1}"` exits non-zero and produces no output for nonexistent paths when `set -uo pipefail` is active, leaving TARGET_DIR empty and producing confusing messages like "ERROR: /go.mod not found"
- **Fix:** Changed to `realpath "${1}" 2>/dev/null || echo "${1}"` — falls back to the raw arg, so error messages show the intended path
- **Files modified:** `bootstrap/pre-flight-checks.sh`
- **Commit:** 13e21f0 (incorporated during Task 1 development)

## Known Stubs

None. All checks are fully wired and functional.

## Threat Flags

None. No new network endpoints, auth paths, or trust boundary changes introduced.

## Self-Check: PASSED

- bootstrap/pre-flight-checks.sh: FOUND and executable
- Makefile preflight target: FOUND (lines 32-35)
- .gitignore .codegraph/ entry: FOUND
- Commit 13e21f0: FOUND (feat(13-02): create bootstrap/pre-flight-checks.sh)
- Commit f94b6cc: FOUND (feat(13-02): add preflight target to root Makefile)
- Commit 7b1a7a7: FOUND (chore(13-02): add .codegraph/ to .gitignore)
