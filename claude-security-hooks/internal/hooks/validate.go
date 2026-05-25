package hooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/invariants"
)

func Validate(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	limited := io.LimitReader(stdin, stdinCapBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("validate: read stdin: %w", err)
	}
	if len(body) > stdinCapBytes {
		return fmt.Errorf("validate: input exceeds %d bytes", stdinCapBytes)
	}
	var ev PostToolUseEvent
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ev); err != nil {
		return EmitBlock(stdout, fmt.Sprintf("H7: PostToolUseEvent parse: %v", err))
	}
	if !IsSecurityAgent(ev.ToolInput.SubagentType) {
		return nil
	}
	root := invariants.ResolveWorkspaceRoot()
	if root == "" {
		return EmitBlock(stdout, "workspace_root_not_found")
	}
	prev, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("validate: chdir(%q): %w", root, err)
	}
	defer func() { _ = os.Chdir(prev) }()

	verdictText, err := extractTextContent(ev.ToolResponse.Content)
	if err != nil {
		return EmitBlock(stdout, fmt.Sprintf("H7: tool_response.content not parseable as text: %v", err))
	}
	if strings.TrimSpace(verdictText) == "" {
		return EmitBlock(stdout, "H7: content empty")
	}
	stripped := stripCodeFences(verdictText)
	// Attempt parse on stripped; if it fails, fall back to raw verdictText.
	reasons, _ := runAgentValidation(ev.ToolInput.SubagentType, stripped, ev.ToolInput.Prompt)
	if len(reasons) == 1 && strings.Contains(reasons[0], "verdict parse") {
		fallback, _ := runAgentValidation(ev.ToolInput.SubagentType, verdictText, ev.ToolInput.Prompt)
		if len(fallback) > 0 && !strings.Contains(fallback[0], "verdict parse") {
			reasons = fallback
		}
	}
	// Synthesis: also run S1 directory check (review-report.{json,md} expected at workspace root).
	if ev.ToolInput.SubagentType == "synthesis" {
		for _, inv := range invariants.SynthesisDirInvariants {
			for _, vio := range inv.Check(".") {
				reasons = append(reasons, FormatViolation(inv.ID, inv.Description, vio))
			}
		}
	}
	if len(reasons) > 0 {
		return EmitBlock(stdout, FormatBlockReason(reasons))
	}
	return nil
}

// extractTextContent handles string OR array per RESEARCH.md Pitfall 4.
func extractTextContent(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", errors.New("empty raw message")
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var arr []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return "", fmt.Errorf("content is neither string nor array: %w", err)
	}
	var sb strings.Builder
	for _, seg := range arr {
		for _, key := range []string{"text", "content"} {
			if rawSeg, ok := seg[key]; ok {
				var sv string
				if err := json.Unmarshal(rawSeg, &sv); err == nil {
					sb.WriteString(sv)
				}
			}
		}
	}
	return sb.String(), nil
}
