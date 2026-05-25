package schema

// TaintInput represents the input to the taint tracer.
type TaintInput struct {
	Source      TaintEndpoint `json:"source"`
	Sink        TaintEndpoint `json:"sink"`
	MaxDepth    int           `json:"max_depth"`
	SemgrepTier string        `json:"semgrep_tier"`
}

type TaintEndpoint struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Expr string `json:"expr"`
	Kind string `json:"kind"`
}

// TaintVerdict represents the output of the taint tracer.
type TaintVerdict struct {
	Verdict                string                 `json:"verdict"`
	Confidence             string                 `json:"confidence"`
	Path                   []TaintPathStep        `json:"path,omitempty"`
	Semgrep                SemgrepEvidence        `json:"semgrep"`
	Gopls                  GoplsEvidence          `json:"gopls"`
	SanitizersUnverified   []SanitizerUnverified  `json:"sanitizers_unverified,omitempty"`
	Notes                  string                 `json:"notes,omitempty"`
}

type TaintPathStep struct {
	File            string  `json:"file"`
	Line            int     `json:"line"`
	Expr            string  `json:"expr"`
	Step            string  `json:"step"`
	Callee          *string `json:"callee,omitempty"`
	SanitizerKind   *string `json:"sanitizer_kind,omitempty"`
	Interface       *string `json:"interface,omitempty"`
	Implementer     *string `json:"implementer,omitempty"`
}

type SemgrepEvidence struct {
	Tier   string `json:"tier"`
	RuleID string `json:"rule_id"`
	Ran    bool   `json:"ran"`
	Finding bool  `json:"finding"`
	Error  *string `json:"error,omitempty"`
}

type GoplsEvidence struct {
	ReferencesCalls         int `json:"references_calls"`
	DefinitionCalls         int `json:"definition_calls"`
	ImplementationCalls     int `json:"implementation_calls"`
	CallHierarchyCalls      int `json:"call_hierarchy_calls"`
	BranchesExplored        int `json:"branches_explored"`
	BranchesUnexplored      int `json:"branches_unexplored"`
	InterfaceFanoutMax      int `json:"interface_fanout_max"`
	GoroutineBoundariesCrossed int `json:"goroutine_boundaries_crossed"`
}

type SanitizerUnverified struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}
