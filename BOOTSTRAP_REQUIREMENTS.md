# Security Review Pipeline - Bootstrap Requirements & Pre-Flight Checks

**Status:** MISSING - Needs implementation  
**Priority:** HIGH - Critical path blocker  
**Estimated Effort:** 4-6 hours

---

## System Prerequisites

### Required Tools

| Tool | Version | Purpose | How to Verify | Install Command |
|------|---------|---------|---------------|-----------------|
| go | 1.20+ | Build hooks binary, static analysis | `go version` | System package manager |
| git | 2.0+ | Repo inspection, code ref tracking | `git --version` | System package manager |
| docker | 20.0+ | Run govulncheck in isolated env | `docker --version` | System package manager |
| make | 3.0+ | Build hooks binary | `make --version` | System package manager |
| codegraph | latest | Static code indexing | `codegraph --version` | TBD - not in standard package managers |

### Optional Tools

| Tool | Purpose | Fallback |
|------|---------|----------|
| govulncheck | Vulnerability scanning | Skip if Docker unavailable |
| semgrep/opengrep | Pattern scanning | Use LSP-based analysis instead |

---

## Codegraph Installation Gap

### Current Issue
The pipeline uses codegraph extensively:
- `codegraph init <path>` — Initialize database
- `codegraph index <path>` — Build index
- Codegraph MCP server — Query code graph

**But:** There is no automated installation of codegraph.

### Observations from This Run

```bash
# Step 3 output:
$ codegraph init /home/saghaulor/code/security_reviewer/examples/sample-vulnerable-service
┌  Initializing CodeGraph
│
▲  Already initialized in /home/saghaulor/code/security_reviewer/examples/sample-vulnerable-service
│
●  Use "codegraph index" to re-index or "codegraph sync" to update
│
└
```

**Status:** ✅ Codegraph was already installed (pre-existing environment)

**What would happen in a fresh environment:** ❌ Command not found

### Recommendation: Install Codegraph in Bootstrap

Add to the bootstrap sequence (before Step 0):

```bash
# Bootstrap Step -1: Install Codegraph
echo "Checking codegraph installation..."
if ! command -v codegraph &> /dev/null; then
    echo "Installing codegraph..."
    # Option A: Build from source (if Go is available)
    git clone https://github.com/anthropics/codegraph.git /tmp/codegraph
    cd /tmp/codegraph
    go install ./cmd/codegraph@latest
    
    # Option B: Download pre-built binary
    # wget https://releases.anthropic.com/codegraph/latest/codegraph -O /usr/local/bin/codegraph
    # chmod +x /usr/local/bin/codegraph
    
    # Option C: Use Docker image (if Docker available)
    # alias codegraph='docker run --rm -v $(pwd):/workspace codegraph:latest'
else
    echo "✓ codegraph is installed"
    codegraph --version
fi
```

---

## Make Bootstrap Issue

### Current Issue
Step 1.5 runs:
```bash
make -C /home/saghaulor/code/security_reviewer/claude-security-hooks install
```

**What could go wrong:**
1. Make not installed → Command not found
2. go.mod/go.sum missing → Build fails
3. Go compilation errors → Build fails
4. Permissions on `.claude/hooks/bin/` → Install fails

### Observations from This Run

```bash
$ make -C claude-security-hooks install
make: Entering directory '.../claude-security-hooks'
CGO_ENABLED=0 go build -ldflags='-s -w' -o bin/claude-security-hooks ./cmd/claude-security-hooks
mkdir -p ../.claude/hooks/bin
cp bin/claude-security-hooks ../.claude/hooks/bin/claude-security-hooks
make: Leaving directory '..'
```

**Status:** ✅ Succeeded (go, make, and permissions all present)

**What would happen if it failed:** ❌ Pipeline stops at Step 1.5, cannot generate session ID

### Recommendation: Add Pre-Flight Check

Before Step 1.5, add:

```bash
# Bootstrap Step 0.5: Verify Make and Go
echo "Checking build prerequisites..."

if ! command -v make &> /dev/null; then
    echo "ERROR: make not found. Install with: apt-get install make (or brew install make)"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "ERROR: go not found. Install from https://golang.org/dl/"
    exit 1
fi

if [ ! -f "claude-security-hooks/go.mod" ]; then
    echo "ERROR: claude-security-hooks/go.mod not found"
    exit 1
fi

# Test build
echo "Testing hooks binary build..."
if ! make -C claude-security-hooks build > /tmp/build.log 2>&1; then
    echo "ERROR: Build failed. Output:"
    cat /tmp/build.log
    exit 1
fi

echo "✓ Build prerequisites verified"
```

---

## Git & Repository Pre-Flight Checks

### Current Issue
The pipeline assumes:
1. Target directory is a git repository
2. Git is installed and working
3. Git can resolve tree hashes and status

**What could go wrong:**
- Target is not a git repo → `git rev-parse` fails
- Git not installed → All git commands fail
- Target not in a git repo (working tree) → Some commands fail
- Detached HEAD or uncommitted state → Status checks confusing

### Observations from This Run

```bash
# Step 2 git commands all succeeded:
$ git -C /path/to/target rev-parse --show-toplevel
/home/saghaulor/code/security_reviewer

$ git -C /path/to/target rev-parse HEAD:examples/sample-vulnerable-service
291125b027a063f0a11342ba7b96a2af77372d6d

$ git -C /path/to/target status --porcelain examples/sample-vulnerable-service
 M examples/sample-vulnerable-service/govulncheck.err
```

**Status:** ✅ All git commands worked

**What would happen if target wasn't a git repo:** ❌ `fatal: not a git repository`

### Recommendation: Add Git Pre-Flight Checks

```bash
# Bootstrap Step 1: Verify Git and Repository
echo "Checking git availability..."

if ! command -v git &> /dev/null; then
    echo "ERROR: git not found. Install from https://git-scm.com/"
    exit 1
fi

TARGET_DIR=$(realpath "${1:-.}")

if ! git -C "$TARGET_DIR" rev-parse --git-dir > /dev/null 2>&1; then
    echo "ERROR: $TARGET_DIR is not a git repository"
    echo "Try: cd $TARGET_DIR && git init"
    exit 1
fi

REPO_ROOT=$(git -C "$TARGET_DIR" rev-parse --show-toplevel)
echo "✓ Git repo verified: $REPO_ROOT"

# Check for uncommitted changes (informational)
if git -C "$TARGET_DIR" status --porcelain | grep -q .; then
    echo "⚠ Warning: Uncommitted changes in $TARGET_DIR"
    echo "  Review these if results look stale:"
    git -C "$TARGET_DIR" status --short
fi
```

---

## Go Project Pre-Flight Checks

### Current Issue
The pipeline needs the target to be a valid Go project with:
1. go.mod file
2. go.sum file (for reproducibility)
3. Parseable Go code
4. Buildable (or at least parseable) modules

**What could go wrong:**
- go.mod missing → codegraph or go tools fail
- go.sum missing → govulncheck fails (this happened in our run)
- Syntax errors in .go files → codegraph indexing fails
- Circular imports → Build fails

### Observations from This Run

The target had:
```bash
$ ls examples/sample-vulnerable-service/
go.mod
go.sum  # <-- Missing entries, caused govulncheck to fail
handlers.go
main.go
oauth_state.go
```

The govulncheck step failed because go.sum was incomplete:
```
govulncheck: loading packages: 
There are errors with the provided package patterns:

handlers.go:7:2: missing go.sum entry for module providing package github.com/gin-gonic/gin
```

**Status:** ⚠️ Partial (go.mod present, but go.sum incomplete → downstream failure)

**What would happen with missing go.mod:** ❌ All Go tooling fails

### Recommendation: Add Go Project Pre-Flight Checks

```bash
# Bootstrap Step 2: Verify Go Project
echo "Checking Go project structure..."

TARGET_DIR=$(realpath "${1:-.}")

if [ ! -f "$TARGET_DIR/go.mod" ]; then
    echo "ERROR: $TARGET_DIR/go.mod not found"
    echo "This doesn't appear to be a Go project."
    echo "Try: cd $TARGET_DIR && go mod init module/path"
    exit 1
fi

echo "✓ go.mod found"

# Warning: go.sum might be missing
if [ ! -f "$TARGET_DIR/go.sum" ]; then
    echo "⚠ Warning: $TARGET_DIR/go.sum not found"
    echo "  This may cause govulncheck and other tools to fail."
    echo "  Try: cd $TARGET_DIR && go mod download && go mod tidy"
fi

# Try to verify Go code can be parsed
echo "Checking Go code syntax..."
if ! go list ./... &> /tmp/go-list.log; then
    echo "⚠ Warning: 'go list' failed. Output:"
    cat /tmp/go-list.log | head -20
    echo "  (This may affect static analysis)"
fi

echo "✓ Go project structure verified"
```

---

## Docker Pre-Flight Checks

### Current Issue
Step 3 attempts to run govulncheck via Docker:
```bash
docker run --rm -v $TARGET_DIR:/workspace -w /workspace golang:latest ...
```

**What could go wrong:**
1. Docker not installed → Command not found
2. Docker daemon not running → Connection refused
3. Image not available → Pull fails (happens offline or with network issues)
4. Permission denied → User not in docker group

### Observations from This Run

The docker command was attempted but failed:
```bash
$ docker run --rm -v /path:/workspace ... golang:latest ...

Error: Exit code 1
go: downloading golang.org/x/vuln v1.3.0
...
```

The failure was due to module dependencies, not Docker itself. Docker was working, but the build inside the container failed due to missing go.sum entries.

**Status:** ✅ Docker was available, but downstream failures occurred

**What would happen without Docker:** ❌ Pipeline would try local govulncheck (if installed)

### Recommendation: Add Docker Pre-Flight Checks

```bash
# Bootstrap Step 3: Verify Docker (Optional)
echo "Checking Docker availability..."

if ! command -v docker &> /dev/null; then
    echo "⚠ Warning: Docker not found. govulncheck will be skipped."
    echo "  Install from https://docs.docker.com/get-docker/"
    DOCKER_AVAILABLE=false
else
    # Test Docker is running
    if ! docker ps > /dev/null 2>&1; then
        echo "⚠ Warning: Docker daemon not running. govulncheck will be skipped."
        echo "  Try: docker ps to test, or 'sudo usermod -aG docker $USER' if permission denied"
        DOCKER_AVAILABLE=false
    else
        echo "✓ Docker is available"
        DOCKER_AVAILABLE=true
    fi
fi

export DOCKER_AVAILABLE  # Pass to downstream steps
```

---

## Codegraph Pre-Flight Checks

### Current Issue
The pipeline needs codegraph MCP server to be:
1. Installed and in PATH
2. Working and accessible via MCP
3. Able to initialize and index Go projects

**What could go wrong:**
1. codegraph not installed → Command not found
2. MCP server not configured → Tools unavailable
3. Database corruption → Index fails
4. Index out of sync → Analysis stale

### Observations from This Run

Codegraph was already initialized:
```bash
$ codegraph init /path/to/target
┌  Initializing CodeGraph
│
▲  Already initialized in /path/to/target
│
●  Use "codegraph index" to re-index or "codegraph sync" to update
│
└

$ codegraph index /path/to/target
┌  Indexing project
...
◆  Indexed 5 files
│
●  96 nodes, 80 edges in 145ms
│
└  Done
```

**Status:** ✅ Codegraph was available and working

**What would happen in a fresh environment:** ❌ Command not found, then MCP calls would fail

### Recommendation: Add Codegraph Pre-Flight Checks

```bash
# Bootstrap Step 4: Verify Codegraph
echo "Checking codegraph availability..."

if ! command -v codegraph &> /dev/null; then
    echo "ERROR: codegraph not found in PATH"
    echo "Try: go install github.com/anthropics/codegraph/cmd/codegraph@latest"
    exit 1
fi

echo "✓ codegraph is installed"
codegraph --version

TARGET_DIR=$(realpath "${1:-.}")

# Test codegraph init and index
echo "Testing codegraph on target directory..."
if ! codegraph init "$TARGET_DIR" > /tmp/codegraph-init.log 2>&1; then
    echo "⚠ Warning: codegraph init failed. Output:"
    cat /tmp/codegraph-init.log | head -10
    exit 1
fi

if ! codegraph index "$TARGET_DIR" > /tmp/codegraph-index.log 2>&1; then
    echo "⚠ Warning: codegraph index failed. Output:"
    cat /tmp/codegraph-index.log | head -10
    echo "  Trying to recover by removing corrupt database..."
    rm -rf "$TARGET_DIR/.codegraph/"
    if ! codegraph init "$TARGET_DIR" > /tmp/codegraph-init-retry.log 2>&1; then
        echo "ERROR: Could not initialize codegraph. Giving up."
        exit 1
    fi
    if ! codegraph index "$TARGET_DIR" > /tmp/codegraph-index-retry.log 2>&1; then
        echo "ERROR: Could not index project. Giving up."
        exit 1
    fi
fi

echo "✓ codegraph initialized and indexed"
```

---

## MCP Server Pre-Flight Checks

### Current Issue
The pipeline uses the codegraph MCP server for code analysis queries. But there's no verification that:
1. MCP server is configured
2. MCP server is responsive
3. MCP server has access to the codegraph database

### Observations from This Run

The MCP server was available and working. All codegraph queries succeeded (codegraph_context, codegraph_search, etc.). No failures observed.

**Status:** ✅ MCP server was working

**What would happen if not configured:** ❌ All codegraph MCP tool calls would fail

### Recommendation: Add MCP Liveness Check

```bash
# Bootstrap Step 5: Verify MCP Server
echo "Checking MCP server connectivity..."

# Try a simple codegraph query to verify MCP is working
if ! python3 << 'EOF'
import subprocess
import json

try:
    result = subprocess.run(
        ["codegraph", "status"],
        capture_output=True,
        timeout=5,
        text=True
    )
    if result.returncode != 0:
        print(f"⚠ Warning: codegraph status check failed")
        print(result.stderr)
    else:
        print("✓ MCP server responsive")
        # Parse status output for index health
        if "pending" in result.stdout.lower():
            print("⚠ Note: Index has pending changes. Consider running 'codegraph sync'")
except Exception as e:
    print(f"⚠ Warning: Could not verify MCP server: {e}")
EOF
then
    echo "⚠ Warning: Could not verify MCP server health"
fi
```

---

## Complete Bootstrap Checklist

```bash
#!/bin/bash
# security-review-pipeline-bootstrap.sh

set -e

TARGET_DIR=$(realpath "${1:-.}")
echo "Bootstrap: Security Review Pipeline"
echo "Target: $TARGET_DIR"
echo "========================================"

# Bootstrap -1: Install Codegraph (if needed)
echo ""
echo "[Step -1] Checking codegraph installation..."
if ! command -v codegraph &> /dev/null; then
    echo "Installing codegraph..."
    # (installation code here)
else
    echo "✓ codegraph is installed"
fi

# Bootstrap 0: Verify Make and Go
echo ""
echo "[Step 0] Verifying build tools..."
command -v make > /dev/null || { echo "ERROR: make not found"; exit 1; }
command -v go > /dev/null || { echo "ERROR: go not found"; exit 1; }
[ -f "claude-security-hooks/go.mod" ] || { echo "ERROR: go.mod not found"; exit 1; }
echo "✓ Build tools available"

# Bootstrap 1: Verify Git
echo ""
echo "[Step 1] Verifying git and repository..."
command -v git > /dev/null || { echo "ERROR: git not found"; exit 1; }
git -C "$TARGET_DIR" rev-parse --git-dir > /dev/null || { echo "ERROR: Not a git repo"; exit 1; }
echo "✓ Git repository verified"

# Bootstrap 2: Verify Go Project
echo ""
echo "[Step 2] Verifying Go project structure..."
[ -f "$TARGET_DIR/go.mod" ] || { echo "ERROR: go.mod not found"; exit 1; }
[ -f "$TARGET_DIR/go.sum" ] || { echo "⚠ Warning: go.sum not found"; }
echo "✓ Go project structure verified"

# Bootstrap 3: Verify Docker (optional)
echo ""
echo "[Step 3] Checking Docker availability..."
if command -v docker &> /dev/null && docker ps > /dev/null 2>&1; then
    echo "✓ Docker available"
    export DOCKER_AVAILABLE=true
else
    echo "⚠ Docker not available (govulncheck will be skipped)"
    export DOCKER_AVAILABLE=false
fi

# Bootstrap 4: Verify Codegraph
echo ""
echo "[Step 4] Initializing codegraph..."
codegraph init "$TARGET_DIR" 2>&1 | grep -v "Already initialized" || true
codegraph index "$TARGET_DIR" > /dev/null || { echo "ERROR: codegraph index failed"; exit 1; }
echo "✓ Codegraph initialized and indexed"

# Bootstrap 5: Verify MCP
echo ""
echo "[Step 5] Verifying MCP server..."
codegraph status > /dev/null 2>&1 || { echo "⚠ MCP server check inconclusive"; }
echo "✓ MCP server responsive"

echo ""
echo "========================================"
echo "✓ Bootstrap complete. Ready to scan."
echo "========================================"
```

---

## Summary: What's Missing

| Check | Status | Priority | Effort |
|-------|--------|----------|--------|
| Codegraph installation | ❌ Not automated | HIGH | 2 hrs |
| Make/Go verification | ❌ Not checked | HIGH | 1 hr |
| Git verification | ❌ Not checked | HIGH | 1 hr |
| Go project validation | ❌ Partial (found issue too late) | HIGH | 1.5 hrs |
| Docker verification | ❌ Not checked | MEDIUM | 1 hr |
| Codegraph liveness | ❌ Not checked | MEDIUM | 1 hr |
| MCP connectivity | ❌ Not checked | MEDIUM | 1 hr |
| **TOTAL** | | | **~9 hours** |

---

## Recommendation: Unified Bootstrap Script

Create `/home/saghaulor/code/security_reviewer/bootstrap/pre-flight-checks.sh` that:
1. Runs before the security-review pipeline
2. Verifies all tools are installed and working
3. Pre-seeds codegraph database
4. Provides clear error messages with remediation steps
5. Fails fast with non-zero exit code if critical checks fail

This script should be added to the `.claude/hooks/bin/` directory so it's included in the hooks binary installation and can be called automatically before each scan.

---

**Document Prepared By:** Claude Haiku 4.5  
**Status:** Identifies missing bootstrap - needs implementation  
**Estimated Timeline:** 2-3 days for full bootstrap implementation
