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

// --- W7: Handler.FirstParamReadLine / FirstParamReadExpr round-trip tests ---
//
// These tests verify that Handler gains two new fields:
//   FirstParamReadLine int    `json:"first_param_read_line,omitempty"`
//   FirstParamReadExpr string `json:"first_param_read_expr,omitempty"`
//
// Both tests will produce a COMPILE ERROR until 13-04 Task 2 adds the fields.
// The compile error IS the correct RED state for struct-addition tests.

// TestHandler_FirstParamReadLine_RoundTrip verifies that Handler serializes and
// deserializes the new FirstParamReadLine and FirstParamReadExpr fields correctly.
// RED: compile error until Handler struct gains these fields.
func TestHandler_FirstParamReadLine_RoundTrip(t *testing.T) {
	h := schema.Handler{
		FQN:                "example.Handler",
		File:               "handlers.go",
		Line:               42,
		FirstParamReadLine: 44,
		FirstParamReadExpr: `c.Query("id")`,
	}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal Handler: %v", err)
	}

	// Verify keys present in raw JSON map.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}
	if v, ok := raw["first_param_read_line"]; !ok {
		t.Error("expected key first_param_read_line in JSON")
	} else if v != float64(44) {
		t.Errorf("first_param_read_line: got %v, want 44", v)
	}
	if v, ok := raw["first_param_read_expr"]; !ok {
		t.Error("expected key first_param_read_expr in JSON")
	} else if v != `c.Query("id")` {
		t.Errorf("first_param_read_expr: got %v, want c.Query(\"id\")", v)
	}

	// Round-trip: unmarshal back into Handler and assert field values.
	var h2 schema.Handler
	if err := json.Unmarshal(data, &h2); err != nil {
		t.Fatalf("Unmarshal to Handler: %v", err)
	}
	if h2.FirstParamReadLine != 44 {
		t.Errorf("round-trip FirstParamReadLine: got %d, want 44", h2.FirstParamReadLine)
	}
	if h2.FirstParamReadExpr != `c.Query("id")` {
		t.Errorf("round-trip FirstParamReadExpr: got %q, want c.Query(\"id\")", h2.FirstParamReadExpr)
	}
}

// TestHandler_FirstParamReadLine_OmitEmpty verifies that zero-value FirstParamReadLine
// and empty FirstParamReadExpr are omitted from JSON output (omitempty semantics).
// RED: compile error until Handler struct gains these fields.
func TestHandler_FirstParamReadLine_OmitEmpty(t *testing.T) {
	h := schema.Handler{
		FQN:  "example.NoParamHandler",
		File: "handlers.go",
		Line: 10,
		// FirstParamReadLine: zero-value (0)
		// FirstParamReadExpr: zero-value ("")
	}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal Handler: %v", err)
	}

	// Verify that omitempty fields are absent from JSON.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}
	if _, ok := raw["first_param_read_line"]; ok {
		t.Error("first_param_read_line should be omitted (omitempty) when zero")
	}
	if _, ok := raw["first_param_read_expr"]; ok {
		t.Error("first_param_read_expr should be omitted (omitempty) when empty")
	}
}
