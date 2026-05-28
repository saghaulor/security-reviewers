package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/invariants"
)

// FormatViolation renders a Violation per D-03 block-reason format.
func FormatViolation(id, description string, v invariants.Violation) string {
	if v.Path != "" {
		return fmt.Sprintf("%s: %s expected '%s' got '%s'", id, v.Path, v.Expected, v.Actual)
	}
	return fmt.Sprintf("%s: %s", id, description)
}

// FormatBlockReason joins per-violation strings with "; " in registration order (D-03).
func FormatBlockReason(parts []string) string {
	return strings.Join(parts, "; ")
}

// EmitBlock writes the {"decision":"block","reason":"..."} JSON to w. Caller exits 0 after (D-14 Class A).
func EmitBlock(w io.Writer, reason string) error {
	return json.NewEncoder(w).Encode(struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}{Decision: "block", Reason: reason})
}
