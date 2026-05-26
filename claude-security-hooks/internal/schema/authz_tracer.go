package schema

// AuthzInput represents the input to the authz tracer.
type AuthzInput struct {
	Routes               []AuthzRoute            `json:"routes"`
	AuthzPrimitives      []AuthzInputPrimitive   `json:"authz_primitives"`
	SensitiveOperations  []SensitiveOp           `json:"sensitive_operations"`
	ReviewSessionID      string                  `json:"review_session_id,omitempty"`
	CodeRef              string                  `json:"code_ref,omitempty"`
	CodeRefDirty         bool                    `json:"code_ref_dirty,omitempty"`
}

type AuthzRoute struct {
	Router          string            `json:"router"`
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	Handler         Handler           `json:"handler"`
	MiddlewareChain []MiddlewareEntry `json:"middleware_chain,omitempty"`
}

type AuthzInputPrimitive struct {
	FQN  string `json:"fqn"`
	Kind string `json:"kind"`
}

type SensitiveOp struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Kind string `json:"kind"`
}

// AuthzVerdict represents the output of the authz tracer.
type AuthzVerdict struct {
	Summary             AuthzSummary    `json:"summary"`
	Findings            []AuthzFinding  `json:"findings,omitempty"`
	WeakPrimitives      []WeakPrimitive `json:"weak_primitives,omitempty"`
	ReviewSessionID     string          `json:"review_session_id,omitempty"`
	CodeRef             string          `json:"code_ref,omitempty"`
	CodeRefDirty        bool            `json:"code_ref_dirty,omitempty"`
}

type AuthzSummary struct {
	RoutesTotal    int `json:"routes_total"`
	Protected      int `json:"protected"`
	Missing        int `json:"missing"`
	Weak           int `json:"weak"`
	IdorRisk       int `json:"idor_risk"`
	PublicIntentional int `json:"public_intentional"`
}

type AuthzFinding struct {
	Route       string    `json:"route"`
	Issue       string    `json:"issue"`
	Evidence    Evidence  `json:"evidence"`
	Confidence  string    `json:"confidence"`
}

type Evidence struct {
	File          string `json:"file"`
	Line          int    `json:"line"`
	Explanation   string `json:"explanation"`
}

type WeakPrimitive struct {
	FQN    string `json:"fqn"`
	Reason string `json:"reason"`
}
