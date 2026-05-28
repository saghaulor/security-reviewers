package hooks

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestContentPreview_UnicodeSafe verifies that contentPreview truncates on rune
// boundaries, not raw bytes, so a multi-byte rune straddling byte index 100 is not
// split into invalid UTF-8 in block-reason previews (IN-04).
func TestContentPreview_UnicodeSafe(t *testing.T) {
	// 99 ASCII runes, then a 3-byte rune at rune index 99 (byte index 99),
	// then filler. Byte-slicing at 100 would split the 世 rune.
	s := strings.Repeat("a", 99) + "世" + strings.Repeat("b", 50)

	quoted := contentPreview(s)
	preview, err := strconv.Unquote(quoted)
	if err != nil {
		t.Fatalf("contentPreview returned an unparseable quoted string %q: %v", quoted, err)
	}
	if !utf8.ValidString(preview) {
		t.Errorf("contentPreview produced invalid UTF-8 (rune split at boundary): %q", quoted)
	}
	if !strings.ContainsRune(preview, '世') {
		t.Errorf("contentPreview should keep the boundary rune 世 intact, got %q", preview)
	}
}

// TestDispatch_Cartographer_Happy — A6 would require graphify-out/graph.json fixture setup,
// so we test a simpler case: valid schema_version passes A2 (sad case below).
func TestDispatch_Cartographer_Happy(t *testing.T) {
	// Use a minimal index that passes structural checks (A1, A2, A3..A5 all pass or are N/A).
	// A6 (graphify check) will fail because we don't have the fixture, but that's OK for unit tests.
	// The dispatch_test is about routing, not file I/O. File I/O is tested separately in cartographer_test.go.
	// For dispatch.go test, just verify the sad path works:
	t.Skip("Cartographer happy path requires graphify-out/graph.json fixture; tested in cartographer_test.go")
}

// TestDispatch_Cartographer_Sad_SchemaVersionWrong
func TestDispatch_Cartographer_Sad_SchemaVersionWrong(t *testing.T) {
	verdict := `{"schema_version":"go-index/v2","graph_version":"","routers_detected":[],"entrypoints":[],"sinks_by_kind":{},"authz_primitives":[],"oauth_locations":{},"payment_surface":{"files":[],"cluster_id":"","confidence":""},"vuln_deps":{"available":false,"findings":[]},"ambiguous_nodes":[],"warnings":[]}`
	reasons, err := runAgentValidation("go-cartographer", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) == 0 {
		t.Fatal("expected non-empty reasons for wrong schema_version")
	}
	if !strings.Contains(reasons[0], "A2") {
		t.Errorf("expected A2 in reasons[0], got %v", reasons[0])
	}
}

// TestDispatch_TaintTracer_Happy
func TestDispatch_TaintTracer_Happy(t *testing.T) {
	verdict := `{"verdict":"unreachable","confidence":"high","path":[],"semgrep":{"tier":"pro","rule_id":"test","ran":true,"finding":true},"gopls":{"references_calls":0,"definition_calls":0,"implementation_calls":0,"call_hierarchy_calls":0,"branches_explored":0,"branches_unexplored":0,"interface_fanout_max":0,"goroutine_boundaries_crossed":0},"sanitizers_unverified":[]}`
	reasons, err := runAgentValidation("go-taint-tracer", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) != 0 {
		t.Errorf("expected empty reasons, got %v", reasons)
	}
}

// TestDispatch_TaintTracer_Sad_BadVerdict
func TestDispatch_TaintTracer_Sad_BadVerdict(t *testing.T) {
	verdict := `{"verdict":"unknown_value","confidence":"high","path":[],"semgrep":{"tier":"","rule_id":"","ran":false,"finding":false},"gopls":{"references_calls":0,"definition_calls":0,"implementation_calls":0,"call_hierarchy_calls":0,"branches_explored":0,"branches_unexplored":0,"interface_fanout_max":0,"goroutine_boundaries_crossed":0},"sanitizers_unverified":[]}`
	reasons, err := runAgentValidation("go-taint-tracer", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) == 0 {
		t.Fatal("expected non-empty reasons for invalid verdict field")
	}
	if !strings.Contains(reasons[0], "T2") {
		t.Errorf("expected T2 in reasons[0], got %v", reasons[0])
	}
}

// TestDispatch_AuthzTracer_Happy
func TestDispatch_AuthzTracer_Happy(t *testing.T) {
	verdict := `{"summary":{"routes_total":1,"protected":1,"missing":0,"weak":0,"idor_risk":0,"public_intentional":0}}`
	reasons, err := runAgentValidation("go-authz-tracer", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) != 0 {
		t.Errorf("expected empty reasons, got %v", reasons)
	}
}

// TestDispatch_AuthzTracer_Sad_MissingSummary
func TestDispatch_AuthzTracer_Sad_MissingSummary(t *testing.T) {
	verdict := `{}`
	reasons, err := runAgentValidation("go-authz-tracer", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) == 0 {
		t.Fatal("expected non-empty reasons for missing summary")
	}
	if !strings.Contains(reasons[0], "AZ") {
		t.Errorf("expected AZ-prefixed reason, got %v", reasons[0])
	}
}

// TestDispatch_OAuthAuditor_Happy
func TestDispatch_OAuthAuditor_Happy(t *testing.T) {
	verdict := `{"profile":"oauth-2.1","checklist":[{"check_id":"O1","spec":"rfc9700","status":"pass","severity":"high","evidence":{"file":"test.go","line":1,"explanation":"test"}}]}`
	reasons, err := runAgentValidation("go-oauth-auditor", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) != 0 {
		t.Errorf("expected empty reasons, got %v", reasons)
	}
}

// TestDispatch_OAuthAuditor_Sad_MissingProfile
func TestDispatch_OAuthAuditor_Sad_MissingProfile(t *testing.T) {
	verdict := `{}`
	reasons, err := runAgentValidation("go-oauth-auditor", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) == 0 {
		t.Fatal("expected non-empty reasons for missing profile")
	}
	if !strings.Contains(reasons[0], "OA") {
		t.Errorf("expected OA-prefixed reason, got %v", reasons[0])
	}
}

// TestDispatch_InvariantChecker_Happy
func TestDispatch_InvariantChecker_Happy(t *testing.T) {
	verdict := `{"flow_name":"test_flow","results":[{"invariant_id":"I1","status":"pass","confidence":"high"}]}`
	reasons, err := runAgentValidation("invariant-checker", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) != 0 {
		t.Errorf("expected empty reasons, got %v", reasons)
	}
}

// TestDispatch_InvariantChecker_Sad_MissingFlowName
func TestDispatch_InvariantChecker_Sad_MissingFlowName(t *testing.T) {
	verdict := `{}`
	reasons, err := runAgentValidation("invariant-checker", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) == 0 {
		t.Fatal("expected non-empty reasons for missing flow_name field")
	}
	if !strings.Contains(reasons[0], "IC") {
		t.Errorf("expected IC-prefixed reason, got %v", reasons[0])
	}
}

// TestDispatch_Synthesis_Happy
func TestDispatch_Synthesis_Happy(t *testing.T) {
	verdict := `{"review_id":"test","timestamp":"2026-05-19T00:00:00Z","schema_version":"review-report/v1","summary":{"total_findings":0,"by_severity":{},"by_class":{}}}`
	reasons, err := runAgentValidation("synthesis", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) != 0 {
		t.Errorf("expected empty reasons, got %v", reasons)
	}
}

// TestDispatch_Synthesis_Sad_WrongSchemaVersion
func TestDispatch_Synthesis_Sad_WrongSchemaVersion(t *testing.T) {
	verdict := `{"review_id":"test","timestamp":"2026-05-19T00:00:00Z","schema_version":"review-report/v2","summary":{"total_findings":0,"by_severity":{},"by_class":{}}}`
	reasons, err := runAgentValidation("synthesis", verdict, "")
	if err != nil {
		t.Fatalf("runAgentValidation: %v", err)
	}
	if len(reasons) == 0 {
		t.Fatal("expected non-empty reasons for wrong schema_version")
	}
	if !strings.Contains(reasons[0], "S") {
		t.Errorf("expected S-prefixed reason, got %v", reasons[0])
	}
}
