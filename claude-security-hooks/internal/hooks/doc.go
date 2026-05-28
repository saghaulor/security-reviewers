// Package hooks declares the event payload structures for Pre/Post-ToolUse
// and SubagentStart hook registrations, plus the verdict emission logic
// (D-14: two-class exit policy).
//
// Hook registration targets:
//   - PreToolUse: task.Task tool, capture subagent dispatch context
//   - PostToolUse: task.Task tool, validate verdict before returning to Claude Code
//   - SubagentStart: security agent name set (D-07), capture subagent spawn context
//
// Verdict emission (D-03):
//   - FormatViolation renders structured violations per assertion ID
//   - EmitBlock writes {"decision":"block","reason":"..."} JSON
//   - Class A (block): exit 0 after writing decision JSON (safe to parse)
//   - Class B (error): exit 1, write prose error to stderr (operator-facing diagnostics)
//
// Live-docs fold-in (D-15): Every field in PreToolUseEvent, PostToolUseEvent,
// SubagentStartEvent must match the Claude Code hook contract verbatim:
//   - cwd, permission_mode, effort, tool_use_id
//   - agent_type (not agent_name — HAND_OFF stale)
// Failure to sync causes live payloads to block on DisallowUnknownFields.
//
// Constraint reminders:
//   - stdin is bounded to 1 MB via io.LimitReader (Wave 2); larger payloads → Class B exit 1
//   - Phase 2 trusts Claude Code's dispatch — the binary does not authenticate subagent identity (T-02-01-04)
package hooks
