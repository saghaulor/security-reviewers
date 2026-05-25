# Installation Guide: security-review Skill

Complete setup instructions for installing the security-review skill on a new machine.

## Prerequisites

Verify you have the following installed:

```bash
go version          # Should be 1.19 or later
docker --version    # Docker must be running
git --version       # For cloning repos
```

If any are missing, install them:

- **Go**: https://golang.org/doc/install
- **Docker**: https://docs.docker.com/get-docker/
- **Git**: https://git-scm.com/downloads
- **Claude Code CLI**: https://claude.com/claude-code (instructions in setup wizard)

## Step 1: Clone the Repository

```bash
git clone git@github.com:saghaulor/security-reviewers.git
cd security-reviewers
```

If you don't have GitHub SSH keys set up, use HTTPS instead:

```bash
git clone https://github.com/saghaulor/security-reviewers.git
cd security-reviewers
```

## Step 2: Build the Hooks Binary

The hooks binary validates agent outputs and enforces security invariants.

```bash
cd claude-security-hooks
make build
cd ..
```

Verify it was built:

```bash
ls -lh claude-security-hooks/bin/claude-security-hooks
ldd claude-security-hooks/bin/claude-security-hooks    # Should show "not a dynamic executable"
```

## Step 3: Set Up OpenGrep MCP Server

The OpenGrep server runs pattern scanning via Semgrep in Docker containers.

### Option A: Automated Setup (Recommended)

The setup script handles this:

```bash
./scripts/install.sh
```

This clones `opengrep-mcp` into a sibling directory and starts the server.

### Option B: Manual Setup

Clone and build the opengrep-mcp repo:

```bash
cd ..
git clone https://github.com/saghaulor/opengrep-mcp.git
cd opengrep-mcp
make build
cd ../security-reviewers
```

Verify the binary was built:

```bash
ls -lh ../opengrep-mcp/bin/opengrep-mcp
```

### Option C: Use Existing Server

If you already have `opengrep-mcp` running elsewhere, ensure it's accessible at `http://localhost:8000`:

```bash
curl http://localhost:8000/health
```

## Step 4: Register the Skill with Claude Code

Claude Code must know about this skill. Register it in your `.claude/settings.json`:

**Option A: Automated (via setup script)**

```bash
./scripts/install.sh
```

**Option B: Manual registration**

Locate your Claude Code settings directory:

```bash
# On macOS/Linux:
~/.claude/

# On Windows:
%APPDATA%\.claude\
```

Add this to `~/.claude/settings.json` (or create the file if it doesn't exist):

```json
{
  "skills": {
    "security-review": {
      "source": "/absolute/path/to/security-reviewers",
      "enabled": true
    }
  }
}
```

Replace `/absolute/path/to/security-reviewers` with the actual path where you cloned the repo.

## Step 5: Verify the Setup

Test that everything is wired correctly:

### 5.1: Test the hooks binary

```bash
echo '{"tool": "Agent", "subagent_type": "go-cartographer"}' | \
  ./claude-security-hooks/bin/claude-security-hooks preflight
```

Expected output: No output (exit code 0).

### 5.2: Test Docker & Semgrep

```bash
docker pull returntocorp/semgrep:1.55.0
docker images | grep semgrep
```

### 5.3: Run the full integration test

From the `security-reviewers` directory, try the test service:

```bash
/security-review examples/sample-vulnerable-service/
```

Wait for completion (~60 seconds). You should see:

- A `review-report.json` file created
- A `review-report.md` file created
- At least 3 findings (SQL injection, authz bypass, OAuth scope tampering)

Check results:

```bash
jq '.findings[] | .title' review-report.json
```

Expected output:

```
"SQL injection in user lookup"
"Authorization bypass in admin endpoints"
"OAuth scope tampering in token exchange"
```

## Step 6: Configure Permissions (Optional)

To reduce permission prompts, add these to `.claude/settings.json`:

```json
{
  "permissions": {
    "allow": [
      "Bash(go *)",
      "Bash(docker *)",
      "Bash(make *)"
    ]
  }
}
```

## Troubleshooting

### "hooks binary not found"

```bash
cd claude-security-hooks && make build && cd ..
```

### "Docker daemon is not running"

Start Docker:

```bash
# macOS (if using Docker Desktop):
open /Applications/Docker.app

# Linux:
sudo systemctl start docker
sudo usermod -aG docker $USER   # Add user to docker group
```

### "opengrep-mcp not responding"

Check if the server is running:

```bash
curl http://localhost:8000/health
```

If not, start it:

```bash
cd ../opengrep-mcp
./bin/opengrep-mcp --port 8000 &
cd ../security-reviewers
```

### "Semgrep container fails to pull"

Ensure Docker can reach Docker Hub:

```bash
docker pull returntocorp/semgrep:1.55.0
```

If it fails, check your network/firewall settings.

### "/security-review command not found"

Ensure Claude Code recognizes the skill:

1. Check that `~/.claude/settings.json` points to the correct path
2. Restart Claude Code (close and reopen)
3. Try again

### "Permission denied" on hooks binary

```bash
chmod +x claude-security-hooks/bin/claude-security-hooks
```

## Next Steps

Once verified, you can:

1. **Review your own Go service:**
   ```bash
   /security-review /path/to/your/go/service
   ```

2. **Parse the results:**
   ```bash
   # List all high-confidence findings
   jq '.findings[] | select(.confidence == "high")' review-report.json
   
   # Get a summary
   jq '.statistics' review-report.json
   ```

3. **View the markdown report:**
   ```bash
   cat review-report.md
   ```

## Support

- **Issues**: https://github.com/saghaulor/security-reviewers/issues
- **Repo**: https://github.com/saghaulor/security-reviewers
