package hooks

import "strings"

// stripCodeFences strips a single ```json or ``` fence (with surrounding
// whitespace) from the leading and trailing edges of s. Internal ``` sequences
// are left untouched. Per CONTEXT.md Claude's Discretion (line 100): the caller
// may fall back to the raw input if the stripped form fails to parse.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
