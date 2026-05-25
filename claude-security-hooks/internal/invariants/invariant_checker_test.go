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
