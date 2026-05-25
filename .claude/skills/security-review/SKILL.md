---
name: security-review
description: Systematic security review of Go services using multi-agent analysis (taint tracing, authorization verification, OAuth auditing, invariant checking)
version: 1.0.0
author: Stephen Aghaulor
keywords: [security, go, vulnerability, taint-analysis, authorization, oauth, code-review]
---

# security-review Skill

A Claude Code plugin for systematic security review of Go services. Dispatches six specialized Claude agents to detect taint flows (SQL injection, command injection, SSRF, path traversal, XXE, unsafe deserialization, template injection), authorization bypasses, OAuth vulnerabilities, and business-logic invariant violations.

## Quick Start

Once installed, use the slash command:

```bash
/security-review /path/to/go/service
```

This produces a `review-report.json` with structured findings:

```json
{
  "findings": [
    {
      "class": "injection|authz|oauth",
      "confidence": "high|medium|low",
      "title": "SQL injection in user lookup",
      "evidence": { ... }
    }
  ]
}
```

Parse with jq:

```bash
jq '.findings[] | select(.confidence == "high") | .title' review-report.json
```

## How It Works

The skill runs through four stages:

1. **Pre-Pass** (one-time setup)
   - Analyze known CVEs with `govulncheck`
   - Build structural code graph with `graphify`

2. **Cartographer** (sequential)
   - Index HTTP handlers, RPC methods (entrypoints)
   - Map SQL queries, exec calls, OAuth endpoints (sinks)
   - Discover middleware chains and authz boundaries
   - Output: `go-index.json`

3. **Parallel Fan-Out** (six agents in parallel)
   - **go-taint-tracer**: Source→sink taint propagation for injections
   - **go-authz-tracer**: Middleware chain verification for IDOR/bypass
   - **go-oauth-auditor**: RFC 9700 / OIDC Core conformance
   - **invariant-checker**: Business-logic assertions (e.g., payment flows)
   - (Two additional specialized tracers for advanced patterns)

4. **Synthesis** (aggregation)
   - Deduplicates findings
   - Ranks by severity and confidence
   - Outputs `review-report.json` + `review-report.md`

**Key insight:** The cartographer runs once and its structural index is reused by all downstream agents. This avoids redundant codebase analysis while parallelizing tracer work.

## Installation

### Prerequisites

- **Claude Code** CLI installed (latest version)
- **Docker** running locally (for Semgrep container)
- **Go** 1.19+ (to build the hooks binary)
- **git** (to clone the repo)

### 1-Minute Setup

```bash
# Clone the repo
git clone git@github.com:saghaulor/security-reviewers.git
cd security-reviewers

# Run the automated setup
./scripts/install.sh

# Verify installation
/security-review examples/sample-vulnerable-service/
```

The setup script will:
1. Build the `claude-security-hooks` binary
2. Clone and build the `opengrep-mcp` server (if not present)
3. Register the skill in Claude Code
4. Test the integration

### Manual Setup (if automated script fails)

See `INSTALL.md` for detailed step-by-step instructions.

## Security Model

- **Hooks binary** (`claude-security-hooks`): Per-agent output validation
  - Checks all tracer verdicts against 51 assertions
  - Blocks malformed JSON
  - Injects context at agent startup
  - Built with `CGO_ENABLED=0` (no C dependencies, portable)

- **OpenGrep MCP server** (`opengrep-mcp`): Pattern scanning
  - Runs Semgrep in isolated Docker containers
  - Mounts code read-only
  - Kills containers on timeout
  - Never logs Pro API tokens

- **Agents**: Claude agents bound to the skill
  - No file-write permissions (read-only code analysis)
  - Confined to provided code scope
  - All verdicts validated before synthesis

## Troubleshooting

**Docker not running?**
```bash
docker ps   # Should list running containers
```

**Semgrep container not available?**
```bash
docker pull returntocorp/semgrep:1.55.0
```

**Hooks binary not found?**
```bash
cd claude-security-hooks && make build
```

**OpenGrep MCP not responding?**
```bash
# Check if server is running on localhost:8000
curl http://localhost:8000/health || echo "Server not running"
```

## Output Structure

### review-report.json

```json
{
  "findings": [
    {
      "class": "injection",
      "confidence": "high",
      "title": "SQL injection in GetUser handler",
      "cwe": "CWE-89",
      "evidence": {
        "source": "http.Request.URL.Query",
        "sink": "database/sql.QueryRow",
        "path": "handlers.go:42 → db.go:15",
        "snippet": "..."
      }
    }
  ],
  "statistics": {
    "total_findings": 1,
    "high_confidence": 1,
    "medium_confidence": 0,
    "low_confidence": 0,
    "analysis_duration_seconds": 45
  }
}
```

### review-report.md

Human-readable summary with recommended fixes.

## Configuration

After installation, you can adjust permissions in `~/.claude/CLAUDE.md` or project-level `.claude/settings.json` to reduce permission prompts:

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

## Example: Review a Sample Vulnerable Service

```bash
/security-review examples/sample-vulnerable-service/
```

This analyzes a deliberately vulnerable Go HTTP service containing:
- SQL injection in user lookup
- Authorization bypass in admin endpoints
- OAuth scope tampering in token exchange

Expected output: all 3 bugs flagged with high confidence.

## Extending the Skill

### Add Custom Patterns

Edit or add `.yara` files to `patterns/`:

```bash
# Create new pattern
cat > patterns/custom-injection.yara << 'EOF'
rule sql_injection_custom {
    meta:
        description = "Custom SQL injection pattern"
    strings:
        $pattern = /sql\.Query\([^)]*\s\+\s/
    condition:
        $pattern
}
EOF
```

Then re-run `/security-review` to use new patterns.

### Add Custom Invariant Checkers

Edit `.claude/agents/invariant-checker.md` to add business-logic assertions:

```
Verify: Payment flows never accept negative amounts
Verify: Authorization checks happen before data access
```

## Support & Contributing

- **Issues**: https://github.com/saghaulor/security-reviewers/issues
- **Contributing**: See `CONTRIBUTING.md` in the repo
- **Design docs**: See `HAND_OFF.md` for architecture rationale

## License

MIT License — See `LICENSE` file.
