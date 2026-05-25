package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

const stdinCapBytes = 1 << 20 // 1 MB

func Preflight(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	limited := io.LimitReader(stdin, stdinCapBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("preflight: read stdin: %w", err)
	}
	if len(body) > stdinCapBytes {
		return fmt.Errorf("preflight: input exceeds %d bytes", stdinCapBytes)
	}
	var ev PreToolUseEvent
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ev); err != nil {
		return EmitBlock(stdout, fmt.Sprintf("H7: PreToolUseEvent parse: %v", err))
	}
	if !IsSecurityAgent(ev.ToolInput.SubagentType) {
		return nil
	}
	// D-09: parse ev.ToolInput.Prompt as per-agent input contract.
	if reason := preflightPerAgent(ev.ToolInput.SubagentType, ev.ToolInput.Prompt); reason != "" {
		return EmitBlock(stdout, reason)
	}
	return nil
}

// preflightPerAgent performs inline typed decode per agent — no interface{}
// helper (preserves D-02 no-any).
func preflightPerAgent(subagentType, promptJSON string) string {
	switch subagentType {
	case "go-taint-tracer":
		var in schema.TaintInput
		dec := json.NewDecoder(strings.NewReader(promptJSON))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return fmt.Sprintf("D-09: taint input parse: %v", err)
		}
	case "go-cartographer":
		// Cartographer input is the working directory + presence of graphify-out — not a JSON schema.
		// Per HAND_OFF §3.1 there is no documented prompt-JSON contract for cartographer.
		// Preflight is a silent pass for cartographer.
		return ""
	case "go-authz-tracer":
		var in schema.AuthzInput
		dec := json.NewDecoder(strings.NewReader(promptJSON))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return fmt.Sprintf("D-09: authz input parse: %v", err)
		}
	case "go-oauth-auditor":
		var in schema.OAuthInput
		dec := json.NewDecoder(strings.NewReader(promptJSON))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return fmt.Sprintf("D-09: oauth input parse: %v", err)
		}
	case "invariant-checker":
		var in schema.InvariantCheckerInput
		dec := json.NewDecoder(strings.NewReader(promptJSON))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return fmt.Sprintf("D-09: invariant-checker input parse: %v", err)
		}
	case "synthesis":
		// Synthesis input is a directory path string. No JSON parse required.
		// The directory existence is verified at validate-time via S1.
		return ""
	}
	return ""
}
