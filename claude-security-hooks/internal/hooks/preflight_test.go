package hooks

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPreflight_NotSecurityAgent_SilentPass(t *testing.T) {
	ev := PreToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "general-purpose",
			Prompt:       "any prompt",
		},
		ToolUseID: "u1",
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout for non-security agent, got: %s", stdout.String())
	}
}

func TestPreflight_TaintAgent_ValidInput_SilentPass(t *testing.T) {
	ev := PreToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       `{"source":{"file":"a.go","line":1,"expr":"x","kind":"param"},"sink":{"file":"b.go","line":2,"expr":"y","kind":"cmd_exec"},"max_depth":10,"semgrep_tier":"pro"}`,
		},
		ToolUseID: "u1",
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout for valid taint input, got: %s", stdout.String())
	}
}

func TestPreflight_TaintAgent_MalformedInput_Block(t *testing.T) {
	ev := PreToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       "not json",
		},
		ToolUseID: "u1",
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for malformed taint input")
	}
	var result struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	if result.Decision != "block" {
		t.Errorf("expected decision=block, got %q", result.Decision)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("D-09")) {
		t.Errorf("expected D-09 in reason, got: %s", result.Reason)
	}
}

func TestPreflight_MalformedEvent_Block(t *testing.T) {
	stdin := bytes.NewReader([]byte("not json"))
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for malformed event")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("H7")) {
		t.Errorf("expected H7 in reason, got: %s", stdout.String())
	}
}

// TestPreflight_H7_Regression_ExtraEnvelopeFields verifies that new harness-added fields
// at any nesting level of PreToolUseEvent do NOT cause an H7 block. This is the
// preflight-side partner to the validate regression tests.
func TestPreflight_H7_Regression_ExtraEnvelopeFields(t *testing.T) {
	// Raw JSON simulating a future harness PreToolUse event with unknown fields.
	// Preflight must tolerate them for non-security agents.
	rawEvent := `{
		"session_id": "s1",
		"transcript_path": "/p",
		"cwd": "/tmp",
		"hook_event_name": "PreToolUse",
		"tool_name": "Task",
		"tool_input": {
			"subagent_type": "general-purpose",
			"prompt": "x",
			"unmodelled_future_field": true
		},
		"tool_use_id": "u1",
		"prompt": "x",
		"new_top_level_field": "some harness value"
	}`
	stdin := strings.NewReader(rawEvent)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("H7")) {
		t.Errorf("H7 block on extra envelope fields (regression): %s", stdout.String())
	}
}

func TestPreflight_InputTooLarge_ClassB(t *testing.T) {
	// Create a 2 MB payload
	largePayload := make([]byte, 2*1024*1024)
	stdin := bytes.NewReader(largePayload)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err == nil {
		t.Fatal("expected non-nil error for input > 1 MB")
	}
}

// TestPreflight_OAuthUnknownField_SchemaDocInError verifies W5 D-09 schema doc injection:
// when go-oauth-auditor receives a prompt with an unknown field (working_directory), the
// block reason MUST contain "Expected schema:" followed by a JSON Schema fragment that
// mentions oauth_locations and additionalProperties:false.
// Regression guard for the W5 D-09 schema-doc injection implemented in preflight.go.
func TestPreflight_OAuthUnknownField_SchemaDocInError(t *testing.T) {
	// Prompt with valid oauth_locations but unexpected working_directory field.
	prompt := `{"oauth_locations": {}, "working_directory": "/tmp/project"}`
	ev := PreToolUseEvent{
		SessionID:      "s1",
		TranscriptPath: "t1",
		CWD:            "/tmp",
		HookEventName:  "PreToolUse",
		ToolName:       "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-oauth-auditor",
			Prompt:       prompt,
		},
		ToolUseID: "u1",
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for oauth prompt with unknown field")
	}
	var result struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	if result.Decision != "block" {
		t.Errorf("expected decision=block, got %q", result.Decision)
	}
	// Must still identify the D-09 code and the offending field name.
	if !strings.Contains(result.Reason, "D-09") {
		t.Errorf("expected D-09 in reason, got: %s", result.Reason)
	}
	if !strings.Contains(result.Reason, "working_directory") {
		t.Errorf("expected 'working_directory' in reason, got: %s", result.Reason)
	}
	// W5: block reason must include the expected schema so the agent can self-correct.
	// RED: current preflight.go does not add "Expected schema:" → assertion fails.
	if !strings.Contains(result.Reason, "Expected schema:") {
		t.Errorf("W5 schema doc: expected block reason to contain \"Expected schema:\" fragment, got: %s", result.Reason)
	}
	// The schema fragment must mention oauth_locations (required field) and additionalProperties.
	if !strings.Contains(result.Reason, "oauth_locations") {
		t.Errorf("W5 schema doc: expected schema fragment to mention oauth_locations, got: %s", result.Reason)
	}
	if !strings.Contains(result.Reason, "additionalProperties") {
		t.Errorf("W5 schema doc: expected schema fragment to mention additionalProperties, got: %s", result.Reason)
	}
}

// TestPreflight_CartographerInput_SilentPass verifies that go-cartographer PreToolUse
// events are always allowed through silently (cartographer has no JSON prompt contract).
// This should PASS already — included to confirm the silent-pass is preserved after W5.
func TestPreflight_CartographerInput_SilentPass(t *testing.T) {
	prompt := `{"working_directory": "/some/path", "review_session_id": "abc"}`
	ev := PreToolUseEvent{
		SessionID:      "s1",
		TranscriptPath: "t1",
		CWD:            "/tmp",
		HookEventName:  "PreToolUse",
		ToolName:       "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-cartographer",
			Prompt:       prompt,
		},
		ToolUseID: "u1",
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout for cartographer (silent pass), got: %s", stdout.String())
	}
}
