package invariants_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/saghaulor/claude-security-hooks/internal/invariants"
	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// locateCartographerCheck finds a predicate by ID from the CartographerInvariants slice.
// This proves registry wiring at test time, per RESEARCH.md Pattern 4.
func locateCartographerCheck(t *testing.T, id string) func(*schema.CartographerIndex) []invariants.Violation {
	t.Helper()
	for _, inv := range invariants.CartographerInvariants {
		if inv.ID == id {
			return inv.Check
		}
	}
	t.Fatalf("%s not registered in CartographerInvariants", id)
	return nil
}

// TestA1_ValidJSONShape — Top-level shape invariants.
// A1: output is valid go-index/v1 JSON with required top-level fields.
// Cases: (1) valid full index → nil; (2) empty SchemaVersion → Violation; (3) nil SinksByKind → Violation
func TestA1_ValidJSONShape(t *testing.T) {
	checkA1 := locateCartographerCheck(t, "A1")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: valid full index with all fields",
			in: &schema.CartographerIndex{
				SchemaVersion:   "go-index/v1",
				GraphVersion:    "abc123",
				RoutersDetected: []string{"chi"},
				Entrypoints: []schema.Entrypoint{
					{
						Router: "chi",
						Method: "GET",
						Path:   "/api/users",
						Handler: schema.Handler{
							FQN:  "pkg/main.Handler",
							File: "handler.go",
							Line: 10,
						},
					},
				},
				SinksByKind: map[string][]schema.SinkLocation{},
			},
			want: nil,
		},
		{
			name: "bad: empty schema_version",
			in: &schema.CartographerIndex{
				SchemaVersion: "",
				SinksByKind:   map[string][]schema.SinkLocation{},
			},
			want: []invariants.Violation{{Path: "schema_version", Expected: "non-empty", Actual: ""}},
		},
		{
			name: "bad: nil sinks_by_kind",
			in: &schema.CartographerIndex{
				SchemaVersion: "go-index/v1",
				SinksByKind:   nil,
			},
			want: []invariants.Violation{{Path: "sinks_by_kind", Expected: "present", Actual: "nil"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA1(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A1 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA2_SchemaVersionLiteral — schema_version must be exactly "go-index/v1".
func TestA2_SchemaVersionLiteral(t *testing.T) {
	checkA2 := locateCartographerCheck(t, "A2")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: schema_version is 'go-index/v1'",
			in: &schema.CartographerIndex{
				SchemaVersion: "go-index/v1",
			},
			want: nil,
		},
		{
			name: "bad: schema_version is 'go-index/v2'",
			in: &schema.CartographerIndex{
				SchemaVersion: "go-index/v2",
			},
			want: []invariants.Violation{{Path: "schema_version", Expected: "go-index/v1", Actual: "go-index/v2"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA2(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A2 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA3_EntrypointsHandlerFieldsPopulated — All required handler fields must be present.
// Cases: 1 positive + 4 negatives (missing FQN, File, Line, Router)
func TestA3_EntrypointsHandlerFieldsPopulated(t *testing.T) {
	checkA3 := locateCartographerCheck(t, "A3")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: all entrypoint handler fields populated",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Router: "chi",
						Method: "POST",
						Path:   "/api/orders",
						Handler: schema.Handler{
							FQN:  "pkg/order.Create",
							File: "order.go",
							Line: 42,
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "bad: entrypoint missing handler.fqn",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Router: "chi",
						Method: "GET",
						Path:   "/api/users",
						Handler: schema.Handler{
							FQN:  "",
							File: "user.go",
							Line: 10,
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "entrypoints[0].handler.fqn", Expected: "non-empty", Actual: ""}},
		},
		{
			name: "bad: entrypoint missing handler.file",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Router: "chi",
						Method: "GET",
						Path:   "/api/users",
						Handler: schema.Handler{
							FQN:  "pkg/user.Get",
							File: "",
							Line: 10,
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "entrypoints[0].handler.file", Expected: "non-empty", Actual: ""}},
		},
		{
			name: "bad: entrypoint handler.line is 0",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Router: "chi",
						Method: "GET",
						Path:   "/api/users",
						Handler: schema.Handler{
							FQN:  "pkg/user.Get",
							File: "user.go",
							Line: 0,
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "entrypoints[0].handler.line", Expected: ">0", Actual: "0"}},
		},
		{
			name: "bad: entrypoint missing router",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Router: "",
						Method: "GET",
						Path:   "/api/users",
						Handler: schema.Handler{
							FQN:  "pkg/user.Get",
							File: "user.go",
							Line: 10,
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "entrypoints[0].router", Expected: "non-empty", Actual: ""}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA3(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A3 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA4_RoutersDetectedInAllowSet — routers_detected must be in {net/http, chi, gin, gorilla/mux, echo, fiber, httprouter, custom}.
func TestA4_RoutersDetectedInAllowSet(t *testing.T) {
	checkA4 := locateCartographerCheck(t, "A4")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: routers are in allow set",
			in: &schema.CartographerIndex{
				RoutersDetected: []string{"chi", "net/http"},
			},
			want: nil,
		},
		{
			name: "bad: contains unknown router",
			in: &schema.CartographerIndex{
				RoutersDetected: []string{"custom", "unknown-router"},
			},
			want: []invariants.Violation{{Path: "routers_detected[1]", Expected: "in {chi,custom,echo,fiber,gin,gorilla/mux,httprouter,net/http}", Actual: "unknown-router"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA4(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A4 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA5_RoutersWithoutEntrypointsHasWarning — If routers non-empty and entrypoints empty, warnings must contain "router_detected_but_no_routes".
func TestA5_RoutersWithoutEntrypointsHasWarning(t *testing.T) {
	checkA5 := locateCartographerCheck(t, "A5")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: routers present and entrypoints present",
			in: &schema.CartographerIndex{
				RoutersDetected: []string{"chi"},
				Entrypoints: []schema.Entrypoint{
					{
						Router: "chi",
						Method: "GET",
						Path:   "/api/users",
						Handler: schema.Handler{
							FQN:  "pkg/user.Get",
							File: "user.go",
							Line: 10,
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "bad: routers detected but no entrypoints, missing warning",
			in: &schema.CartographerIndex{
				RoutersDetected: []string{"chi"},
				Entrypoints:     []schema.Entrypoint{},
				Warnings:        []string{},
			},
			want: []invariants.Violation{{Path: "warnings", Expected: "contains 'router_detected_but_no_routes'", Actual: "missing"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA5(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A5 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA6_CitedNodeIDsExistInGraph — All cited node IDs must exist in graphify-out/graph.json.
func TestA6_CitedNodeIDsExistInGraph(t *testing.T) {
	checkA6 := locateCartographerCheck(t, "A6")

	// Override the graph path for testing
	originalPath := invariants.CartographerGraphPath
	t.Cleanup(func() { invariants.CartographerGraphPath = originalPath })

	cases := []struct {
		name      string
		graphPath string
		in        *schema.CartographerIndex
		want      []invariants.Violation
	}{
		{
			name:      "ok: all cited node IDs exist in graph",
			graphPath: "testdata/graphify-out/graph.json",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Handler: schema.Handler{
							FQN: "pkg/main.main",
						},
					},
				},
				AuthzPrimitives: []schema.AuthzPrimitive{
					{
						FQN: "pkg/handler.Login",
					},
				},
			},
			want: nil,
		},
		{
			name:      "bad: cited node ID not in graph",
			graphPath: "testdata/graphify-out/graph.json",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Handler: schema.Handler{
							FQN: "pkg/fake.Symbol",
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "entrypoints[0].handler.fqn", Expected: "present in testdata/graphify-out/graph.json", Actual: "pkg/fake.Symbol"}},
		},
		{
			name:      "bad: graph file missing",
			graphPath: "testdata/missing/graph.json",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Handler: schema.Handler{
							FQN: "pkg/main.main",
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "testdata/missing/graph.json", Expected: "exists", Actual: "missing"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			invariants.CartographerGraphPath = tc.graphPath
			got := checkA6(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A6 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA7_CitedFilePathsExist — Every cited file path must exist in the workspace.
func TestA7_CitedFilePathsExist(t *testing.T) {
	checkA7 := locateCartographerCheck(t, "A7")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: cited file path exists",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Handler: schema.Handler{
							File: "testdata/workspace/handler.go",
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "bad: cited file path does not exist",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Handler: schema.Handler{
							File: "testdata/workspace/missing.go",
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "entrypoints[0].handler.file", Expected: "exists", Actual: "testdata/workspace/missing.go"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA7(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A7 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA8_CitedLineNumbersWithinFileBounds — Every cited line number must be within the file's bounds.
func TestA8_CitedLineNumbersWithinFileBounds(t *testing.T) {
	checkA8 := locateCartographerCheck(t, "A8")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: line number within file bounds",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Handler: schema.Handler{
							File: "testdata/workspace/handler.go",
							Line: 10,
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "bad: line number exceeds file bounds",
			in: &schema.CartographerIndex{
				Entrypoints: []schema.Entrypoint{
					{
						Handler: schema.Handler{
							File: "testdata/workspace/handler.go",
							Line: 28, // handler.go has 27 lines
						},
					},
				},
			},
			want: []invariants.Violation{{Path: "entrypoints[0].handler.line", Expected: "<=27", Actual: "28"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA8(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A8 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA9_AuthzPrimitivesOnlyBlocking — All authz_primitives must have blocking=true.
func TestA9_AuthzPrimitivesOnlyBlocking(t *testing.T) {
	checkA9 := locateCartographerCheck(t, "A9")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: all authz primitives blocking",
			in: &schema.CartographerIndex{
				AuthzPrimitives: []schema.AuthzPrimitive{
					{
						FQN:      "pkg/auth.Guard",
						Blocking: true,
					},
					{
						FQN:      "pkg/auth.Middleware",
						Blocking: true,
					},
				},
			},
			want: nil,
		},
		{
			name: "bad: authz primitive with blocking=false",
			in: &schema.CartographerIndex{
				AuthzPrimitives: []schema.AuthzPrimitive{
					{
						FQN:      "pkg/auth.Middleware",
						Blocking: false,
					},
				},
			},
			want: []invariants.Violation{{Path: "authz_primitives[0].blocking", Expected: "true", Actual: "false"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA9(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A9 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA10_AgentDidNotModifyTrackedFiles_NoOpStub — No-op stub; returns nil regardless of input.
// Per Phase 2 decision, tool-usage telemetry not available in verdict JSON.
// TODO(phase-5): enforce via SubagentStop telemetry or tool-call summary in verdict.
func TestA10_AgentDidNotModifyTrackedFiles_NoOpStub(t *testing.T) {
	checkA10 := locateCartographerCheck(t, "A10")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: valid index (stub returns nil)",
			in: &schema.CartographerIndex{
				SchemaVersion: "go-index/v1",
			},
			want: nil,
		},
		{
			name: "ok: stub does not produce false positives",
			in: &schema.CartographerIndex{
				SchemaVersion: "go-index/v1",
			},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA10(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A10 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestA11_AgentBashCommandsAllowed_NoOpStub — No-op stub; returns nil regardless of input.
// Per Phase 2 decision, tool-usage telemetry not available in verdict JSON.
// TODO(phase-5): enforce via SubagentStop telemetry or tool-call summary in verdict.
func TestA11_AgentBashCommandsAllowed_NoOpStub(t *testing.T) {
	checkA11 := locateCartographerCheck(t, "A11")

	cases := []struct {
		name string
		in   *schema.CartographerIndex
		want []invariants.Violation
	}{
		{
			name: "ok: valid index (stub returns nil)",
			in: &schema.CartographerIndex{
				SchemaVersion: "go-index/v1",
			},
			want: nil,
		},
		{
			name: "ok: stub does not produce false positives",
			in: &schema.CartographerIndex{
				SchemaVersion: "go-index/v1",
			},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkA11(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("A11 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
