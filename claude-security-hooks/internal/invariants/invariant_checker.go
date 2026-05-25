package invariants

import (
	"fmt"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

type InvariantCheckerInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.InvariantCheckerVerdict) []Violation
}

type InvariantCheckerJointInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.InvariantCheckerInput, *schema.InvariantCheckerVerdict) []Violation
}

var InvariantCheckerInvariants = []InvariantCheckerInvariant{
	{ID: "IC1", Description: "verdict has required fields (flow_name + results)", Severity: SeverityCritical, Check: checkIC1},
	{ID: "IC3", Description: "status=violated requires evidence.files non-empty AND evidence.explanation non-empty", Severity: SeverityHigh, Check: checkIC3},
}

var InvariantCheckerJointInvariants = []InvariantCheckerJointInvariant{
	{ID: "IC2", Description: "exactly one result per input invariant; IDs match input set", Severity: SeverityHigh, Check: checkIC2Joint},
	{ID: "IC4", Description: "results introduce no new invariant IDs (D-16: no discovery)", Severity: SeverityHigh, Check: checkIC4Joint},
	{ID: "IC5", Description: "if input has review_session_id, verdict review_session_id must match exactly", Severity: SeverityHigh, Check: checkIC5Joint},
}

// --- IC1 ---
func checkIC1(v *schema.InvariantCheckerVerdict) []Violation {
	var out []Violation
	if v.FlowName == "" {
		out = append(out, Violation{Path: "flow_name", Expected: "non-empty", Actual: ""})
	}
	if v.Results == nil {
		out = append(out, Violation{Path: "results", Expected: "present", Actual: "nil"})
	}
	return out
}

// --- IC2 (joint) ---
func checkIC2Joint(in *schema.InvariantCheckerInput, v *schema.InvariantCheckerVerdict) []Violation {
	inputIDs := make(map[string]int, len(in.Invariants))
	for _, inv := range in.Invariants {
		inputIDs[inv.ID]++
	}
	resultIDs := make(map[string]int, len(v.Results))
	for _, r := range v.Results {
		resultIDs[r.InvariantID]++
	}
	var out []Violation
	// Every input invariant must appear exactly once in results.
	for id := range inputIDs {
		switch resultIDs[id] {
		case 0:
			out = append(out, Violation{Path: "results", Expected: fmt.Sprintf("contains invariant_id %q", id), Actual: "missing"})
		case 1:
			// ok
		default:
			out = append(out, Violation{Path: "results", Expected: "exactly one result per input invariant", Actual: fmt.Sprintf("duplicate %q", id)})
		}
	}
	return out
}

// --- IC3 ---
func checkIC3(v *schema.InvariantCheckerVerdict) []Violation {
	var out []Violation
	for i, r := range v.Results {
		if r.Status != "violated" {
			continue
		}
		if len(r.Evidence.Files) == 0 {
			out = append(out, Violation{Path: fmt.Sprintf("results[%d].evidence.files", i), Expected: "non-empty (status=violated)", Actual: "empty"})
		}
		if r.Evidence.Explanation == "" {
			out = append(out, Violation{Path: fmt.Sprintf("results[%d].evidence.explanation", i), Expected: "non-empty (status=violated)", Actual: ""})
		}
	}
	return out
}

// --- IC4 (joint) ---
func checkIC4Joint(in *schema.InvariantCheckerInput, v *schema.InvariantCheckerVerdict) []Violation {
	inputSet := make(map[string]struct{}, len(in.Invariants))
	for _, inv := range in.Invariants {
		inputSet[inv.ID] = struct{}{}
	}
	var out []Violation
	for i, r := range v.Results {
		if _, ok := inputSet[r.InvariantID]; !ok {
			out = append(out, Violation{Path: fmt.Sprintf("results[%d].invariant_id", i), Expected: "present in input.invariants ids", Actual: r.InvariantID})
		}
	}
	return out
}

// --- IC5 (JOINT) ---
func checkIC5Joint(in *schema.InvariantCheckerInput, v *schema.InvariantCheckerVerdict) []Violation {
	// If input does not specify a session ID, no check is performed (field is optional).
	if in.ReviewSessionID == "" {
		return nil
	}
	// If input specifies a session ID, verdict must echo it exactly.
	if v.ReviewSessionID != in.ReviewSessionID {
		return []Violation{{
			Path:     "review_session_id",
			Expected: in.ReviewSessionID,
			Actual:   v.ReviewSessionID,
		}}
	}
	return nil
}
