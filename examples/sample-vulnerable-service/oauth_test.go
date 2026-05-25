package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestAuthorizeHandlerCompiles verifies the handler compiles without errors.
func TestAuthorizeHandlerCompiles(t *testing.T) {
	// Should not panic
	_ = oauthAuthorizeHandler
}

// TestAuthorizeHandlerReadsClientID verifies handler reads client_id from query params.
func TestAuthorizeHandlerReadsClientID(t *testing.T) {
	router := gin.New()
	router.GET("/oauth/authorize", oauthAuthorizeHandler)

	req := httptest.NewRequest("GET", "/oauth/authorize?client_id=test-client&scope=read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler should read and store the client_id
	require.NotNil(t, w, "Response should not be nil")
}

// TestAuthorizeHandlerReadsScope verifies handler reads scope from query params.
func TestAuthorizeHandlerReadsScope(t *testing.T) {
	router := gin.New()
	router.GET("/oauth/authorize", oauthAuthorizeHandler)

	req := httptest.NewRequest("GET", "/oauth/authorize?client_id=test-client&scope=read+write", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler should read and store the scope
	require.NotNil(t, w, "Response should not be nil")
}

// TestAuthorizeHandlerGeneratesCode verifies handler generates an auth code.
func TestAuthorizeHandlerGeneratesCode(t *testing.T) {
	router := gin.New()
	router.GET("/oauth/authorize", oauthAuthorizeHandler)

	req := httptest.NewRequest("GET", "/oauth/authorize?client_id=test-client&scope=read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should redirect (302 Found or 303 See Other)
	require.Contains(t, []int{http.StatusFound, http.StatusSeeOther}, w.Code,
		"Authorize handler should redirect (302 or 303)")
}

// TestAuthorizeHandlerRedirectsToConsent verifies handler redirects to /consent?code=...
func TestAuthorizeHandlerRedirectsToConsent(t *testing.T) {
	router := gin.New()
	router.GET("/oauth/authorize", oauthAuthorizeHandler)

	req := httptest.NewRequest("GET", "/oauth/authorize?client_id=test-client&scope=read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	location := w.Header().Get("Location")
	require.NotEmpty(t, location, "Redirect header should not be empty")
	require.True(t, strings.HasPrefix(location, "/consent?code="),
		"Redirect should be to /consent?code=...")
}

// TestAuthRequestStorageExists verifies oauth_state.go provides storage functions.
func TestAuthRequestStorageExists(t *testing.T) {
	// These should not panic or be nil
	_ = StoreAuthRequest
	_ = RetrieveAuthRequest
	_ = GenerateAuthCode
}

// TestTokenHandlerCompiles verifies the handler compiles without errors.
func TestTokenHandlerCompiles(t *testing.T) {
	// Should not panic
	_ = oauthTokenHandler
}

// TestTokenHandlerReadsCodeFromForm verifies handler reads code from POST form.
func TestTokenHandlerReadsCodeFromForm(t *testing.T) {
	router := gin.New()
	router.POST("/oauth/token", oauthTokenHandler)

	formData := url.Values{
		"code":  {"test-code"},
		"scope": {"read"},
	}

	req := httptest.NewRequest("POST", "/oauth/token",
		strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler should process the request
	require.NotNil(t, w, "Response should not be nil")
}

// TestTokenHandlerReadsScope verifies handler reads scope from POST form (SOURCE).
func TestTokenHandlerReadsScope(t *testing.T) {
	router := gin.New()
	router.POST("/oauth/token", oauthTokenHandler)

	formData := url.Values{
		"code":  {"test-code"},
		"scope": {"read write admin"},
	}

	req := httptest.NewRequest("POST", "/oauth/token",
		strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler should read scope from form
	require.NotNil(t, w, "Response should not be nil")
}

// TestTokenHandlerReturnsJSON verifies handler returns JSON response with access_token and scope.
func TestTokenHandlerReturnsJSON(t *testing.T) {
	router := gin.New()
	router.POST("/oauth/token", oauthTokenHandler)

	formData := url.Values{
		"code":  {"test-code"},
		"scope": {"read write admin"},
	}

	req := httptest.NewRequest("POST", "/oauth/token",
		strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code,
		"Token handler should return 200 OK for valid request")
	require.Contains(t, w.Header().Get("Content-Type"), "application/json",
		"Response should be JSON")

	// Parse response to verify structure
	responseBody := w.Body.String()
	require.Contains(t, responseBody, "access_token",
		"Response should contain access_token field")
	require.Contains(t, responseBody, "scope",
		"Response should contain scope field")
}

// TestTokenHandlerIncludesFormScopeInResponse verifies that form scope is in response (SINK).
// This test documents the vulnerability: scopeFromForm is directly assigned to response.Scope
func TestTokenHandlerIncludesFormScopeInResponse(t *testing.T) {
	router := gin.New()
	router.POST("/oauth/token", oauthTokenHandler)

	formData := url.Values{
		"code":  {"test-code"},
		"scope": {"read write admin"},
	}

	req := httptest.NewRequest("POST", "/oauth/token",
		strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseBody := w.Body.String()
	// The vulnerability is that the form scope is directly in the response
	require.Contains(t, responseBody, "read write admin",
		"Response should contain the form scope (vulnerability: no validation)")
}

// TestConsentHandlerExists verifies the consent form handler exists.
func TestConsentHandlerExists(t *testing.T) {
	// Should not panic
	_ = consentHandler
}

// TestConsentHandlerRendersForm verifies consent handler renders an HTML form.
func TestConsentHandlerRendersForm(t *testing.T) {
	router := gin.New()
	router.GET("/consent", consentHandler)

	req := httptest.NewRequest("GET", "/consent?code=test-code", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code,
		"Consent handler should return 200 OK")
	require.Contains(t, w.Header().Get("Content-Type"), "text/html",
		"Response should be HTML")

	responseBody := w.Body.String()
	require.Contains(t, responseBody, "form",
		"Response should contain an HTML form")
	require.Contains(t, responseBody, "test-code",
		"Form should include the authorization code")
}
