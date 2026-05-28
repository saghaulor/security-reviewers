// Package skillcheck verifies that Claude Code skill files are well-formed
// and contain operational content, not just documentation.
//
// These tests guard against the "doc-as-skill" anti-pattern: an agent
// accidentally writing a README into SKILL.md instead of execution
// instructions, causing the skill to fire but do nothing.
package skillcheck
