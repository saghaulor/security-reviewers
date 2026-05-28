package skillcheck_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// skillMdPath resolves the absolute path to the security-review SKILL.md.
//
// It uses runtime.Caller to find the source file location, then walks up three
// directories to reach the project root:
//
//	.../claude-security-hooks/internal/skillcheck/skill_md_test.go  (thisFile)
//	→ .../claude-security-hooks/internal/skillcheck/               (Dir × 1)
//	→ .../claude-security-hooks/internal/                          (Dir × 2)
//	→ .../claude-security-hooks/                                   (Dir × 3)
//	→ .../<project-root>/                                          (Dir × 4)
//
// This approach is CWD-independent and works when tests are run from any
// directory (e.g. `go test ./...` from the module root or `make test`).
func skillMdPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed — cannot resolve test file path")
	}
	// Walk up 4 levels: skillcheck/ → internal/ → claude-security-hooks/ → project root
	dir := thisFile
	for i := 0; i < 4; i++ {
		dir = filepath.Dir(dir)
	}
	return filepath.Join(dir, ".claude", "skills", "security-review", "SKILL.md")
}

// readSkillMd reads and returns the content of SKILL.md, failing the test if it
// cannot be read.
func readSkillMd(t *testing.T) string {
	t.Helper()
	path := skillMdPath(t)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read SKILL.md at %s: %v", path, err)
	}
	return string(content)
}

// TestSKILLMdExists verifies the skill file is present at its expected location.
// Guards against accidental deletion or path changes.
func TestSKILLMdExists(t *testing.T) {
	path := skillMdPath(t)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("SKILL.md not found at %s: %v", path, err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("SKILL.md at %s is not a regular file", path)
	}
}

// TestSKILLMdHasObjectiveBlock verifies the skill contains an <objective> block.
// Without this, the harness injects the file as context but Claude has no
// instructions to follow, causing a silent no-op.
func TestSKILLMdHasObjectiveBlock(t *testing.T) {
	content := readSkillMd(t)
	if !strings.Contains(content, "<objective>") {
		t.Error("SKILL.md is missing <objective> block\n" +
			"Hint: add an <objective>…</objective> section describing what the skill does. " +
			"Without it, Claude reads the file but has no instructions to execute.")
	}
}

// TestSKILLMdHasAllowedToolsTask verifies the frontmatter grants the Task tool.
// The security-review pipeline spawns agents via Task; without this permission
// the skill cannot dispatch any sub-agents.
func TestSKILLMdHasAllowedToolsTask(t *testing.T) {
	content := readSkillMd(t)
	// Look for "Task" in the allowed-tools frontmatter list.
	// The frontmatter is YAML and can list tools with or without leading "  - ".
	if !strings.Contains(content, "Task") {
		t.Error("SKILL.md allowed-tools frontmatter does not include 'Task'\n" +
			"Hint: the security-review pipeline dispatches agents via Task; " +
			"add '  - Task' to the allowed-tools list in the frontmatter.")
	}
}

// TestSKILLMdHasWorkflowSteps verifies the skill contains pipeline step references.
// Guards against the instructions being stripped or replaced with pure documentation.
// Checks for "Step 1", "Step 4", and "Step 7" as representative waypoints covering
// the start, middle, and end of the 8-step pipeline.
func TestSKILLMdHasWorkflowSteps(t *testing.T) {
	content := readSkillMd(t)
	missing := []string{}
	for _, marker := range []string{"Step 1", "Step 4", "Step 7"} {
		if !strings.Contains(content, marker) {
			missing = append(missing, marker)
		}
	}
	if len(missing) > 0 {
		t.Errorf("SKILL.md is missing pipeline step markers: %v\n"+
			"The 8-step pipeline (Step 1 through Step 8) must be present, either inline\n"+
			"or via an @path include that resolves to .claude/commands/security-review.md.\n"+
			"Missing: %s", missing, strings.Join(missing, ", "))
	}
}

// TestSKILLMdIsNotPureDocumentation guards against the specific regression where
// an agent writes README-style content into SKILL.md with no operational instructions.
//
// Rule: if "## Quick Start" appears in the file, "<objective>" must appear before it.
// A file that opens with Quick Start / How It Works / Installation sections and no
// objective block is a documentation file masquerading as a skill.
func TestSKILLMdIsNotPureDocumentation(t *testing.T) {
	content := readSkillMd(t)

	quickStartIdx := strings.Index(content, "## Quick Start")
	if quickStartIdx == -1 {
		// No Quick Start section — nothing to check.
		return
	}

	objectiveIdx := strings.Index(content, "<objective>")
	if objectiveIdx == -1 {
		t.Error("SKILL.md contains '## Quick Start' but no '<objective>' block at all.\n" +
			"The file appears to be documentation rather than a skill definition.\n" +
			"Fix: add <objective>…</objective> before any documentation sections, " +
			"or move documentation to README.md/INSTALL.md.")
		return
	}

	if objectiveIdx > quickStartIdx {
		t.Errorf("SKILL.md has '<objective>' at offset %d but '## Quick Start' appears earlier at offset %d.\n"+
			"The <objective> block must come before any documentation sections.\n"+
			"Fix: move <objective>…</objective> to immediately after the YAML frontmatter.",
			objectiveIdx, quickStartIdx)
	}
}
