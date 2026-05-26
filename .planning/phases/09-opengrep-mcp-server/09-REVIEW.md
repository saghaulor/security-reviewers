---
phase: 09-opengrep-mcp-server
reviewed: 2026-05-26T13:30:59Z
depth: standard
files_reviewed: 3
files_reviewed_list:
  - Makefile
  - .claude/commands/security-review.md
  - .mcp.json
findings:
  critical: 2
  warning: 4
  info: 2
  total: 8
status: issues_found
---

# Phase 9: Code Review Report

**Reviewed:** 2026-05-26T13:30:59Z
**Depth:** standard
**Files Reviewed:** 3
**Status:** issues_found

## Summary

Reviewed three files delivering the Phase 9 opengrep-mcp integration: the top-level `Makefile`, the `/security-review` orchestration command, and the `.mcp.json` MCP server configuration. Two critical defects were found: a literal-string bug that breaks the graphify MCP server pointer on every run, and hardcoded absolute paths in two files that prevent any other developer from using the tool without manual edits. Four additional warnings cover a loose Docker image tag (supply-chain risk in a security tool), missing bash safety flags, a relative-path fragility in the install script, and a non-portable static-link check in the Makefile.

---

## Critical Issues

### CR-01: `security-review.md` writes the literal string "TARGET_DIR" to `.current-review` instead of the variable value

**File:** `.claude/commands/security-review.md:25`

**Issue:** Step 2 instructs the agent to run:
```
Bash: echo "TARGET_DIR" > /home/saghaulor/code/security_reviewer/.current-review
```
The token `TARGET_DIR` is quoted and not prefixed with `$`, so the literal four-word string `TARGET_DIR` is written to the file instead of the resolved path. `graphify-mcp.sh` reads this file (line 16) and passes the value verbatim to `exec graphify "$TARGET_DIR" --mcp`. The MCP server will therefore always attempt to load the graph for a directory named literally `TARGET_DIR`, which does not exist. Every subsequent MCP tool call will fail or silently operate on the wrong target for the entire session.

**Fix:**
```markdown
Bash: echo "$TARGET_DIR" > /home/saghaulor/code/security_reviewer/.current-review
```
The variable must be expanded — replace `"TARGET_DIR"` with `"$TARGET_DIR"`.

---

### CR-02: Hardcoded absolute paths to a single user's home directory appear in both `.mcp.json` and `Makefile`

**File:** `.mcp.json:5,9` and `Makefile:28,29`

**Issue:** `.mcp.json` embeds `/home/saghaulor/code/security_reviewer/...` for both MCP server executables. The `Makefile` hardcodes `/home/saghaulor/code/opengrep-mcp` as the source tree for `build-opengrep-mcp`. These paths are machine-specific. Any developer, CI runner, or container that checks out this repo to a different location will get immediate failures:
- Claude Code will refuse to start the `graphify` and `opengrep` MCP servers (file not found).
- `make build-opengrep-mcp` will fail because the sibling repo is expected at a hardcoded absolute path.

The SKILL.md installation guide instructs users to clone to an arbitrary directory (`git clone ... && cd security-reviewers`), making the hardcoded paths directly contradictory to the documented install flow.

**Fix for `.mcp.json`** — use paths relative to the project root. Claude Code supports project-relative paths prefixed with `./`:
```json
{
  "mcpServers": {
    "graphify": {
      "type": "stdio",
      "command": "./scripts/graphify-mcp.sh"
    },
    "opengrep": {
      "type": "stdio",
      "command": "./.claude/hooks/bin/opengrep-mcp",
      "env": {
        "SAST_ENGINE": "opengrep"
      }
    }
  }
}
```

**Fix for `Makefile`** — parameterise the sibling repo path:
```makefile
OPENGREP_SRC ?= $(abspath ../opengrep-mcp)

build-opengrep-mcp:
	@mkdir -p .claude/hooks/bin
	$(MAKE) -C $(OPENGREP_SRC) build
	@cp $(OPENGREP_SRC)/bin/opengrep-mcp .claude/hooks/bin/opengrep-mcp
```
This allows `make build-opengrep-mcp OPENGREP_SRC=/custom/path` overrides while defaulting to the expected sibling-repo layout.

---

## Warnings

### WR-01: `golang:latest` Docker image is unpinned — supply-chain risk in a security tool

**File:** `.claude/commands/security-review.md:48`

**Issue:** The govulncheck step pulls and runs `golang:latest`:
```
docker run --rm -v TARGET_DIR:/workspace -w /workspace golang:latest \
  sh -c "go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck -json ./... ..."
```
Both the Go base image and `govulncheck` itself are resolved at `@latest` with no digest pin. In a security scanning tool this is especially problematic: a compromised or malicious update to either artifact could silently alter scan results or exfiltrate source code. The Docker mount gives the container full read access to `TARGET_DIR`.

**Fix:** Pin both the base image and the tool to known-good versions:
```
docker run --rm -v "$TARGET_DIR":/workspace -w /workspace golang:1.23.4 \
  sh -c "go install golang.org/x/vuln/cmd/govulncheck@v1.1.3 && govulncheck -json ./..."
```
Use a digest pin (`golang:1.23.4@sha256:...`) for the strongest guarantee.

---

### WR-02: `graphify-mcp.sh` is missing `set -eu` — silent failures can serve wrong target

**File:** `scripts/graphify-mcp.sh` (no line number — missing from top of file)

**Issue:** The script has no `set -e` or `set -u`. If `cat "$TARGET_FILE"` returns an empty string (e.g., the file exists but is empty — which will happen when CR-01 is not yet fixed since `echo ""` truncates the file), `TARGET_DIR` becomes an empty string and `exec graphify "" --mcp` is silently passed an invalid argument. With `set -eu`, an unset or empty variable would abort with a clear error message rather than passing garbage to `graphify`.

Additionally, the TARGET_DIR value read from `.current-review` is never validated or sanitized before being passed to `exec`. A crafted value containing shell metacharacters or an absolute path like `/etc` would be passed directly.

**Fix:**
```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TARGET_FILE="$PROJECT_ROOT/.current-review"

if [ -f "$TARGET_FILE" ]; then
  TARGET_DIR="$(cat "$TARGET_FILE")"
  # Guard against empty content
  if [ -z "$TARGET_DIR" ]; then
    echo "graphify-mcp: .current-review is empty, falling back to sample service" >&2
    TARGET_DIR="$PROJECT_ROOT/examples/sample-vulnerable-service"
  fi
else
  TARGET_DIR="$PROJECT_ROOT/examples/sample-vulnerable-service"
fi

exec graphify "$TARGET_DIR" --mcp
```

---

### WR-03: `install.sh` uses a relative path for `OPENGREP_DIR` that breaks when the script is run from outside the repo root

**File:** `scripts/install.sh:48`

**Issue:** `OPENGREP_DIR` is set to `"../opengrep-mcp"` — a path relative to the current working directory at the time of evaluation, not relative to `$REPO_ROOT`. The script does `cd "$REPO_ROOT"` at line 36 and 43, so by the time line 65 evaluates `[ -f "$OPENGREP_DIR/bin/opengrep-mcp" ]`, the CWD is `$REPO_ROOT` and the relative path resolves correctly. However, if any intermediate `cd` call in the else-branch at line 54 (`cd ..`) is followed by a `git clone` failure and the error path does `cd "$REPO_ROOT"` (line 59) fails for some reason, the subsequent `[ -f "$OPENGREP_DIR/bin/opengrep-mcp" ]` on line 65 evaluates the relative path from an unexpected directory.

More directly: if a user runs `bash scripts/install.sh` from a directory other than the repo root (e.g., `bash /abs/path/scripts/install.sh`), the `OPENGREP_DIR` assignment evaluates before any `cd`, so `../opengrep-mcp` resolves relative to the caller's CWD, not the repo root.

**Fix:** Derive `OPENGREP_DIR` from the already-computed `REPO_ROOT`:
```bash
OPENGREP_DIR="$(dirname "$REPO_ROOT")/opengrep-mcp"
```

---

### WR-04: Static-link check in `verify-opengrep-mcp` uses `readelf` which is Linux-only

**File:** `Makefile:22`

**Issue:** The verify target runs `readelf -d $(OPENGREP_BIN)` to confirm no dynamic section, falling back to `ldd ... | grep -q "not a dynamic executable"`. On macOS, `readelf` is not installed by default (it requires binutils or llvm), so the first command exits non-zero and `grep -q "no dynamic section"` gets no input, failing. The logic then falls through to the `ldd` check — but macOS `ldd` does not exist either (it uses `otool -L`). On macOS, the verify target will therefore always print "FAIL: not statically linked" even for correct binaries.

The SKILL.md prerequisites do not mention Linux as a requirement, implying macOS developers are expected to be able to use this tool.

**Fix:** Add a platform guard or use a portable check. For Go binaries, `go tool nm` provides a portable alternative:
```makefile
verify-opengrep-mcp:
	@test -f $(OPENGREP_BIN) || (echo "FAIL: $(OPENGREP_BIN) does not exist. Run: make build-opengrep-mcp" && exit 1)
	@test -x $(OPENGREP_BIN) || (echo "FAIL: $(OPENGREP_BIN) is not executable" && exit 1)
	@if command -v readelf >/dev/null 2>&1; then \
	  readelf -d $(OPENGREP_BIN) 2>&1 | grep -q "no dynamic section" || \
	    (echo "FAIL: $(OPENGREP_BIN) is not statically linked" && exit 1); \
	elif command -v otool >/dev/null 2>&1; then \
	  otool -L $(OPENGREP_BIN) 2>&1 | grep -q "not a dynamic executable" || \
	    (echo "FAIL: $(OPENGREP_BIN) is not statically linked" && exit 1); \
	else \
	  echo "WARN: cannot verify static linking (no readelf or otool found)"; \
	fi
	@echo "OK: opengrep-mcp binary verified"
```

---

## Info

### IN-01: `security-review.md` Step 3 Docker command contains un-substituted `TARGET_DIR` placeholder in a code block

**File:** `.claude/commands/security-review.md:48`

**Issue:** The govulncheck Docker command in the Step 3 code block still reads `TARGET_DIR` (unsubstituted placeholder) in the `-v` volume mount flag:
```
docker run --rm -v TARGET_DIR:/workspace ...
```
The command text says "Replace TARGET_DIR with the actual resolved path" only at the end of Step 2 (line 36), not Step 3. An LLM agent following these instructions literally could mount the literal string `TARGET_DIR` as a Docker volume rather than the resolved path. This is the same substitution-instruction pattern that caused CR-01, making it a systemic risk. All code blocks in the document should use the `$TARGET_DIR` shell variable form, not the placeholder form.

**Fix:** Replace `TARGET_DIR` with `"$TARGET_DIR"` in all Bash code-block examples in the document, removing the separate "Replace TARGET_DIR" prose instructions. Consistent variable syntax throughout prevents literal-string substitution errors.

---

### IN-02: `install.sh` calls `/security-review` as if it were a shell command at line 150

**File:** `scripts/install.sh:150`

**Issue:** The integration test at the end of `install.sh` runs:
```bash
if /security-review examples/sample-vulnerable-service/ 2>&1 | tail -20; then
```
`/security-review` is a Claude Code slash command, not a filesystem executable at `/security-review`. This will always exit with "No such file or directory" on any system. The integration test therefore never actually validates the skill and always falls through to the yellow "this is OK" message, meaning install failures go undetected.

**Fix:** Either remove the integration test from the install script entirely (document it as a manual post-install step using the Claude Code CLI), or invoke the Claude Code CLI correctly:
```bash
claude --command security-review examples/sample-vulnerable-service/
```

---

_Reviewed: 2026-05-26T13:30:59Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
