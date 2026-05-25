package schema

// OAuthInput represents the input to the OAuth auditor.
type OAuthInput struct {
	OAuthLocations OAuthLocations `json:"oauth_locations"`
	TargetProfile  string         `json:"target_profile"`
	FeaturesInUse  []string       `json:"features_in_use,omitempty"`
}

// OAuthVerdict represents the output of the OAuth auditor.
type OAuthVerdict struct {
	Profile    string             `json:"profile"`
	Checklist  []ChecklistEntry   `json:"checklist,omitempty"`
	TaintPairs []TaintPair        `json:"taint_pairs,omitempty"`
}

type ChecklistEntry struct {
	CheckID     string              `json:"check_id"`
	Spec        string              `json:"spec"`
	Status      string              `json:"status"`
	Evidence    ChecklistEvidence   `json:"evidence,omitempty"`
	Severity    string              `json:"severity"`
}

type ChecklistEvidence struct {
	File          string `json:"file"`
	Line          int    `json:"line"`
	Explanation   string `json:"explanation"`
}

type TaintPair struct {
	Source    OAuthEndpoint `json:"source"`
	Sink      OAuthEndpoint `json:"sink"`
	Rationale string        `json:"rationale"`
}

type OAuthEndpoint struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Expr string `json:"expr"`
	Kind string `json:"kind"`
}
