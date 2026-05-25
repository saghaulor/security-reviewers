package hooks_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/saghaulor/claude-security-hooks/internal/hooks"
	"github.com/saghaulor/claude-security-hooks/internal/invariants"
)

func TestFormatViolation_PathStyle(t *testing.T) {
	v := invariants.Violation{
		Path:     "path[0].step",
		Expected: "source",
		Actual:   "sink",
	}
	result := hooks.FormatViolation("T3", "ignored", v)
	expected := "T3: path[0].step expected 'source' got 'sink'"
	if result != expected {
		t.Errorf("FormatViolation = %q, want %q", result, expected)
	}
}

func TestFormatViolation_BooleanStyle(t *testing.T) {
	v := invariants.Violation{
		Path: "",
	}
	result := hooks.FormatViolation("T6", "neither semgrep.ran nor gopls.references_calls > 0", v)
	expected := "T6: neither semgrep.ran nor gopls.references_calls > 0"
	if result != expected {
		t.Errorf("FormatViolation = %q, want %q", result, expected)
	}
}

func TestFormatBlockReason_JoinsWithSemicolonSpace(t *testing.T) {
	parts := []string{
		"T3: path[0].step expected 'source' got 'sink'",
		"T6: neither semgrep.ran nor gopls.references_calls > 0",
	}
	result := hooks.FormatBlockReason(parts)
	expected := "T3: path[0].step expected 'source' got 'sink'; T6: neither semgrep.ran nor gopls.references_calls > 0"
	if result != expected {
		t.Errorf("FormatBlockReason = %q, want %q", result, expected)
	}
}

func TestEmitBlock_EmitsValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := hooks.EmitBlock(&buf, "test reason")
	if err != nil {
		t.Fatalf("EmitBlock error = %v, want nil", err)
	}

	output := buf.String()
	// json.Encoder adds a trailing newline
	output = strings.TrimSpace(output)

	var decoded struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Decision != "block" {
		t.Errorf("decision = %q, want 'block'", decoded.Decision)
	}
	if decoded.Reason != "test reason" {
		t.Errorf("reason = %q, want 'test reason'", decoded.Reason)
	}
}
