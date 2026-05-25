package hooks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/invariants"
	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// runAgentValidation dispatches based on subagentType. Returns the list of
// formatted block-reason fragments (one per Violation) in registration order.
// Empty []string means the verdict passed all invariants.
// err != nil signals Class B failure.
func runAgentValidation(subagentType, verdictText, inputText string) ([]string, error) {
	switch subagentType {
	case "go-cartographer":
		return runCartographer(verdictText)
	case "go-taint-tracer":
		return runTaintTracer(verdictText, inputText)
	case "go-authz-tracer":
		return runAuthzTracer(verdictText, inputText)
	case "go-oauth-auditor":
		return runOAuthAuditor(verdictText, inputText)
	case "invariant-checker":
		return runInvariantChecker(verdictText, inputText)
	case "synthesis":
		return runSynthesis(verdictText)
	default:
		return nil, fmt.Errorf("unknown security subagent: %s", subagentType)
	}
}

// newStrictDecoder returns a *json.Decoder configured with DisallowUnknownFields
// bound to text. Each per-agent helper holds onto a concrete typed local
// variable and calls dec.Decode(&typedVar) — no interface{} parameter ever.
func newStrictDecoder(text string) *json.Decoder {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.DisallowUnknownFields()
	return dec
}

func runCartographer(verdictText string) ([]string, error) {
	var idx schema.CartographerIndex
	if err := newStrictDecoder(verdictText).Decode(&idx); err != nil {
		return []string{FormatViolation("A1", "cartographer verdict parse", invariants.Violation{Path: "", Expected: "", Actual: err.Error()})}, nil
	}
	var reasons []string
	for _, inv := range invariants.CartographerInvariants {
		for _, vio := range inv.Check(&idx) {
			reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
		}
	}
	return reasons, nil
}

func runTaintTracer(verdictText, inputText string) ([]string, error) {
	var v schema.TaintVerdict
	if err := newStrictDecoder(verdictText).Decode(&v); err != nil {
		return []string{FormatViolation("T1", "taint verdict parse", invariants.Violation{Path: "", Expected: "", Actual: err.Error()})}, nil
	}
	var reasons []string
	for _, inv := range invariants.TaintTracerInvariants {
		for _, vio := range inv.Check(&v) {
			reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
		}
	}
	if inputText != "" {
		var in schema.TaintInput
		if err := newStrictDecoder(inputText).Decode(&in); err == nil {
			for _, inv := range invariants.TaintTracerJointInvariants {
				for _, vio := range inv.Check(&in, &v) {
					reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
				}
			}
		}
	}
	return reasons, nil
}

func runAuthzTracer(verdictText, inputText string) ([]string, error) {
	var v schema.AuthzVerdict
	if err := newStrictDecoder(verdictText).Decode(&v); err != nil {
		return []string{FormatViolation("AZ1", "authz verdict parse", invariants.Violation{Path: "", Expected: "", Actual: err.Error()})}, nil
	}
	var reasons []string
	for _, inv := range invariants.AuthzInvariants {
		for _, vio := range inv.Check(&v) {
			reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
		}
	}
	if inputText != "" {
		var in schema.AuthzInput
		if err := newStrictDecoder(inputText).Decode(&in); err == nil {
			for _, inv := range invariants.AuthzJointInvariants {
				for _, vio := range inv.Check(&in, &v) {
					reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
				}
			}
		}
	}
	return reasons, nil
}

func runOAuthAuditor(verdictText, inputText string) ([]string, error) {
	var v schema.OAuthVerdict
	if err := newStrictDecoder(verdictText).Decode(&v); err != nil {
		return []string{FormatViolation("OA1", "oauth verdict parse", invariants.Violation{Path: "", Expected: "", Actual: err.Error()})}, nil
	}
	var reasons []string
	for _, inv := range invariants.OAuthInvariants {
		for _, vio := range inv.Check(&v) {
			reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
		}
	}
	if inputText != "" {
		var in schema.OAuthInput
		if err := newStrictDecoder(inputText).Decode(&in); err == nil {
			for _, inv := range invariants.OAuthJointInvariants {
				for _, vio := range inv.Check(&in, &v) {
					reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
				}
			}
		}
	}
	return reasons, nil
}

func runInvariantChecker(verdictText, inputText string) ([]string, error) {
	var v schema.InvariantCheckerVerdict
	if err := newStrictDecoder(verdictText).Decode(&v); err != nil {
		return []string{FormatViolation("IC1", "invariant-checker verdict parse", invariants.Violation{Path: "", Expected: "", Actual: err.Error()})}, nil
	}
	var reasons []string
	for _, inv := range invariants.InvariantCheckerInvariants {
		for _, vio := range inv.Check(&v) {
			reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
		}
	}
	if inputText != "" {
		var in schema.InvariantCheckerInput
		if err := newStrictDecoder(inputText).Decode(&in); err == nil {
			for _, inv := range invariants.InvariantCheckerJointInvariants {
				for _, vio := range inv.Check(&in, &v) {
					reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
				}
			}
		}
	}
	return reasons, nil
}

func runSynthesis(verdictText string) ([]string, error) {
	var r schema.SynthesisReport
	if err := newStrictDecoder(verdictText).Decode(&r); err != nil {
		return []string{FormatViolation("S2", "synthesis verdict parse", invariants.Violation{Path: "", Expected: "", Actual: err.Error()})}, nil
	}
	var reasons []string
	for _, inv := range invariants.SynthesisInvariants {
		for _, vio := range inv.Check(&r) {
			reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
		}
	}
	return reasons, nil
}
