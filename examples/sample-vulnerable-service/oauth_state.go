package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// AuthRequest represents an OAuth authorization request stored in server-side state.
// Per the OAuth 2.0 spec, the authorization server should store the original scope
// and validate it at the token endpoint. This implementation intentionally stores it
// but tokenHandler doesn't validate against it (the vulnerability).
type AuthRequest struct {
	ClientID string
	Scope    string
	UserID   string
	Code     string
}

// authRequests holds in-memory storage of OAuth authorization requests.
// Using sync.Map for thread-safe concurrent access (no external DB required).
var authRequests sync.Map

// StoreAuthRequest stores an OAuth authorization request in memory.
// This stores the CORRECT scope from the initial authorization request.
// The tokenHandler SHOULD validate against this scope, but doesn't (vulnerability).
func StoreAuthRequest(code, clientID, scope, userID string) {
	authReq := &AuthRequest{
		ClientID: clientID,
		Scope:    scope,
		UserID:   userID,
		Code:     code,
	}
	authRequests.Store(code, authReq)
}

// RetrieveAuthRequest retrieves a stored OAuth authorization request by code.
// Returns an error if the code is not found (e.g., invalid or expired code).
func RetrieveAuthRequest(code string) (*AuthRequest, error) {
	value, ok := authRequests.Load(code)
	if !ok {
		return nil, errors.New("invalid_code: authorization code not found")
	}

	authReq, ok := value.(*AuthRequest)
	if !ok {
		return nil, errors.New("internal_error: invalid auth request state")
	}

	return authReq, nil
}

// GenerateAuthCode generates a pseudo-random authorization code.
// No cryptographic randomness required for this demo (OAuth 2.0 typically uses
// higher-entropy codes in production, but for vulnerability demonstration
// a simple approach is sufficient).
func GenerateAuthCode() string {
	// Seed random number generator if not already seeded.
	// In production, use crypto/rand for cryptographic randomness.
	rand.Seed(time.Now().UnixNano())

	// Generate random code component
	randomID := rand.Intn(1000000)
	code := fmt.Sprintf("code_%d_%d", time.Now().UnixNano(), randomID)
	return code
}
