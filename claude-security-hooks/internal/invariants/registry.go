package invariants

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Violation struct {
	Path     string // e.g. "path[0].step", or "" for boolean-style
	Expected string
	Actual   string
}
