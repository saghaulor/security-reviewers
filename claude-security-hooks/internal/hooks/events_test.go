package hooks_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/hooks"
)

func TestPreToolUseEvent_RoundTrip(t *testing.T) {
	payload := `{
		"session_id": "test-session",
		"transcript_path": "/path/to/transcript",
		"cwd": "/home/user/project",
		"hook_event_name": "PreToolUse",
		"tool_name": "Task",
		"tool_input": {
			"subagent_type": "go-taint-tracer",
			"prompt": "Review this code",
			"description": "Security review"
		},
		"tool_use_id": "tool-123",
		"permission_mode": "manual",
		"effort": {"level": "high"}
	}`

	var event hooks.PreToolUseEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if event.ToolInput.SubagentType != "go-taint-tracer" {
		t.Errorf("SubagentType = %q, want 'go-taint-tracer'", event.ToolInput.SubagentType)
	}
	if event.CWD != "/home/user/project" {
		t.Errorf("CWD = %q, want '/home/user/project'", event.CWD)
	}
	if event.PermissionMode != "manual" {
		t.Errorf("PermissionMode = %q, want 'manual'", event.PermissionMode)
	}
	if event.Effort.Level != "high" {
		t.Errorf("Effort.Level = %q, want 'high'", event.Effort.Level)
	}
	if event.ToolUseID != "tool-123" {
		t.Errorf("ToolUseID = %q, want 'tool-123'", event.ToolUseID)
	}
}

func TestPreToolUseEvent_RejectsUnknownField(t *testing.T) {
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "PreToolUse",
		"tool_name": "Task",
		"tool_input": {"subagent_type": "test", "prompt": "test"},
		"tool_use_id": "123",
		"bogus": "unknown field"
	}`

	var event hooks.PreToolUseEvent
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&event)
	if err == nil {
		t.Errorf("Decode with unknown field returned nil error, want error")
	}
}

func TestPostToolUseEvent_ContentString(t *testing.T) {
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "PostToolUse",
		"tool_name": "Task",
		"tool_input": {"subagent_type": "test", "prompt": "test"},
		"tool_use_id": "123",
		"tool_response": {
			"content": "response text"
		}
	}`

	var event hooks.PostToolUseEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	var content string
	if err := json.Unmarshal(event.ToolResponse.Content, &content); err != nil {
		t.Fatalf("Unmarshal content failed: %v", err)
	}
	if content != "response text" {
		t.Errorf("content = %q, want 'response text'", content)
	}
}

func TestPostToolUseEvent_ContentArray(t *testing.T) {
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "PostToolUse",
		"tool_name": "Task",
		"tool_input": {"subagent_type": "test", "prompt": "test"},
		"tool_use_id": "123",
		"tool_response": {
			"content": [{"type": "text", "text": "output"}]
		}
	}`

	var event hooks.PostToolUseEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	var content []map[string]interface{}
	if err := json.Unmarshal(event.ToolResponse.Content, &content); err != nil {
		t.Fatalf("Unmarshal content array failed: %v", err)
	}
	if len(content) != 1 {
		t.Errorf("content length = %d, want 1", len(content))
	}
}

func TestPostToolUseEvent_ToolResponseWithStatus(t *testing.T) {
	// Claude Code now includes tool_response.status; DisallowUnknownFields must not reject it.
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "PostToolUse",
		"tool_name": "Task",
		"tool_input": {"subagent_type": "go-taint-tracer", "prompt": "test"},
		"tool_use_id": "123",
		"tool_response": {
			"content": "verdict output",
			"status": "success"
		}
	}`

	var event hooks.PostToolUseEvent
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		t.Fatalf("Decode failed with tool_response.status present: %v", err)
	}
	if event.ToolResponse.Status != "success" {
		t.Errorf("ToolResponse.Status = %q, want 'success'", event.ToolResponse.Status)
	}
}

func TestPostToolUseEvent_AgentToolWithTopLevelPrompt(t *testing.T) {
	// Claude Code Agent tool PostToolUse events include a top-level "prompt" field.
	// DisallowUnknownFields must not reject it (H7 fix).
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "PostToolUse",
		"tool_name": "Agent",
		"tool_input": {"subagent_type": "gsd-executor", "prompt": "do the thing", "description": "Execute plan"},
		"tool_use_id": "123",
		"tool_response": {"content": "result", "status": "success"},
		"prompt": "do the thing"
	}`

	var event hooks.PostToolUseEvent
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		t.Fatalf("Decode failed with top-level prompt (Agent tool): %v", err)
	}
	if event.Prompt != "do the thing" {
		t.Errorf("Prompt = %q, want 'do the thing'", event.Prompt)
	}
}

func TestPreToolUseEvent_AgentToolWithTopLevelPrompt(t *testing.T) {
	// Claude Code Agent tool PreToolUse events may include a top-level "prompt" field.
	// DisallowUnknownFields must not reject it (H7 fix, symmetric with PostToolUse).
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "PreToolUse",
		"tool_name": "Agent",
		"tool_input": {"subagent_type": "gsd-executor", "prompt": "do the thing", "description": "Execute plan"},
		"tool_use_id": "123",
		"prompt": "do the thing"
	}`

	var event hooks.PreToolUseEvent
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		t.Fatalf("Decode failed with top-level prompt (Agent PreToolUse): %v", err)
	}
	if event.Prompt != "do the thing" {
		t.Errorf("Prompt = %q, want 'do the thing'", event.Prompt)
	}
}

func TestSubagentStartEvent_UsesAgentType(t *testing.T) {
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "SubagentStart",
		"agent_type": "go-taint-tracer",
		"agent_id": "agent-123",
		"prompt": "Review this"
	}`

	var event hooks.SubagentStartEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if event.AgentType != "go-taint-tracer" {
		t.Errorf("AgentType = %q, want 'go-taint-tracer'", event.AgentType)
	}
}

func TestSubagentStartEvent_RejectsAgentName(t *testing.T) {
	// Test that agent_name (HAND_OFF stale) is rejected
	payload := `{
		"session_id": "test",
		"transcript_path": "/path",
		"cwd": "/home",
		"hook_event_name": "SubagentStart",
		"agent_name": "go-taint-tracer",
		"agent_id": "agent-123",
		"prompt": "Review this"
	}`

	var event hooks.SubagentStartEvent
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&event)
	if err == nil {
		t.Errorf("Decode with agent_name returned nil error, want error for unknown field")
	}
}
