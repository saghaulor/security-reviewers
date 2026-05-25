package agentcheck

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Frontmatter represents the parsed YAML frontmatter of an agent definition file.
type Frontmatter struct {
	Name  string
	Model string
	Tools []string
}

// ParseFrontmatter parses the YAML frontmatter block from a .md file and returns
// a Frontmatter struct with Name, Model, and Tools fields extracted.
// Returns an error if the file cannot be read or if the frontmatter block is malformed.
func ParseFrontmatter(path string) (*Frontmatter, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Extract frontmatter block between --- delimiters
	text := string(content)
	if !strings.HasPrefix(text, "---") {
		return nil, fmt.Errorf("file does not start with frontmatter delimiter (---)")
	}

	// Find the closing --- delimiter
	rest := strings.TrimPrefix(text, "---")
	parts := strings.SplitN(rest, "---", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("closing frontmatter delimiter (---) not found")
	}

	frontmatterText := parts[0]

	// Parse YAML-like key: value format
	fm := &Frontmatter{}
	for _, line := range strings.Split(frontmatterText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			fm.Name = value
		case "model":
			fm.Model = value
		case "tools":
			// Parse comma-separated tools list
			if value != "" {
				tools := strings.Split(value, ",")
				for _, t := range tools {
					t = strings.TrimSpace(t)
					if t != "" {
						fm.Tools = append(fm.Tools, t)
					}
				}
			}
		}
	}

	return fm, nil
}

// ContainsForbidden returns a list of forbidden tools that appear in the tools list.
func ContainsForbidden(tools []string, forbidden []string) []string {
	var violations []string
	for _, tool := range tools {
		for _, forbiddenTool := range forbidden {
			if tool == forbiddenTool {
				violations = append(violations, tool)
				break
			}
		}
	}
	return violations
}

// HookSettings represents the parsed structure of .claude/settings.json.
// The structure is: {"hooks": {"PreToolUse": [...], "PostToolUse": [...], ...}}
type HookSettings struct {
	Hooks map[string][]HookEntry `json:"hooks"`
}

// HookEntry represents a single event registration entry (with matcher and nested hooks array).
type HookEntry struct {
	Matcher string       `json:"matcher"`
	Hooks   []HookCommand `json:"hooks"`
}

// HookCommand represents a command to execute when a hook fires.
type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

// ParseHookSettings parses .claude/settings.json and returns the HookSettings.
func ParseHookSettings(path string) (*HookSettings, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings.json: %w", err)
	}

	var settings HookSettings
	if err := json.Unmarshal(content, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &settings, nil
}
