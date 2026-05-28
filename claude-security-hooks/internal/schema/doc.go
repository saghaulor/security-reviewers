// Package schema declares typed Go structs mirroring each security
// subagent's verdict JSON schema. All decoders in internal/hooks use
// json.NewDecoder(input).DisallowUnknownFields() per D-10; therefore
// any new field in an agent's verdict that is not declared on the
// corresponding struct here causes the validator to block at runtime.
//
// Schema-bump protocol: when a verdict gains a field, the agent prompt
// AND the matching struct here MUST change in the same commit.
//
// Field names are JSON-tag-driven; Go field names are CamelCase mirrors
// of the snake_case JSON keys defined by CON-schema-* in
// .planning/intel/constraints.md.
//
// Do NOT define custom UnmarshalJSON methods on these structs — they
// interact badly with DisallowUnknownFields for embedded types
// (golang/go#22533). Use json.RawMessage + second-stage typed decode if
// polymorphism is needed.
package schema
