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
