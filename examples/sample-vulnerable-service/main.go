package main

import (
	"github.com/gin-gonic/gin"
)

// authMiddleware is a simple authentication middleware that sets the user_id in the context.
// For Phase 7 purposes, this simulates a basic auth check (identity-only, no resource-level authz).
// The handlers use this to determine if the user is authenticated, but still need to verify
// that the user has permission to perform the specific operation.
func authMiddleware(c *gin.Context) {
	c.Set("user_id", "user123")
	c.Next()
}

func main() {
	// Per D-01: Use gin framework for the sample vulnerable service.
	// Per D-02: Service minimal but complete with three routes (one per bug class).
	router := gin.Default()

	// Bug 1: SQLi vulnerability
	// GET /api/user?id=<userID> → getUserHandler (raw SQL concatenation)
	router.GET("/api/user", getUserHandler)

	// Bug 2: Authorization bypass vulnerability
	// DELETE /user/:id → deleteUserHandler (missing authz middleware)
	router.DELETE("/user/:id", deleteUserHandler)

	// Bug 3: OAuth scope-tampering vulnerability
	// GET /oauth/authorize → oauthAuthorizeHandler (simplified OAuth flow)
	// GET /consent → consentHandler (consent form for user approval)
	// POST /oauth/token → oauthTokenHandler (reads scopes from form POST body)
	router.GET("/oauth/authorize", oauthAuthorizeHandler)
	router.GET("/consent", consentHandler)
	router.POST("/oauth/token", oauthTokenHandler)

	// Bug 4: Call-chain SQL injection (4+ hop pattern)
	// GET /api/advanced-search?search=<userInput> → callChainSQLiHandler
	router.GET("/api/advanced-search", callChainSQLiHandler)

	// Bug 5: Identity-only weak authorization check
	// POST /transfer?to=<recipient>&amount=<amount> → transferFundsHandler
	// (With authMiddleware to make the pattern clear: middleware present, but handler insufficient)
	router.POST("/transfer", authMiddleware, transferFundsHandler)

	// Bug 6: Implicit permission assumption weak authorization
	// POST /admin/config → adminConfigHandler
	// (With authMiddleware to make the pattern clear: middleware present, handler trusts unvalidated claim)
	router.POST("/admin/config", authMiddleware, adminConfigHandler)

	// Start the server (will listen on 0.0.0.0:8080 by default with gin.Default())
	// Per D-03: Service is source code only; no requirement to run as a live HTTP service.
	// This allows /security-review to analyze the code statically.
	router.Run()
}
