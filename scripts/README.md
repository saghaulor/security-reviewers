# Scripts

Automation scripts for the security-review skill.

## install.sh

Automated setup script for new machines.

**Usage:**

```bash
./scripts/install.sh
```

**What it does:**

1. Verifies prerequisites (Go, Docker, Git)
2. Builds the `claude-security-hooks` binary
3. Clones and builds `opengrep-mcp` (if needed)
4. Pulls the Semgrep Docker image
5. Registers the skill with Claude Code
6. Runs integration test against sample-vulnerable-service

**Prerequisite check:**

```bash
go version   # 1.19+
docker ps    # Docker running
git --version
```

**See also:**

- `../INSTALL.md` — Detailed manual installation steps
- `../.claude/skills/security-review/SKILL.md` — Skill documentation
