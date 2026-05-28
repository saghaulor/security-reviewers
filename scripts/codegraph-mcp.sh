#!/usr/bin/env bash
# codegraph-mcp.sh — Stdio MCP wrapper for codegraph.
#
# The target directory is read from .current-review (written by the
# /security-review orchestration command at Step 2). This lets the
# codegraph MCP server stays pointed at the currently-reviewed service
# without requiring a session restart.
#
# Falls back to examples/sample-vulnerable-service if the file is absent.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TARGET_FILE="$PROJECT_ROOT/.current-review"

if [ -f "$TARGET_FILE" ]; then
  TARGET_DIR="$(tr -d '\n' < "$TARGET_FILE")"
else
  TARGET_DIR="$PROJECT_ROOT/examples/sample-vulnerable-service"
fi

if [ -z "$TARGET_DIR" ]; then
  echo "codegraph-mcp.sh: TARGET_DIR is empty after reading $TARGET_FILE" >&2
  exit 1
fi
if [ ! -d "$TARGET_DIR" ]; then
  echo "codegraph-mcp.sh: TARGET_DIR '$TARGET_DIR' is not a directory" >&2
  exit 1
fi

exec codegraph serve --mcp --path "$TARGET_DIR"
