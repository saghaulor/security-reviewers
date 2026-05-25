package hooks

import (
	"bytes"
	"encoding/json"
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

func TestPreflight_UnknownFieldInEvent_Block(t *testing.T) {
	body := []byte(`{"session_id":"s1","transcript_path":"t1","cwd":"/tmp","hook_event_name":"PreToolUse","tool_name":"Task","tool_input":{"subagent_type":"general-purpose","prompt":"p"},"tool_use_id":"u1","bogus_field":"x"}`)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := Preflight(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("Preflight returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for unknown field")
	}
	var result struct {
		Decision string `json:"decision"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	if result.Decision != "block" {
		t.Errorf("expected decision=block, got %q", result.Decision)
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
