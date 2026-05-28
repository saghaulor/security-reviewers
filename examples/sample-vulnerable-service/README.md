# Sample Vulnerable Service — Phase 7 Extended Test Cases

This is a small Go HTTP service (using Gin framework) with six intentionally planted security bugs. It serves as a test fixture for verifying the `/security-review` workflow. Phase 5 proves the system works with three basic bugs. Phase 7 extends to six total bugs by adding three moderate-difficulty vulnerabilities (call-chain SQLi, identity-only weak authz, implicit permission assumption), testing whether the tracer agents can handle complex multi-layer patterns.

## What Is This Service

The service implements a simplified OAuth 2.0 flow for a web API:

- **Main entry points:** `GET /api/user`, `DELETE /user/:id`, `GET /oauth/authorize`, `GET /consent`, `POST /oauth/token`
- **Framework:** Gin (Go HTTP framework)
- **Purpose:** Demonstrate real-world bug classes in a runnable example
- **Not for production:** This code is intentionally vulnerable; do not run it publicly

The six planted bugs test different analysis capabilities:

**Phase 5 Basic Bugs (3):**
1. **Injection-class bug** (SQL injection) — Tests taint-flow tracing from HTTP input to database query
2. **Authorization-class bug** (authz bypass) — Tests middleware/control-flow analysis
3. **OAuth-class bug** (scope tampering) — Tests OAuth-specific data-flow analysis

**Phase 7 Moderate Bugs (3):**
4. **Injection-class bug** (call-chain SQLi) — Tests multi-hop taint tracing across function boundaries (handler → service → repo → builder → db)
5. **Authorization-class bug** (identity-only weak authz) — Tests detection of middleware-only auth (missing resource-level checks)
6. **Authorization-class bug** (implicit permission assumption) — Tests detection of unvalidated JWT claims trusted directly

## All Six Bugs (Summary Table)

| # | Class | Difficulty | Location (File:Lines) | Description | Expected Finding |
|---|-------|------------|-------|-------------|------------------|
| 1 | SQLi | Basic | handlers.go:111–127 | Raw SQL concat in `getUserHandler` | class: injection, confidence: high, data-flow path (2+ hops) |
| 2 | Authz | Basic | handlers.go:129–141 | Missing authz in `deleteUserHandler` | class: authz, confidence: high, missing primitive |
| 3 | OAuth | Basic | handlers.go:210–251 | Form scope bypass in `oauthTokenHandler` | class: oauth, confidence: high, data-flow path (2+ hops) |
| 4 | SQLi | Moderate | handlers.go:254–271 | 4+ hop call-chain in `callChainSQLiHandler` | class: injection, confidence: high, data-flow path (4+ hops) |
| 5 | Authz | Moderate | handlers.go:274–309 | Identity-only check in `transferFundsHandler` | class: authz, confidence: high, IDOR / missing resource check |
| 6 | Authz | Moderate | handlers.go:312–337 | Implicit permission in `adminConfigHandler` | class: authz, confidence: high, unvalidated claim / weak primitive |

All six bugs are intentionally planted for testing the `/security-review` system's ability to detect diverse vulnerability classes and patterns. The Phase 7 additions (bugs 4–6) test whether the system can handle multi-layer call chains, weak authorization patterns, and unvalidated JWT claims with high confidence.

---

## The Six Bugs (Detailed Explanations)

### Bug 1: SQL Injection (Injection Class)

**Location:** `handlers.go`, function `getUserHandler()`, lines 16–29

**Vulnerability:** User input from HTTP query parameter is directly concatenated into a SQL query without parameterization.

```go
userID := c.Query("id")  // SOURCE: untrusted HTTP input
query := "SELECT * FROM users WHERE id=" + userID  // PROPAGATION: raw concat
db.Query(query)  // SINK: execute query with untrusted input
```

**Expected finding in review-report.json:**
- `"class": "injection"` — SQL injection vulnerability class
- `"confidence": "high"` — Direct concatenation is unambiguous
- `"evidence.data_flow_path"` — Shows 2+ steps: HTTP query parameter → concatenation → db.Query()

**Why the system should detect it:** The taint-tracer receives the pair (source: `c.Query("id")`, sink: `db.Query()`), searches for data flow, and confirms untrusted input reaches the sink without a recognized sanitizer (parameterized queries).

---

### Bug 2: Authorization Bypass (Authz Class)

**Location:** `handlers.go`, function `deleteUserHandler()`, lines 34–44

**Vulnerability:** DELETE endpoint performs a sensitive operation (delete any user) without any authorization check or middleware guard.

```go
// No middleware, no authorization check before deletion
db.Exec("DELETE FROM users WHERE id=?", userID)  // Executes for any request
```

**Expected finding in review-report.json:**
- `"class": "authz"` — Authorization missing or insufficient
- `"confidence": "high"` — Absence of control-flow guard is clear
- `"evidence.files"` — Citations showing the endpoint handler and absence of middleware

**Why the system should detect it:** The authz-tracer receives the sensitive operation (DELETE) and checks if the call path includes authorization middleware. Finding none, it reports an authz bypass.

---

### Bug 3: OAuth Scope Tampering (OAuth Class)

**Location:** `handlers.go`, function `oauthTokenHandler()`, lines 115–149

**Vulnerability:** The token endpoint reads the scopes from the form POST (attacker-controlled) instead of validating them against the server-side authorization request. This allows an attacker to grant themselves additional privileges.

```go
// Line 118: SOURCE — read untrusted form POST
scopeFromForm, _ := c.GetPostForm("scope")

// Line 125: Retrieve stored auth request but then IGNORE its scope
authRequest, err := RetrieveAuthRequest(code)
// BUG: No validation like: if scopeFromForm != authRequest.Scope { ... }

// Line 144: SINK — form scope directly assigned to token response
c.JSON(200, gin.H{
    "access_token": token,
    "scope": scopeFromForm,  // Attacker-controlled scope in response
    "token_type": "Bearer",
})
```

**Expected finding in review-report.json:**
- `"class": "oauth"` — OAuth-specific vulnerability (scope tampering)
- `"confidence": "high"` — Form input flows directly to token response without validation
- `"evidence.data_flow_path"` — Shows 2+ steps: form input → token response

**Why the system should detect it:** The oauth-auditor receives the token endpoint and checks for scope validation. The tracer traces form-input scope to the response and finds no validation check (missing `authRequest.Scope` comparison). This is reported as scope-tampering.

---

### Bug 4: Call-Chain SQL Injection (Moderate - Multi-Layer Pattern)

**Location:** `handlers.go`, function `callChainSQLiHandler()`, lines 254–271

**Vulnerability:** User input from HTTP query parameter flows through multiple layers (service → repo → query builder) before reaching the database query. Each layer has real business logic, making the taint chain non-trivial to trace. The WHERE clause is constructed via string concatenation without parameterization.

```go
// Hop 1: HTTP handler
userInput := c.Query("search")  // SOURCE: untrusted HTTP input

// Hop 2: Service layer
results := service.SearchUsers(userInput)  // Passes to service without sanitization

// (Inside Service.SearchUsers)
// Hop 3: Repository layer
return repo.QueryByName(searchTerm)  // Passes to repo without sanitization

// (Inside Repo.QueryByName)
// Hop 4: Query builder
builder.Select("id","name","email").From("users").Where("name", name).Build()

// (Inside QueryBuilder.Where)
// Hop 5: String concatenation (vulnerable)
whereClause = " WHERE " + column + " = '" + value + "'"  // PROPAGATION: raw concat

// Hop 5 (continued): Database query
// SINK: execute concatenated query
db.Query(builder.Build())  // Query with untrusted input
```

**Expected finding in review-report.json:**
- `"class": "injection"` — SQL injection vulnerability
- `"confidence": "high"` — String concatenation is unambiguous
- `"evidence.data_flow_path"` — Shows 4+ steps: HTTP query parameter → service → repo → builder.Where() → db.Query()

**Why the system should detect it:** The taint-tracer must follow data-flow across function boundaries (service → repo → builder methods) while maintaining taint context. The injection sink (db.Query) is multiple hops away from the source (c.Query), testing whether the tracer can reconstruct the complete call chain.

---

### Bug 5: Identity-Only Weak Authorization (Moderate - Missing Resource Check)

**Location:** `handlers.go`, function `transferFundsHandler()`, lines 274–309

**Vulnerability:** The endpoint has authentication middleware that verifies the user is logged in. However, the handler performs a sensitive operation (transferring funds) without checking if the user has permission to perform that specific action on that specific resource (recipient account).

```go
// Authentication middleware sets user_id in context (identity verified)
userID := c.GetString("user_id")  // user123 is authenticated

// Attacker can specify any recipient
recipientID := c.Query("to")  // Could be "victim_account"

// BUG: No resource ownership check! Any authenticated user can transfer to any recipient
// Missing check: if userID != recipientID { c.JSON(403, ...); return }

db.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?", amount, userID)
db.Exec("UPDATE accounts SET balance = balance + ? WHERE user_id = ?", amount, recipientID)
```

**Expected finding in review-report.json:**
- `"class": "authz"` — Authorization missing or insufficient
- `"confidence": "high"` — Absence of resource-level check is clear
- `"evidence"` — May cite the endpoint handler and absence of resource ownership validation

**Why the system should detect it:** The authz-tracer receives the sensitive operation (fund transfer) and checks if the call path includes sufficient authorization checks. Finding only identity verification but no resource-level check, it reports weak authorization (IDOR vulnerability).

---

### Bug 6: Implicit Permission Assumption Weak Authorization (Moderate - Unvalidated JWT Claim)

**Location:** `handlers.go`, function `adminConfigHandler()`, lines 312–337

**Vulnerability:** The endpoint has authentication middleware that verifies the user is logged in. The handler checks a user claim from the JWT token (IsAdmin) without verifying that claim against the server's authorization service. This allows an attacker to forge a JWT with IsAdmin=true and gain admin privileges.

```go
// Authentication middleware verifies user is logged in
user, _ := extractUserFromToken(c)  // Extracts claims from JWT

// BUG: Trusts unvalidated JWT claim without server-side verification
// Missing: if !authzService.IsAdmin(user.ID) { c.JSON(403, ...); return }

if user.IsAdmin {  // UNVALIDATED CLAIM from token
    newConfig := c.PostForm("config")
    db.Exec("UPDATE system_config SET value = ? WHERE key = 'admin_setting'", newConfig)
    c.JSON(200, gin.H{"status": "config updated"})
    return
}
```

**Expected finding in review-report.json:**
- `"class": "authz"` — Authorization missing or insufficient
- `"confidence": "high"` — Trusting unvalidated claims is a known weak primitive
- `"evidence"` — May cite the endpoint handler and the unvalidated claim check

**Why the system should detect it:** The authz-tracer receives the sensitive operation (config update) and traces the authorization check back to the IsAdmin claim from the JWT. Finding no server-side validation call (authz service), it reports weak authorization (privilege escalation via unvalidated claim).

---

## How to Run /security-review and Verify Findings

### Prerequisites

Before running the security review, ensure:

1. **opengrep-mcp is running:**
   ```bash
   # Check if the server is already running
   curl -s http://localhost:8000/health || echo "Server not running"

   # If not running, start it (in a separate terminal)
   /path/to/opengrep-mcp/bin/opengrep-mcp &
   ```

2. **Agent definitions are registered:** See `.planning/STATE.md` or `.claude/settings.json` to verify six agents are registered.

3. **You're in the repo root:** The `/security-review` command outputs `review-report.json` to the current directory.

### Run the Review

In the repo root:

```bash
/security-review examples/sample-vulnerable-service/
```

**Output:** The command generates:
- `review-report.json` — Machine-readable findings (JSON)
- `review-report.md` — Human-readable summary (Markdown, optional)

### Verify the Six Findings

Parse the report and verify you see six findings: two injection (basic + call-chain), three authz (basic + identity-only + implicit assumption), and one oauth (scope tampering):

```bash
# Count findings (expect 6)
cat review-report.json | jq '.findings | length'
# Expected output: 6

# List finding classes (expect: injection, injection, authz, authz, authz, oauth)
cat review-report.json | jq '.findings[] | .class'
# Expected output (order may vary):
# "injection"
# "injection"
# "authz"
# "authz"
# "authz"
# "oauth"

# Verify all are high confidence
cat review-report.json | jq '.findings[] | {class, confidence}'
# Expected output: all with confidence="high"

# Verify basic injection finding has 2+ hop data-flow path
cat review-report.json | jq '.findings[] | select(.class=="injection") | .evidence.data_flow_path | length'
# Expected output: At least one with >= 2 hops (basic pattern)

# Verify call-chain injection finding has 4+ hop data-flow path
cat review-report.json | jq '.findings[] | select(.class=="injection") | .evidence.data_flow_path | select(length >= 4) | length'
# Expected output: At least one with >= 4 hops (call-chain pattern)

# Verify authz findings (expect 3)
cat review-report.json | jq '.findings[] | select(.class=="authz") | length'
# Expected output: 3 (basic bypass + identity-only + implicit assumption)

# Verify oauth finding has 2+ hop data-flow path
cat review-report.json | jq '.findings[] | select(.class=="oauth") | .evidence.data_flow_path | length'
# Expected output: >= 2 (scope tampering pattern)
```

If all checks pass with the expected outputs, the security review system is working correctly and has detected all six planted bugs, including the moderate-difficulty call-chain and weak authorization patterns.

---

## Future Extensions

Phase 7 completes the basic test suite with six representative bugs covering three classes (injection, authz, oauth) and two difficulty levels (basic, moderate). For future phases or additional contributions, here are additional variants that would make good test cases:

### Potential Future Extensions

1. **Redirect URI Tampering (OAuth)** — Token endpoint allows custom redirect_uri, leaking tokens to attacker-controlled domains. Tests OAuth endpoint validation beyond scope tampering.

2. **Cross-request forgery via OAuth state** — OAuth flow doesn't validate the state parameter, allowing state injection attacks. Tests OAuth state validation.

3. **Path traversal in file serving** — Handler serves files with path concatenation without validation. Tests path traversal detection.

4. **JWT signature bypass** — Handler accepts unsigned JWTs or skips signature validation. Tests JWT validation detection.

5. **Type confusion in authz checks** — Authz check compares types inconsistently (e.g., string "1" vs integer 1). Tests type-safety in authorization.

**Implementation note:** These are suggestions for future contributors. Phase 7 establishes the patterns; additional classes and difficulty levels can extend the test suite as new analysis capabilities are added to the system.

## See Also

- `.claude/agents/README.md` — How agents are defined and how they trace through code like this
- `examples/sample-vulnerable-service/oauth_test.go` — Programmatic test fixture (verifies review-report.json contains expected findings)
- `HAND_OFF.md` §8 — Source/sink catalogs that the tracers use to recognize these bug classes
