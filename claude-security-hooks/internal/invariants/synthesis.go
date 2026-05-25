// Package invariants provides assertion checks for security review reports.
// It implements H5 (zero non-stdlib deps), H6 (mechanical coverage), and
// 51 per-agent assertions (A1-A11, T1-T11, AZ1-AZ6, OA1-OA7, IC1-IC4, S1-S6).
package invariants

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// SynthesisInvariant represents a check over a SynthesisReport verdict.
// S2-S6 assertions are verdict-only checks that validate the synthesized
// review report's structural invariants (schema, accounting, provenance, dedup).
type SynthesisInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.SynthesisReport) []Violation
}

// SynthesisDirInvariant represents a check over a directory path (S1 special case).
// S1 is unique: it validates filesystem output (both review-report.json and .md exist)
// rather than validating the deserialized verdict struct. SynthesisDirInvariant
// is a separate registry to preserve type safety (Check signature differs).
type SynthesisDirInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(dirPath string) []Violation
}

// SynthesisInvariants is the registry of verdict-only checks (S2–S6).
var SynthesisInvariants = []SynthesisInvariant{
	{
		ID:          "S2",
		Description: "review-report JSON is valid review-report/v1",
		Severity:    SeverityCritical,
		Check:       checkS2,
	},
	{
		ID:          "S3",
		Description: "summary.total_findings equals len(findings)",
		Severity:    SeverityHigh,
		Check:       checkS3,
	},
	{
		ID:          "S4",
		Description: "sum of summary.by_severity values equals total_findings",
		Severity:    SeverityHigh,
		Check:       checkS4,
	},
	{
		ID:          "S5",
		Description: "every finding has at least one source_agents entry",
		Severity:    SeverityHigh,
		Check:       checkS5,
	},
	{
		ID:          "S6",
		Description: "no two findings share (file,line,class) unless noted in deduplication_notes",
		Severity:    SeverityHigh,
		Check:       checkS6,
	},
}

// SynthesisDirInvariants is the registry of directory-path checks (S1 only).
var SynthesisDirInvariants = []SynthesisDirInvariant{
	{
		ID:          "S1",
		Description: "synthesis directory contains both review-report.json and review-report.md",
		Severity:    SeverityCritical,
		Check:       checkS1Dir,
	},
}

// checkS1Dir verifies that both review-report.json and review-report.md exist in dirPath.
func checkS1Dir(dirPath string) []Violation {
	var out []Violation
	for _, name := range []string{"review-report.json", "review-report.md"} {
		p := filepath.Join(dirPath, name)
		if !FileExists(p) {
			out = append(out, Violation{
				Path:     p,
				Expected: "exists",
				Actual:   "missing",
			})
		}
	}
	return out
}

// checkS2 verifies that ReviewID is non-empty and SchemaVersion is "review-report/v1".
func checkS2(r *schema.SynthesisReport) []Violation {
	var out []Violation
	if r.ReviewID == "" {
		out = append(out, Violation{
			Path:     "review_id",
			Expected: "non-empty",
			Actual:   "",
		})
	}
	const wantSchema = "review-report/v1"
	if r.SchemaVersion != wantSchema {
		out = append(out, Violation{
			Path:     "scma_version",
			Expected: wantSchema,
			Actual:   r.SchemaVersion,
		})
	}
	return out
}

// checkS3 verifies that summary.total_findings == len(findings).
func checkS3(r *schema.SynthesisReport) []Violation {
	if r.Summary.TotalFindings != len(r.Findings) {
		return []Violation{{
			Path:     "summary.total_findings",
			Expected: fmt.Sprintf("%d", len(r.Findings)),
			Actual:   fmt.Sprintf("%d", r.Summary.TotalFindings),
		}}
	}
	return nil
}

// checkS4 verifies that sum of summary.by_severity values equals total_findings.
func checkS4(r *schema.SynthesisReport) []Violation {
	sum := 0
	for _, v := range r.Summary.BySeverity {
		sum += v
	}
	if sum != r.Summary.TotalFindings {
		return []Violation{{
			Path:     "summary.by_severity",
			Expected: fmt.Sprintf("sum==%d", r.Summary.TotalFindings),
			Actual:   fmt.Sprintf("%d", sum),
		}}
	}
	return nil
}

// checkS5 verifies that every finding has at least one source_agents entry.
func checkS5(r *schema.SynthesisReport) []Violation {
	var out []Violation
	for i, f := range r.Findings {
		if len(f.SourceAgents) == 0 {
			out = append(out, Violation{
				Path:     fmt.Sprintf("findings[%d].source_agents", i),
				Expected: "at least one",
				Actual:   "empty",
			})
		}
	}
	return out
}

// checkS6 verifies deduplication: no two findings share (class, file, line) without being noted.
// A finding key is constructed as: class + "|" + first evidence file + "|" + first evidence line.
// Duplicate keys are allowed ONLY if deduplication_notes mentions at least one of the IDs.
func checkS6(r *schema.SynthesisReport) []Violation {
	notes := strings.Join(r.DeduplicationNotes, " ")
	keyToIDs := make(map[string][]string)

	// Build the key-to-IDs map.
	for _, f := range r.Findings {
		file := ""
		line := 0
		if len(f.Evidence.Files) > 0 {
			file = f.Evidence.Files[0]
		}
		if len(f.Evidence.Lines) > 0 {
			line = f.Evidence.Lines[0]
		}
		key := f.Class + "|" + file + "|" + fmt.Sprintf("%d", line)
		keyToIDs[key] = append(keyToIDs[key], f.ID)
	}

	// Check for unnoted duplicates.
	var out []Violation
	for key, ids := range keyToIDs {
		if len(ids) < 2 {
			continue // Not a duplicate.
		}
		// Duplicate detected — check if any ID is referenced in notes.
		mentioned := false
		for _, id := range ids {
			if id != "" && strings.Contains(notes, id) {
				mentioned = true
				break
			}
		}
		if !mentioned {
			out = append(out, Violation{
				Path:     "findings",
				Expected: fmt.Sprintf("duplicate (%s) referenced in deduplication_notes", key),
				Actual:   fmt.Sprintf("%s unnoted", strings.Join(ids, " + ")),
			})
		}
	}
	return out
}
