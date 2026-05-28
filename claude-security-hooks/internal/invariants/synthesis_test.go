package invariants_test

import (
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/invariants"
	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// locateSynthesisCheck finds a SynthesisInvariant by ID in the registry.
func locateSynthesisCheck(id string) *invariants.SynthesisInvariant {
	for i := range invariants.SynthesisInvariants {
		if invariants.SynthesisInvariants[i].ID == id {
			return &invariants.SynthesisInvariants[i]
		}
	}
	return nil
}

// locateSynthesisDirCheck finds a SynthesisDirInvariant by ID in the registry.
func locateSynthesisDirCheck(id string) *invariants.SynthesisDirInvariant {
	for i := range invariants.SynthesisDirInvariants {
		if invariants.SynthesisDirInvariants[i].ID == id {
			return &invariants.SynthesisDirInvariants[i]
		}
	}
	return nil
}

// TestS1_BothOutputFilesExist tests the S1 directory-path predicate.
// S1 verifies that synthesis output contains both review-report.json AND review-report.md.
func TestS1_BothOutputFilesExist(t *testing.T) {
	check := locateSynthesisDirCheck("S1")
	if check == nil {
		t.Fatalf("S1 check not found in SynthesisDirInvariants")
	}

	tests := []struct {
		name    string
		dirPath string
		wantErr int // number of violations expected
		wantPath string // path that should appear in a violation if wantErr > 0
	}{
		{
			name:     "both files exist",
			dirPath:  "testdata/synthesis_out/both",
			wantErr:  0,
			wantPath: "",
		},
		{
			name:     "md missing",
			dirPath:  "testdata/synthesis_out/no_md",
			wantErr:  1,
			wantPath: "testdata/synthesis_out/no_md/review-report.md",
		},
		{
			name:     "json missing",
			dirPath:  "testdata/synthesis_out/no_json",
			wantErr:  1,
			wantPath: "testdata/synthesis_out/no_json/review-report.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.dirPath)
			if len(violations) != tt.wantErr {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantErr)
			}
			if tt.wantErr > 0 && len(violations) > 0 {
				if violations[0].Path != tt.wantPath {
					t.Errorf("got path %q, want %q", violations[0].Path, tt.wantPath)
				}
				if violations[0].Expected != "exists" {
					t.Errorf("got Expected %q, want %q", violations[0].Expected, "exists")
				}
				if violations[0].Actual != "missing" {
					t.Errorf("got Actual %q, want %q", violations[0].Actual, "missing")
				}
			}
		})
	}
}

// TestS2_VerdictRequiredFields tests the S2 verdict predicate.
// S2 enforces that SchemaVersion is "review-report/v1" and ReviewID is non-empty.
func TestS2_VerdictRequiredFields(t *testing.T) {
	check := locateSynthesisCheck("S2")
	if check == nil {
		t.Fatalf("S2 check not found in SynthesisInvariants")
	}

	tests := []struct {
		name    string
		report  *schema.SynthesisReport
		wantErr int
	}{
		{
			name: "valid schema version",
			report: &schema.SynthesisReport{
				ReviewID:      "test-review-001",
				SchemaVersion: "review-report/v1",
			},
			wantErr: 0,
		},
		{
			name: "wrong schema version",
			report: &schema.SynthesisReport{
				ReviewID:      "test-review-001",
				SchemaVersion: "review-report/v2",
			},
			wantErr: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.report)
			if len(violations) != tt.wantErr {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantErr)
			}
			if tt.wantErr > 0 && len(violations) > 0 {
				if violations[0].Path != "schema_version" {
					t.Errorf("got path %q, want %q", violations[0].Path, "schema_version")
				}
				if violations[0].Expected != "review-report/v1" {
					t.Errorf("got Expected %q, want %q", violations[0].Expected, "review-report/v1")
				}
				if violations[0].Actual != "review-report/v2" {
					t.Errorf("got Actual %q, want %q", violations[0].Actual, "review-report/v2")
				}
			}
		})
	}
}

// TestS3_TotalFindingsMatchesLen tests the S3 verdict predicate.
// S3 enforces that summary.total_findings equals len(findings).
func TestS3_TotalFindingsMatchesLen(t *testing.T) {
	check := locateSynthesisCheck("S3")
	if check == nil {
		t.Fatalf("S3 check not found in SynthesisInvariants")
	}

	tests := []struct {
		name    string
		report  *schema.SynthesisReport
		wantErr int
	}{
		{
			name: "counts match",
			report: &schema.SynthesisReport{
				Summary: schema.SynthesisSummary{
					TotalFindings: 2,
				},
				Findings: []schema.SynthesisFinding{
					{ID: "f1"},
					{ID: "f2"},
				},
			},
			wantErr: 0,
		},
		{
			name: "counts mismatch",
			report: &schema.SynthesisReport{
				Summary: schema.SynthesisSummary{
					TotalFindings: 5,
				},
				Findings: []schema.SynthesisFinding{
					{ID: "f1"},
					{ID: "f2"},
				},
			},
			wantErr: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.report)
			if len(violations) != tt.wantErr {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantErr)
			}
			if tt.wantErr > 0 && len(violations) > 0 {
				if violations[0].Path != "summary.total_findings" {
					t.Errorf("got path %q, want %q", violations[0].Path, "summary.total_findings")
				}
			}
		})
	}
}

// TestS4_BySeveritySumsToTotal tests the S4 verdict predicate.
// S4 enforces that sum of summary.by_severity values equals total_findings.
func TestS4_BySeveritySumsToTotal(t *testing.T) {
	check := locateSynthesisCheck("S4")
	if check == nil {
		t.Fatalf("S4 check not found in SynthesisInvariants")
	}

	tests := []struct {
		name    string
		report  *schema.SynthesisReport
		wantErr int
	}{
		{
			name: "sum matches total",
			report: &schema.SynthesisReport{
				Summary: schema.SynthesisSummary{
					TotalFindings: 3,
					BySeverity: map[string]int{
						"critical": 1,
						"high":     2,
						"medium":   0,
						"low":      0,
						"info":     0,
					},
				},
			},
			wantErr: 0,
		},
		{
			name: "sum does not match total",
			report: &schema.SynthesisReport{
				Summary: schema.SynthesisSummary{
					TotalFindings: 3,
					BySeverity: map[string]int{
						"critical": 1,
						"high":     1,
					},
				},
			},
			wantErr: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.report)
			if len(violations) != tt.wantErr {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantErr)
			}
			if tt.wantErr > 0 && len(violations) > 0 {
				if violations[0].Path != "summary.by_severity" {
					t.Errorf("got path %q, want %q", violations[0].Path, "summary.by_severity")
				}
			}
		})
	}
}

// TestS5_EveryFindingHasSourceAgents tests the S5 verdict predicate.
// S5 enforces that every finding has at least one source_agents entry.
func TestS5_EveryFindingHasSourceAgents(t *testing.T) {
	check := locateSynthesisCheck("S5")
	if check == nil {
		t.Fatalf("S5 check not found in SynthesisInvariants")
	}

	tests := []struct {
		name    string
		report  *schema.SynthesisReport
		wantErr int
	}{
		{
			name: "finding with source agents",
			report: &schema.SynthesisReport{
				Findings: []schema.SynthesisFinding{
					{
						ID:           "f1",
						SourceAgents: []string{"go-taint-tracer"},
					},
				},
			},
			wantErr: 0,
		},
		{
			name: "finding without source agents",
			report: &schema.SynthesisReport{
				Findings: []schema.SynthesisFinding{
					{
						ID:           "f1",
						SourceAgents: nil,
					},
				},
			},
			wantErr: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.report)
			if len(violations) != tt.wantErr {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantErr)
			}
			if tt.wantErr > 0 && len(violations) > 0 {
				if violations[0].Path != "findings[0].source_agents" {
					t.Errorf("got path %q, want %q", violations[0].Path, "findings[0].source_agents")
				}
				if violations[0].Expected != "at least one" {
					t.Errorf("got Expected %q, want %q", violations[0].Expected, "at least one")
				}
				if violations[0].Actual != "empty" {
					t.Errorf("got Actual %q, want %q", violations[0].Actual, "empty")
				}
			}
		})
	}
}

// TestS6_NoDuplicateFindingKeyWithoutNote tests the S6 verdict predicate.
// S6 enforces that no two findings share (file, line, class) unless noted.
// A "finding key" is class + "|" + first evidence file + "|" + first evidence line.
// This test covers 3 cases:
// 1. Different class, same file/line → no violation (different key)
// 2. Same class, file, line but no dedup note → violation
// 3. Same class, file, line with dedup note mentioning one finding ID → no violation
func TestS6_NoDuplicateFindingKeyWithoutNote(t *testing.T) {
	check := locateSynthesisCheck("S6")
	if check == nil {
		t.Fatalf("S6 check not found in SynthesisInvariants")
	}

	tests := []struct {
		name    string
		report  *schema.SynthesisReport
		wantErr int
		wantMsg string // substring to check in violation if wantErr > 0
	}{
		{
			name: "different class, same file/line",
			report: &schema.SynthesisReport{
				Findings: []schema.SynthesisFinding{
					{
						ID:    "finding-001",
						Class: "injection",
						Evidence: schema.SynthesisEvidence{
							Files: []string{"x.go"},
							Lines: []int{10},
						},
					},
					{
						ID:    "finding-002",
						Class: "authz",
						Evidence: schema.SynthesisEvidence{
							Files: []string{"x.go"},
							Lines: []int{10},
						},
					},
				},
				DeduplicationNotes: []string{},
			},
			wantErr: 0,
		},
		{
			name: "unnoted duplicate key",
			report: &schema.SynthesisReport{
				Findings: []schema.SynthesisFinding{
					{
						ID:    "finding-001",
						Class: "injection",
						Evidence: schema.SynthesisEvidence{
							Files: []string{"x.go"},
							Lines: []int{10},
						},
					},
					{
						ID:    "finding-007",
						Class: "injection",
						Evidence: schema.SynthesisEvidence{
							Files: []string{"x.go"},
							Lines: []int{10},
						},
					},
				},
				DeduplicationNotes: []string{},
			},
			wantErr: 1,
			wantMsg: "finding-001 + finding-007 unnoted",
		},
		{
			name: "noted duplicate key recovery",
			report: &schema.SynthesisReport{
				Findings: []schema.SynthesisFinding{
					{
						ID:    "finding-001",
						Class: "injection",
						Evidence: schema.SynthesisEvidence{
							Files: []string{"x.go"},
							Lines: []int{10},
						},
					},
					{
						ID:    "finding-007",
						Class: "injection",
						Evidence: schema.SynthesisEvidence{
							Files: []string{"x.go"},
							Lines: []int{10},
						},
					},
				},
				DeduplicationNotes: []string{"finding-001 and finding-007 merged: same source/sink"},
			},
			wantErr: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := check.Check(tt.report)
			if len(violations) != tt.wantErr {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantErr)
			}
			if tt.wantErr > 0 && len(violations) > 0 {
				if violations[0].Path != "findings" {
					t.Errorf("got path %q, want %q", violations[0].Path, "findings")
				}
				if tt.wantMsg != "" && violations[0].Actual != tt.wantMsg {
					t.Errorf("got Actual %q, want substring %q", violations[0].Actual, tt.wantMsg)
				}
			}
		})
	}
}
