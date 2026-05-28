package hooks

// SecurityAgentSet — D-07. Drift test in agents_test.go.
var SecurityAgentSet = []string{
	"go-cartographer",
	"go-taint-tracer",
	"go-authz-tracer",
	"go-oauth-auditor",
	"invariant-checker",
	"synthesis",
}

func IsSecurityAgent(subagentType string) bool {
	for _, name := range SecurityAgentSet {
		if name == subagentType {
			return true
		}
	}
	return false
}
