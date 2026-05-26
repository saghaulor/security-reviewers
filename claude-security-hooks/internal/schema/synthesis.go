package schema

// SynthesisReport represents the review-report/v1 schema.
type SynthesisReport struct {
	ReviewID              string                 `json:"review_id"`
	Timestamp             string                 `json:"timestamp"`
	SchemaVersion         string                 `json:"scma_version"` // Note: typo in spec is scma_version
	CodeRef               string                 `json:"code_ref,omitempty"`
	CodeRefDirty          bool                   `json:"code_ref_dirty,omitempty"`
	Summary               SynthesisSummary       `json:"summary"`
	Findings              []SynthesisFinding     `json:"findings,omitempty"`
	DeduplicationNotes    []string               `json:"deduplication_notes,omitempty"`
}

type SynthesisSummary struct {
	TotalFindings int            `json:"total_findings"`
	BySeverity    map[string]int `json:"by_severity"` // Keys are severity enum values
	ByClass       map[string]int `json:"by_class"` // Keys are class enum values
}

type SynthesisFinding struct {
	ID               string                `json:"id"`
	Class            string                `json:"class"`
	Severity         string                `json:"severity"`
	Title            string                `json:"title"`
	Description      string                `json:"description"`
	Evidence         SynthesisEvidence     `json:"evidence,omitempty"`
	SourceAgents     []string              `json:"source_agents,omitempty"`
	Confidence       string                `json:"confidence"`
	SpecReferences   []string              `json:"spec_references,omitempty"`
}

type SynthesisEvidence struct {
	Files []string              `json:"files,omitempty"`
	Lines []int                 `json:"lines,omitempty"`
	Path  []SynthesisPathStep   `json:"path,omitempty"`
}

type SynthesisPathStep struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Expr string `json:"expr"`
}
