package schema

// CartographerIndex represents the go-index/v1 schema.
type CartographerIndex struct {
	SchemaVersion    string                     `json:"schema_version"`
	GraphVersion     string                     `json:"graph_version"`
	RoutersDetected  []string                   `json:"routers_detected,omitempty"`
	Entrypoints      []Entrypoint               `json:"entrypoints,omitempty"`
	SinksByKind      map[string][]SinkLocation  `json:"sinks_by_kind,omitempty"` // Keys are kind enum values
	AuthzPrimitives  []AuthzPrimitive           `json:"authz_primitives,omitempty"`
	OAuthLocations   OAuthLocations             `json:"oauth_locations,omitempty"`
	PaymentSurface   PaymentSurface             `json:"payment_surface,omitempty"`
	VulnDeps         VulnDeps                   `json:"vuln_deps,omitempty"`
	AmbiguousNodes   []AmbiguousNode            `json:"ambiguous_nodes,omitempty"`
	Warnings         []string                   `json:"warnings,omitempty"`
}

type Entrypoint struct {
	Router           string              `json:"router"`
	Method           string              `json:"method"`
	Path             string              `json:"path"`
	Handler          Handler             `json:"handler"`
	MiddlewareChain  []MiddlewareEntry   `json:"middleware_chain,omitempty"`
}

type Handler struct {
	FQN  string `json:"fqn"`
	File string `json:"file"`
	Line int    `json:"line"`
}

type MiddlewareEntry struct {
	FQN  string `json:"fqn"`
	Kind string `json:"kind"` // global|group|route
	File string `json:"file"`
	Line int    `json:"line"`
}

type SinkLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Callee string `json:"callee"`
}

type AuthzPrimitive struct {
	FQN         string `json:"fqn"`
	Kind        string `json:"kind"` // middleware|guard|decorator
	Confidence  string `json:"confidence"` // extracted|inferred
	Blocking    bool   `json:"blocking"`
}

type OAuthLocations struct {
	AuthorizeEndpoint *EndpointLoc `json:"authorize_endpoint,omitempty"`
	TokenEndpoint     *EndpointLoc `json:"token_endpoint,omitempty"`
	CallbackHandler   *EndpointLoc `json:"callback_handler,omitempty"`
	TokenStorage      *EndpointLoc `json:"token_storage,omitempty"`
	RefreshPath       *EndpointLoc `json:"refresh_path,omitempty"`
}

type EndpointLoc struct {
	FQN  string `json:"fqn"`
	File string `json:"file"`
	Line int    `json:"line"`
}

type PaymentSurface struct {
	Files      []string `json:"files,omitempty"`
	ClusterID  string   `json:"cluster_id"`
	Confidence string   `json:"confidence"`
}

type VulnDeps struct {
	Available bool          `json:"available"`
	Findings  []VulnFinding `json:"findings,omitempty"`
}

type VulnFinding struct {
	OsvID       string   `json:"osv_id"`
	Package     string   `json:"package"`
	Symbol      string   `json:"symbol"`
	Reachable   bool     `json:"reachable"`
	CallStack   []string `json:"call_stack,omitempty"`
}

type AmbiguousNode struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}
