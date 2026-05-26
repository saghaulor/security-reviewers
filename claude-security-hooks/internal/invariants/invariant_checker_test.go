package invariants_test

import (
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/invariants"
	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// locateICCheck finds a check in the verdict-only IC list by ID.
func locateICCheck(id string) *invariants.InvariantCheckerInvariant {
	for i := range invariants.InvariantCheckerInvariants {
		if invariants.InvariantCheckerInvariants[i].ID == id {
			return &invariants.InvariantCheckerInvariants[i]
		}
	}
	return nil
}

// locateICJointCheck finds a check in the joint IC list by ID.
func locateICJointCheck(id string) *invariants.InvariantCheckerJointInvariant {
	for i := range invariants.InvariantCheckerJointInvariants {
		if invariants.InvariantCheckerJointInvariants[i].ID == id {
			return &invariants.InvariantCheckerJointInvariants[i]
		}
	}
	return nil
}

// --- IC1: Verdict required fields ---

func TestIC1_VerdictRequiredFields(t *testing.T) {
	check := locateICCheck("IC1")
	if check == nil {
		t.Fatalf("IC1 check not found")
	}

	tests := []struct {
		name           string
		verdict        *schema.InvariantCheckerVerdict
		wantViolations bool
	}{
		{
			name: "valid verdict",
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results:  []schema.InvariantResult{},
			},
			wantViolations: false,
		},
		{
			name: "empty flow_name",
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "",
				Results:  []schema.InvariantResult{},
			},
			wantViolations: true,
		},
		{
			name: "nil results",
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results:  nil,
			},
			wantViolations: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.verdict)
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("got violations=%v, want %v; violations=%v", hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// --- IC2: Results match input invariants (joint) ---

func TestIC2_ResultsMatchInputInvariants(t *testing.T) {
	check := locateICJointCheck("IC2")
	if check == nil {
		t.Fatalf("IC2 check not found")
	}

	tests := []struct {
		name           string
		input          *schema.InvariantCheckerInput
		verdict        *schema.InvariantCheckerVerdict
		wantViolations bool
	}{
		{
			name: "exact match: two invariants, two results",
			input: &schema.InvariantCheckerInput{
				FlowName: "checkout",
				Invariants: []schema.InputInvariant{
					{ID: "inv-a", Statement: "test a"},
					{ID: "inv-b", Statement: "test b"},
				},
			},
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{InvariantID: "inv-a", Status: "holds"},
					{InvariantID: "inv-b", Status: "holds"},
				},
			},
			wantViolations: false,
		},
		{
			name: "missing result: input has inv-b but verdict has only inv-a",
			input: &schema.InvariantCheckerInput{
				FlowName: "checkout",
				Invariants: []schema.InputInvariant{
					{ID: "inv-a", Statement: "test a"},
					{ID: "inv-b", Statement: "test b"},
				},
			},
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{InvariantID: "inv-a", Status: "holds"},
				},
			},
			wantViolations: true,
		},
		{
			name: "duplicate result: inv-a appears twice",
			input: &schema.InvariantCheckerInput{
				FlowName: "checkout",
				Invariants: []schema.InputInvariant{
					{ID: "inv-a", Statement: "test a"},
				},
			},
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{InvariantID: "inv-a", Status: "holds"},
					{InvariantID: "inv-a", Status: "violated"},
				},
			},
			wantViolations: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.input, tt.verdict)
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("got violations=%v, want %v; violations=%v", hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// --- IC3: Violated status requires evidence ---

func TestIC3_ViolatedRequiresEvidence(t *testing.T) {
	check := locateICCheck("IC3")
	if check == nil {
		t.Fatalf("IC3 check not found")
	}

	tests := []struct {
		name           string
		verdict        *schema.InvariantCheckerVerdict
		wantViolations bool
	}{
		{
			name: "violated with full evidence",
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{
						InvariantID: "test-1",
						Status:      "violated",
						Evidence: schema.InvariantEvidence{
							Files:       []string{"handler.go"},
							Explanation: "price reads from client input",
						},
					},
				},
			},
			wantViolations: false,
		},
		{
			name: "violated with empty files",
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{
						InvariantID: "test-1",
						Status:      "violated",
						Evidence: schema.InvariantEvidence{
							Files:       nil,
							Explanation: "some explanation",
						},
					},
				},
			},
			wantViolations: true,
		},
		{
			name: "violated with empty explanation",
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{
						InvariantID: "test-1",
						Status:      "violated",
						Evidence: schema.InvariantEvidence{
							Files:       []string{"handler.go"},
							Explanation: "",
						},
					},
				},
			},
			wantViolations: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.verdict)
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("got violations=%v, want %v; violations=%v", hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// --- IC4: No new invariant IDs (joint) ---

func TestIC4_NoNewInvariantIDs(t *testing.T) {
	check := locateICJointCheck("IC4")
	if check == nil {
		t.Fatalf("IC4 check not found")
	}

	tests := []struct {
		name           string
		input          *schema.InvariantCheckerInput
		verdict        *schema.InvariantCheckerVerdict
		wantViolations bool
	}{
		{
			name: "no new IDs: result matches input",
			input: &schema.InvariantCheckerInput{
				FlowName: "checkout",
				Invariants: []schema.InputInvariant{
					{ID: "inv-a", Statement: "test a"},
				},
			},
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{InvariantID: "inv-a", Status: "holds"},
				},
			},
			wantViolations: false,
		},
		{
			name: "new ID discovered: inv-NEW-discovered not in input",
			input: &schema.InvariantCheckerInput{
				FlowName: "checkout",
				Invariants: []schema.InputInvariant{
					{ID: "inv-a", Statement: "test a"},
				},
			},
			verdict: &schema.InvariantCheckerVerdict{
				FlowName: "checkout",
				Results: []schema.InvariantResult{
					{InvariantID: "inv-NEW-discovered", Status: "holds"},
				},
			},
			wantViolations: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.input, tt.verdict)
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("got violations=%v, want %v; violations=%v", hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestIC5_SessionIDMatchesInput validates IC5 (JOINT): if input has review_session_id,
// output review_session_id must match exactly.
func TestIC5_SessionIDMatchesInput(t *testing.T) {
	check := locateICJointCheck("IC5")
	if check == nil {
		t.Fatalf("IC5 check not found")
	}
	cases := []struct {
		name string
		in   *schema.InvariantCheckerInput
		v    *schema.InvariantCheckerVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: both have matching session ID",
			in: &schema.InvariantCheckerInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.InvariantCheckerVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "bad: input has session ID but output mismatches",
			in: &schema.InvariantCheckerInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.InvariantCheckerVerdict{
				ReviewSessionID: "uuid-456",
			},
			want: []invariants.Violation{{Path: "review_session_id", Expected: "uuid-123", Actual: "uuid-456"}},
		},
		{
			name: "ok: input omits session ID (optional field)",
			in: &schema.InvariantCheckerInput{
				ReviewSessionID: "",
			},
			v: &schema.InvariantCheckerVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "ok: both omit session ID",
			in: &schema.InvariantCheckerInput{
				ReviewSessionID: "",
			},
			v: &schema.InvariantCheckerVerdict{
				ReviewSessionID: "",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check.Check(tc.in, tc.v)
			if len(got) != len(tc.want) {
				t.Errorf("IC5 check got %d violations, want %d", len(got), len(tc.want))
				return
			}
			if len(got) == 0 {
				return
			}
			if got[0].Path != tc.want[0].Path || got[0].Expected != tc.want[0].Expected || got[0].Actual != tc.want[0].Actual {
				t.Errorf("IC5 check mismatch: got %+v, want %+v", got[0], tc.want[0])
			}
		})
	}
}

// --- E1: code_ref Joint Invariant Tests (compile-error RED until Wave 1) ---

// TestIC_CodeRefMismatch_Blocked verifies that a new IC-CodeRef joint invariant
// detects when InvariantCheckerInput.CodeRef and InvariantCheckerVerdict.CodeRef differ.
// This test is RED (compile-error) until Wave 1 adds CodeRef field to schema.
func TestIC_CodeRefMismatch_Blocked(t *testing.T) {
	// After Wave 1: locate the IC-CodeRef joint invariant
	// For now, this test documents the intended behavior:
	// - Input has CodeRef: "abc123"
	// - Verdict has CodeRef: "wronghash"
	// - Joint invariant should fire → violations non-empty

	// Input with non-empty code_ref
	in := &schema.InvariantCheckerInput{
		FlowName: "payment_processing",
		Invariants: []schema.InputInvariant{
			{
				ID:        "I1",
				Statement: "Payment amount must be validated before processing",
				AnchorSymbols: []string{"processPayment"},
			},
		},
		// CodeRef: "abc123", // WILL BE ADDED IN WAVE 1
	}

	// Verdict with mismatched code_ref
	v := &schema.InvariantCheckerVerdict{
		FlowName: "payment_processing",
		Results: []schema.InvariantResult{
			{
				InvariantID: "I1",
				Status:      "satisfied",
				Confidence:  "high",
				Evidence: schema.InvariantEvidence{
					Files:       []string{"payment.go"},
					Explanation: "Amount validation found before sink",
				},
			},
		},
		// CodeRef: "wronghash", // WILL BE ADDED IN WAVE 1; mismatch from input
	}

	// After Wave 1 adds the field and joint invariant:
	// check := locateICJointCheck("IC-CodeRef")
	// if check == nil {
	//    t.Fatalf("IC-CodeRef joint invariant not found")
	// }
	// violations := check.Check(in, v)
	// if len(violations) == 0 {
	//    t.Errorf("IC-CodeRef joint invariant should fire on code_ref mismatch")
	// }

	// For now, this test documents the intended behavior and will be GREEN after Wave 1.
	t.Skip("E1 compile-error RED: CodeRef field not yet added to schema (Wave 1)")
}
