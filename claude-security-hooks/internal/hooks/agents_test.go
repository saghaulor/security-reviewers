package hooks_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/saghaulor/claude-security-hooks/internal/hooks"
)

func TestSecurityAgentSet_MatchesSpec(t *testing.T) {
	expected := []string{
		"go-cartographer",
		"go-taint-tracer",
		"go-authz-tracer",
		"go-oauth-auditor",
		"invariant-checker",
		"synthesis",
	}
	if diff := cmp.Diff(expected, hooks.SecurityAgentSet); diff != "" {
		t.Errorf("SecurityAgentSet mismatch:\n%s", diff)
	}
}

func TestIsSecurityAgent_KnownAgent(t *testing.T) {
	tests := []string{
		"go-cartographer",
		"go-taint-tracer",
		"go-authz-tracer",
		"go-oauth-auditor",
		"invariant-checker",
		"synthesis",
	}
	for _, agent := range tests {
		t.Run(agent, func(t *testing.T) {
			if !hooks.IsSecurityAgent(agent) {
				t.Errorf("IsSecurityAgent(%q) = false, want true", agent)
			}
		})
	}
}

func TestIsSecurityAgent_UnknownAgent(t *testing.T) {
	tests := []string{
		"general-purpose",
		"main",
		"",
		"unknown-agent",
	}
	for _, agent := range tests {
		t.Run(agent, func(t *testing.T) {
			if hooks.IsSecurityAgent(agent) {
				t.Errorf("IsSecurityAgent(%q) = true, want false", agent)
			}
		})
	}
}
