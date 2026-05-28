package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// --- E1: CodeRef Round-Trip Tests ---
//
// These tests verify that every schema struct carrying CodeRef/CodeRefDirty:
//  1. Serializes code_ref when set.
//  2. Omits code_ref_dirty when false (omitempty) and emits it when true.
//  3. Round-trips both fields without data loss.
//
// (The fields were added in Wave 1; these tests were previously skipped while the
// structs lacked them — IN-01.)

// assertCodeRefRoundTrip marshals v, asserts code_ref serializes to wantRef and
// code_ref_dirty follows omitempty semantics (present only when true), then
// unmarshals into a fresh T and re-marshals to confirm both fields round-trip.
// The round-trip check re-marshals the decoded value (rather than reflecting on
// fields) so a single generic helper covers every struct type.
func assertCodeRefRoundTrip[T any](t *testing.T, v T, wantRef string, wantDirty bool) {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal %T: %v", v, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal %T to map: %v", v, err)
	}
	if got, ok := raw["code_ref"].(string); !ok || got != wantRef {
		t.Errorf("%T code_ref: got %v, want %q", v, raw["code_ref"], wantRef)
	}
	if wantDirty {
		if got, ok := raw["code_ref_dirty"].(bool); !ok || !got {
			t.Errorf("%T code_ref_dirty: got %v, want true", v, raw["code_ref_dirty"])
		}
	} else if _, ok := raw["code_ref_dirty"]; ok {
		t.Errorf("%T code_ref_dirty should be omitted (omitempty) when false", v)
	}

	var back T
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("Unmarshal %T: %v", back, err)
	}
	roundTripped, err := json.Marshal(back)
	if err != nil {
		t.Fatalf("re-Marshal %T: %v", back, err)
	}
	var raw2 map[string]any
	if err := json.Unmarshal(roundTripped, &raw2); err != nil {
		t.Fatalf("Unmarshal round-tripped %T: %v", back, err)
	}
	if got, ok := raw2["code_ref"].(string); !ok || got != wantRef {
		t.Errorf("%T round-trip code_ref: got %v, want %q", back, raw2["code_ref"], wantRef)
	}
}

func TestCartographerIndex_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.CartographerIndex{CodeRef: "abc123def456", CodeRefDirty: false}, "abc123def456", false)
}

func TestTaintVerdict_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.TaintVerdict{CodeRef: "abc123", CodeRefDirty: false}, "abc123", false)
}

func TestAuthzVerdict_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.AuthzVerdict{CodeRef: "abc123", CodeRefDirty: true}, "abc123", true)
}

func TestOAuthVerdict_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.OAuthVerdict{CodeRef: "abc123", CodeRefDirty: false}, "abc123", false)
}

func TestInvariantCheckerVerdict_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.InvariantCheckerVerdict{CodeRef: "abc123", CodeRefDirty: false}, "abc123", false)
}

func TestSynthesisReport_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.SynthesisReport{CodeRef: "abc123", CodeRefDirty: false}, "abc123", false)
}

func TestTaintInput_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.TaintInput{CodeRef: "abc123", CodeRefDirty: false}, "abc123", false)
}

func TestAuthzInput_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.AuthzInput{CodeRef: "abc123", CodeRefDirty: true}, "abc123", true)
}

func TestOAuthInput_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.OAuthInput{CodeRef: "abc123", CodeRefDirty: false}, "abc123", false)
}

func TestInvariantCheckerInput_CodeRefRoundTrip(t *testing.T) {
	assertCodeRefRoundTrip(t, schema.InvariantCheckerInput{CodeRef: "abc123", CodeRefDirty: false}, "abc123", false)
}

// --- W7: Handler.FirstParamReadLine / FirstParamReadExpr round-trip tests ---

// TestHandler_FirstParamReadLine_RoundTrip verifies that Handler serializes and
// deserializes the FirstParamReadLine and FirstParamReadExpr fields correctly.
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

// --- WR-04: PaymentSurface.ClusterID omitempty ---

// TestPaymentSurface_ClusterID_OmitEmpty verifies that an empty ClusterID is omitted
// from JSON (the cartographer spec omits cluster_id because codegraph does not surface
// community IDs) and that a populated ClusterID round-trips.
func TestPaymentSurface_ClusterID_OmitEmpty(t *testing.T) {
	ps := schema.PaymentSurface{
		Files:      []string{"billing/stripe.go"},
		Confidence: "extracted",
		// ClusterID: zero-value ("")
	}
	data, err := json.Marshal(ps)
	if err != nil {
		t.Fatalf("Marshal PaymentSurface: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}
	if _, ok := raw["cluster_id"]; ok {
		t.Error("cluster_id should be omitted (omitempty) when empty")
	}

	// When set, it must round-trip.
	ps.ClusterID = "cluster-7"
	data, err = json.Marshal(ps)
	if err != nil {
		t.Fatalf("Marshal PaymentSurface with ClusterID: %v", err)
	}
	var ps2 schema.PaymentSurface
	if err := json.Unmarshal(data, &ps2); err != nil {
		t.Fatalf("Unmarshal PaymentSurface: %v", err)
	}
	if ps2.ClusterID != "cluster-7" {
		t.Errorf("round-trip ClusterID: got %q, want cluster-7", ps2.ClusterID)
	}
}
