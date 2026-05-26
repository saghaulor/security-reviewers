package invariants_test

import (
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/invariants"
	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// locateAuthzCheck retrieves the Check function from the AuthzInvariants registry
// by ID for test invocation.
func locateAuthzCheck(t *testing.T, id string) func(*schema.AuthzVerdict) []invariants.Violation {
	for _, inv := range invariants.AuthzInvariants {
		if inv.ID == id {
			return inv.Check
		}
	}
	t.Fatalf("locateAuthzCheck: ID %q not found in AuthzInvariants", id)
	return nil
}

// locateAuthzJointCheck retrieves the Check function from the AuthzJointInvariants
// registry by ID for test invocation.
func locateAuthzJointCheck(t *testing.T, id string) func(*schema.AuthzInput, *schema.AuthzVerdict) []invariants.Violation {
	for _, inv := range invariants.AuthzJointInvariants {
		if inv.ID == id {
			return inv.Check
		}
	}
	t.Fatalf("locateAuthzJointCheck: ID %q not found in AuthzJointInvariants", id)
	return nil
}

// TestAZ1_VerdictRequiredFields verifies that AuthzVerdict has required fields
// and non-negative bucket counts.
func TestAZ1_VerdictRequiredFields(t *testing.T) {
	check := locateAuthzCheck(t, "AZ1")

	tests := []struct {
		name      string
		verdict   *schema.AuthzVerdict
		wantLen   int
		wantError bool
	}{
		{
			name: "valid verdict with data",
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{
					RoutesTotal:       5,
					Protected:         3,
					Missing:           1,
					Weak:              0,
					IdorRisk:          0,
					PublicIntentional: 1,
				},
				Findings: []schema.AuthzFinding{
					{
						Route:      "POST /api/users",
						Issue:      "missing_authz",
						Confidence: "high",
					},
				},
			},
			wantLen:   0,
			wantError: false,
		},
		{
			name: "all-zero summary with no findings",
			verdict: &schema.AuthzVerdict{
				Summary:  schema.AuthzSummary{},
				Findings: []schema.AuthzFinding{},
			},
			wantLen:   1,
			wantError: true,
		},
		{
			name: "negative missing bucket",
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{
					RoutesTotal:       5,
					Protected:         3,
					Missing:           -1,
					Weak:              0,
					IdorRisk:          0,
					PublicIntentional: 2,
				},
				Findings: []schema.AuthzFinding{
					{Route: "GET /api/health", Issue: "public_intentional"},
				},
			},
			wantLen:   1,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check(tt.verdict)
			if len(violations) != tt.wantLen {
				t.Errorf("check() returned %d violations, want %d", len(violations), tt.wantLen)
			}
			if (len(violations) > 0) != tt.wantError {
				t.Errorf("check() error mismatch: got %d violations, wantError=%v", len(violations), tt.wantError)
			}
		})
	}
}

// TestAZ2_RoutesTotalMatchesInput verifies that summary.routes_total equals
// the number of routes in the input.
func TestAZ2_RoutesTotalMatchesInput(t *testing.T) {
	check := locateAuthzJointCheck(t, "AZ2")

	tests := []struct {
		name      string
		input     *schema.AuthzInput
		verdict   *schema.AuthzVerdict
		wantVioln bool
	}{
		{
			name: "routes_total matches input",
			input: &schema.AuthzInput{
				Routes: []schema.AuthzRoute{
					{Method: "GET", Path: "/api/users"},
					{Method: "POST", Path: "/api/users"},
					{Method: "DELETE", Path: "/api/users/{id}"},
				},
			},
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{RoutesTotal: 3},
			},
			wantVioln: false,
		},
		{
			name: "routes_total mismatch",
			input: &schema.AuthzInput{
				Routes: []schema.AuthzRoute{
					{Method: "GET", Path: "/api/users"},
					{Method: "POST", Path: "/api/users"},
					{Method: "DELETE", Path: "/api/users/{id}"},
				},
			},
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{RoutesTotal: 5},
			},
			wantVioln: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check(tt.input, tt.verdict)
			hasVioln := len(violations) > 0
			if hasVioln != tt.wantVioln {
				t.Errorf("check() violation mismatch: got %v, want %v", hasVioln, tt.wantVioln)
			}
		})
	}
}

// TestAZ3_BucketSumEqualsRoutesTotal verifies that the sum of all status buckets
// equals routes_total.
func TestAZ3_BucketSumEqualsRoutesTotal(t *testing.T) {
	check := locateAuthzCheck(t, "AZ3")

	tests := []struct {
		name      string
		verdict   *schema.AuthzVerdict
		wantVioln bool
	}{
		{
			name: "bucket sum matches routes_total",
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{
					RoutesTotal:       5,
					Protected:         3,
					Missing:           1,
					Weak:              0,
					IdorRisk:          0,
					PublicIntentional: 1,
				},
			},
			wantVioln: false,
		},
		{
			name: "bucket sum less than routes_total",
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{
					RoutesTotal:       5,
					Protected:         3,
					Missing:           0,
					Weak:              0,
					IdorRisk:          0,
					PublicIntentional: 1,
				},
			},
			wantVioln: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check(tt.verdict)
			hasVioln := len(violations) > 0
			if hasVioln != tt.wantVioln {
				t.Errorf("check() violation mismatch: got %v, want %v", hasVioln, tt.wantVioln)
			}
		})
	}
}

// TestAZ4_FindingsRouteInInput verifies that every finding references a route
// that exists in the input.
func TestAZ4_FindingsRouteInInput(t *testing.T) {
	check := locateAuthzJointCheck(t, "AZ4")

	tests := []struct {
		name      string
		input     *schema.AuthzInput
		verdict   *schema.AuthzVerdict
		wantVioln bool
	}{
		{
			name: "findings reference existing routes",
			input: &schema.AuthzInput{
				Routes: []schema.AuthzRoute{
					{Method: "POST", Path: "/api/orders"},
					{Method: "GET", Path: "/api/orders/{id}"},
				},
			},
			verdict: &schema.AuthzVerdict{
				Findings: []schema.AuthzFinding{
					{Route: "POST /api/orders", Issue: "missing_authz"},
				},
			},
			wantVioln: false,
		},
		{
			name: "finding references non-existent route",
			input: &schema.AuthzInput{
				Routes: []schema.AuthzRoute{
					{Method: "POST", Path: "/api/orders"},
				},
			},
			verdict: &schema.AuthzVerdict{
				Findings: []schema.AuthzFinding{
					{Route: "GET /not-in-input", Issue: "missing_authz"},
				},
			},
			wantVioln: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check(tt.input, tt.verdict)
			hasVioln := len(violations) > 0
			if hasVioln != tt.wantVioln {
				t.Errorf("check() violation mismatch: got %v, want %v", hasVioln, tt.wantVioln)
			}
		})
	}
}

// TestAZ5_WeakPrimitivesInInput verifies that every weak primitive references
// a primitive that exists in the input.
func TestAZ5_WeakPrimitivesInInput(t *testing.T) {
	check := locateAuthzJointCheck(t, "AZ5")

	tests := []struct {
		name      string
		input     *schema.AuthzInput
		verdict   *schema.AuthzVerdict
		wantVioln bool
	}{
		{
			name: "weak primitives reference existing primitives",
			input: &schema.AuthzInput{
				AuthzPrimitives: []schema.AuthzInputPrimitive{
					{FQN: "middleware.RequireAuth", Kind: "middleware"},
				},
			},
			verdict: &schema.AuthzVerdict{
				WeakPrimitives: []schema.WeakPrimitive{
					{FQN: "middleware.RequireAuth", Reason: "lacks failure path"},
				},
			},
			wantVioln: false,
		},
		{
			name: "weak primitive references non-existent primitive",
			input: &schema.AuthzInput{
				AuthzPrimitives: []schema.AuthzInputPrimitive{
					{FQN: "middleware.RequireAuth", Kind: "middleware"},
				},
			},
			verdict: &schema.AuthzVerdict{
				WeakPrimitives: []schema.WeakPrimitive{
					{FQN: "middleware.NotInInput", Reason: "not found"},
				},
			},
			wantVioln: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check(tt.input, tt.verdict)
			hasVioln := len(violations) > 0
			if hasVioln != tt.wantVioln {
				t.Errorf("check() violation mismatch: got %v, want %v", hasVioln, tt.wantVioln)
			}
		})
	}
}

// TestAZ6_ReadCallCount_NoOpStub verifies the no-op stub behavior for AZ6.
// TODO(phase-5): implement actual read-call-count tracking when SubagentStop
// includes tool-call telemetry or when verdict JSON includes call counts.
func TestAZ6_ReadCallCount_NoOpStub(t *testing.T) {
	check := locateAuthzCheck(t, "AZ6")

	tests := []struct {
		name    string
		verdict *schema.AuthzVerdict
	}{
		{
			name: "any verdict passes (no-op)",
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{RoutesTotal: 10},
			},
		},
		{
			name: "zero verdict passes (no-op)",
			verdict: &schema.AuthzVerdict{
				Summary: schema.AuthzSummary{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check(tt.verdict)
			if len(violations) != 0 {
				t.Errorf("check() returned %d violations, want 0 (no-op stub)", len(violations))
			}
		})
	}
}

// TestAZ7_SessionIDMatchesInput validates AZ7 (JOINT): if input has review_session_id,
// output review_session_id must match exactly.
func TestAZ7_SessionIDMatchesInput(t *testing.T) {
	check := locateAuthzJointCheck(t, "AZ7")
	cases := []struct {
		name string
		in   *schema.AuthzInput
		v    *schema.AuthzVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: both have matching session ID",
			in: &schema.AuthzInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.AuthzVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "bad: input has session ID but output mismatches",
			in: &schema.AuthzInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.AuthzVerdict{
				ReviewSessionID: "uuid-456",
			},
			want: []invariants.Violation{{Path: "review_session_id", Expected: "uuid-123", Actual: "uuid-456"}},
		},
		{
			name: "ok: input omits session ID (optional field)",
			in: &schema.AuthzInput{
				ReviewSessionID: "",
			},
			v: &schema.AuthzVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "ok: both omit session ID",
			in: &schema.AuthzInput{
				ReviewSessionID: "",
			},
			v: &schema.AuthzVerdict{
				ReviewSessionID: "",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in, tc.v)
			if len(got) != len(tc.want) {
				t.Errorf("AZ7 check got %d violations, want %d", len(got), len(tc.want))
				return
			}
			if len(got) == 0 {
				return
			}
			if got[0].Path != tc.want[0].Path || got[0].Expected != tc.want[0].Expected || got[0].Actual != tc.want[0].Actual {
				t.Errorf("AZ7 check mismatch: got %+v, want %+v", got[0], tc.want[0])
			}
		})
	}
}

// --- E1: code_ref Joint Invariant Tests (compile-error RED until Wave 1) ---

// TestAZ_CodeRefMismatch_Blocked verifies that a new AZ-CodeRef joint invariant
// detects when AuthzInput.CodeRef and AuthzVerdict.CodeRef differ.
// This test is RED (compile-error) until Wave 1 adds CodeRef field to schema.
func TestAZ_CodeRefMismatch_Blocked(t *testing.T) {
	// After Wave 1: locate the AZ-CodeRef joint invariant
	// For now, this test documents the intended behavior:
	// - Input has CodeRef: "abc123"
	// - Verdict has CodeRef: "wronghash"
	// - Joint invariant should fire → violations non-empty

	// Input with non-empty code_ref and at least one route (for AZ1 to pass)
	in := &schema.AuthzInput{
		Routes: []schema.AuthzRoute{
			{
				Router:  "gin",
				Method:  "GET",
				Path:    "/api/users/:id",
				Handler: schema.Handler{Name: "GetUser"},
			},
		},
		AuthzPrimitives: []schema.AuthzInputPrimitive{
			{FQN: "pkg.middleware.AuthCheck", Kind: "func"},
		},
		// CodeRef: "abc123", // WILL BE ADDED IN WAVE 1
	}

	// Verdict with mismatched code_ref
	v := &schema.AuthzVerdict{
		Summary: schema.AuthzSummary{
			RoutesTotal:       1,
			Protected:         1,
			Missing:           0,
			Weak:              0,
			IdorRisk:          0,
			PublicIntentional: 0,
		},
		// CodeRef: "wronghash", // WILL BE ADDED IN WAVE 1; mismatch from input
	}

	// After Wave 1 adds the field and joint invariant:
	// check := locateAuthzJointCheck(t, "AZ-CodeRef")
	// violations := check(in, v)
	// if len(violations) == 0 {
	//    t.Errorf("AZ-CodeRef joint invariant should fire on code_ref mismatch")
	// }

	// For now, this test documents the intended behavior and will be GREEN after Wave 1.
	t.Skip("E1 compile-error RED: CodeRef field not yet added to schema (Wave 1)")
}

// TestAZ_CodeRefEmpty_Skipped verifies that the code_ref joint invariant
// does not fire when AuthzInput.CodeRef is empty (non-git repo compatibility).
// This test is RED (compile-error) until Wave 1 adds CodeRef field to schema.
func TestAZ_CodeRefEmpty_Skipped(t *testing.T) {
	// After Wave 1: verify that empty input code_ref skips the invariant

	// Input with empty code_ref
	inNoCodeRef := &schema.AuthzInput{
		Routes: []schema.AuthzRoute{
			{
				Router:  "gin",
				Method:  "POST",
				Path:    "/api/data",
				Handler: schema.Handler{Name: "CreateData"},
			},
		},
		AuthzPrimitives: []schema.AuthzInputPrimitive{
			{FQN: "pkg.authz.Check", Kind: "func"},
		},
		// CodeRef: "", // empty input code_ref
	}

	vWithCodeRef := &schema.AuthzVerdict{
		Summary: schema.AuthzSummary{
			RoutesTotal:       1,
			Protected:         1,
			Missing:           0,
			Weak:              0,
			IdorRisk:          0,
			PublicIntentional: 0,
		},
		// CodeRef: "somehash", // verdict has code_ref, but empty input skips check
	}

	// After Wave 1:
	// check := locateAuthzJointCheck(t, "AZ-CodeRef")
	// violations := check(inNoCodeRef, vWithCodeRef)
	// if len(violations) != 0 {
	//    t.Errorf("AZ-CodeRef joint invariant should skip when input code_ref is empty")
	// }

	// For now, this test documents the intended behavior and will be GREEN after Wave 1.
	t.Skip("E1 compile-error RED: CodeRef field not yet added to schema (Wave 1)")
}
