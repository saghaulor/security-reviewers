#!/usr/bin/env bash
# graphify-mcp.sh — Stdio MCP wrapper for graphify.
#
# The target directory is read from .current-review (written by the
# /security-review orchestration command at Step 2). This lets the
# graphify MCP server stay pointed at the currently-reviewed service
# without requiring a session restart.
#
# Falls back to examples/sample-vulnerable-service if the file is absent.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TARGET_FILE="$PROJECT_ROOT/.current-review"

if [ -f "$TARGET_FILE" ]; then
  TARGET_DIR="$(cat "$TARGET_FILE")"
else
  TARGET_DIR="$PROJECT_ROOT/examples/sample-vulnerable-service"
fi

exec graphify "$TARGET_DIR" --mcp
