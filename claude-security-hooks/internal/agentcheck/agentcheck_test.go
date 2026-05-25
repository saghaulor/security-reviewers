package agentcheck_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/agentcheck"
	"github.com/saghaulor/claude-security-hooks/internal/hooks"
)

// repoRootRel is the relative path from the test's working directory
// (claude-security-hooks/internal/agentcheck/) to the repository root.
const repoRootRel = "../../../"

// tracerAgents are the four agent files subject to the D-12 forbidden-tools check.
// go-cartographer legitimately has Bash; synthesis legitimately has Write.
// Only these four must exclude all forbidden tools.
var tracerAgents = []string{
	"go-taint-tracer",
	"go-authz-tracer",
	"go-oauth-auditor",
	"invariant-checker",
}

// forbiddenTools lists the tools that D-12 prohibits from tracer agent allowlists.
var forbiddenTools = []string{"Grep", "Bash", "Edit", "Write"}

// TestTracerAgentsExcludeForbiddenTools parses each tracer agent file and asserts
// that none of the D-12 forbidden tools (Grep, Bash, Edit, Write) appear in
// their tools: frontmatter field. This test is RED until Wave 1 writes the agent
// files under .claude/agents/.
func TestTracerAgentsExcludeForbiddenTools(t *testing.T) {
	agentsDir := filepath.Join(repoRootRel, ".claude", "agents")

	for _, agentName := range tracerAgents {
		agentName := agentName // capture range variable
		t.Run(agentName, func(t *testing.T) {
			path := filepath.Join(agentsDir, agentName+".md")

			fm, err := agentcheck.ParseFrontmatter(path)
			if err != nil {
				t.Fatalf("ParseFrontmatter(%q): %v — agent file missing or unreadable (expected RED until Wave 1 writes agent files)", path, err)
			}

			violations := agentcheck.ContainsForbidden(fm.Tools, forbiddenTools)
			if len(violations) > 0 {
				t.Errorf("agent %q contains D-12 forbidden tools: %v", agentName, violations)
			}
		})
	}
}

// TestAgentNamesMatchSecurityAgentSet parses all .claude/agents/*.md files and
// asserts that the set of name: fields exactly matches hooks.SecurityAgentSet.
// This test is RED until Wave 1 writes all six agent files.
func TestAgentNamesMatchSecurityAgentSet(t *testing.T) {
	agentsDir := filepath.Join(repoRootRel, ".claude", "agents")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("ReadDir(%q): %v — .claude/agents/ directory missing or unreadable", agentsDir, err)
	}

	var parsedNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// skip non-.md files and placeholders
		if filepath.Ext(name) != ".md" || name == ".gitkeep" {
			continue
		}

		path := filepath.Join(agentsDir, name)
		fm, err := agentcheck.ParseFrontmatter(path)
		if err != nil {
			t.Errorf("ParseFrontmatter(%q): %v", path, err)
			continue
		}
		if fm.Name == "" {
			t.Errorf("agent file %q has empty name: field in frontmatter", path)
			continue
		}
		parsedNames = append(parsedNames, fm.Name)
	}

	// Build expected set from SecurityAgentSet (sort both for deterministic comparison).
	expected := make([]string, len(hooks.SecurityAgentSet))
	copy(expected, hooks.SecurityAgentSet)
	sort.Strings(expected)
	sort.Strings(parsedNames)

	if len(parsedNames) != len(expected) {
		t.Fatalf("agent name count mismatch: got %d (%v), want %d (%v)",
			len(parsedNames), parsedNames, len(expected), expected)
	}
	for i := range expected {
		if parsedNames[i] != expected[i] {
			t.Errorf("agent name[%d]: got %q, want %q", i, parsedNames[i], expected[i])
		}
	}
}

// TestParseFrontmatter_TempDirFixture tests ParseFrontmatter using a synthetic
// fixture file in t.TempDir(). This test does NOT depend on real agent files and
// therefore can go GREEN as soon as agent_check.go exists.
func TestParseFrontmatter_TempDirFixture(t *testing.T) {
	t.Run("valid frontmatter", func(t *testing.T) {
		content := `---
name: go-test-agent
model: claude-sonnet-4-6
tools: Bash, Read, ListDirectory
---
# Agent Body

Some agent instructions here.
`
		path := filepath.Join(t.TempDir(), "test-agent.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		fm, err := agentcheck.ParseFrontmatter(path)
		if err != nil {
			t.Fatalf("ParseFrontmatter: unexpected error: %v", err)
		}

		if fm.Name != "go-test-agent" {
			t.Errorf("Name: got %q, want %q", fm.Name, "go-test-agent")
		}
		if fm.Model != "claude-sonnet-4-6" {
			t.Errorf("Model: got %q, want %q", fm.Model, "claude-sonnet-4-6")
		}

		wantTools := []string{"Bash", "Read", "ListDirectory"}
		if len(fm.Tools) != len(wantTools) {
			t.Fatalf("Tools length: got %d (%v), want %d (%v)", len(fm.Tools), fm.Tools, len(wantTools), wantTools)
		}
		for i, tool := range wantTools {
			if fm.Tools[i] != tool {
				t.Errorf("Tools[%d]: got %q, want %q", i, fm.Tools[i], tool)
			}
		}
	})

	t.Run("tools with spaces around commas", func(t *testing.T) {
		content := `---
name: go-taint-tracer
model: claude-sonnet-4-6
tools: Read , ListDirectory , computer
---
`
		path := filepath.Join(t.TempDir(), "taint-tracer.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		fm, err := agentcheck.ParseFrontmatter(path)
		if err != nil {
			t.Fatalf("ParseFrontmatter: unexpected error: %v", err)
		}

		wantTools := []string{"Read", "ListDirectory", "computer"}
		if len(fm.Tools) != len(wantTools) {
			t.Fatalf("Tools length: got %d (%v), want %d (%v)", len(fm.Tools), fm.Tools, len(wantTools), wantTools)
		}
		for i, tool := range wantTools {
			if fm.Tools[i] != tool {
				t.Errorf("Tools[%d]: got %q, want %q", i, fm.Tools[i], tool)
			}
		}
	})

	t.Run("missing frontmatter block returns error", func(t *testing.T) {
		content := `# Agent Body Without Frontmatter

This file has no --- delimiters.
`
		path := filepath.Join(t.TempDir(), "no-frontmatter.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		_, err := agentcheck.ParseFrontmatter(path)
		if err == nil {
			t.Fatal("ParseFrontmatter: expected an error for file with no frontmatter block, got nil")
		}
	})

	t.Run("empty tools list", func(t *testing.T) {
		content := `---
name: synthesis
model: claude-sonnet-4-6
tools:
---
`
		path := filepath.Join(t.TempDir(), "synthesis.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		fm, err := agentcheck.ParseFrontmatter(path)
		if err != nil {
			t.Fatalf("ParseFrontmatter: unexpected error: %v", err)
		}

		if fm.Name != "synthesis" {
			t.Errorf("Name: got %q, want %q", fm.Name, "synthesis")
		}
		// Empty tools line: Tools should be empty or contain a single empty string
		// After trimming, empty entries should be filtered out.
		for _, tool := range fm.Tools {
			if tool == "" {
				t.Errorf("Tools contains empty string entry; should be filtered out")
			}
		}
	})
}

// TestContainsForbidden verifies the helper function directly.
func TestContainsForbidden(t *testing.T) {
	t.Run("no violations", func(t *testing.T) {
		tools := []string{"Read", "ListDirectory", "computer"}
		forbidden := []string{"Grep", "Bash", "Edit", "Write"}
		violations := agentcheck.ContainsForbidden(tools, forbidden)
		if len(violations) != 0 {
			t.Errorf("ContainsForbidden: expected no violations, got %v", violations)
		}
	})

	t.Run("single violation", func(t *testing.T) {
		tools := []string{"Read", "Bash", "ListDirectory"}
		forbidden := []string{"Grep", "Bash", "Edit", "Write"}
		violations := agentcheck.ContainsForbidden(tools, forbidden)
		if len(violations) != 1 || violations[0] != "Bash" {
			t.Errorf("ContainsForbidden: expected [Bash], got %v", violations)
		}
	})

	t.Run("multiple violations", func(t *testing.T) {
		tools := []string{"Grep", "Read", "Write", "Bash"}
		forbidden := []string{"Grep", "Bash", "Edit", "Write"}
		violations := agentcheck.ContainsForbidden(tools, forbidden)
		if len(violations) != 3 {
			t.Errorf("ContainsForbidden: expected 3 violations (Grep, Write, Bash), got %v", violations)
		}
	})

	t.Run("empty tools list", func(t *testing.T) {
		tools := []string{}
		forbidden := []string{"Grep", "Bash", "Edit", "Write"}
		violations := agentcheck.ContainsForbidden(tools, forbidden)
		if len(violations) != 0 {
			t.Errorf("ContainsForbidden: expected no violations for empty tools, got %v", violations)
		}
	})
}

// TestSettingsJsonHookRegistration parses .claude/settings.json and asserts:
// 1. PreToolUse event exists with at least one command containing "claude-security-hooks preflight" with timeout 1.
// 2. PostToolUse event exists with at least one command containing "claude-security-hooks validate" with timeout 5.
// 3. SubagentStart event exists with at least one command containing "claude-security-hooks inject-context" with timeout 1.
// 4. SubagentStop event must NOT be registered.
// This test is RED until Task 2 writes .claude/settings.json.
func TestSettingsJsonHookRegistration(t *testing.T) {
	settingsPath := filepath.Join(repoRootRel, ".claude", "settings.json")

	settings, err := agentcheck.ParseHookSettings(settingsPath)
	if err != nil {
		t.Fatalf("ParseHookSettings(%q): %v — settings.json missing or unreadable (expected RED until Task 2)", settingsPath, err)
	}

	// Check for required events
	requiredEvents := map[string]struct {
		subcommand string
		timeout    int
	}{
		"PreToolUse":    {"preflight", 1},
		"PostToolUse":   {"validate", 5},
		"SubagentStart": {"inject-context", 1},
	}

	for event, expected := range requiredEvents {
		entries, exists := settings.Hooks[event]
		if !exists {
			t.Errorf("event %q not found in hooks", event)
			continue
		}

		found := false
		for _, entry := range entries {
			for _, cmd := range entry.Hooks {
				if strings.Contains(cmd.Command, expected.subcommand) && cmd.Timeout == expected.timeout {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("event %q: no command with subcommand %q and timeout %d found", event, expected.subcommand, expected.timeout)
		}
	}

	// Ensure SubagentStop is NOT registered
	if _, exists := settings.Hooks["SubagentStop"]; exists {
		t.Errorf("SubagentStop event should NOT be registered (D-05)")
	}
}
