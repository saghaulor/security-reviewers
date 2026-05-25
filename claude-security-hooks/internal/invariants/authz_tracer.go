package invariants

import (
	"fmt"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// AuthzInvariant represents a verdict-only invariant check for the authz tracer.
type AuthzInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.AuthzVerdict) []Violation
}

// AuthzJointInvariant represents a joint invariant check that requires both
// input and verdict for the authz tracer.
type AuthzJointInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.AuthzInput, *schema.AuthzVerdict) []Violation
}

// AuthzInvariants is the registry of verdict-only authz assertions.
var AuthzInvariants = []AuthzInvariant{
	{
		ID:          "AZ1",
		Description: "verdict has required fields and non-negative bucket counts",
		Severity:    SeverityCritical,
		Check:       checkAZ1,
	},
	{
		ID:          "AZ3",
		Description: "summary buckets sum to routes_total (accounting invariant)",
		Severity:    SeverityHigh,
		Check:       checkAZ3,
	},
	{
		ID:          "AZ6",
		Description: "agent read-call count within budget (no-op stub Phase 2; real check Phase 5)",
		Severity:    SeverityInfo,
		Check:       checkAZ6,
	},
}

// AuthzJointInvariants is the registry of joint authz assertions.
var AuthzJointInvariants = []AuthzJointInvariant{
	{
		ID:          "AZ2",
		Description: "summary.routes_total equals len(input.routes)",
		Severity:    SeverityHigh,
		Check:       checkAZ2Joint,
	},
	{
		ID:          "AZ4",
		Description: "every findings[*].route exists in input.routes",
		Severity:    SeverityHigh,
		Check:       checkAZ4Joint,
	},
	{
		ID:          "AZ5",
		Description: "every weak_primitives[*].fqn exists in input.authz_primitives",
		Severity:    SeverityHigh,
		Check:       checkAZ5Joint,
	},
	{
		ID:          "AZ7",
		Description: "if input has review_session_id, verdict review_session_id must match exactly",
		Severity:    SeverityHigh,
		Check:       checkAZ7Joint,
	},
}

// checkAZ1 verifies that the verdict has required fields and non-negative counts.
func checkAZ1(v *schema.AuthzVerdict) []Violation {
	var out []Violation

	// Check if summary is all-zero and no findings present
	if v.Summary.RoutesTotal == 0 && len(v.Findings) == 0 {
		out = append(out, Violation{
			Path:     "summary",
			Expected: "populated",
			Actual:   "all-zero",
		})
	}

	// Check all bucket counts for non-negative values
	type bucket struct {
		name string
		val  int
	}
	for _, b := range []bucket{
		{"protected", v.Summary.Protected},
		{"missing", v.Summary.Missing},
		{"weak", v.Summary.Weak},
		{"idor_risk", v.Summary.IdorRisk},
		{"public_intentional", v.Summary.PublicIntentional},
		{"routes_total", v.Summary.RoutesTotal},
	} {
		if b.val < 0 {
			out = append(out, Violation{
				Path:     "summary." + b.name,
				Expected: ">=0",
				Actual:   fmt.Sprintf("%d", b.val),
			})
		}
	}

	return out
}

// checkAZ2Joint verifies that summary.routes_total equals the number of input routes.
func checkAZ2Joint(in *schema.AuthzInput, v *schema.AuthzVerdict) []Violation {
	want := len(in.Routes)
	if v.Summary.RoutesTotal != want {
		return []Violation{{
			Path:     "summary.routes_total",
			Expected: fmt.Sprintf("%d", want),
			Actual:   fmt.Sprintf("%d", v.Summary.RoutesTotal),
		}}
	}
	return nil
}

// checkAZ3 verifies that the sum of all status buckets equals routes_total.
func checkAZ3(v *schema.AuthzVerdict) []Violation {
	sum := v.Summary.Protected + v.Summary.Missing + v.Summary.Weak +
		v.Summary.IdorRisk + v.Summary.PublicIntentional
	if sum != v.Summary.RoutesTotal {
		return []Violation{{
			Path: "summary",
			Expected: fmt.Sprintf(
				"protected+missing+weak+idor_risk+public_intentional==routes_total (%d)",
				v.Summary.RoutesTotal,
			),
			Actual: fmt.Sprintf("%d", sum),
		}}
	}
	return nil
}

// checkAZ4Joint verifies that every finding references a route present in the input.
func checkAZ4Joint(in *schema.AuthzInput, v *schema.AuthzVerdict) []Violation {
	// Build a map of input routes by their composite key (Method + " " + Path)
	inputRoutes := make(map[string]struct{}, len(in.Routes))
	for _, r := range in.Routes {
		inputRoutes[r.Method+" "+r.Path] = struct{}{}
	}

	var out []Violation
	for i, f := range v.Findings {
		if _, ok := inputRoutes[f.Route]; !ok {
			out = append(out, Violation{
				Path:     fmt.Sprintf("findings[%d].route", i),
				Expected: "present in input.routes",
				Actual:   f.Route,
			})
		}
	}
	return out
}

// checkAZ5Joint verifies that every weak primitive references a primitive
// present in the input.
func checkAZ5Joint(in *schema.AuthzInput, v *schema.AuthzVerdict) []Violation {
	// Build a map of input primitives by FQN
	primSet := make(map[string]struct{}, len(in.AuthzPrimitives))
	for _, p := range in.AuthzPrimitives {
		primSet[p.FQN] = struct{}{}
	}

	var out []Violation
	for i, wp := range v.WeakPrimitives {
		if _, ok := primSet[wp.FQN]; !ok {
			out = append(out, Violation{
				Path:     fmt.Sprintf("weak_primitives[%d].fqn", i),
				Expected: "present in input.authz_primitives",
				Actual:   wp.FQN,
			})
		}
	}
	return out
}

// checkAZ6 is a no-op stub per Phase 2 decision.
// TODO(phase-5): enforce via SubagentStop telemetry or tool-call summary in verdict.
func checkAZ6(v *schema.AuthzVerdict) []Violation {
	_ = v
	return nil
}

// --- AZ7 (JOINT) ---
func checkAZ7Joint(in *schema.AuthzInput, v *schema.AuthzVerdict) []Violation {
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
