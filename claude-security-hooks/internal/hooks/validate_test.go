package hooks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupWorkspaceTemp creates a temporary workspace directory with .planning marker.
// Returns the temp dir path; caller must defer removal.
func setupWorkspaceTemp(t *testing.T) string {
	tmpdir := t.TempDir()
	planningDir := filepath.Join(tmpdir, ".planning")
	if err := os.MkdirAll(planningDir, 0755); err != nil {
		t.Fatalf("failed to create .planning dir: %v", err)
	}
	return tmpdir
}

func TestValidate_NotSecurityAgent_SilentPass(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "general-purpose",
			Prompt:       "any prompt",
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(`"any content"`),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout for non-security agent, got: %s", stdout.String())
	}
}

func TestValidate_TaintAgent_PassingVerdict(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	verdict := `{"verdict":"unreachable","confidence":"high","path":[],"semgrep":{"tier":"pro","rule_id":"test","ran":true,"finding":true},"gopls":{"references_calls":0,"definition_calls":0,"implementation_calls":0,"call_hierarchy_calls":0,"branches_explored":0,"branches_unexplored":0,"interface_fanout_max":0,"goroutine_boundaries_crossed":0}}`
	contentBytes, _ := json.Marshal(verdict)
	// Input must have semgrep_tier matching the verdict's semgrep.tier
	input := `{"source":{"file":"a.go","line":1,"expr":"x","kind":"param"},"sink":{"file":"b.go","line":2,"expr":"y","kind":"cmd_exec"},"max_depth":10,"semgrep_tier":"pro"}`
	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       input,
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(contentBytes),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout for passing verdict, got: %s", stdout.String())
	}
}

func TestValidate_TaintAgent_FailingVerdict_Block(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	// Create a verdict that violates T2 (bad verdict value)
	verdict := `{"verdict":"unknown","confidence":"high","path":[],"semgrep":{"tier":"pro","rule_id":"test","ran":true,"finding":true},"gopls":{"references_calls":0,"definition_calls":0,"implementation_calls":0,"call_hierarchy_calls":0,"branches_explored":0,"branches_unexplored":0,"interface_fanout_max":0,"goroutine_boundaries_crossed":0}}`
	contentBytes, _ := json.Marshal(verdict)
	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       "{}",
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(contentBytes),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for failing verdict")
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
	if !bytes.Contains(stdout.Bytes(), []byte("T2")) {
		t.Errorf("expected T2 in reason, got: %s", result.Reason)
	}
}

func TestValidate_CodeFenceStripping(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	verdict := `{"verdict":"unreachable","confidence":"high","path":[],"semgrep":{"tier":"pro","rule_id":"test","ran":true,"finding":true},"gopls":{"references_calls":0,"definition_calls":0,"implementation_calls":0,"call_hierarchy_calls":0,"branches_explored":0,"branches_unexplored":0,"interface_fanout_max":0,"goroutine_boundaries_crossed":0}}`
	fencedVerdict := "```json\n" + verdict + "\n```"
	contentBytes, _ := json.Marshal(fencedVerdict)
	input := `{"source":{"file":"a.go","line":1,"expr":"x","kind":"param"},"sink":{"file":"b.go","line":2,"expr":"y","kind":"cmd_exec"},"max_depth":10,"semgrep_tier":"pro"}`
	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       input,
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(contentBytes),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout after fence stripping, got: %s", stdout.String())
	}
}

func TestValidate_H7_MalformedContent_InvalidJSON(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	contentBytes, _ := json.Marshal("not valid json as verdict")
	input := `{"source":{"file":"a.go","line":1,"expr":"x","kind":"param"},"sink":{"file":"b.go","line":2,"expr":"y","kind":"cmd_exec"},"max_depth":10,"semgrep_tier":"pro"}`
	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       input,
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(contentBytes),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for invalid JSON verdict")
	}
	// Could be H7 (parse fail) or T1 (missing verdict field) - both are acceptable for invalid JSON
	if !bytes.Contains(stdout.Bytes(), []byte("H7")) && !bytes.Contains(stdout.Bytes(), []byte("T1")) {
		t.Errorf("expected H7 or T1 in reason, got: %s", stdout.String())
	}
}

func TestValidate_H7_MalformedContent_WrongType(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       "{}",
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(`42`), // number, not string
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for wrong content type")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("H7")) {
		t.Errorf("expected H7 in reason, got: %s", stdout.String())
	}
}

func TestValidate_H7_MalformedContent_Empty(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	contentBytes, _ := json.Marshal("")
	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       "{}",
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(contentBytes),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for empty content")
	}
	if !bytes.Contains(stdout.Bytes(), []byte("empty")) {
		t.Errorf("expected 'empty' in reason, got: %s", stdout.String())
	}
}

func TestValidate_H7_MalformedContent_ArraySegments(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	verdict := `{"verdict":"unreachable","confidence":"high","path":[],"semgrep":{"tier":"pro","rule_id":"test","ran":true,"finding":true},"gopls":{"references_calls":0,"definition_calls":0,"implementation_calls":0,"call_hierarchy_calls":0,"branches_explored":0,"branches_unexplored":0,"interface_fanout_max":0,"goroutine_boundaries_crossed":0}}`
	arrayContent := []map[string]interface{}{
		{"type": "text", "text": verdict},
	}
	contentBytes, _ := json.Marshal(arrayContent)
	input := `{"source":{"file":"a.go","line":1,"expr":"x","kind":"param"},"sink":{"file":"b.go","line":2,"expr":"y","kind":"cmd_exec"},"max_depth":10,"semgrep_tier":"pro"}`
	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       input,
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(contentBytes),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout for valid array segments, got: %s", stdout.String())
	}
}

func TestValidate_MultiViolation_OrderPreserved(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	// Create a verdict that violates T2 (bad verdict) AND T5 (bad confidence)
	verdict := `{"verdict":"unknown","confidence":"invalid","path":[],"semgrep":{"tier":"pro","rule_id":"test","ran":true,"finding":true},"gopls":{"references_calls":0,"definition_calls":0,"implementation_calls":0,"call_hierarchy_calls":0,"branches_explored":0,"branches_unexplored":0,"interface_fanout_max":0,"goroutine_boundaries_crossed":0}}`
	contentBytes, _ := json.Marshal(verdict)
	ev := PostToolUseEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           tmpdir,
		HookEventName: "PostToolUse",
		ToolName:      "Task",
		ToolInput: TaskToolInput{
			SubagentType: "go-taint-tracer",
			Prompt:       "{}",
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(contentBytes),
			Type:    "text",
		},
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for multiple violations")
	}
	var result struct {
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	// T2 comes before T5 in registration order
	if !bytes.Contains([]byte(result.Reason), []byte("T2")) {
		t.Errorf("expected T2 in multi-violation reason, got: %s", result.Reason)
	}
	if !bytes.Contains([]byte(result.Reason), []byte("T5")) {
		t.Errorf("expected T5 in multi-violation reason, got: %s", result.Reason)
	}
}

func TestValidate_InputTooLarge_ClassB(t *testing.T) {
	// Create a 2 MB payload
	largePayload := make([]byte, 2*1024*1024)
	stdin := bytes.NewReader(largePayload)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err == nil {
		t.Fatal("expected non-nil error for input > 1 MB")
	}
}

func TestValidate_PostToolUseEventWithStatus_NoH7Block(t *testing.T) {
	// H7 regression test: PostToolUseEvent now includes a "status" field from the harness.
	// The struct should accept it without DisallowUnknownFields causing an H7 block.
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	// Create a non-security agent event (should pass silently regardless)
	ev := PostToolUseEvent{
		SessionID:      "s1",
		TranscriptPath: "t1",
		CWD:            tmpdir,
		HookEventName:  "PostToolUse",
		ToolName:       "Task",
		ToolInput: TaskToolInput{
			SubagentType: "general-purpose",
			Prompt:       "any prompt",
		},
		ToolUseID: "u1",
		ToolResponse: ToolResponse{
			Content: json.RawMessage(`"any content"`),
			Type:    "text",
		},
		Status: "success", // New field from harness
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	// Should not contain H7 parse error
	if bytes.Contains(stdout.Bytes(), []byte("H7")) {
		t.Errorf("expected no H7 block, got: %s", stdout.String())
	}
}

// TestValidate_H7_Regression_ToolResponsePrompt is the direct regression test for the
// recurring H7 "unknown field prompt" bug (round 4). The harness echoes the original
// prompt inside tool_response; Validate() must not H7-block on it.
// Uses raw JSON — not json.Marshal — to replicate the actual harness event shape.
func TestValidate_H7_Regression_ToolResponsePrompt(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	rawEvent := `{
		"session_id": "s1",
		"transcript_path": "/p",
		"cwd": "` + tmpdir + `",
		"hook_event_name": "PostToolUse",
		"tool_name": "Task",
		"tool_input": {"subagent_type": "general-purpose", "prompt": "x"},
		"tool_use_id": "u1",
		"tool_response": {
			"content": "ok",
			"status": "success",
			"prompt": "x"
		}
	}`
	stdin := strings.NewReader(rawEvent)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("H7")) {
		t.Errorf("H7 block on tool_response.prompt (regression): %s", stdout.String())
	}
}

// TestValidate_H7_Regression_FutureEnvelopeFields verifies that any new harness field
// at any nesting level never causes an H7 parse block. This guards against round 5+
// of the recurring "unknown field" problem.
func TestValidate_H7_Regression_FutureEnvelopeFields(t *testing.T) {
	tmpdir := setupWorkspaceTemp(t)
	prev, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(prev)

	// Simulate a future harness event with completely novel fields at multiple levels.
	rawEvent := `{
		"session_id": "s1",
		"transcript_path": "/p",
		"cwd": "` + tmpdir + `",
		"hook_event_name": "PostToolUse",
		"tool_name": "Task",
		"tool_input": {
			"subagent_type": "general-purpose",
			"prompt": "x",
			"not_yet_modelled_input_field": true
		},
		"tool_use_id": "u1",
		"tool_response": {
			"content": "ok",
			"status": "success",
			"prompt": "x",
			"new_field_v42": "some future value"
		},
		"some_new_top_level_field": "value",
		"another_new_field": 99
	}`
	stdin := strings.NewReader(rawEvent)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Validate(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Validate returned error: %v", err)
	}
	if bytes.Contains(stdout.Bytes(), []byte("H7")) {
		t.Errorf("H7 block on future envelope fields (regression): %s", stdout.String())
	}
}
