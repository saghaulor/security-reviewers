package main

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Placeholder database connection (nil for static analysis; not executed at runtime)
var db *sql.DB

// QueryBuilder is a helper struct for constructing SQL queries.
// Used in Phase 7 to create a 4+ hop call chain for SQLi detection.
type QueryBuilder struct {
	selectClause string
	fromClause   string
	whereClause  string
}

// Select sets the SELECT clause of the query.
// Returns *QueryBuilder for method chaining.
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		qb.selectClause = "SELECT "
		for i, col := range columns {
			if i > 0 {
				qb.selectClause += ", "
			}
			qb.selectClause += col
		}
	}
	return qb
}

// From sets the FROM clause of the query.
// Returns *QueryBuilder for method chaining.
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.fromClause = " FROM " + table
	return qb
}

// Where sets the WHERE clause using string concatenation (vulnerable).
// VULNERABILITY: This function concatenates the value directly into the WHERE clause
// without parameterization, creating a 3-hop data flow to the SQL sink.
func (qb *QueryBuilder) Where(column, value string) *QueryBuilder {
	// BUG(D-01,D-02): String concatenation without parameterization
	// This is hop 3 in the call chain: handler → service → repo → builder.Where() → db.Query()
	qb.whereClause = " WHERE " + column + " = '" + value + "'"
	return qb
}

// Build constructs and returns the complete SQL query string.
// Returns the concatenated query (vulnerable due to lack of parameterization).
func (qb *QueryBuilder) Build() string {
	return qb.selectClause + qb.fromClause + qb.whereClause
}

// Repo is a data access layer that constructs and executes queries.
// Used in Phase 7 to add a repository layer to the call chain.
type Repo struct {
	db *sql.DB
}

// QueryByName searches for users by name using the QueryBuilder.
// This is hop 2 in the call chain (service → repo → builder → db.Query).
// The name parameter is tainted from the HTTP input and flows through
// QueryBuilder.Where() without sanitization.
func (r *Repo) QueryByName(name string) []map[string]interface{} {
	// Build query using QueryBuilder (vulnerable chain)
	builder := &QueryBuilder{}
	builder.Select("id", "name", "email").
		From("users").
		Where("name", name)

	query := builder.Build()

	// SINK: db.Query() receives the concatenated (tainted) query
	// This is hop 4: the actual database query execution
	r.db.Query(query)

	// Return empty slice (for demo; no actual execution)
	return []map[string]interface{}{}
}

// Service is a business logic layer that performs user searches.
// Used in Phase 7 to add a service layer to the call chain.
type Service struct {
	repo *Repo
}

// SearchUsers searches for users matching a search term.
// This is hop 1 in the call chain (handler → service → repo).
// The searchTerm parameter comes from untrusted HTTP input and is passed
// to the repo layer without sanitization, allowing SQLi through the call chain.
func (s *Service) SearchUsers(searchTerm string) []map[string]interface{} {
	// Do some validation of OTHER business rules (not the input itself)
	if searchTerm == "" {
		return []map[string]interface{}{}
	}

	// Call repo without sanitizing the search term
	// This allows the untrusted input to flow through to the database
	return s.repo.QueryByName(searchTerm)
}

// getUserHandler handles GET /api/user?id=<id> with a SQL injection vulnerability.
//
// BUG(D-04): Raw string concatenation. userID is untrusted HTTP input.
// Direct concatenation into SQL query allows injection.
func getUserHandler(c *gin.Context) {
	// SOURCE: Read untrusted HTTP query parameter
	userID := c.Query("id")

	// PROPAGATION: Direct string concatenation without sanitizer
	query := "SELECT * FROM users WHERE id=" + userID

	// SINK: Pass concatenated query to Query() call
	// This allows SQL injection: c.Query("id") → concatenation → db.Query()
	db.Query(query)

	// Respond with stub response
	c.JSON(200, gin.H{"user": "data"})
}

// deleteUserHandler handles DELETE /user/:id with an authorization bypass vulnerability.
//
// BUG(D-04): DELETE endpoint with no authz middleware. Any request can delete any user.
func deleteUserHandler(c *gin.Context) {
	// Get user ID from URL path parameter
	userID := c.Param("id")

	// Execute sensitive operation (DELETE) without any authorization check
	// No middleware guard, no authz primitive in the call path
	db.Exec("DELETE FROM users WHERE id=?", userID)

	// Respond with success
	c.JSON(200, gin.H{"status": "deleted"})
}

// oauthAuthorizeHandler handles GET /oauth/authorize
// Part of a simplified OAuth 2.0 flow. This endpoint initiates the authorization request
// by reading client_id and scope from query parameters and storing them in server-side state.
//
// CORRECT pattern: The scope is stored in server-side state (authRequest) via StoreAuthRequest.
// The token endpoint SHOULD read from this stored request, not from form input.
func oauthAuthorizeHandler(c *gin.Context) {
	// SOURCE: Read untrusted HTTP query parameters
	clientID := c.Query("client_id")
	scope := c.Query("scope")

	// Generate a temporary authorization code
	code := GenerateAuthCode()

	// Store the authorization request in server-side state
	// This stores the CORRECT scope from the authorization request.
	StoreAuthRequest(code, clientID, scope, "user123")

	// Redirect to consent form with the authorization code
	// The consent form will POST back to /oauth/token with the code
	c.Redirect(302, "/consent?code="+code)
}

// consentHandler handles GET /consent
// Renders a simple HTML form for the user to approve the authorization request.
// This form will POST back to /oauth/token with the authorization code and scopes.
func consentHandler(c *gin.Context) {
	code := c.Query("code")

	// Simple HTML form for consent
	// NOTE: In a real OAuth flow, this would be a proper HTML template with security measures.
	// For this demo, we use a simple string for clarity.
	html := `<!DOCTYPE html>
<html>
<head><title>OAuth Consent</title></head>
<body>
<h1>Authorization Request</h1>
<p>An application has requested the following permissions:</p>
<form method="POST" action="/oauth/token">
  <input type="hidden" name="code" value="` + code + `">
  <label>
    <input type="checkbox" name="scope" value="read"> Read access
  </label>
  <label>
    <input type="checkbox" name="scope" value="write"> Write access
  </label>
  <label>
    <input type="checkbox" name="scope" value="admin"> Admin access
  </label>
  <button type="submit">Approve</button>
</form>
</body>
</html>`

	c.Data(200, "text/html; charset=utf-8", []byte(html))
}

// oauthTokenHandler handles POST /oauth/token
// Part of the OAuth 2.0 token endpoint. This endpoint exchanges an authorization code
// for an access token.
//
// BUG(D-06): Token endpoint reads scope from form POST instead of stored auth request.
// The vulnerability is that scopeFromForm is attacker-controlled and is NOT validated
// against the original authorization request scope stored by oauthAuthorizeHandler.
// This allows scope tampering / token privilege escalation.
//
// CORRECT pattern: Would validate scope matches original request:
//   if scopeFromForm != authRequest.Scope { return error("scope_mismatch") }
// Then use authRequest.Scope in the token response.
func oauthTokenHandler(c *gin.Context) {
	// Get the authorization code from form POST
	code, _ := c.GetPostForm("code")

	// SOURCE: Read untrusted form POST data (scope parameter)
	// This is where the attacker can tamper with scopes
	scopeFromForm, _ := c.GetPostForm("scope")

	// Retrieve the stored authorization request (to validate it exists)
	// NOTE: We retrieve it but then IGNORE the scope stored in it (the bug)
	authRequest, err := RetrieveAuthRequest(code)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid_code"})
		return
	}

	// BUG: No validation comparing scopeFromForm against authRequest.Scope
	// A correct implementation would be:
	//   if scopeFromForm != authRequest.Scope {
	//     c.JSON(400, gin.H{"error": "scope_mismatch"})
	//     return
	//   }

	// Generate a simple token (no cryptography needed for vulnerability demonstration)
	token := "token_" + code + "_" + authRequest.UserID

	// SINK: Form scope is directly assigned to response scope (taint propagation)
	// scopeFromForm comes from untrusted form input and flows directly to the response
	// This is the vulnerability: c.PostFormValue("scope") → response.Scope
	c.JSON(200, gin.H{
		"access_token": token,
		"scope":        scopeFromForm,
		"token_type":   "Bearer",
	})
}

// callChainSQLiHandler handles GET /api/advanced-search with a 4+ hop SQL injection vulnerability.
//
// BUG(D-01,D-02): Service layer doesn't sanitize input; call chain: handler → ServiceLayer → RepoLayer → builder → db.Query
// FIX: Use parameterized queries at every layer or sanitize at the source before passing to service
//
// This handler demonstrates a call-chain SQLi where the vulnerability is separated from the HTTP source
// by multiple function boundaries (service → repo → builder), testing whether the tracer can follow the
// data-flow across layers and recognize the injection sink even when indirect.
func callChainSQLiHandler(c *gin.Context) {
	// SOURCE: Read untrusted HTTP query parameter
	userInput := c.Query("search")

	// Create service with repo and db
	repo := &Repo{db: db}
	service := &Service{repo: repo}

	// Call service without sanitizing input (vulnerability flows through layers)
	results := service.SearchUsers(userInput)

	// Respond with search results
	c.JSON(200, gin.H{"results": results})
}

// transferFundsHandler handles POST /transfer with a weak authorization vulnerability.
//
// BUG(D-03): Authz middleware checks isAuthenticated(); handler assumes permission without resource ownership check
// FIX: Add resource ownership verification: if userID != recipientID { return 403 }
// Vulnerability: Identity-only check allows any authenticated user to transfer funds to any recipient (IDOR)
func transferFundsHandler(c *gin.Context) {
	// From auth middleware (identity verified)
	userID := c.GetString("user_id")

	// Attacker can set this to anyone
	recipientID := c.Query("to")
	amount, _ := strconv.ParseFloat(c.Query("amount"), 64)

	// BUG: No ownership check! Any authenticated user can transfer to any recipient
	// Missing: if userID != recipientID { c.JSON(403, ...); return }

	db.Exec("UPDATE accounts SET balance = balance - ? WHERE user_id = ?", amount, userID)
	db.Exec("UPDATE accounts SET balance = balance + ? WHERE user_id = ?", amount, recipientID)

	c.JSON(200, gin.H{"status": "transferred"})
}

// UserToken represents a user extracted from a JWT token.
// Includes both verified claims (ID) and unverified claims (IsAdmin).
type UserToken struct {
	ID      string
	IsAdmin bool // Unvalidated claim from token
}

// extractUserFromToken extracts a user from a JWT token.
// In a real implementation, this would verify the JWT signature and claims.
// For this demo, it returns a simple user token.
func extractUserFromToken(c *gin.Context) (*UserToken, error) {
	// For demo: decode JWT without validating custom claims against authz service
	// An attacker could craft a token with IsAdmin: true
	return &UserToken{ID: "attacker123", IsAdmin: false}, nil
}

// adminConfigHandler handles POST /admin/config with a weak authorization vulnerability.
//
// BUG(D-03): Handler trusts IsAdmin() claim from JWT without server-side validation
// FIX: Call authz service to verify claim: authzService.IsAdmin(user.ID)
// Vulnerability: Unvalidated JWT claims allow privilege escalation (attacker sets IsAdmin:true in token)
func adminConfigHandler(c *gin.Context) {
	// Decode user from JWT token
	user, _ := extractUserFromToken(c)

	// BUG: Trusts unvalidated JWT claim without authz service verification
	// Missing: if !authzService.IsAdmin(user.ID) { c.JSON(403, ...); return }

	if user.IsAdmin { // UNVALIDATED CLAIM from token
		newConfig := c.PostForm("config")
		db.Exec("UPDATE system_config SET value = ? WHERE key = 'admin_setting'", newConfig)
		c.JSON(200, gin.H{"status": "config updated"})
		return
	}

	c.JSON(403, gin.H{"error": "forbidden"})
}
