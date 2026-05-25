package hooks

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestInject_ValidSubagentStart_SilentPass(t *testing.T) {
	ev := SubagentStartEvent{
		SessionID:     "s1",
		TranscriptPath: "t1",
		CWD:           "/tmp",
		HookEventName: "SubagentStart",
		AgentType:     "go-taint-tracer",
		AgentID:       "a1",
		Prompt:        "test prompt",
	}
	body, _ := json.Marshal(ev)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := InjectContext(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("InjectContext returned error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("expected empty stdout for valid SubagentStart, got: %s", stdout.String())
	}
}

func TestInject_MalformedJSON_Block(t *testing.T) {
	stdin := bytes.NewReader([]byte("not json"))
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := InjectContext(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("InjectContext returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for malformed JSON")
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
	if !bytes.Contains(stdout.Bytes(), []byte("H7")) {
		t.Errorf("expected H7 in reason, got: %s", result.Reason)
	}
}

func TestInject_AgentNameRejected_Block(t *testing.T) {
	// D-15 fold-in: agent_name (HAND_OFF stale) must be rejected as an unknown field
	body := []byte(`{"session_id":"s1","transcript_path":"t1","cwd":"/tmp","hook_event_name":"SubagentStart","agent_name":"go-taint-tracer","agent_id":"a1","prompt":"test"}`)
	stdin := bytes.NewReader(body)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := InjectContext(stdin, stdout, stderr)
	if err != nil {
		t.Errorf("InjectContext returned error: %v", err)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected stdout block for unknown field (agent_name)")
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

func TestInject_InputTooLarge_ClassB(t *testing.T) {
	// Create a 2 MB payload
	largePayload := make([]byte, 2*1024*1024)
	stdin := bytes.NewReader(largePayload)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	err := InjectContext(stdin, stdout, stderr)
	if err == nil {
		t.Fatal("expected non-nil error for input > 1 MB")
	}
}
