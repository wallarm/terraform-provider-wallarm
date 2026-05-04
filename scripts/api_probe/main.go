// API Probe — discovers actual Wallarm API constraints for each rule type.
//
// For every rule type in `probes` below, the program:
//
//  1. POSTs /v1/objects/hint/create with a starting body (Type + minimal Action).
//  2. If 400, parses the per-field error map and adds candidate values for
//     "can't be blank" or "should be in N..M" errors, then retries.
//  3. On 200, reads back the response body, diffs it against what we sent,
//     logs which fields the API filled in (the API defaults), and deletes
//     the rule.
//  4. Logs "stuck" if no progress between retries (max 20 attempts).
//
// Output is a Markdown table at .claude/api_probe_results.md.
//
// Required env vars:
//
//	WALLARM_API_HOST       e.g. https://api.wallarm.com
//	WALLARM_API_TOKEN      API token with rule-create permission
//	WALLARM_API_CLIENT_ID  numeric client/tenant ID
//
// Optional:
//
//	API_PROBE_RULES        comma-separated list to limit probes (default: all)
//	API_PROBE_VERBOSE      set to 1 to log every request/response
//
// Run:
//
//	cd .claude/scripts/api_probe && go run .
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Probe defines a starting body for a single rule type. base is mutated
// during discovery — start with the minimum every rule type needs.
//
// TokenEnv overrides the token env var name for rule types that need
// elevated permissions (e.g. disable_stamp requires Administrator (extended)).
//
// Label disambiguates two probes that share a RuleType but probe different
// shapes (e.g. enumerated_parameters mode=exact vs mode=regexp). Defaults
// to RuleType when empty.
type Probe struct {
	Label    string
	RuleType string
	Base     map[string]any
	TokenEnv string
}

var probes = []Probe{
	// Plain rules
	{RuleType: "vpatch", Base: map[string]any{"attack_type": "xss", "point": [][]string{{"get_all"}}}},
	{RuleType: "wallarm_mode", Base: map[string]any{"mode": "monitoring"}},
	{RuleType: "api_abuse_mode", Base: map[string]any{"mode": "disabled"}},
	{RuleType: "disable_attack_type", Base: map[string]any{"attack_type": "xss", "point": [][]string{{"get_all"}}}},
	{RuleType: "disable_stamp", Base: map[string]any{"stamp": 1, "point": [][]string{{"get_all"}}}, TokenEnv: "WALLARM_API_TOKEN_EXTENDED"},
	{RuleType: "regex", Base: map[string]any{"regex": ".*", "attack_type": "xss", "point": [][]string{{"get_all"}}}},
	{RuleType: "experimental_regex", Base: map[string]any{"regex": ".*", "attack_type": "xss", "point": [][]string{{"get_all"}}}},
	// disable_regex needs an existing regex_id from the same client. Skipped
	// by default — set DISABLE_REGEX_ID env var to probe it.
	// {RuleType: "disable_regex", Base: map[string]any{"regex_id": ?, "point": [][]string{{"get_all"}}}},
	{RuleType: "binary_data", Base: map[string]any{"point": [][]string{{"get_all"}}}},
	{RuleType: "sensitive_data", Base: map[string]any{"point": [][]string{{"get_all"}}}},
	{RuleType: "uploads", Base: map[string]any{"point": [][]string{{"get_all"}}, "file_type": "docs"}},
	{RuleType: "set_response_header", Base: map[string]any{"name": "X-Probe", "values": []string{"1"}, "mode": "append"}},
	{RuleType: "parser_state", Base: map[string]any{"parser": "json_doc", "state": "enabled", "point": [][]string{{"post"}}}},
	{RuleType: "overlimit_res_settings", Base: map[string]any{"overlimit_time": 1000}},

	// Planned rule types (not yet exposed by the provider — probing API ground truth).
	{RuleType: "detailed_export", Base: map[string]any{"mode": "keep_headers"}},
	{RuleType: "response_conds", Base: map[string]any{"enable_parsers": false}},

	// Rate limiting (plain — non-mitigation)
	{RuleType: "rate_limit", Base: map[string]any{"point": [][]string{{"get_all"}}}},

	// Mitigation controls — regexp mode (default candidate)
	{Label: "brute_regexp", RuleType: "brute", Base: map[string]any{"mode": "block"}},
	{Label: "bola_regexp", RuleType: "bola", Base: map[string]any{"mode": "block"}},
	{Label: "enum_regexp", RuleType: "enum", Base: map[string]any{"mode": "block"}},
	{RuleType: "forced_browsing", Base: map[string]any{"mode": "block"}},
	{RuleType: "rate_limit_enum", Base: map[string]any{"mode": "block"}},
	{RuleType: "graphql_detection", Base: map[string]any{"mode": "block"}},
	{RuleType: "file_upload_size_limit", Base: map[string]any{}},
	// size=0 boundary probe — does the API accept zero?
	{Label: "file_upload_size_limit_size0", RuleType: "file_upload_size_limit", Base: map[string]any{"size": 0}},
	// size=very large boundary probe — discover the upper bound.
	{Label: "file_upload_size_limit_sizelarge", RuleType: "file_upload_size_limit", Base: map[string]any{"size": 999999999}},

	// Mitigation controls — exact mode (different fieldset for enumerated_parameters)
	{Label: "brute_exact", RuleType: "brute", Base: map[string]any{
		"mode": "block",
		"enumerated_parameters": map[string]any{
			"mode":   "exact",
			"points": []map[string]any{{"point": []string{"header", "REFERER"}, "sensitive": false}},
		},
	}},
	{Label: "bola_exact", RuleType: "bola", Base: map[string]any{
		"mode": "block",
		"enumerated_parameters": map[string]any{
			"mode":   "exact",
			"points": []map[string]any{{"point": []string{"header", "REFERER"}, "sensitive": false}},
		},
	}},
	{Label: "enum_exact", RuleType: "enum", Base: map[string]any{
		"mode": "block",
		"enumerated_parameters": map[string]any{
			"mode":   "exact",
			"points": []map[string]any{{"point": []string{"header", "REFERER"}, "sensitive": false}},
		},
	}},

	// Counters — none of the three accept a `point` field.
	{RuleType: "brute_counter", Base: map[string]any{}},
	{RuleType: "dirbust_counter", Base: map[string]any{}},
	{RuleType: "bola_counter", Base: map[string]any{}},

	// Credential stuffing (skipped by default — needs login_point/regex setup)
	// {RuleType: "credentials_point", Base: map[string]any{...}},
	// {RuleType: "credentials_regex", Base: map[string]any{...}},
}

// candidateValues maps known field names to a value to try when the API
// reports "can't be blank" or similar. Range errors override these.
var candidateValues = map[string]any{
	"mode":                  "block",
	"attack_type":           "xss",
	"stamp":                 1,
	"regex":                 ".*",
	"login_regex":           ".*",
	"case_sensitive":        false,
	"cred_stuff_type":       "regex",
	"size":                  1,
	"size_unit":             "kb",
	"max_depth":             5,
	"max_value_size_kb":     5,
	"max_doc_size_kb":       50,
	"max_alias_size_kb":     3,
	"max_doc_per_batch":     5,
	"introspection":         true,
	"debug_enabled":         false,
	"overlimit_time":        0, // probe: range allows 0; do we silently drop?
	"parser":                "json_doc",
	"state":                 "enabled",
	"delay":                 0, // probe: range allows 0; do we silently drop?
	"burst":                 0, // probe: range allows 0; do we silently drop?
	"rate":                  0, // probe: range allows 0; do we silently drop?
	"rsp_status":            429,
	"time_unit":             "rps",
	"name":                  "X-Probe",
	"values":                []string{"1"},
	"file_type":             "docs",
	"point":                 [][]string{{"get_all"}},
	"login_point":           [][]string{{"post"}, {"form_urlencoded", "login"}},
	"threshold":             map[string]any{"count": 5, "period": 30},
	"reaction":              map[string]any{"block_by_ip": 600},
	"enumerated_parameters": map[string]any{"mode": "regexp", "name_regexps": []string{"foo"}, "value_regexps": []string{"bar"}, "additional_parameters": false, "plain_parameters": false},
	"action":                nil, // handled specially
	"comment":               "api-probe",
	"title":                 "api-probe",
	"enable_parsers":        false,
	"limit_parsers":         []string{},
	"limit_size":            0,
	"limit_deep":            0,
	"limit_chance":          0.0,
	"validated":             false,
	"variativity_disabled":  true,
	"active":                true,
	"set":                   "",
}

// rangeRE matches "should be in N..M" / "should be in N..M, can't be blank".
var rangeRE = regexp.MustCompile(`should be in (-?\d+)\s*\.\.\s*(\d+)`)

type APIResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

type Result struct {
	RuleType        string
	Success         bool
	MinimalFields   []string       // fields beyond Type+Action that we sent on the successful Create
	APIDefaults     map[string]any // API-filled fields not in our request
	RequiredFields  []string       // fields the API said "can't be blank" / similar
	OptionalFields  []string       // fields not required and not API-defaulted (we never sent them)
	RangeErrors     map[string]string
	Attempts        int
	FinalErrorBody  string
	WirePayloadSent map[string]any
	WirePayloadGot  map[string]any

	// Update probe — single PUT /v3/hint/{id} attempt with the same minimal
	// body that succeeded on Create. Reveals Update-side schema drift, e.g.
	// fields the Update endpoint requires/rejects that Create accepts.
	UpdateAttempted   bool
	UpdateStatus      int
	UpdateOK          bool
	UpdateBody        string // raw response body, truncated
	UpdateFieldErrors map[string]string

	// FieldMutability — per-field: "mutable" (PUT accepted + Read confirms new
	// value), "immutable_silent" (PUT 200 but value reverts on Read),
	// "rejected" (PUT 4xx with field error), "untested" (no alternate value
	// known). Populated by probePerFieldMutability when API_PROBE_MUTABILITY=1.
	FieldMutability map[string]string
}

func main() {
	host := mustEnv("WALLARM_API_HOST")
	token := mustEnv("WALLARM_API_TOKEN")
	clientIDStr := mustEnv("WALLARM_API_CLIENT_ID")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		fatalf("WALLARM_API_CLIENT_ID is not numeric: %v", err)
	}

	verbose := os.Getenv("API_PROBE_VERBOSE") == "1"
	allow := splitFilter(os.Getenv("API_PROBE_RULES"))

	results := make([]Result, 0, len(probes))
	for _, p := range probes {
		label := p.Label
		if label == "" {
			label = p.RuleType
		}
		if len(allow) > 0 && !allow[p.RuleType] && !allow[label] {
			continue
		}
		fmt.Printf("\n=== probing %s ===\n", label)
		probeToken := token
		if p.TokenEnv != "" {
			if v := os.Getenv(p.TokenEnv); v != "" {
				probeToken = v
			} else {
				fmt.Printf("  (skipped: %s requires env %s)\n", p.RuleType, p.TokenEnv)
				continue
			}
		}
		r := discover(host, probeToken, clientID, p, verbose)
		r.RuleType = label // override with the disambiguated label for report output
		results = append(results, r)
		// Be polite to the API.
		time.Sleep(300 * time.Millisecond)
	}

	out := os.Getenv("API_PROBE_OUT")
	if out == "" {
		out = "api_probe_results.md"
	}
	if err := writeReport(results, out); err != nil {
		fatalf("failed to write report: %v", err)
	}
	fmt.Printf("\n→ wrote %s\n", out)
}

func discover(host, token string, clientID int, p Probe, verbose bool) Result {
	r := Result{RuleType: p.RuleType, RangeErrors: map[string]string{}, APIDefaults: map[string]any{}}

	// Unique action scope so we don't collide with anything else.
	hostHeader := fmt.Sprintf("api-probe-%s-%d.example.com", p.RuleType, time.Now().UnixNano())
	action := []map[string]any{
		{"type": "iequal", "point": []string{"header", "HOST"}, "value": strings.ToLower(hostHeader)},
	}

	body := cloneMap(p.Base)
	body["type"] = p.RuleType
	body["clientid"] = clientID
	body["action"] = action

	const maxAttempts = 20
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		r.Attempts = attempt
		resp, raw, err := postCreate(host, token, body, verbose)
		if err != nil {
			r.FinalErrorBody = err.Error()
			return r
		}

		if resp.Status == 200 {
			r.Success = true
			r.WirePayloadSent = body
			// Capture all fields the API echoed that weren't in our request.
			var respBody map[string]any
			_ = json.Unmarshal(resp.Body, &respBody)
			r.WirePayloadGot = respBody
			for k, v := range respBody {
				if isMetaField(k) {
					continue
				}
				if _, sent := body[k]; !sent {
					r.APIDefaults[k] = v
				}
			}
			r.MinimalFields = sortedKeysExcl(body, "type", "clientid", "action", "validated")
			// Update probe — single shot, no retry. Re-send the same body that
			// got us here on the Create endpoint, but as PUT /v3/hint/{id}.
			if id, ok := respBody["id"].(float64); ok {
				ruleID := int(id)
				probeUpdate(host, token, ruleID, body, &r, verbose)
				if os.Getenv("API_PROBE_MUTABILITY") == "1" {
					probePerFieldMutability(host, token, ruleID, body, &r, verbose)
				}
				_ = deleteHint(host, token, clientID, ruleID, verbose)
			}
			return r
		}

		// 4xx — parse field-level errors.
		errs := parseFieldErrors(resp.Body)
		if len(errs) == 0 {
			r.FinalErrorBody = string(raw)
			fmt.Printf("  [attempt %d] no field errors parsed; body: %s\n", attempt, truncate(string(raw), 300))
			return r
		}

		progress := false
		for field, msg := range errs {
			isRange, lo, hi := parseRange(msg)
			if isRange {
				val := pickInRange(lo, hi)
				body[field] = val
				r.RangeErrors[field] = msg
				progress = true
				continue
			}
			// Generic "can't be blank" / "is missing" / "is required".
			if v, ok := candidateValues[field]; ok && v != nil {
				if _, already := body[field]; !already {
					body[field] = v
					if !contains(r.RequiredFields, field) {
						r.RequiredFields = append(r.RequiredFields, field)
					}
					progress = true
				}
			} else {
				// Unknown field — log and try a string.
				if _, already := body[field]; !already {
					body[field] = "probe"
					if !contains(r.RequiredFields, field) {
						r.RequiredFields = append(r.RequiredFields, field)
					}
					progress = true
				}
			}
		}
		if !progress {
			r.FinalErrorBody = string(raw)
			fmt.Printf("  [attempt %d] stuck: %v\n", attempt, errs)
			return r
		}
		fmt.Printf("  [attempt %d] errors=%v → adding fields, retrying\n", attempt, fieldList(errs))
	}
	r.FinalErrorBody = "max attempts exceeded"
	return r
}

func postCreate(host, token string, body map[string]any, verbose bool) (*APIResponse, []byte, error) {
	bb, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	if verbose {
		fmt.Printf("  → POST %s\n", string(bb))
	}
	req, _ := http.NewRequest("POST", host+"/v1/objects/hint/create", bytes.NewReader(bb))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-WallarmAPI-Token", token)
	req.Header.Set("User-Agent", "wallarm-tf-api-probe/0.1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if verbose {
		fmt.Printf("  ← %d %s\n", resp.StatusCode, truncate(string(raw), 400))
	}
	var ar APIResponse
	_ = json.Unmarshal(raw, &ar)
	if ar.Status == 0 {
		ar.Status = resp.StatusCode
	}
	return &ar, raw, nil
}

// probeUpdate sends a single PUT /v3/hint/{id} with the same minimal body
// the rule was created with (less type/clientid/action — those aren't
// updatable via this route). Records status, parsed field errors, and the
// raw response body. Single attempt — never retries.
func probeUpdate(host, token string, ruleID int, createBody map[string]any, r *Result, verbose bool) {
	r.UpdateAttempted = true
	body := cloneMap(createBody)
	// PUT /v3/hint/{id} doesn't take type/clientid/action — strip them.
	delete(body, "type")
	delete(body, "clientid")
	delete(body, "action")
	delete(body, "validated")

	bb, _ := json.Marshal(body)
	url := fmt.Sprintf("%s/v3/hint/%d", host, ruleID)
	if verbose {
		fmt.Printf("  → PUT %s %s\n", url, string(bb))
	}
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(bb))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-WallarmAPI-Token", token)
	req.Header.Set("User-Agent", "wallarm-tf-api-probe/0.1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.UpdateBody = "transport error: " + err.Error()
		return
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if verbose {
		fmt.Printf("  ← %d %s\n", resp.StatusCode, truncate(string(raw), 400))
	}
	r.UpdateStatus = resp.StatusCode

	// v3 wraps the payload as {"status":N,"body":...}. Inspect both layers.
	var ar APIResponse
	_ = json.Unmarshal(raw, &ar)
	switch {
	case resp.StatusCode == 200 && ar.Status == 200:
		r.UpdateOK = true
	case resp.StatusCode == 200 && ar.Status >= 400:
		// Wrapped error — parse the inner body.
		r.UpdateFieldErrors = parseFieldErrors(ar.Body)
	default:
		r.UpdateFieldErrors = parseFieldErrors(raw)
	}
	r.UpdateBody = truncate(string(raw), 500)
}

// probePerFieldMutability tries to mutate each API-defaulted field individually
// via PUT /v3/hint/{id}, then reads the rule back via GET /v3/hint/{id} (when
// available) and compares. Records "mutable" / "immutable_silent" / "rejected"
// / "untested" per field. Best-effort; never fails the parent probe.
//
// The PUT body is the full createBody with just the target field flipped to
// `alt`. Sending a partial body (target field alone) caused the API to silently
// drop most fields — this matters because the Wallarm Update endpoint treats
// missing fields differently from explicit ones in some cases. Provider Update
// always sends a full body via resourcerule.Update; mirror that here.
func probePerFieldMutability(host, token string, ruleID int, createBody map[string]any, r *Result, verbose bool) {
	r.FieldMutability = map[string]string{}
	// Strip Create-only fields the PUT endpoint doesn't accept — same as
	// probeUpdate above.
	baseBody := cloneMap(createBody)
	delete(baseBody, "type")
	delete(baseBody, "clientid")
	delete(baseBody, "action")
	delete(baseBody, "validated")
	// Augment with non-nil API-echoed defaults — gives the PUT body more
	// context than the bare Create body. Note: this still doesn't fully
	// reproduce what the provider's resourcerule.Update sends (which
	// includes empty-string-as-pointer for unset commonResourceRuleFields).
	// Adding empty-string echoes for nil fields turned out to make the
	// Update endpoint reject bodies — the API is sensitive to extra-keys
	// shape in ways that aren't fully understood from outside. As a result,
	// "immutable_silent" / "rejected" results for common fields like `set`
	// or `attack_type` are likely probe-noise from this body-shape mismatch,
	// NOT a real claim that the field is server-immutable.
	for k, v := range r.APIDefaults {
		if v == nil {
			continue
		}
		if _, exists := baseBody[k]; !exists {
			baseBody[k] = v
		}
	}

	for field, current := range r.APIDefaults {
		alt, ok := alternateValue(field, current)
		if !ok {
			r.FieldMutability[field] = "untested"
			continue
		}
		// PUT the full body with just this one field overridden to `alt`.
		body := cloneMap(baseBody)
		body[field] = alt
		bb, _ := json.Marshal(body)
		url := fmt.Sprintf("%s/v3/hint/%d", host, ruleID)
		if verbose {
			fmt.Printf("    [mutability] PUT %s %s → %v\n", url, field, alt)
		}
		req, _ := http.NewRequest("PUT", url, bytes.NewReader(bb))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-WallarmAPI-Token", token)
		req.Header.Set("User-Agent", "wallarm-tf-api-probe/0.1")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			r.FieldMutability[field] = "transport_error"
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var ar APIResponse
		_ = json.Unmarshal(raw, &ar)
		if resp.StatusCode != 200 || ar.Status != 200 {
			r.FieldMutability[field] = "rejected"
			continue
		}
		// PUT body echoes the persisted state — read the field from ar.Body.
		var persisted map[string]any
		_ = json.Unmarshal(ar.Body, &persisted)
		if reflect.DeepEqual(persisted[field], alt) {
			r.FieldMutability[field] = "mutable"
		} else {
			r.FieldMutability[field] = "immutable_silent"
		}
	}
}

// enumAlternates lists known-good alternates for enum fields, used by
// alternateValue when the current value is in the list (pick a different one)
// or when the current value is nil/empty (use the first). Keeps mutability
// probes from sending invalid values to enum-validated fields.
var enumAlternates = map[string][]string{
	"mode":            {"block", "monitoring", "off", "default", "enabled", "disabled"},
	"attack_type":     {"sqli", "xss", "rce", "any", "ldapi", "redir"},
	"parser":          {"json_doc", "xml", "jwt", "gql"},
	"state":           {"enabled", "disabled"},
	"size_unit":       {"b", "kb", "mb", "gb"},
	"time_unit":       {"rps", "rpm"},
	"cred_stuff_type": {"custom", "default"},
	"file_type":       {"docs", "html", "images", "music", "video"},
}

// alternateValue picks a "different" value of the same kind for a per-field
// mutability test. When the current value is nil (API echoed `<nil>`), falls
// back to enumAlternates for known enums and to candidateValues for everything
// else — that way fields with no API default still get probed for mutability
// instead of returning "untested". Returns ok=false only when no safe
// alternative is known.
func alternateValue(field string, current any) (any, bool) {
	if current == nil {
		// Fallback for nil API echoes — try enum first, then candidateValues.
		if alts, ok := enumAlternates[field]; ok && len(alts) > 0 {
			return alts[0], true
		}
		if v, ok := candidateValues[field]; ok && v != nil {
			return v, true
		}
		return nil, false
	}
	switch v := current.(type) {
	case bool:
		return !v, true
	case float64: // JSON numbers
		alt := v + 1
		if alt == v {
			return nil, false
		}
		return alt, true
	case int:
		return v + 1, true
	case string:
		if v == "" {
			// Same fallback path as nil — empty string is not informative.
			if alts, ok := enumAlternates[field]; ok && len(alts) > 0 {
				return alts[0], true
			}
			if cv, ok := candidateValues[field].(string); ok && cv != "" {
				return cv, true
			}
			return "probe-mutated", true
		}
		// Enums: pick a different value from the known set.
		if alts, ok := enumAlternates[field]; ok {
			for _, a := range alts {
				if a != v {
					return a, true
				}
			}
			return nil, false
		}
		return v + "-x", true
	}
	return nil, false
}

func deleteHint(host, token string, clientID, ruleID int, verbose bool) error {
	body := map[string]any{
		"filter": map[string]any{"clientid": []int{clientID}, "id": []int{ruleID}},
	}
	bb, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", host+"/v1/objects/hint/delete", bytes.NewReader(bb))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-WallarmAPI-Token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if verbose {
		fmt.Printf("  ✗ deleted hint %d (status %d)\n", ruleID, resp.StatusCode)
	}
	return nil
}

// parseFieldErrors expects the v1 error shape:
//
//	{"<field>": {"error": "..."}, ...}
//
// Returns a flat map of field → message. Skips structurally weird responses.
func parseFieldErrors(body []byte) map[string]string {
	out := map[string]string{}
	var generic map[string]any
	if err := json.Unmarshal(body, &generic); err != nil {
		return out
	}
	for k, v := range generic {
		// Sometimes errors are nested: {"enumerated_parameters": {"name_regexps": {...}}}
		switch vv := v.(type) {
		case map[string]any:
			if msg, ok := vv["error"].(string); ok {
				out[k] = msg
				continue
			}
			// Recurse one level deep.
			for k2, v2 := range vv {
				if m2, ok := v2.(map[string]any); ok {
					if msg, ok := m2["error"].(string); ok {
						out[k+"."+k2] = msg
					}
				}
			}
		case string:
			out[k] = vv
		}
	}
	return out
}

func parseRange(msg string) (bool, int, int) {
	m := rangeRE.FindStringSubmatch(msg)
	if len(m) != 3 {
		return false, 0, 0
	}
	lo, _ := strconv.Atoi(m[1])
	hi, _ := strconv.Atoi(m[2])
	return true, lo, hi
}

func pickInRange(lo, hi int) int {
	if lo <= 0 && hi > 0 {
		// pick a small positive number, in range
		if hi >= 1 {
			return 1
		}
	}
	if lo > 0 {
		return lo
	}
	return (lo + hi) / 2
}

// emptyMark is the placeholder shown in the report for fields without a value.
const emptyMark = "—"

func writeReport(results []Result, outPath string) error {
	var sb strings.Builder
	sb.WriteString("# Wallarm API Rule-Type Probe Results\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("| Rule type | Create | Update (same body) | Min fields sent | API defaults filled | Range constraints | Notes |\n")
	sb.WriteString("|---|---|---|---|---|---|---|\n")
	for _, r := range results {
		status := "❌ failed"
		if r.Success {
			status = "✅ ok"
		}
		updateStatus := emptyMark
		if r.UpdateAttempted {
			switch {
			case r.UpdateOK:
				updateStatus = "✅ ok"
			case len(r.UpdateFieldErrors) > 0:
				keys := mapKeys(r.UpdateFieldErrors)
				sort.Strings(keys)
				parts := []string{}
				for _, k := range keys {
					parts = append(parts, fmt.Sprintf("`%s`", k))
				}
				updateStatus = fmt.Sprintf("❌ %d (%s)", r.UpdateStatus, strings.Join(parts, ", "))
			default:
				updateStatus = fmt.Sprintf("❌ %d", r.UpdateStatus)
			}
		}
		ranges := []string{}
		for f, msg := range r.RangeErrors {
			ranges = append(ranges, fmt.Sprintf("`%s`: %s", f, msg))
		}
		sort.Strings(ranges)
		notes := r.FinalErrorBody
		if len(notes) > 200 {
			notes = notes[:200] + "…"
		}
		notes = strings.ReplaceAll(notes, "|", "\\|")
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s | %s | %s |\n",
			r.RuleType,
			status,
			updateStatus,
			joinBackticked(r.MinimalFields),
			joinKVDefaults(r.APIDefaults),
			strings.Join(ranges, "<br>"),
			notes,
		))
	}
	sb.WriteString("\n## Per-rule detail\n")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("\n### `%s`\n\n", r.RuleType))
		sb.WriteString(fmt.Sprintf("- Success: %v\n", r.Success))
		sb.WriteString(fmt.Sprintf("- Attempts: %d\n", r.Attempts))
		if len(r.MinimalFields) > 0 {
			sb.WriteString(fmt.Sprintf("- Minimal Create body fields: %s\n", joinBackticked(r.MinimalFields)))
		}
		if len(r.APIDefaults) > 0 {
			sb.WriteString("- API-default values for omitted fields:\n")
			keys := mapKeys(r.APIDefaults)
			sort.Strings(keys)
			for _, k := range keys {
				sb.WriteString(fmt.Sprintf("  - `%s` = `%v`\n", k, r.APIDefaults[k]))
			}
		}
		if len(r.RangeErrors) > 0 {
			sb.WriteString("- Range constraints discovered:\n")
			keys := mapKeys(r.RangeErrors)
			sort.Strings(keys)
			for _, k := range keys {
				sb.WriteString(fmt.Sprintf("  - `%s`: %s\n", k, r.RangeErrors[k]))
			}
		}
		if r.FinalErrorBody != "" {
			sb.WriteString(fmt.Sprintf("- Last error body: `%s`\n", strings.ReplaceAll(r.FinalErrorBody, "`", "")))
		}
		if r.UpdateAttempted {
			sb.WriteString(fmt.Sprintf("- Update probe (PUT /v3/hint/{id} with same body): HTTP %d, ok=%v\n", r.UpdateStatus, r.UpdateOK))
			if len(r.UpdateFieldErrors) > 0 {
				keys := mapKeys(r.UpdateFieldErrors)
				sort.Strings(keys)
				sb.WriteString("  - Field errors:\n")
				for _, k := range keys {
					sb.WriteString(fmt.Sprintf("    - `%s`: %s\n", k, r.UpdateFieldErrors[k]))
				}
			}
			if r.UpdateBody != "" && !r.UpdateOK {
				sb.WriteString(fmt.Sprintf("  - Raw body: `%s`\n", strings.ReplaceAll(r.UpdateBody, "`", "")))
			}
		}
		if len(r.FieldMutability) > 0 {
			sb.WriteString("- Per-field mutability (PUT field=alt → GET, compare):\n")
			keys := mapKeys(r.FieldMutability)
			sort.Strings(keys)
			for _, k := range keys {
				sb.WriteString(fmt.Sprintf("    - `%s`: %s\n", k, r.FieldMutability[k]))
			}
		}
	}
	return os.WriteFile(outPath, []byte(sb.String()), 0o600)
}

// helpers

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		fatalf("env var %s is required", k)
	}
	return v
}

func fatalf(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "FATAL: "+f+"\n", a...)
	os.Exit(1)
}

func splitFilter(s string) map[string]bool {
	out := map[string]bool{}
	if s == "" {
		return out
	}
	for _, p := range strings.Split(s, ",") {
		out[strings.TrimSpace(p)] = true
	}
	return out
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func sortedKeysExcl(m map[string]any, excl ...string) []string {
	skip := map[string]bool{}
	for _, e := range excl {
		skip[e] = true
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		if skip[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func joinBackticked(xs []string) string {
	if len(xs) == 0 {
		return emptyMark
	}
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = "`" + x + "`"
	}
	return strings.Join(out, ", ")
}

func joinKVDefaults(m map[string]any) string {
	if len(m) == 0 {
		return emptyMark
	}
	keys := mapKeys(m)
	sort.Strings(keys)
	parts := []string{}
	for _, k := range keys {
		v := fmt.Sprintf("%v", m[k])
		if len(v) > 50 {
			v = v[:50] + "…"
		}
		parts = append(parts, fmt.Sprintf("`%s`=%s", k, v))
	}
	return strings.Join(parts, "<br>")
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func fieldList(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// isMetaField returns true for response-only fields (IDs, timestamps, system
// flags) that aren't user-settable and shouldn't count as "API defaults".
func isMetaField(name string) bool {
	switch name {
	case "id", "actionid", "clientid", "action", "create_time", "create_userid",
		"updated_at", "system", "validated", "regex_id", "group_uuid", "uuid",
		"type", "mitigation":
		return true
	}
	return false
}
