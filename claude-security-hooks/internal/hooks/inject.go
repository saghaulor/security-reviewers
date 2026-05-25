package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// InjectContext is a no-op stub per D-08: read the SubagentStartEvent from stdin,
// verify it parses against the live-docs schema (D-15 fold-in: agent_type, not
// agent_name), and exit 0 silently. Phase 3 or Phase 5 will populate real
// injection logic once context requirements are known.
func InjectContext(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	limited := io.LimitReader(stdin, stdinCapBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("inject-context: read stdin: %w", err)
	}
	if len(body) > stdinCapBytes {
		return fmt.Errorf("inject-context: input exceeds %d bytes", stdinCapBytes)
	}
	var ev SubagentStartEvent
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&ev); err != nil {
		return EmitBlock(stdout, fmt.Sprintf("H7: SubagentStartEvent parse: %v", err))
	}
	_ = ev // D-08: no further work in Phase 2.
	return nil
}
