#!/usr/bin/env bash
# bootstrap/pre-flight-checks.sh
#
# Run 6 sequential environment checks before executing the security review pipeline.
# Usage: bash bootstrap/pre-flight-checks.sh [TARGET_DIR]
#   TARGET_DIR defaults to "." if not provided.
#
# Exits 0 if all hard checks pass (warnings allowed).
# Exits 1 if any hard check fails.
#
# Environment exported:
#   DOCKER_AVAILABLE=true|false

set -uo pipefail

# ---------------------------------------------------------------------------
# Resolve paths
# ---------------------------------------------------------------------------
if [ -n "${1:-}" ]; then
  TARGET_DIR="$(realpath "${1}" 2>/dev/null || echo "${1}")"
else
  TARGET_DIR="$(realpath . 2>/dev/null || echo ".")"
fi

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo "$PWD")"

FAILURES=0

echo "pre-flight-checks: TARGET_DIR=$TARGET_DIR"
echo "pre-flight-checks: REPO_ROOT=$REPO_ROOT"
echo ""

# ---------------------------------------------------------------------------
# Check 1 — Codegraph
# ---------------------------------------------------------------------------
echo "==> Check 1: codegraph"
CODEGRAPH_OK=false
if command -v codegraph > /dev/null 2>&1; then
  CODEGRAPH_OK=true
  echo "OK: codegraph found at $(which codegraph)"
else
  echo "ERROR: codegraph not found. Install with: go install github.com/saghaulor/codegraph/cmd/codegraph@latest" >&2
  FAILURES=$((FAILURES+1))
fi
echo ""

# ---------------------------------------------------------------------------
# Check 2 — Make + Go + hooks module
# ---------------------------------------------------------------------------
echo "==> Check 2: make + go + hooks module"
if command -v make > /dev/null 2>&1; then
  echo "OK: make found at $(which make)"
else
  echo "ERROR: make not found. Install build-essential or equivalent." >&2
  FAILURES=$((FAILURES+1))
fi

if command -v go > /dev/null 2>&1; then
  echo "OK: go found at $(which go)"
else
  echo "ERROR: go not found. Install from https://golang.org/dl/" >&2
  FAILURES=$((FAILURES+1))
fi

if [ -f "$REPO_ROOT/claude-security-hooks/go.mod" ]; then
  echo "OK: claude-security-hooks/go.mod found"
else
  echo "ERROR: claude-security-hooks/go.mod not found. Are you running from the security_reviewer repo root?" >&2
  FAILURES=$((FAILURES+1))
fi
echo ""

# ---------------------------------------------------------------------------
# Check 3 — Git + target repo
# ---------------------------------------------------------------------------
echo "==> Check 3: git + target repo"
if command -v git > /dev/null 2>&1; then
  echo "OK: git found at $(which git)"
else
  echo "ERROR: git not found." >&2
  FAILURES=$((FAILURES+1))
fi

if git -C "$TARGET_DIR" rev-parse --git-dir > /dev/null 2>&1; then
  echo "OK: $TARGET_DIR is a git repository"
else
  echo "WARNING: $TARGET_DIR is not a git repository. code_ref will be empty." >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Check 4 — Go project structure in TARGET_DIR
# ---------------------------------------------------------------------------
echo "==> Check 4: Go project structure in $TARGET_DIR"
if [ -f "$TARGET_DIR/go.mod" ]; then
  echo "OK: $TARGET_DIR/go.mod found"
else
  echo "ERROR: $TARGET_DIR/go.mod not found. TARGET_DIR must be a Go module." >&2
  FAILURES=$((FAILURES+1))
fi

if [ -f "$TARGET_DIR/go.sum" ]; then
  echo "OK: $TARGET_DIR/go.sum found"
else
  echo "WARNING: $TARGET_DIR/go.sum not found. Run: go mod tidy" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Check 5 — Docker
# ---------------------------------------------------------------------------
echo "==> Check 5: docker"
if command -v docker > /dev/null 2>&1; then
  if docker ps > /dev/null 2>&1; then
    export DOCKER_AVAILABLE=true
    echo "OK: docker available."
  else
    export DOCKER_AVAILABLE=false
    echo "WARNING: docker daemon not running. govulncheck will use fallback mode." >&2
  fi
else
  export DOCKER_AVAILABLE=false
  echo "WARNING: docker not found. govulncheck will use fallback mode." >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Check 6 — Codegraph init + index with retry
# ---------------------------------------------------------------------------
echo "==> Check 6: codegraph init + index"
if [ "$CODEGRAPH_OK" = "true" ]; then
  output="$(codegraph init "$TARGET_DIR" 2>&1)"
  if [ $? -ne 0 ]; then
    echo "WARNING: codegraph init failed: $output" >&2
  fi

  output="$(codegraph index "$TARGET_DIR" 2>&1)"
  if [ $? -ne 0 ]; then
    echo "WARNING: codegraph index failed. Attempting retry after removing .codegraph/" >&2
    rm -rf "$TARGET_DIR/.codegraph/"

    output="$(codegraph init "$TARGET_DIR" 2>&1)"
    if [ $? -ne 0 ]; then
      echo "WARNING: codegraph init (retry) failed: $output" >&2
    fi

    output="$(codegraph index "$TARGET_DIR" 2>&1)"
    if [ $? -ne 0 ]; then
      echo "ERROR: codegraph index retry failed. Pipeline cannot proceed." >&2
      FAILURES=$((FAILURES+1))
    else
      echo "OK: codegraph index complete (after retry)."
    fi
  else
    echo "OK: codegraph index complete."
  fi
else
  echo "SKIP: codegraph check (Check 1) failed — skipping init/index." >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
if [ "$FAILURES" -gt 0 ]; then
  echo "pre-flight-checks: FAILED ($FAILURES hard failure(s))" >&2
  exit 1
else
  echo "pre-flight-checks: PASSED"
  exit 0
fi
