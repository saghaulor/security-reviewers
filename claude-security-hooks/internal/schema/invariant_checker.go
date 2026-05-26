package schema

// InvariantCheckerInput represents the input to the invariant checker.
type InvariantCheckerInput struct {
	FlowName          string            `json:"flow_name"`
	Invariants        []InputInvariant  `json:"invariants"`
	ReviewSessionID   string            `json:"review_session_id,omitempty"`
	CodeRef           string            `json:"code_ref,omitempty"`
	CodeRefDirty      bool              `json:"code_ref_dirty,omitempty"`
}

type InputInvariant struct {
	ID             string   `json:"id"`
	Statement      string   `json:"statement"`
	AnchorSymbols  []string `json:"anchor_symbols,omitempty"`
}

// InvariantCheckerVerdict represents the output of the invariant checker.
type InvariantCheckerVerdict struct {
	FlowName          string            `json:"flow_name"`
	Results           []InvariantResult `json:"results,omitempty"`
	ReviewSessionID   string            `json:"review_session_id,omitempty"`
	CodeRef           string            `json:"code_ref,omitempty"`
	CodeRefDirty      bool              `json:"code_ref_dirty,omitempty"`
}

type InvariantResult struct {
	InvariantID string              `json:"invariant_id"`
	Status      string              `json:"status"`
	Evidence    InvariantEvidence   `json:"evidence,omitempty"`
	Confidence  string              `json:"confidence"`
}

type InvariantEvidence struct {
	Files       []string `json:"files,omitempty"`
	Explanation string   `json:"explanation"`
}
