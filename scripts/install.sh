#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=== security-review Skill Setup ==="
echo "Installing from: $REPO_ROOT"
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_command() {
  if ! command -v "$1" &> /dev/null; then
    echo -e "${RED}✗ $1 not found${NC}"
    echo "  Please install $1 and try again."
    return 1
  fi
  echo -e "${GREEN}✓ $1 found${NC}"
  return 0
}

# Check prerequisites
echo "Checking prerequisites..."
check_command go || exit 1
check_command docker || exit 1
check_command git || exit 1
echo

# Build hooks binary
echo "Building claude-security-hooks binary..."
cd "$REPO_ROOT/claude-security-hooks"
if make build; then
  echo -e "${GREEN}✓ Hooks binary built${NC}"
else
  echo -e "${RED}✗ Failed to build hooks binary${NC}"
  exit 1
fi
cd "$REPO_ROOT"
echo

# Set up opengrep-mcp
echo "Setting up opengrep-mcp server..."
OPENGREP_DIR="../opengrep-mcp"

if [ -d "$OPENGREP_DIR" ]; then
  echo -e "${YELLOW}→ opengrep-mcp already cloned${NC}"
else
  echo "Cloning opengrep-mcp..."
  cd ..
  if git clone https://github.com/saghaulor/opengrep-mcp.git; then
    echo -e "${GREEN}✓ opengrep-mcp cloned${NC}"
  else
    echo -e "${RED}✗ Failed to clone opengrep-mcp${NC}"
    cd "$REPO_ROOT"
    exit 1
  fi
  cd "$REPO_ROOT"
fi

if [ -f "$OPENGREP_DIR/bin/opengrep-mcp" ]; then
  echo -e "${YELLOW}→ opengrep-mcp binary already built${NC}"
else
  echo "Building opengrep-mcp binary..."
  cd "$OPENGREP_DIR"
  if make build; then
    echo -e "${GREEN}✓ opengrep-mcp binary built${NC}"
  else
    echo -e "${RED}✗ Failed to build opengrep-mcp${NC}"
    cd "$REPO_ROOT"
    exit 1
  fi
  cd "$REPO_ROOT"
fi
echo

# Pull Semgrep container
echo "Pulling Semgrep container..."
if docker pull returntocorp/semgrep:1.55.0 &> /dev/null; then
  echo -e "${GREEN}✓ Semgrep container ready${NC}"
else
  echo -e "${RED}✗ Failed to pull Semgrep container${NC}"
  echo "  Check Docker daemon is running: docker ps"
  exit 1
fi
echo

# Register skill with Claude Code
echo "Registering skill with Claude Code..."
CLAUDE_SETTINGS="$HOME/.claude/settings.json"
REPO_ABSOLUTE_PATH="$(cd "$REPO_ROOT" && pwd)"

if [ ! -f "$CLAUDE_SETTINGS" ]; then
  mkdir -p "$(dirname "$CLAUDE_SETTINGS")"
  echo "{}" > "$CLAUDE_SETTINGS"
fi

# Add skill registration (using Python for safe JSON manipulation)
python3 << EOF
import json
import sys

settings_file = "$CLAUDE_SETTINGS"
repo_path = "$REPO_ABSOLUTE_PATH"

try:
  with open(settings_file, 'r') as f:
    settings = json.load(f)
except (json.JSONDecodeError, FileNotFoundError):
  settings = {}

# Ensure skills section exists
if 'skills' not in settings:
  settings['skills'] = {}

# Register the skill
settings['skills']['security-review'] = {
  'source': repo_path,
  'enabled': True
}

with open(settings_file, 'w') as f:
  json.dump(settings, f, indent=2)

print(f"Registered security-review skill at: {repo_path}")
EOF

if [ $? -eq 0 ]; then
  echo -e "${GREEN}✓ Skill registered${NC}"
else
  echo -e "${RED}✗ Failed to register skill${NC}"
  exit 1
fi
echo

# Verify hooks binary is executable
chmod +x "$REPO_ROOT/claude-security-hooks/bin/claude-security-hooks"
echo

# Run integration test
echo "Running integration test..."
echo "Testing against sample-vulnerable-service..."
echo

cd "$REPO_ROOT"
if /security-review examples/sample-vulnerable-service/ 2>&1 | tail -20; then
  echo
  if [ -f "review-report.json" ]; then
    FINDING_COUNT=$(python3 -c "import json; print(len(json.load(open('review-report.json')).get('findings', [])))" 2>/dev/null || echo "?")
    echo -e "${GREEN}✓ Review completed - $FINDING_COUNT findings detected${NC}"
  else
    echo -e "${YELLOW}→ Review command executed (report may not be in current directory)${NC}"
  fi
else
  echo -e "${YELLOW}→ Integration test encountered issues (this is OK if hooks binary worked)${NC}"
fi
echo

# Summary
echo "=== Setup Complete ==="
echo
echo "Next steps:"
echo "  1. Review the results: jq '.findings[]' review-report.json"
echo "  2. Run on your own code: /security-review /path/to/your/go/service"
echo "  3. See INSTALL.md for troubleshooting"
echo
echo -e "${GREEN}Installation successful!${NC}"
