package invariants

import (
	"fmt"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// TaintInvariant represents a verdict-only invariant check (T1..T6, T9..T11).
type TaintInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.TaintVerdict) []Violation
}

// TaintInputJointInvariant represents a joint predicate (T7, T8) that checks
// both input and verdict. Separate type per RESEARCH.md option (a) to keep
// types narrow and avoid any.
type TaintInputJointInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.TaintInput, *schema.TaintVerdict) []Violation
}

var TaintTracerInvariants = []TaintInvariant{
	{ID: "T1", Description: "verdict has required top-level fields (verdict, confidence)", Severity: SeverityCritical, Check: checkT1},
	{ID: "T2", Description: "verdict in {exploitable,sanitized,unreachable,ambiguous,input_mismatch}", Severity: SeverityCritical, Check: checkT2},
	{ID: "T3", Description: "sanitized/exploitable requires non-empty path with source first and sink last", Severity: SeverityCritical, Check: checkT3},
	{ID: "T4", Description: "input_mismatch permits empty path", Severity: SeverityInfo, Check: checkT4},
	{ID: "T5", Description: "confidence in {high,medium,low}", Severity: SeverityHigh, Check: checkT5},
	{ID: "T6", Description: "either semgrep.ran or gopls.references_calls > 0 (unless input_mismatch)", Severity: SeverityHigh, Check: checkT6},
	{ID: "T9", Description: "every (file,line) in path corresponds to a real workspace location", Severity: SeverityHigh, Check: checkT9},
	{ID: "T10", Description: "confidence=high requires Semgrep pro/intrafile finding OR full LSP path with no sanitizers_unverified", Severity: SeverityHigh, Check: checkT10},
	{ID: "T11", Description: "agent did not call Grep/Bash/Edit/Write (no-op stub Phase 2; real check Phase 5)", Severity: SeverityInfo, Check: checkT11},
}

var TaintTracerJointInvariants = []TaintInputJointInvariant{
	{ID: "T7", Description: "semgrep.tier in verdict equals semgrep_tier in input", Severity: SeverityHigh, Check: checkT7Joint},
	{ID: "T8", Description: "interface-typed source requires implementation_calls>0 OR notes explaining skip", Severity: SeverityHigh, Check: checkT8Joint},
}

// --- T1 ---
func checkT1(v *schema.TaintVerdict) []Violation {
	var out []Violation
	if v.Verdict == "" {
		out = append(out, Violation{Path: "verdict", Expected: "non-empty", Actual: ""})
	}
	if v.Confidence == "" {
		out = append(out, Violation{Path: "confidence", Expected: "non-empty", Actual: ""})
	}
	return out
}

// --- T2 ---
var taintVerdictEnum = map[string]struct{}{
	"exploitable": {}, "sanitized": {}, "unreachable": {}, "ambiguous": {}, "input_mismatch": {},
}

func checkT2(v *schema.TaintVerdict) []Violation {
	if _, ok := taintVerdictEnum[v.Verdict]; !ok {
		return []Violation{{Path: "verdict", Expected: "in {exploitable,sanitized,unreachable,ambiguous,input_mismatch}", Actual: v.Verdict}}
	}
	return nil
}

// --- T3 (verbatim RESEARCH.md Pattern 2) ---
func checkT3(v *schema.TaintVerdict) []Violation {
	if v.Verdict != "sanitized" && v.Verdict != "exploitable" {
		return nil
	}
	var out []Violation
	if len(v.Path) == 0 {
		return []Violation{{Path: "path", Expected: "non-empty", Actual: "empty"}}
	}
	if v.Path[0].Step != "source" {
		out = append(out, Violation{Path: "path[0].step", Expected: "source", Actual: v.Path[0].Step})
	}
	if last := v.Path[len(v.Path)-1]; last.Step != "sink" {
		out = append(out, Violation{Path: fmt.Sprintf("path[%d].step", len(v.Path)-1), Expected: "sink", Actual: last.Step})
	}
	return out
}

// --- T4 ---
func checkT4(v *schema.TaintVerdict) []Violation {
	// T4 is purely permissive — empty path on input_mismatch is allowed.
	// Implementation: always returns nil. Test confirms the case.
	_ = v
	return nil
}

// --- T5 ---
var taintConfidenceEnum = map[string]struct{}{"high": {}, "medium": {}, "low": {}}

func checkT5(v *schema.TaintVerdict) []Violation {
	if _, ok := taintConfidenceEnum[v.Confidence]; !ok {
		return []Violation{{Path: "confidence", Expected: "in {high,medium,low}", Actual: v.Confidence}}
	}
	return nil
}

// --- T6 ---
func checkT6(v *schema.TaintVerdict) []Violation {
	if v.Verdict == "input_mismatch" {
		return nil
	}
	if v.Semgrep.Ran || v.Gopls.ReferencesCalls > 0 {
		return nil
	}
	return []Violation{{
		Path:     "semgrep+gopls",
		Expected: "semgrep.ran=true OR gopls.references_calls>0 (unless input_mismatch)",
		Actual:   "neither",
	}}
}

// --- T7 (joint) ---
func checkT7Joint(in *schema.TaintInput, v *schema.TaintVerdict) []Violation {
	if in.SemgrepTier != v.Semgrep.Tier {
		return []Violation{{Path: "semgrep.tier", Expected: in.SemgrepTier, Actual: v.Semgrep.Tier}}
	}
	return nil
}

// --- T8 (joint) ---
// Heuristic for "interface-typed source": Source.Kind == "iface" OR Source.Expr contains "interface".
// Documented limitation: HAND_OFF source kind enum (line 247) does not include "iface" explicitly;
// this heuristic is conservative — flagging only obvious interface-typed sources.
func checkT8Joint(in *schema.TaintInput, v *schema.TaintVerdict) []Violation {
	isIface := in.Source.Kind == "iface" || strings.Contains(strings.ToLower(in.Source.Expr), "interface")
	if !isIface {
		return nil
	}
	if v.Gopls.ImplementationCalls > 0 {
		return nil
	}
	if strings.TrimSpace(v.Notes) != "" {
		// Any non-empty notes string counts as "explaining the skip" — agents
		// produce free-text justifications and the validator does not parse them.
		return nil
	}
	return []Violation{{
		Path:     "gopls.implementation_calls+notes",
		Expected: "implementation_calls>0 OR notes explain skip",
		Actual:   "both empty",
	}}
}

// --- T9 ---
func checkT9(v *schema.TaintVerdict) []Violation {
	var out []Violation
	lineCache := make(map[string]int)
	for i, step := range v.Path {
		if step.File == "" || step.Line <= 0 {
			continue
		}
		if !FileExists(step.File) {
			out = append(out, Violation{Path: fmt.Sprintf("path[%d].file", i), Expected: "exists", Actual: step.File})
			continue
		}
		n, ok := lineCache[step.File]
		if !ok {
			lc, err := LineCount(step.File)
			if err != nil {
				continue
			}
			n = lc
			lineCache[step.File] = n
		}
		if step.Line > n {
			out = append(out, Violation{Path: fmt.Sprintf("path[%d].line", i), Expected: fmt.Sprintf("<=%d", n), Actual: fmt.Sprintf("%d", step.Line)})
		}
	}
	return out
}

// --- T10 ---
func checkT10(v *schema.TaintVerdict) []Violation {
	if v.Confidence != "high" {
		return nil
	}
	// Branch A: Semgrep Pro/intrafile finding.
	if v.Semgrep.Finding && (v.Semgrep.Tier == "pro" || v.Semgrep.Tier == "intrafile") {
		return nil
	}
	// Branch B: full LSP path with no sanitizers_unverified.
	if v.Gopls.ReferencesCalls > 0 && len(v.SanitizersUnverified) == 0 {
		return nil
	}
	return []Violation{{
		Path:     "confidence",
		Expected: "low/medium (no strong evidence and sanitizers_unverified present)",
		Actual:   "high",
	}}
}

// --- T11 ---
// No-op stub per Phase 2 decision; verdict has no tool-call telemetry.
// TODO(phase-5): enforce via SubagentStop telemetry or tool-call summary in verdict.
func checkT11(v *schema.TaintVerdict) []Violation {
	_ = v
	return nil
}
