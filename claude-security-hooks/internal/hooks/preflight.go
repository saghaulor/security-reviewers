package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

const stdinCapBytes = 1 << 20 // 1 MB

// Compact JSON Schema summaries for each security agent's input contract.
// Used in D-09 block reasons to help agents self-correct on schema mismatch (W5).
const taintInputSchemaDoc = `{"type":"object","required":["source","sink"],"properties":{"source":{"type":"object","required":["file","line","expr","kind"]},"sink":{"type":"object","required":["file","line","expr","kind"]},"max_depth":{"type":"integer"},"semgrep_tier":{"type":"string"}},"additionalProperties":false}`
const authzInputSchemaDoc = `{"type":"object","required":["routes"],"properties":{"routes":{"type":"array"},"authz_primitives":{"type":"array"},"sensitive_operations":{"type":"array"},"review_session_id":{"type":"string"},"code_ref":{"type":"string"},"code_ref_dirty":{"type":"boolean"}},"additionalProperties":false}`
const oauthInputSchemaDoc = `{"type":"object","required":["oauth_locations"],"properties":{"oauth_locations":{"type":"object"},"target_profile":{"type":"string","enum":["oauth_2_1","oauth_2_0","oauth_2_0_with_9700_bcp"]},"features_in_use":{"type":"array"},"review_session_id":{"type":"string"},"code_ref":{"type":"string"},"code_ref_dirty":{"type":"boolean"}},"additionalProperties":false}`
const invariantInputSchemaDoc = `{"type":"object","required":["flow_name","invariants"],"properties":{"flow_name":{"type":"string"},"invariants":{"type":"array","items":{"type":"object","required":["id","statement"]}},"review_session_id":{"type":"string"},"code_ref":{"type":"string"},"code_ref_dirty":{"type":"boolean"}},"additionalProperties":false}`

// Note: envelope parsing uses json.Unmarshal (lenient). See validate.go for rationale.

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
	if err := json.Unmarshal(body, &ev); err != nil {
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
			return fmt.Sprintf("D-09: taint input parse: %v\nExpected schema: %s", err, taintInputSchemaDoc)
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
			return fmt.Sprintf("D-09: authz input parse: %v\nExpected schema: %s", err, authzInputSchemaDoc)
		}
	case "go-oauth-auditor":
		var in schema.OAuthInput
		dec := json.NewDecoder(strings.NewReader(promptJSON))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return fmt.Sprintf("D-09: oauth input parse: %v\nExpected schema: %s", err, oauthInputSchemaDoc)
		}
	case "invariant-checker":
		var in schema.InvariantCheckerInput
		dec := json.NewDecoder(strings.NewReader(promptJSON))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return fmt.Sprintf("D-09: invariant-checker input parse: %v\nExpected schema: %s", err, invariantInputSchemaDoc)
		}
	case "synthesis":
		// Synthesis input is a directory path string. No JSON parse required.
		// The directory existence is verified at validate-time via S1.
		return ""
	}
	return ""
}
