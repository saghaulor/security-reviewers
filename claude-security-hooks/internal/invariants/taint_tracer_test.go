package invariants_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/saghaulor/claude-security-hooks/internal/invariants"
	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// locateTaintCheck locates a verdict-only predicate from TaintTracerInvariants by ID.
func locateTaintCheck(t *testing.T, id string) func(*schema.TaintVerdict) []invariants.Violation {
	t.Helper()
	for _, inv := range invariants.TaintTracerInvariants {
		if inv.ID == id {
			return inv.Check
		}
	}
	t.Fatalf("%s not registered in TaintTracerInvariants", id)
	return nil
}

// locateTaintJointCheck locates a joint predicate from TaintTracerJointInvariants by ID.
func locateTaintJointCheck(t *testing.T, id string) func(*schema.TaintInput, *schema.TaintVerdict) []invariants.Violation {
	t.Helper()
	for _, inv := range invariants.TaintTracerJointInvariants {
		if inv.ID == id {
			return inv.Check
		}
	}
	t.Fatalf("%s not registered in TaintTracerJointInvariants", id)
	return nil
}

// TestT1_VerdictHasRequiredTopLevelFields validates T1: verdict and confidence must be non-empty.
func TestT1_VerdictHasRequiredTopLevelFields(t *testing.T) {
	check := locateTaintCheck(t, "T1")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: verdict and confidence both present",
			in: &schema.TaintVerdict{
				Verdict:    "exploitable",
				Confidence: "high",
			},
			want: nil,
		},
		{
			name: "bad: verdict empty",
			in: &schema.TaintVerdict{
				Verdict:    "",
				Confidence: "high",
			},
			want: []invariants.Violation{{Path: "verdict", Expected: "non-empty", Actual: ""}},
		},
		{
			name: "bad: confidence empty",
			in: &schema.TaintVerdict{
				Verdict:    "exploitable",
				Confidence: "",
			},
			want: []invariants.Violation{{Path: "confidence", Expected: "non-empty", Actual: ""}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T1 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT2_VerdictEnumValues validates T2: verdict is one of the canonical enum values.
func TestT2_VerdictEnumValues(t *testing.T) {
	check := locateTaintCheck(t, "T2")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: exploitable",
			in:   &schema.TaintVerdict{Verdict: "exploitable"},
			want: nil,
		},
		{
			name: "ok: sanitized",
			in:   &schema.TaintVerdict{Verdict: "sanitized"},
			want: nil,
		},
		{
			name: "ok: unreachable",
			in:   &schema.TaintVerdict{Verdict: "unreachable"},
			want: nil,
		},
		{
			name: "ok: ambiguous",
			in:   &schema.TaintVerdict{Verdict: "ambiguous"},
			want: nil,
		},
		{
			name: "ok: input_mismatch",
			in:   &schema.TaintVerdict{Verdict: "input_mismatch"},
			want: nil,
		},
		{
			name: "bad: unknown verdict",
			in:   &schema.TaintVerdict{Verdict: "maybe"},
			want: []invariants.Violation{{Path: "verdict", Expected: "in {exploitable,sanitized,unreachable,ambiguous,input_mismatch}", Actual: "maybe"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T2 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT3_PathNonEmptySourceFirstSinkLast validates T3: sanitized/exploitable verdicts require
// non-empty path with source first and sink last. This test is VERBATIM from RESEARCH.md Pattern 4.
func TestT3_PathNonEmptySourceFirstSinkLast(t *testing.T) {
	check := locateTaintCheck(t, "T3")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: source first, sink last, exploitable",
			in: &schema.TaintVerdict{
				Verdict: "exploitable",
				Path: []schema.TaintPathStep{{Step: "source"}, {Step: "call"}, {Step: "sink"}},
			},
			want: nil,
		},
		{
			name: "bad: first step is not source",
			in: &schema.TaintVerdict{
				Verdict: "exploitable",
				Path: []schema.TaintPathStep{{Step: "sink"}, {Step: "sink"}},
			},
			want: []invariants.Violation{{Path: "path[0].step", Expected: "source", Actual: "sink"}},
		},
		{
			name: "bad: empty path on sanitized verdict",
			in:   &schema.TaintVerdict{Verdict: "sanitized", Path: nil},
			want: []invariants.Violation{{Path: "path", Expected: "non-empty", Actual: "empty"}},
		},
		{
			name: "ok: input_mismatch permits empty path (T4 covers this)",
			in:   &schema.TaintVerdict{Verdict: "input_mismatch", Path: nil},
			want: nil,
		},
		{
			name: "bad: last step is not sink",
			in: &schema.TaintVerdict{
				Verdict: "sanitized",
				Path: []schema.TaintPathStep{{Step: "source"}, {Step: "call"}},
			},
			want: []invariants.Violation{{Path: "path[1].step", Expected: "sink", Actual: "call"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T3 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT4_InputMismatchPermitsEmptyPath validates T4: input_mismatch permits empty path.
func TestT4_InputMismatchPermitsEmptyPath(t *testing.T) {
	check := locateTaintCheck(t, "T4")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: input_mismatch with empty path",
			in:   &schema.TaintVerdict{Verdict: "input_mismatch", Path: nil},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T4 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT5_ConfidenceEnum validates T5: confidence is one of {high, medium, low}.
func TestT5_ConfidenceEnum(t *testing.T) {
	check := locateTaintCheck(t, "T5")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: high",
			in:   &schema.TaintVerdict{Confidence: "high"},
			want: nil,
		},
		{
			name: "ok: medium",
			in:   &schema.TaintVerdict{Confidence: "medium"},
			want: nil,
		},
		{
			name: "ok: low",
			in:   &schema.TaintVerdict{Confidence: "low"},
			want: nil,
		},
		{
			name: "bad: very high",
			in:   &schema.TaintVerdict{Confidence: "very high"},
			want: []invariants.Violation{{Path: "confidence", Expected: "in {high,medium,low}", Actual: "very high"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T5 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT6_SemgrepOrGoplsEvidence validates T6: require semgrep.ran or gopls evidence (unless input_mismatch).
func TestT6_SemgrepOrGoplsEvidence(t *testing.T) {
	check := locateTaintCheck(t, "T6")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: semgrep ran",
			in: &schema.TaintVerdict{
				Verdict: "exploitable",
				Semgrep: schema.SemgrepEvidence{Ran: true},
				Gopls:   schema.GoplsEvidence{ReferencesCalls: 0},
			},
			want: nil,
		},
		{
			name: "ok: gopls references",
			in: &schema.TaintVerdict{
				Verdict: "exploitable",
				Semgrep: schema.SemgrepEvidence{Ran: false},
				Gopls:   schema.GoplsEvidence{ReferencesCalls: 5},
			},
			want: nil,
		},
		{
			name: "ok: both semgrep and gopls",
			in: &schema.TaintVerdict{
				Verdict: "exploitable",
				Semgrep: schema.SemgrepEvidence{Ran: true},
				Gopls:   schema.GoplsEvidence{ReferencesCalls: 5},
			},
			want: nil,
		},
		{
			name: "ok: input_mismatch exemption (neither tool)",
			in: &schema.TaintVerdict{
				Verdict: "input_mismatch",
				Semgrep: schema.SemgrepEvidence{Ran: false},
				Gopls:   schema.GoplsEvidence{ReferencesCalls: 0},
			},
			want: nil,
		},
		{
			name: "bad: exploitable with neither tool",
			in: &schema.TaintVerdict{
				Verdict: "exploitable",
				Semgrep: schema.SemgrepEvidence{Ran: false},
				Gopls:   schema.GoplsEvidence{ReferencesCalls: 0},
			},
			want: []invariants.Violation{{
				Path:     "semgrep+gopls",
				Expected: "semgrep.ran=true OR gopls.references_calls>0 (unless input_mismatch or ambiguous)",
				Actual:   "neither",
			}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T6 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT7_SemgrepTierEqualsInputTier validates T7 (JOINT): semgrep.tier matches input tier.
func TestT7_SemgrepTierEqualsInputTier(t *testing.T) {
	check := locateTaintJointCheck(t, "T7")
	cases := []struct {
		name string
		in   *schema.TaintInput
		v    *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: pro tier matches",
			in: &schema.TaintInput{
				SemgrepTier: "pro",
			},
			v: &schema.TaintVerdict{
				Semgrep: schema.SemgrepEvidence{Tier: "pro"},
			},
			want: nil,
		},
		{
			name: "bad: pro vs ce",
			in: &schema.TaintInput{
				SemgrepTier: "pro",
			},
			v: &schema.TaintVerdict{
				Semgrep: schema.SemgrepEvidence{Tier: "ce"},
			},
			want: []invariants.Violation{{Path: "semgrep.tier", Expected: "pro", Actual: "ce"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T7 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT8_InterfaceSourceRequiresImplementationOrNotes validates T8 (JOINT): interface sources
// require implementation_calls or explanatory notes.
func TestT8_InterfaceSourceRequiresImplementationOrNotes(t *testing.T) {
	check := locateTaintJointCheck(t, "T8")
	cases := []struct {
		name string
		in   *schema.TaintInput
		v    *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: interface with implementation calls",
			in: &schema.TaintInput{
				Source: schema.TaintEndpoint{Kind: "iface"},
			},
			v: &schema.TaintVerdict{
				Gopls: schema.GoplsEvidence{ImplementationCalls: 3},
			},
			want: nil,
		},
		{
			name: "ok: interface without impl calls but notes explain",
			in: &schema.TaintInput{
				Source: schema.TaintEndpoint{Kind: "iface"},
			},
			v: &schema.TaintVerdict{
				Gopls: schema.GoplsEvidence{ImplementationCalls: 0},
				Notes: "implementation walk skipped: anonymous interface with no implementers found",
			},
			want: nil,
		},
		{
			name: "bad: interface without impl calls and no notes",
			in: &schema.TaintInput{
				Source: schema.TaintEndpoint{Kind: "iface"},
			},
			v: &schema.TaintVerdict{
				Gopls: schema.GoplsEvidence{ImplementationCalls: 0},
				Notes: "",
			},
			want: []invariants.Violation{{
				Path:     "gopls.implementation_calls+notes",
				Expected: "implementation_calls>0 OR notes explain skip",
				Actual:   "both empty",
			}},
		},
		{
			name: "ok: non-interface source (vacuous)",
			in: &schema.TaintInput{
				Source: schema.TaintEndpoint{Kind: "http_query"},
			},
			v: &schema.TaintVerdict{
				Gopls: schema.GoplsEvidence{ImplementationCalls: 0},
				Notes: "",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T8 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT9_PathFileLinesExistInWorkspace validates T9: every (file, line) in path exists in the workspace.
func TestT9_PathFileLinesExistInWorkspace(t *testing.T) {
	check := locateTaintCheck(t, "T9")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: handler.go line 10 exists",
			in: &schema.TaintVerdict{
				Path: []schema.TaintPathStep{
					{File: "testdata/workspace/handler.go", Line: 10},
				},
			},
			want: nil,
		},
		{
			name: "bad: missing file",
			in: &schema.TaintVerdict{
				Path: []schema.TaintPathStep{
					{File: "testdata/workspace/missing.go", Line: 1},
				},
			},
			want: []invariants.Violation{{Path: "path[0].file", Expected: "exists", Actual: "testdata/workspace/missing.go"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T9 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT10_HighConfidenceRequiresStrongEvidence validates T10: high confidence requires
// either Semgrep Pro/intrafile finding OR full LSP path with no sanitizers_unverified.
func TestT10_HighConfidenceRequiresStrongEvidence(t *testing.T) {
	check := locateTaintCheck(t, "T10")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: high confidence with Semgrep Pro finding",
			in: &schema.TaintVerdict{
				Confidence: "high",
				Semgrep: schema.SemgrepEvidence{
					Tier:    "pro",
					Finding: true,
				},
				SanitizersUnverified: nil,
			},
			want: nil,
		},
		{
			name: "ok: high confidence with full LSP path (no sanitizers)",
			in: &schema.TaintVerdict{
				Confidence: "high",
				Gopls:      schema.GoplsEvidence{ReferencesCalls: 5},
				SanitizersUnverified: nil,
			},
			want: nil,
		},
		{
			name: "bad: high confidence without strong evidence",
			in: &schema.TaintVerdict{
				Confidence: "high",
				Semgrep: schema.SemgrepEvidence{
					Tier:    "ce",
					Finding: false,
				},
				Gopls:   schema.GoplsEvidence{ReferencesCalls: 5},
				SanitizersUnverified: []schema.SanitizerUnverified{{File: "x", Line: 1, Reason: "y"}},
			},
			want: []invariants.Violation{{
				Path:     "confidence",
				Expected: "low/medium (no strong evidence and sanitizers_unverified present)",
				Actual:   "high",
			}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T10 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT11_AgentToolUsageCompliance_NoOpStub validates T11: no-op stub per Phase 2 decision
// (verdict has no tool-call telemetry; deferred to Phase 5).
// TODO(phase-5): enforce via SubagentStop telemetry or tool-call summary in verdict.
func TestT11_AgentToolUsageCompliance_NoOpStub(t *testing.T) {
	check := locateTaintCheck(t, "T11")
	cases := []struct {
		name string
		in   *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: no-op stub always passes (phase 2)",
			in:   &schema.TaintVerdict{},
			want: nil,
		},
		{
			name: "ok: no-op stub always passes (phase 2, populated)",
			in: &schema.TaintVerdict{
				Verdict: "exploitable",
				Path:    []schema.TaintPathStep{{Step: "source"}},
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T11 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestT12_SessionIDMatchesInput validates T12 (JOINT): if input has review_session_id,
// output review_session_id must match exactly.
func TestT12_SessionIDMatchesInput(t *testing.T) {
	check := locateTaintJointCheck(t, "T12")
	cases := []struct {
		name string
		in   *schema.TaintInput
		v    *schema.TaintVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: both have matching session ID",
			in: &schema.TaintInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.TaintVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "bad: input has session ID but output mismatches",
			in: &schema.TaintInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.TaintVerdict{
				ReviewSessionID: "uuid-456",
			},
			want: []invariants.Violation{{Path: "review_session_id", Expected: "uuid-123", Actual: "uuid-456"}},
		},
		{
			name: "ok: input omits session ID (optional field)",
			in: &schema.TaintInput{
				ReviewSessionID: "",
			},
			v: &schema.TaintVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "ok: both omit session ID",
			in: &schema.TaintInput{
				ReviewSessionID: "",
			},
			v: &schema.TaintVerdict{
				ReviewSessionID: "",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("T12 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// --- B2: T6 Ambiguous Verdict Tests ---

// TestT6_AmbiguousVerdictNoEvidence_ShouldPass verifies that T6 should NOT block
// an honest "ambiguous" verdict when tools could not run (e.g., tools unavailable).
// Currently T6 blocks this verdict → test is RED (B2 blocker).
// After Wave 1, T6 will be modified to skip the check for verdict="ambiguous".
func TestT6_AmbiguousVerdictNoEvidence_ShouldPass(t *testing.T) {
	check := locateTaintCheck(t, "T6")

	// Case: verdict=ambiguous, no tools ran, no gopls calls made.
	// This is an honest verdict: the agent could not determine taintability.
	// T6 should NOT fire in this case.
	v := &schema.TaintVerdict{
		Verdict:    "ambiguous",
		Confidence: "low",
		Semgrep: schema.SemgrepEvidence{
			Ran:     false,
			Finding: false,
		},
		Gopls: schema.GoplsEvidence{
			ReferencesCalls: 0,
		},
	}

	violations := check(v)

	// Currently this is RED: T6 fires on "ambiguous" with no evidence.
	// The test asserts violations should be empty → fails because len(violations) > 0.
	if len(violations) != 0 {
		t.Logf("B2 RED: T6 blocks ambiguous verdict with no evidence (currently firing)")
		t.Logf("  violations: %v", violations)
	}

	// After Wave 1 adds the "ambiguous" exemption to T6, this assertion will pass.
	if len(violations) != 0 {
		t.Errorf("T6 should permit ambiguous verdict with no evidence, got violations: %v", violations)
	}
}

// --- E1: code_ref Joint Invariant Tests (compile-error RED until Wave 1) ---

// TestT_CodeRefMismatch_Blocked verifies that a new T-CodeRef joint invariant
// detects when TaintInput.CodeRef and TaintVerdict.CodeRef differ.
// This test is RED (compile-error) until Wave 1 adds CodeRef field to schema.
func TestT_CodeRefMismatch_Blocked(t *testing.T) {
	// After Wave 1: locate the T-CodeRef joint invariant
	// For now, this test is a placeholder that will compile-error on CodeRef reference.
	// The test demonstrates the intended behavior:
	// - Input has CodeRef: "abc123def456"
	// - Verdict has CodeRef: "999wronghash"
	// - Joint invariant should fire → violations non-empty

	// Input with non-empty code_ref
	_ = &schema.TaintInput{
		Source:        schema.TaintEndpoint{File: "main.go", Line: 10, Expr: "userInput", Kind: "param"},
		Sink:          schema.TaintEndpoint{File: "db.go", Line: 50, Expr: "query", Kind: "sink"},
		MaxDepth:      10,
		SemgrepTier:   "pro",
		// CodeRef: "abc123def456", // WILL BE ADDED IN WAVE 1
	}

	// Verdict with mismatched code_ref
	_ = &schema.TaintVerdict{
		Verdict:    "exploitable",
		Confidence: "high",
		Semgrep: schema.SemgrepEvidence{
			Ran:     true,
			Finding: true,
			Tier:    "pro",
			RuleID:  "go/sql-injection",
		},
		Gopls: schema.GoplsEvidence{
			ReferencesCalls: 5,
		},
		// CodeRef: "999wronghash", // WILL BE ADDED IN WAVE 1; mismatch from input
	}

	// After Wave 1 adds the field and joint invariant:
	// check := locateTaintJointCheck(t, "T-CodeRef")
	// violations := check(in, v)
	// if len(violations) == 0 {
	//    t.Errorf("T-CodeRef joint invariant should fire on code_ref mismatch")
	// }

	// Test empty input CodeRef: invariant should skip (non-git repo compat)
	_ = &schema.TaintInput{
		Source:        schema.TaintEndpoint{File: "main.go", Line: 10, Expr: "userInput", Kind: "param"},
		Sink:          schema.TaintEndpoint{File: "db.go", Line: 50, Expr: "query", Kind: "sink"},
		MaxDepth:      10,
		SemgrepTier:   "pro",
		// CodeRef: "", // empty input code_ref
	}

	_ = &schema.TaintVerdict{
		Verdict:    "exploitable",
		Confidence: "high",
		Semgrep: schema.SemgrepEvidence{
			Ran:     true,
			Finding: true,
			Tier:    "pro",
			RuleID:  "go/sql-injection",
		},
		Gopls: schema.GoplsEvidence{
			ReferencesCalls: 5,
		},
		// CodeRef: "somehash", // even if verdict has code_ref, empty input skips check
	}

	// After Wave 1:
	// violations = check(inNoCodeRef, vWithCodeRef)
	// if len(violations) != 0 {
	//    t.Errorf("T-CodeRef joint invariant should skip when input code_ref is empty")
	// }

	// For now, this test documents the intended behavior and will be GREEN after Wave 1.
	t.Skip("E1 compile-error RED: CodeRef field not yet added to schema (Wave 1)")
}
