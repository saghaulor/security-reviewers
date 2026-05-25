package invariants

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/saghaulor/claude-security-hooks/internal/schema"
)

type OAuthInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.OAuthVerdict) []Violation
}

type OAuthJointInvariant struct {
	ID          string
	Description string
	Severity    Severity
	Check       func(*schema.OAuthInput, *schema.OAuthVerdict) []Violation
}

var OAuthInvariants = []OAuthInvariant{
	{ID: "OA1", Description: "verdict has required fields (profile + checklist)", Severity: SeverityCritical, Check: checkOA1},
	{ID: "OA2", Description: "every checklist[*].spec matches RFC <num> or draft-<id>", Severity: SeverityHigh, Check: checkOA2},
	{ID: "OA3", Description: "status=pass requires evidence.file AND evidence.line populated", Severity: SeverityHigh, Check: checkOA3},
}

var OAuthJointInvariants = []OAuthJointInvariant{
	{ID: "OA4", Description: "taint_pairs source/sink file+line must match input.oauth_locations", Severity: SeverityHigh, Check: checkOA4Joint},
	{ID: "OA5", Description: "oauth_2_1 + pkce feature requires PKCE checks with non-not_applicable status", Severity: SeverityHigh, Check: checkOA5Joint},
	{ID: "OA6", Description: "oauth_2_0 + authorize_endpoint requires implicit-flow-disallowed check evaluated", Severity: SeverityHigh, Check: checkOA6Joint},
	{ID: "OA7", Description: "absent feature checks must be not_applicable, never pass", Severity: SeverityHigh, Check: checkOA7Joint},
	{ID: "OA8", Description: "if input has review_session_id, verdict review_session_id must match exactly", Severity: SeverityHigh, Check: checkOA8Joint},
}

// oauthCheckFeature maps check_id substring → feature name.
// Used by OA7 to determine whether a check's target feature is present in input.
var oauthCheckFeature = map[string]string{
	"PKCE":          "pkce",
	"DPoP":          "dpop",
	"PAR":           "par",
	"CIBA":          "ciba",
	"Exchange":      "token_exchange",
	"DCR":           "dynamic_client_registration",
	"Introspection": "introspection",
	"Revocation":    "revocation",
}

// featureForCheckID returns the feature this check_id targets, or "" if no mapping.
func featureForCheckID(checkID string) string {
	for needle, feat := range oauthCheckFeature {
		if strings.Contains(checkID, needle) {
			return feat
		}
	}
	return ""
}

var specRefRegex = regexp.MustCompile(`(?i)^(rfc\s*\d+|draft-[a-z0-9-]+)`)

// --- OA1 ---
func checkOA1(v *schema.OAuthVerdict) []Violation {
	var out []Violation
	if v.Profile == "" {
		out = append(out, Violation{Path: "profile", Expected: "non-empty", Actual: ""})
	}
	if v.Checklist == nil {
		out = append(out, Violation{Path: "checklist", Expected: "present", Actual: "nil"})
	}
	return out
}

// --- OA2 ---
func checkOA2(v *schema.OAuthVerdict) []Violation {
	var out []Violation
	for i, c := range v.Checklist {
		if !specRefRegex.MatchString(strings.TrimSpace(c.Spec)) {
			out = append(out, Violation{Path: fmt.Sprintf("checklist[%d].spec", i), Expected: "matches RFC <num> or draft-<id>", Actual: c.Spec})
		}
	}
	return out
}

// --- OA3 ---
func checkOA3(v *schema.OAuthVerdict) []Violation {
	var out []Violation
	for i, c := range v.Checklist {
		if c.Status != "pass" {
			continue
		}
		if c.Evidence.File == "" || c.Evidence.Line == 0 {
			out = append(out, Violation{
				Path:     fmt.Sprintf("checklist[%d].evidence", i),
				Expected: "file and line populated (status=pass)",
				Actual:   fmt.Sprintf("file=%q line=%d", c.Evidence.File, c.Evidence.Line),
			})
		}
	}
	return out
}

// --- OA4 (joint) ---
func checkOA4Joint(in *schema.OAuthInput, v *schema.OAuthVerdict) []Violation {
	locKey := func(file string, line int) string { return file + ":" + strconv.Itoa(line) }
	locSet := make(map[string]struct{})
	for _, loc := range []*schema.EndpointLoc{
		in.OAuthLocations.AuthorizeEndpoint, in.OAuthLocations.TokenEndpoint,
		in.OAuthLocations.CallbackHandler, in.OAuthLocations.TokenStorage, in.OAuthLocations.RefreshPath,
	} {
		if loc == nil {
			continue
		}
		locSet[locKey(loc.File, loc.Line)] = struct{}{}
	}
	var out []Violation
	check := func(path, file string, line int) {
		if _, ok := locSet[locKey(file, line)]; !ok {
			out = append(out, Violation{Path: path, Expected: "file+line in input.oauth_locations", Actual: fmt.Sprintf("%s:%d", file, line)})
		}
	}
	for i, p := range v.TaintPairs {
		check(fmt.Sprintf("taint_pairs[%d].source", i), p.Source.File, p.Source.Line)
		check(fmt.Sprintf("taint_pairs[%d].sink", i), p.Sink.File, p.Sink.Line)
	}
	return out
}

// --- OA5 (joint) ---
func checkOA5Joint(in *schema.OAuthInput, v *schema.OAuthVerdict) []Violation {
	if in.TargetProfile != "oauth_2_1" {
		return nil
	}
	usesPKCE := false
	for _, f := range in.FeaturesInUse {
		if f == "pkce" {
			usesPKCE = true
			break
		}
	}
	if !usesPKCE {
		return nil
	}
	// Require at least one PKCE-classified check with Status != "not_applicable".
	for _, c := range v.Checklist {
		if strings.Contains(c.CheckID, "PKCE") && c.Status != "not_applicable" {
			return nil
		}
	}
	return []Violation{{
		Path:     "checklist",
		Expected: "PKCE-classified check with status != not_applicable (oauth_2_1 + pkce in features)",
		Actual:   "missing or N/A",
	}}
}

// --- OA6 (joint) ---
func checkOA6Joint(in *schema.OAuthInput, v *schema.OAuthVerdict) []Violation {
	if in.TargetProfile != "oauth_2_0" {
		return nil
	}
	if in.OAuthLocations.AuthorizeEndpoint == nil {
		return nil
	}
	// Require an implicit-flow check evaluated (status in {pass, fail}).
	for _, c := range v.Checklist {
		if strings.Contains(strings.ToLower(c.CheckID), "implicit") {
			if c.Status == "pass" || c.Status == "fail" {
				return nil
			}
		}
	}
	return []Violation{{
		Path:     "checklist",
		Expected: "implicit-flow-disallowed check evaluated for oauth_2_0 with authorize_endpoint set",
		Actual:   "missing",
	}}
}

// --- OA7 (joint) ---
func checkOA7Joint(in *schema.OAuthInput, v *schema.OAuthVerdict) []Violation {
	featSet := make(map[string]struct{}, len(in.FeaturesInUse))
	for _, f := range in.FeaturesInUse {
		featSet[f] = struct{}{}
	}
	var out []Violation
	for i, c := range v.Checklist {
		feat := featureForCheckID(c.CheckID)
		if feat == "" {
			continue // unknown mapping; cannot enforce OA7 for this check
		}
		if _, present := featSet[feat]; present {
			continue
		}
		if c.Status == "pass" {
			out = append(out, Violation{
				Path:     fmt.Sprintf("checklist[%d].status", i),
				Expected: fmt.Sprintf("not_applicable (%s not in features_in_use)", feat),
				Actual:   "pass",
			})
		}
	}
	return out
}

// --- OA8 (JOINT) ---
func checkOA8Joint(in *schema.OAuthInput, v *schema.OAuthVerdict) []Violation {
	// If input does not specify a session ID, no check is performed (field is optional).
	if in.ReviewSessionID == "" {
		return nil
	}
	// If input specifies a session ID, verdict must echo it exactly.
	if v.ReviewSessionID != in.ReviewSessionID {
		return []Violation{{
			Path:     "review_session_id",
			Expected: in.ReviewSessionID,
			Actual:   v.ReviewSessionID,
		}}
	}
	return nil
}
