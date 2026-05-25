package invariants_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/saghaulor/claude-security-hooks/internal/invariants"
	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

// locateOAuthCheck finds a check in the OAuthVerdict.Checklist by ID.
func locateOAuthCheck(v *schema.OAuthVerdict, checkID string) *schema.ChecklistEntry {
	if v == nil || v.Checklist == nil {
		return nil
	}
	for i := range v.Checklist {
		if v.Checklist[i].CheckID == checkID {
			return &v.Checklist[i]
		}
	}
	return nil
}

// locateOAuthJointCheck finds a check in the OAuthVerdict.Checklist by ID (for joint tests).
func locateOAuthJointCheck(v *schema.OAuthVerdict, checkID string) *schema.ChecklistEntry {
	return locateOAuthCheck(v, checkID)
}

// locateOAuthJointInvariant retrieves the Check function from OAuthJointInvariants registry by ID.
func locateOAuthJointInvariant(t *testing.T, id string) func(*schema.OAuthInput, *schema.OAuthVerdict) []invariants.Violation {
	t.Helper()
	for _, inv := range invariants.OAuthJointInvariants {
		if inv.ID == id {
			return inv.Check
		}
	}
	t.Fatalf("%s not registered in OAuthJointInvariants", id)
	return nil
}

func TestOA1_VerdictRequiredFields(t *testing.T) {
	cases := []struct {
		name string
		in   *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: valid verdict with profile and checklist",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "TEST-001", Spec: "RFC 6749", Status: "pass"},
				},
			},
			want: nil,
		},
		{
			name: "bad: profile is empty",
			in: &schema.OAuthVerdict{
				Profile:   "",
				Checklist: []schema.ChecklistEntry{},
			},
			want: []invariants.Violation{
				{Path: "profile", Expected: "non-empty", Actual: ""},
			},
		},
		{
			name: "bad: checklist is nil",
			in: &schema.OAuthVerdict{
				Profile:   "oauth_2_1",
				Checklist: nil,
			},
			want: []invariants.Violation{
				{Path: "checklist", Expected: "present", Actual: "nil"},
			},
		},
	}

	var checkOA1 func(*schema.OAuthVerdict) []invariants.Violation
	for _, inv := range invariants.OAuthInvariants {
		if inv.ID == "OA1" {
			checkOA1 = inv.Check
			break
		}
	}
	if checkOA1 == nil {
		t.Fatal("OA1 not registered in OAuthInvariants")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOA1(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA1 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOA2_SpecReferenceFormat(t *testing.T) {
	cases := []struct {
		name string
		in   *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: RFC 6749 with section",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "TEST-001", Spec: "RFC 6749 §2.3.1", Status: "pass"},
				},
			},
			want: nil,
		},
		{
			name: "ok: RFC 9700",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "TEST-002", Spec: "RFC 9700", Status: "pass"},
				},
			},
			want: nil,
		},
		{
			name: "ok: draft-ietf-oauth-v2-1-15",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "TEST-003", Spec: "draft-ietf-oauth-v2-1-15", Status: "pass"},
				},
			},
			want: nil,
		},
		{
			name: "bad: made-up-spec",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "TEST-004", Spec: "made-up-spec", Status: "pass"},
				},
			},
			want: []invariants.Violation{
				{Path: "checklist[0].spec", Expected: "matches RFC <num> or draft-<id>", Actual: "made-up-spec"},
			},
		},
	}

	var checkOA2 func(*schema.OAuthVerdict) []invariants.Violation
	for _, inv := range invariants.OAuthInvariants {
		if inv.ID == "OA2" {
			checkOA2 = inv.Check
			break
		}
	}
	if checkOA2 == nil {
		t.Fatal("OA2 not registered in OAuthInvariants")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOA2(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA2 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOA3_PassRequiresEvidence(t *testing.T) {
	cases := []struct {
		name string
		in   *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: pass with evidence file and line",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{
						CheckID: "TEST-001",
						Spec:    "RFC 6749",
						Status:  "pass",
						Evidence: schema.ChecklistEvidence{
							File: "x.go",
							Line: 42,
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "bad: pass without evidence",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{
						CheckID: "TEST-002",
						Spec:    "RFC 6749",
						Status:  "pass",
						Evidence: schema.ChecklistEvidence{
							File: "",
							Line: 0,
						},
					},
				},
			},
			want: []invariants.Violation{
				{
					Path:     "checklist[0].evidence",
					Expected: "file and line populated (status=pass)",
					Actual:   `file="" line=0`,
				},
			},
		},
		{
			name: "ok: fail without evidence",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{
						CheckID: "TEST-003",
						Spec:    "RFC 6749",
						Status:  "fail",
						Evidence: schema.ChecklistEvidence{
							File: "",
							Line: 0,
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "ok: not_applicable without evidence",
			in: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{
						CheckID: "TEST-004",
						Spec:    "RFC 6749",
						Status:  "not_applicable",
						Evidence: schema.ChecklistEvidence{
							File: "",
							Line: 0,
						},
					},
				},
			},
			want: nil,
		},
	}

	var checkOA3 func(*schema.OAuthVerdict) []invariants.Violation
	for _, inv := range invariants.OAuthInvariants {
		if inv.ID == "OA3" {
			checkOA3 = inv.Check
			break
		}
	}
	if checkOA3 == nil {
		t.Fatal("OA3 not registered in OAuthInvariants")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOA3(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA3 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOA4_TaintPairsInOAuthLocations(t *testing.T) {
	cases := []struct {
		name string
		in   *schema.OAuthInput
		v    *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: taint pair source/sink match oauth locations",
			in: &schema.OAuthInput{
				OAuthLocations: schema.OAuthLocations{
					AuthorizeEndpoint: &schema.EndpointLoc{
						File: "x.go",
						Line: 10,
					},
				},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "TEST-001", Spec: "RFC 6749", Status: "pass"},
				},
				TaintPairs: []schema.TaintPair{
					{
						Source: schema.OAuthEndpoint{File: "x.go", Line: 10},
						Sink:   schema.OAuthEndpoint{File: "x.go", Line: 10},
					},
				},
			},
			want: nil,
		},
		{
			name: "bad: taint pair source not in locations",
			in: &schema.OAuthInput{
				OAuthLocations: schema.OAuthLocations{
					AuthorizeEndpoint: &schema.EndpointLoc{
						File: "x.go",
						Line: 10,
					},
				},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "TEST-002", Spec: "RFC 6749", Status: "pass"},
				},
				TaintPairs: []schema.TaintPair{
					{
						Source: schema.OAuthEndpoint{File: "other.go", Line: 99},
						Sink:   schema.OAuthEndpoint{File: "x.go", Line: 10},
					},
				},
			},
			want: []invariants.Violation{
				{Path: "taint_pairs[0].source", Expected: "file+line in input.oauth_locations", Actual: "other.go:99"},
			},
		},
	}

	var checkOA4 func(*schema.OAuthInput, *schema.OAuthVerdict) []invariants.Violation
	for _, inv := range invariants.OAuthJointInvariants {
		if inv.ID == "OA4" {
			checkOA4 = inv.Check
			break
		}
	}
	if checkOA4 == nil {
		t.Fatal("OA4 not registered in OAuthJointInvariants")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOA4(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA4 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOA5_OAuth21PKCEChecksPresent(t *testing.T) {
	cases := []struct {
		name string
		in   *schema.OAuthInput
		v    *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: oauth_2_1 + pkce feature has PKCE check with pass status",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_1",
				FeaturesInUse: []string{"pkce"},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "OAUTH21-PKCE-S256", Spec: "RFC 7636", Status: "pass"},
				},
			},
			want: nil,
		},
		{
			name: "bad: oauth_2_1 + pkce feature but no PKCE check",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_1",
				FeaturesInUse: []string{"pkce"},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "OTHER-001", Spec: "RFC 6749", Status: "pass"},
				},
			},
			want: []invariants.Violation{
				{Path: "checklist", Expected: "PKCE-classified check with status != not_applicable (oauth_2_1 + pkce in features)", Actual: "missing or N/A"},
			},
		},
		{
			name: "ok: oauth_2_0 (not 2_1) skips PKCE requirement",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_0",
				FeaturesInUse: []string{"pkce"},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_0",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "OTHER-001", Spec: "RFC 6749", Status: "pass"},
				},
			},
			want: nil,
		},
	}

	var checkOA5 func(*schema.OAuthInput, *schema.OAuthVerdict) []invariants.Violation
	for _, inv := range invariants.OAuthJointInvariants {
		if inv.ID == "OA5" {
			checkOA5 = inv.Check
			break
		}
	}
	if checkOA5 == nil {
		t.Fatal("OA5 not registered in OAuthJointInvariants")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOA5(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA5 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOA6_OAuth20ImplicitFlowCheck(t *testing.T) {
	cases := []struct {
		name string
		in   *schema.OAuthInput
		v    *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: oauth_2_0 + authorize_endpoint has implicit-flow check evaluated",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_0",
				OAuthLocations: schema.OAuthLocations{
					AuthorizeEndpoint: &schema.EndpointLoc{
						File: "x.go",
						Line: 10,
					},
				},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_0",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "OAUTH20-IMPLICIT-FLOW", Spec: "RFC 6749 §4.2", Status: "fail"},
				},
			},
			want: nil,
		},
		{
			name: "bad: oauth_2_0 + authorize_endpoint but no implicit-flow check",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_0",
				OAuthLocations: schema.OAuthLocations{
					AuthorizeEndpoint: &schema.EndpointLoc{
						File: "x.go",
						Line: 10,
					},
				},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_0",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "OTHER-001", Spec: "RFC 6749", Status: "pass"},
				},
			},
			want: []invariants.Violation{
				{Path: "checklist", Expected: "implicit-flow-disallowed check evaluated for oauth_2_0 with authorize_endpoint set", Actual: "missing"},
			},
		},
		{
			name: "ok: oauth_2_1 (not 2_0) skips implicit check",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_1",
				OAuthLocations: schema.OAuthLocations{
					AuthorizeEndpoint: &schema.EndpointLoc{
						File: "x.go",
						Line: 10,
					},
				},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "OTHER-001", Spec: "RFC 6749", Status: "pass"},
				},
			},
			want: nil,
		},
	}

	var checkOA6 func(*schema.OAuthInput, *schema.OAuthVerdict) []invariants.Violation
	for _, inv := range invariants.OAuthJointInvariants {
		if inv.ID == "OA6" {
			checkOA6 = inv.Check
			break
		}
	}
	if checkOA6 == nil {
		t.Fatal("OA6 not registered in OAuthJointInvariants")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOA6(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA6 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOA7_AbsentFeatureMustBeNotApplicable_NotPass(t *testing.T) {
	cases := []struct {
		name string
		in   *schema.OAuthInput
		v    *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: dpop check not_applicable when dpop not in features",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_1",
				FeaturesInUse: []string{"pkce"},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "RFC9449-DPoP-htm", Spec: "RFC 9449", Status: "not_applicable"},
				},
			},
			want: nil,
		},
		{
			name: "bad: dpop check pass when dpop not in features",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_1",
				FeaturesInUse: []string{"pkce"},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "RFC9449-DPoP-htm", Spec: "RFC 9449", Status: "pass"},
				},
			},
			want: []invariants.Violation{
				{Path: "checklist[0].status", Expected: "not_applicable (dpop not in features_in_use)", Actual: "pass"},
			},
		},
		{
			name: "ok: dpop check pass when dpop IS in features",
			in: &schema.OAuthInput{
				TargetProfile: "oauth_2_1",
				FeaturesInUse: []string{"dpop", "pkce"},
			},
			v: &schema.OAuthVerdict{
				Profile: "oauth_2_1",
				Checklist: []schema.ChecklistEntry{
					{CheckID: "RFC9449-DPoP-htm", Spec: "RFC 9449", Status: "pass"},
				},
			},
			want: nil,
		},
	}

	var checkOA7 func(*schema.OAuthInput, *schema.OAuthVerdict) []invariants.Violation
	for _, inv := range invariants.OAuthJointInvariants {
		if inv.ID == "OA7" {
			checkOA7 = inv.Check
			break
		}
	}
	if checkOA7 == nil {
		t.Fatal("OA7 not registered in OAuthJointInvariants")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOA7(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA7 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestOA8_SessionIDMatchesInput validates OA8 (JOINT): if input has review_session_id,
// output review_session_id must match exactly.
func TestOA8_SessionIDMatchesInput(t *testing.T) {
	check := locateOAuthJointInvariant(t, "OA8")
	cases := []struct {
		name string
		in   *schema.OAuthInput
		v    *schema.OAuthVerdict
		want []invariants.Violation
	}{
		{
			name: "ok: both have matching session ID",
			in: &schema.OAuthInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.OAuthVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "bad: input has session ID but output mismatches",
			in: &schema.OAuthInput{
				ReviewSessionID: "uuid-123",
			},
			v: &schema.OAuthVerdict{
				ReviewSessionID: "uuid-456",
			},
			want: []invariants.Violation{{Path: "review_session_id", Expected: "uuid-123", Actual: "uuid-456"}},
		},
		{
			name: "ok: input omits session ID (optional field)",
			in: &schema.OAuthInput{
				ReviewSessionID: "",
			},
			v: &schema.OAuthVerdict{
				ReviewSessionID: "uuid-123",
			},
			want: nil,
		},
		{
			name: "ok: both omit session ID",
			in: &schema.OAuthInput{
				ReviewSessionID: "",
			},
			v: &schema.OAuthVerdict{
				ReviewSessionID: "",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := check(tc.in, tc.v)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("OA8 check mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
