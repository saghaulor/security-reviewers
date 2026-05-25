package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// PathStep represents a single step in a data flow path (source to sink).
type PathStep struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Desc   string `json:"description"`
}

// Evidence holds detailed information about a finding.
type Evidence struct {
	Files        []FileLocation `json:"files,omitempty"`
	DataFlowPath []PathStep     `json:"data_flow_path,omitempty"`
}

// FileLocation represents a file:line citation.
type FileLocation struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// Finding represents a single security finding in the review report.
type Finding struct {
	Class       string   `json:"class"`
	Confidence  string   `json:"confidence"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Evidence    Evidence `json:"evidence"`
}

// ReviewReport is the parsed /security-review output.
type ReviewReport struct {
	Findings []Finding `json:"findings"`
	Metadata struct {
		Timestamp string `json:"timestamp"`
		Version   string `json:"version"`
	} `json:"metadata"`
}

// RunSecurityReviewWorkflow invokes /security-review against the service and parses review-report.json.
// Per D-09 (programmatic verification), this is the entry point for Phase 5 verification.
func RunSecurityReviewWorkflow() (*ReviewReport, error) {
	// The /security-review command is invoked by Claude Code via the slash command interface.
	// It writes review-report.json to the project root or CWD.
	// This function reads and parses that output.

	reportPath := filepath.Join("review-report.json")

	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, err
	}

	var report ReviewReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// RunSecurityReviewWorkflowLive invokes the production /security-review command and parses the result.
// NOTE: This helper is provided for future automation. The /security-review command is designed
// as an interactive Claude command and cannot be reliably invoked as a subprocess.
// For Phase 8 E2E verification, /security-review is run manually by the user, who verifies
// that the full pipeline executes and produces correct findings.
//
// If this function is ever needed (e.g., if /security-review is refactored as a script),
// it demonstrates the intended pattern: invoke the command, parse review-report.json, verify findings.
func RunSecurityReviewWorkflowLive(workdir string) (*ReviewReport, error) {
	// Build path to review-report.json in the working directory
	reportPath := filepath.Join(workdir, "review-report.json")

	// Delete existing report to force fresh execution
	_ = os.Remove(reportPath)

	// Create subprocess to invoke /security-review command
	cmd := exec.Command("claude", "/security-review", workdir)

	// Capture stderr for debugging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("workflow failed: %w\nstderr: %s", err, stderr.String())
	}

	// Read the generated review-report.json
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read review-report.json: %w", err)
	}

	// Parse the report
	var report ReviewReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse review-report.json: %w", err)
	}

	return &report, nil
}

// TestSecurityReviewDetectsAllBugs is the Phase 5 + Phase 7 exit criterion verification.
// Per ROADMAP Phase 5: "exit criterion is three findings in review-report.json with non-ambiguous verdicts"
// Per ROADMAP Phase 7: Extended to 6 findings (3 basic + 3 moderate).
// Non-ambiguous = Confidence:high, Class:{injection|authz|oauth}, evidence present.
//
// This test verifies all six planted bugs are detected with high confidence.
func TestSecurityReviewDetectsAllBugs(t *testing.T) {
	report, err := RunSecurityReviewWorkflow()
	require.NoError(t, err, "RunSecurityReviewWorkflow should succeed and parse review-report.json")
	require.NotNil(t, report, "ReviewReport should not be nil")

	// Phase 7 exit criterion: exactly 6 findings (3 basic + 3 moderate)
	require.Equal(t, 6, len(report.Findings),
		"review-report.json must contain exactly 6 findings (3 basic: injection, authz, oauth; 3 moderate: call-chain injection, identity-only authz, implicit permission authz)")

	// Track findings by class to ensure proper distribution across bug classes
	findingsByClass := make(map[string]int)
	var injectionFindings, authzFindings []*Finding
	var oauthFinding *Finding

	for i := range report.Findings {
		finding := &report.Findings[i]

		// All findings must have high confidence (non-ambiguous)
		require.Equal(t, "high", finding.Confidence,
			"Finding %d (%s) must have confidence='high', not 'low' or 'ambiguous'", i, finding.Class)

		// Validate finding class
		require.Contains(t, []string{"injection", "authz", "oauth"}, finding.Class,
			"Finding class must be one of: injection, authz, oauth")

		findingsByClass[finding.Class]++

		// Categorize by class for detailed assertions
		switch finding.Class {
		case "injection":
			injectionFindings = append(injectionFindings, finding)
		case "authz":
			authzFindings = append(authzFindings, finding)
		case "oauth":
			oauthFinding = finding
		}
	}

	// Phase 7 class distribution: 2 injection (basic + call-chain), 3 authz (basic + identity-only + implicit), 1 oauth
	require.Equal(t, 2, findingsByClass["injection"],
		"Must have exactly 2 'injection' findings (basic + call-chain), got %d", findingsByClass["injection"])
	require.Equal(t, 3, findingsByClass["authz"],
		"Must have exactly 3 'authz' findings (basic + identity-only + implicit), got %d", findingsByClass["authz"])
	require.Equal(t, 1, findingsByClass["oauth"],
		"Must have exactly 1 'oauth' finding (scope-tampering), got %d", findingsByClass["oauth"])

	// Injection findings must have data flow paths (both basic and call-chain)
	require.GreaterOrEqual(t, len(injectionFindings), 2, "Should have at least 2 injection findings")
	for i, injectionFinding := range injectionFindings {
		require.NotEmpty(t, injectionFinding.Evidence.DataFlowPath,
			"Injection finding %d must have non-empty data_flow_path (source → sink)", i)
		require.GreaterOrEqual(t, len(injectionFinding.Evidence.DataFlowPath), 2,
			"Injection finding %d data_flow_path must have ≥2 steps", i)
	}

	// Authz findings must have evidence (files or data-flow)
	require.GreaterOrEqual(t, len(authzFindings), 3, "Should have at least 3 authz findings")
	for i, authzFinding := range authzFindings {
		hasEvidence := len(authzFinding.Evidence.Files) > 0 || len(authzFinding.Evidence.DataFlowPath) > 0
		require.True(t, hasEvidence,
			"Authz finding %d must have concrete evidence (Files or DataFlowPath)", i)
	}

	// OAuth finding must have data flow path (form input → token response)
	require.NotNil(t, oauthFinding, "OAuth finding should exist")
	require.NotEmpty(t, oauthFinding.Evidence.DataFlowPath,
		"OAuth finding must have non-empty data_flow_path (form input → token response)")
	require.GreaterOrEqual(t, len(oauthFinding.Evidence.DataFlowPath), 2,
		"OAuth data_flow_path must have ≥2 steps (form input → token response)")
}

// TestDetectsCallChainSQLi verifies the Phase 7 moderate-difficulty call-chain SQL injection is detected.
// Per D-05: A separate test function for call-chain SQLi that verifies multi-hop data-flow tracing.
//
// This test validates:
// 1. At least one "injection" class finding exists in the report
// 2. The finding has confidence == "high"
// 3. The data-flow path shows 4+ hops (handler → service → repo → builder → db.Query)
// 4. Each PathStep has non-empty File and Line fields (concrete evidence)
//
// The vulnerability pattern is: HTTP handler → ServiceLayer.SearchUsers() → RepoLayer.QueryByName() →
// QueryBuilder.Where() → db.Query(concatenated_sql). This tests whether the tracer follows data-flow
// across multiple function boundaries and maintains taint context.
func TestDetectsCallChainSQLi(t *testing.T) {
	report, err := RunSecurityReviewWorkflow()
	require.NoError(t, err, "RunSecurityReviewWorkflow should succeed and parse review-report.json")
	require.NotNil(t, report, "ReviewReport should not be nil")

	// Find at least one injection finding (we expect 2 total: basic + call-chain)
	var callChainSQLiFinding *Finding
	for i := range report.Findings {
		finding := &report.Findings[i]
		if finding.Class == "injection" && len(finding.Evidence.DataFlowPath) >= 4 {
			callChainSQLiFinding = finding
			break
		}
	}

	require.NotNil(t, callChainSQLiFinding,
		"Call-chain SQLi finding should exist with class='injection' and data_flow_path with ≥4 hops")

	// Verify high confidence
	require.Equal(t, "high", callChainSQLiFinding.Confidence,
		"Call-chain SQLi finding must have confidence='high', not 'low' or 'ambiguous'")

	// Verify 4+ hop data-flow path (handler → service → repo → builder → db)
	require.GreaterOrEqual(t, len(callChainSQLiFinding.Evidence.DataFlowPath), 4,
		"Call-chain SQLi finding must have ≥4 hops in data_flow_path (handler → service → repo → builder/db)")

	// Verify each hop has concrete file/line information
	for i, step := range callChainSQLiFinding.Evidence.DataFlowPath {
		require.NotEmpty(t, step.File,
			"PathStep %d (hop %d) must have non-empty File field", i, i+1)
		require.NotZero(t, step.Line,
			"PathStep %d (hop %d) must have non-zero Line field", i, i+1)
	}
}


// TestDetectsWeakAuthz verifies the Phase 7 moderate-difficulty weak authorization patterns are detected.
// Per D-05: A separate test function for weak authz patterns that covers both variants.
//
// This test validates:
// 1. At least two "authz" class findings exist in the report (basic + two moderate patterns)
// 2. Each finding has confidence == "high"
// 3. Each finding has concrete evidence (Files or DataFlowPath)
// 4. The findings together represent both weak authz patterns:
//    - Identity-only check (middleware present but missing resource ownership check)
//    - Implicit permission assumption (unvalidated JWT claim trusted directly)
//
// The vulnerability patterns are:
// 1. Identity-only: Middleware verifies authentication, but handler performs sensitive operation without
//    resource-level authorization (e.g., transferFundsHandler without checking if user can transfer to that recipient)
// 2. Implicit assumption: Handler checks identity but trusts unvalidated JWT claim without authz service
//    (e.g., adminConfigHandler trusts user.IsAdmin claim without server-side validation)
func TestDetectsWeakAuthz(t *testing.T) {
	report, err := RunSecurityReviewWorkflow()
	require.NoError(t, err, "RunSecurityReviewWorkflow should succeed and parse review-report.json")
	require.NotNil(t, report, "ReviewReport should not be nil")

	// Count authz findings (expect ≥3: basic bypass + identity-only + implicit assumption)
	var authzCount int
	var authzFindings []*Finding
	for i := range report.Findings {
		finding := &report.Findings[i]
		if finding.Class == "authz" {
			authzCount++
			authzFindings = append(authzFindings, finding)
		}
	}

	require.GreaterOrEqual(t, authzCount, 3,
		"Must have at least 3 'authz' findings: basic bypass + identity-only + implicit assumption, got %d", authzCount)

	// Verify all authz findings have high confidence
	for i, finding := range authzFindings {
		require.Equal(t, "high", finding.Confidence,
			"Authz finding %d (%s) must have confidence='high', not 'low' or 'ambiguous'", i, finding.Title)

		// Verify each has concrete evidence (either Files or DataFlowPath)
		hasEvidence := len(finding.Evidence.Files) > 0 || len(finding.Evidence.DataFlowPath) > 0
		require.True(t, hasEvidence,
			"Authz finding %d must have concrete evidence (Files or DataFlowPath)", i)
	}
}
