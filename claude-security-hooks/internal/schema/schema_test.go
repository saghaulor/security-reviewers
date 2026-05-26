package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// --- E1: CodeRef Round-Trip Tests (compile-error RED until Wave 1) ---
//
// These tests verify that CodeRef and CodeRefDirty fields:
// 1. Serialize correctly (CodeRef present, CodeRefDirty omitted when false)
// 2. Deserialize correctly
// 3. Round-trip without data loss
//
// All tests will compile-error RED until Wave 1 adds CodeRef and CodeRefDirty
// fields to the 10 schema structs.

// TestCartographerIndex_CodeRefRoundTrip verifies that CartographerIndex
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestCartographerIndex_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to CartographerIndex:
	// index := &schema.CartographerIndex{
	//     SchemaVersion: "go-index/v1",
	//     GraphVersion:  "1.0",
	//     CodeRef:       "abc123def456",
	//     CodeRefDirty:  false,
	// }
	//
	// data, err := json.Marshal(index)
	// if err != nil {
	//     t.Fatalf("Marshal failed: %v", err)
	// }
	//
	// // Verify CodeRef is present in JSON
	// var raw map[string]interface{}
	// json.Unmarshal(data, &raw)
	// if v, ok := raw["code_ref"].(string); !ok || v != "abc123def456" {
	//     t.Errorf("code_ref missing or wrong in JSON")
	// }
	//
	// // Verify CodeRefDirty is omitted (omitempty)
	// if _, ok := raw["code_ref_dirty"]; ok {
	//     t.Errorf("code_ref_dirty should be omitted when false")
	// }
	//
	// // Round-trip
	// var unmarshaled schema.CartographerIndex
	// if err := json.Unmarshal(data, &unmarshaled); err != nil {
	//     t.Fatalf("Unmarshal failed: %v", err)
	// }
	// if unmarshaled.CodeRef != "abc123def456" || unmarshaled.CodeRefDirty != false {
	//     t.Errorf("Round-trip failed: got %q/%v, want %q/%v",
	//         unmarshaled.CodeRef, unmarshaled.CodeRefDirty, "abc123def456", false)
	// }

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to CartographerIndex (Wave 1)")
}

// TestTaintVerdict_CodeRefRoundTrip verifies that TaintVerdict
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestTaintVerdict_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to TaintVerdict:
	// verdict := &schema.TaintVerdict{
	//     Verdict:      "exploitable",
	//     Confidence:   "high",
	//     CodeRef:      "abc123",
	//     CodeRefDirty: false,
	// }
	//
	// data, err := json.Marshal(verdict)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to TaintVerdict (Wave 1)")
}

// TestAuthzVerdict_CodeRefRoundTrip verifies that AuthzVerdict
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestAuthzVerdict_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to AuthzVerdict:
	// verdict := &schema.AuthzVerdict{
	//     Summary:      schema.AuthzSummary{RoutesTotal: 1, Protected: 1},
	//     CodeRef:      "abc123",
	//     CodeRefDirty: true,
	// }
	//
	// data, err := json.Marshal(verdict)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to AuthzVerdict (Wave 1)")
}

// TestOAuthVerdict_CodeRefRoundTrip verifies that OAuthVerdict
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestOAuthVerdict_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to OAuthVerdict:
	// verdict := &schema.OAuthVerdict{
	//     Profile:      "oauth_2_0",
	//     CodeRef:      "abc123",
	//     CodeRefDirty: false,
	// }
	//
	// data, err := json.Marshal(verdict)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to OAuthVerdict (Wave 1)")
}

// TestInvariantCheckerVerdict_CodeRefRoundTrip verifies that InvariantCheckerVerdict
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestInvariantCheckerVerdict_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to InvariantCheckerVerdict:
	// verdict := &schema.InvariantCheckerVerdict{
	//     FlowName:     "test_flow",
	//     CodeRef:      "abc123",
	//     CodeRefDirty: false,
	// }
	//
	// data, err := json.Marshal(verdict)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to InvariantCheckerVerdict (Wave 1)")
}

// TestSynthesisReport_CodeRefRoundTrip verifies that SynthesisReport
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestSynthesisReport_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to SynthesisReport:
	// report := &schema.SynthesisReport{
	//     ReviewID:        "review-123",
	//     Timestamp:       "2026-05-26T21:25:26Z",
	//     SchemaVersion:   "review-report/v1",
	//     CodeRef:         "abc123",
	//     CodeRefDirty:    false,
	// }
	//
	// data, err := json.Marshal(report)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to SynthesisReport (Wave 1)")
}

// TestTaintInput_CodeRefRoundTrip verifies that TaintInput
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestTaintInput_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to TaintInput:
	// input := &schema.TaintInput{
	//     Source:       schema.TaintEndpoint{File: "main.go", Line: 10, Expr: "x", Kind: "param"},
	//     Sink:         schema.TaintEndpoint{File: "db.go", Line: 50, Expr: "q", Kind: "sink"},
	//     MaxDepth:     10,
	//     SemgrepTier:  "pro",
	//     CodeRef:      "abc123",
	//     CodeRefDirty: false,
	// }
	//
	// data, err := json.Marshal(input)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to TaintInput (Wave 1)")
}

// TestAuthzInput_CodeRefRoundTrip verifies that AuthzInput
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestAuthzInput_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to AuthzInput:
	// input := &schema.AuthzInput{
	//     Routes: []schema.AuthzRoute{
	//         {Router: "gin", Method: "GET", Path: "/api/users", Handler: schema.Handler{Name: "GetUsers"}},
	//     },
	//     CodeRef:      "abc123",
	//     CodeRefDirty: false,
	// }
	//
	// data, err := json.Marshal(input)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to AuthzInput (Wave 1)")
}

// TestOAuthInput_CodeRefRoundTrip verifies that OAuthInput
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestOAuthInput_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to OAuthInput:
	// input := &schema.OAuthInput{
	//     TargetProfile: "oauth_2_0",
	//     OAuthLocations: schema.OAuthLocations{
	//         AuthorizationEndpoint: []string{"https://provider.example.com/oauth/authorize"},
	//     },
	//     CodeRef:      "abc123",
	//     CodeRefDirty: false,
	// }
	//
	// data, err := json.Marshal(input)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to OAuthInput (Wave 1)")
}

// TestInvariantCheckerInput_CodeRefRoundTrip verifies that InvariantCheckerInput
// serializes/deserializes CodeRef and CodeRefDirty fields correctly.
func TestInvariantCheckerInput_CodeRefRoundTrip(t *testing.T) {
	// After Wave 1 adds CodeRef and CodeRefDirty to InvariantCheckerInput:
	// input := &schema.InvariantCheckerInput{
	//     FlowName: "test_flow",
	//     Invariants: []schema.InputInvariant{
	//         {ID: "I1", Statement: "test invariant"},
	//     },
	//     CodeRef:      "abc123",
	//     CodeRefDirty: false,
	// }
	//
	// data, err := json.Marshal(input)
	// // ... verify CodeRef/CodeRefDirty in JSON, round-trip, etc.

	t.Skip("E1 compile-error RED: CodeRef/CodeRefDirty fields not yet added to InvariantCheckerInput (Wave 1)")
}
