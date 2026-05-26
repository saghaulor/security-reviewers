package hooks

import "encoding/json"

// EffortLevel mirrors the live docs effort object {"level": "low|medium|high|xhigh|max"}.
type EffortLevel struct {
	Level string `json:"level,omitempty"`
}

// TaskToolInput — tool_input shape when tool_name == "Task".
type TaskToolInput struct {
	SubagentType    string `json:"subagent_type"`
	Prompt          string `json:"prompt"`
	Description     string `json:"description,omitempty"`
	RunInBackground bool   `json:"run_in_background,omitempty"`
	Model           string `json:"model,omitempty"`
	Isolation       string `json:"isolation,omitempty"`
}

// PreToolUseEvent — Claude Code hook input for PreToolUse on Task/Agent.
// Prompt: Claude Code Agent tool includes top-level "prompt" in hook events (H7 fix).
type PreToolUseEvent struct {
	SessionID      string        `json:"session_id"`
	TranscriptPath string        `json:"transcript_path"`
	CWD            string        `json:"cwd"`
	HookEventName  string        `json:"hook_event_name"`
	ToolName       string        `json:"tool_name"`
	ToolInput      TaskToolInput `json:"tool_input"`
	ToolUseID      string        `json:"tool_use_id"`
	Prompt         string        `json:"prompt,omitempty"`
	PermissionMode string        `json:"permission_mode,omitempty"`
	Effort         EffortLevel   `json:"effort,omitempty"`
	Status         string        `json:"status,omitempty"`
}

// ToolResponse — content is string OR array per live docs; RawMessage handles both.
// Status field added: Claude Code runtime now includes tool_response.status (e.g. "success", "error").
type ToolResponse struct {
	Content json.RawMessage `json:"content"`
	Type    string          `json:"type,omitempty"`
	Status  string          `json:"status,omitempty"`
}

// PostToolUseEvent adds tool_response.
// Prompt: Claude Code Agent tool includes top-level "prompt" in hook events (H7 fix).
type PostToolUseEvent struct {
	SessionID      string        `json:"session_id"`
	TranscriptPath string        `json:"transcript_path"`
	CWD            string        `json:"cwd"`
	HookEventName  string        `json:"hook_event_name"`
	ToolName       string        `json:"tool_name"`
	ToolInput      TaskToolInput `json:"tool_input"`
	ToolUseID      string        `json:"tool_use_id"`
	ToolResponse   ToolResponse  `json:"tool_response"`
	Prompt         string        `json:"prompt,omitempty"`
	PermissionMode string        `json:"permission_mode,omitempty"`
	Effort         EffortLevel   `json:"effort,omitempty"`
	Status         string        `json:"status,omitempty"`
}

// SubagentStartEvent — D-15 fold-in: agent_type per live docs, NOT agent_name (HAND_OFF stale).
type SubagentStartEvent struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	AgentType      string `json:"agent_type"`
	AgentID        string `json:"agent_id"`
	Prompt         string `json:"prompt"`
}
