package resourcerule

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"
)

// ReverseMapResult contains the extracted path/domain/etc. from action conditions.
type ReverseMapResult struct {
	Path     string
	Domain   string
	Instance string
	Method   string
	Scheme   string
	Proto    string
	Query    []QueryParam
	Headers  []HeaderParam
}

// QueryParam represents a query parameter condition.
type QueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

// HeaderParam represents a header condition (not HOST).
type HeaderParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

// ReverseMapActions converts API action conditions back to the user-friendly
// path/domain/method/etc. format used in YAML rule configs.
//
// Mapping from API conditions to user-friendly fields:
//
//	point=["header","HOST"]              -> domain
//	point=["instance"]                   -> instance
//	point=["method"]                     -> method
//	point=["scheme"]                     -> scheme
//	point=["proto"]                      -> proto
//	point=["get","key"]                  -> query param
//	point=["header","X-Custom"]          -> custom header
//	point=["action_name"] present        -> last path segment name
//	point=["action_name"] MISSING        -> wildcard * (match any name)
//	point=["action_ext"] type="absent"   -> no extension
//	point=["action_ext"] type="equal"    -> specific extension
//	point=["action_ext"] MISSING         -> wildcard .* (match any extension)
//	point=["path",N] type="equal"        -> directory segment at index N
//	point=["path",N] type="absent"       -> path depth limiter
//	NO path limiter + has path segments  -> globstar ** (any depth)
//	point=["uri"]                        -> too-deep fallback (full URI)
//	NO path conditions at all            -> /**/*.* (match everything)
func ReverseMapActions(actions []wallarm.ActionDetails) ReverseMapResult {
	var result ReverseMapResult

	var actionName string
	var actionNameFound bool
	var actionExt string
	var actionExtFound bool
	var actionExtAbsent bool
	var pathSegments = make(map[int]string) // index -> value for "equal" path conditions
	var limiterIndex = -1
	var hasURI bool
	var uriValue string

	for _, a := range actions {
		pointKey := ActionPointKey(a)
		condType := a.Type
		value := ActionValueString(a)

		switch pointKey {
		case "header":
			headerName := ActionPointSecond(a)
			if strings.EqualFold(headerName, "HOST") {
				result.Domain = value
			} else {
				result.Headers = append(result.Headers, HeaderParam{
					Name:  headerName,
					Value: value,
					Type:  condType,
				})
			}

		case "instance":
			result.Instance = value

		case "method":
			result.Method = value

		case "scheme":
			result.Scheme = value

		case "proto":
			result.Proto = value

		case "get":
			paramName := ActionPointSecond(a)
			result.Query = append(result.Query, QueryParam{
				Key:   paramName,
				Value: value,
				Type:  condType,
			})

		case "action_name":
			actionName = value
			actionNameFound = true

		case "action_ext":
			actionExtFound = true
			if condType == "absent" {
				actionExtAbsent = true
			} else {
				actionExt = value
			}

		case "path":
			idx := ActionPointIndex(a)
			if idx < 0 {
				continue
			}
			if condType == "absent" {
				limiterIndex = idx
			} else {
				pathSegments[idx] = value
			}

		case "uri":
			hasURI = true
			uriValue = value
		}
	}

	// Build the path string.
	if hasURI {
		result.Path = uriValue
		return result
	}

	// No path-related conditions at all -> match everything: /**/*.*
	hasAnyPathInfo := actionNameFound || actionExtFound || len(pathSegments) > 0 || limiterIndex >= 0
	if !hasAnyPathInfo {
		result.Path = "/**/*.*"
		return result
	}

	// Root path: action_name="" + action_ext=absent + path[0]=absent
	if actionNameFound && actionName == "" && actionExtAbsent && limiterIndex == 0 {
		result.Path = "/"
		return result
	}

	result.Path = buildPathFromComponents(
		pathSegments, limiterIndex,
		actionName, actionNameFound,
		actionExt, actionExtFound, actionExtAbsent,
	)

	return result
}

// buildPathFromComponents reconstructs a URL path from its decomposed action conditions.
func buildPathFromComponents(
	segments map[int]string,
	limiterIndex int,
	actionName string, actionNameFound bool,
	actionExt string, actionExtFound bool, actionExtAbsent bool,
) string {
	// Determine how many directory segments exist.
	maxSegIdx := -1
	for idx := range segments {
		if idx > maxSegIdx {
			maxSegIdx = idx
		}
	}

	// The limiter tells us the expected path depth.
	// If limiter is at index N, there are N directory segments (0..N-1).
	// If no limiter and segments exist, globstar ** was used (any depth).
	var dirCount int
	if limiterIndex >= 0 {
		dirCount = limiterIndex
	} else if maxSegIdx >= 0 {
		dirCount = maxSegIdx + 1
	}

	// Build directory parts. Gaps in indices -> wildcard *.
	parts := make([]string, 0, dirCount+2)
	for i := 0; i < dirCount; i++ {
		if val, ok := segments[i]; ok {
			parts = append(parts, val)
		} else {
			parts = append(parts, "*")
		}
	}

	// No limiter + has path segments -> globstar ** (any depth).
	if limiterIndex < 0 && maxSegIdx >= 0 {
		parts = append(parts, "**")
	}

	// Build the final segment: action_name[.action_ext]
	var lastSegment string
	if actionNameFound {
		lastSegment = actionName
	} else {
		// action_name missing from conditions -> wildcard * (match any name)
		lastSegment = "*"
	}

	// Append extension.
	if actionExtFound {
		if !actionExtAbsent && actionExt != "" {
			// Specific extension
			lastSegment = lastSegment + "." + actionExt
		}
		// actionExtAbsent -> no extension (don't add dot)
	} else {
		// action_ext missing from conditions entirely -> wildcard .* (match any extension)
		lastSegment = lastSegment + ".*"
	}

	parts = append(parts, lastSegment)

	return "/" + strings.Join(parts, "/")
}

// ExpandPathToActions converts user-friendly path/domain/method/etc. fields
// into API action conditions ([]wallarm.ActionDetails).
//
// This is the forward mapping -- the inverse of ReverseMapActions.
// Used by CustomizeDiff to compute action conditions from scope fields.
func ExpandPathToActions(path, domain, instance, method, scheme, proto string, query []QueryParam, headers []HeaderParam) []wallarm.ActionDetails {
	var actions []wallarm.ActionDetails

	// Instance
	if instance != "" {
		actions = append(actions, wallarm.ActionDetails{
			Type:  "equal",
			Value: instance,
			Point: []interface{}{"instance"},
		})
	}

	// Domain (HOST header) -- skip when "*" (match any)
	if domain != "" && domain != "*" {
		actions = append(actions, wallarm.ActionDetails{
			Type:  "iequal",
			Value: domain,
			Point: []interface{}{"header", "HOST"},
		})
	}

	// Custom headers -- uppercase name to match API convention
	for _, h := range headers {
		t := h.Type
		if t == "" {
			t = "equal"
		}
		actions = append(actions, wallarm.ActionDetails{
			Type:  t,
			Value: h.Value,
			Point: []interface{}{"header", strings.ToUpper(h.Name)},
		})
	}

	// Parse path
	actions = append(actions, expandPath(path)...)

	// Method, scheme, proto
	if method != "" {
		actions = append(actions, wallarm.ActionDetails{
			Type:  "equal",
			Value: method,
			Point: []interface{}{"method"},
		})
	}
	if scheme != "" {
		actions = append(actions, wallarm.ActionDetails{
			Type:  "equal",
			Value: scheme,
			Point: []interface{}{"scheme"},
		})
	}
	if proto != "" {
		actions = append(actions, wallarm.ActionDetails{
			Type:  "equal",
			Value: proto,
			Point: []interface{}{"proto"},
		})
	}

	// Query parameters
	for _, q := range query {
		t := q.Type
		if t == "" {
			t = "equal"
		}
		actions = append(actions, wallarm.ActionDetails{
			Type:  t,
			Value: q.Value,
			Point: []interface{}{"get", q.Key},
		})
	}

	return actions
}

// expandPath parses a path string into action conditions.
func expandPath(path string) []wallarm.ActionDetails {
	if path == "" {
		return nil
	}

	// Global wildcard: /**/*.* -> no conditions (match everything)
	if path == "/**/*.*" {
		return nil
	}

	// Root path "/"
	if path == "/" {
		return []wallarm.ActionDetails{
			{Type: "equal", Value: "", Point: []interface{}{"action_name"}},
			{Type: "absent", Value: nil, Point: []interface{}{"action_ext"}},
			{Type: "absent", Value: nil, Point: []interface{}{"path", float64(0)}},
		}
	}

	rawParts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(rawParts) == 0 {
		return nil
	}

	// Split into directory segments and final segment (action component).
	var dirSegments []string
	lastSegment := rawParts[len(rawParts)-1]
	if len(rawParts) > 1 {
		dirSegments = rawParts[:len(rawParts)-1]
	}

	// Check for ** globstar (last directory segment).
	hasGlobstar := false
	if len(dirSegments) > 0 && dirSegments[len(dirSegments)-1] == "**" {
		hasGlobstar = true
		dirSegments = dirSegments[:len(dirSegments)-1]
	}

	// Parse last segment into action_name and action_ext.
	actionName, actionExt, hasDot := parseLastSegment(lastSegment)
	actionNameIsWildcard := actionName == "*"
	actionExtIsWildcard := hasDot && actionExt == "*"

	var actions []wallarm.ActionDetails

	// action_name -- skip when wildcard *
	if !actionNameIsWildcard {
		actions = append(actions, wallarm.ActionDetails{
			Type:  "equal",
			Value: actionName,
			Point: []interface{}{"action_name"},
		})
	}

	// action_ext
	if hasDot {
		if !actionExtIsWildcard {
			// Specific extension
			actions = append(actions, wallarm.ActionDetails{
				Type:  "equal",
				Value: actionExt,
				Point: []interface{}{"action_ext"},
			})
		}
		// Wildcard extension *.* -> skip action_ext condition
	} else {
		// No dot -> extension is absent
		actions = append(actions, wallarm.ActionDetails{
			Type:  "absent",
			Value: nil,
			Point: []interface{}{"action_ext"},
		})
	}

	// Path segments -- skip * wildcards (match any value at that position)
	for i, seg := range dirSegments {
		if seg != "*" {
			actions = append(actions, wallarm.ActionDetails{
				Type:  "equal",
				Value: seg,
				Point: []interface{}{"path", float64(i)},
			})
		}
	}

	// Path limiter -- suppressed when ** is present
	if !hasGlobstar {
		limiterIdx := len(dirSegments)
		actions = append(actions, wallarm.ActionDetails{
			Type:  "absent",
			Value: nil,
			Point: []interface{}{"path", float64(limiterIdx)},
		})
	}

	return actions
}

// parseLastSegment splits "name.ext" into name and ext parts.
func parseLastSegment(seg string) (name, ext string, hasDot bool) {
	dotIdx := strings.LastIndex(seg, ".")
	if dotIdx < 0 {
		return seg, "", false
	}
	return seg[:dotIdx], seg[dotIdx+1:], true
}

// --- Helper functions ---

// ActionPointKey extracts the first element of ActionDetails.Point as a string.
func ActionPointKey(a wallarm.ActionDetails) string {
	if len(a.Point) == 0 {
		return ""
	}
	if s, ok := a.Point[0].(string); ok {
		return s
	}
	return ""
}

// ActionPointSecond extracts the second element of ActionDetails.Point as a string.
func ActionPointSecond(a wallarm.ActionDetails) string {
	if len(a.Point) < 2 {
		return ""
	}
	if s, ok := a.Point[1].(string); ok {
		return s
	}
	return ""
}

// ActionPointIndex extracts the second element of ActionDetails.Point as an int.
func ActionPointIndex(a wallarm.ActionDetails) int {
	if len(a.Point) < 2 {
		return -1
	}
	switch v := a.Point[1].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return -1
		}
		return n
	}
	return -1
}

// ActionValueString extracts the Value field as a string.
func ActionValueString(a wallarm.ActionDetails) string {
	if a.Value == nil {
		return ""
	}
	switch v := a.Value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// --- Export helpers ---

// APITypeToTerraformResource maps Wallarm API rule types to Terraform resource names.
var APITypeToTerraformResource = map[string]string{
	"binary_data":            "wallarm_rule_binary_data",
	"bola":                   "wallarm_rule_bola",
	"bola_counter":           "wallarm_rule_bola_counter",
	"brute":                  "wallarm_rule_brute",
	"brute_counter":          "wallarm_rule_bruteforce_counter",
	"credentials_point":      "wallarm_rule_credential_stuffing_point",
	"credentials_regex":      "wallarm_rule_credential_stuffing_regex",
	"dirbust_counter":        "wallarm_rule_dirbust_counter",
	"disable_attack_type":    "wallarm_rule_disable_attack_type",
	"disable_regex":          "wallarm_rule_ignore_regex",
	"disable_stamp":          "wallarm_rule_disable_stamp",
	"enum":                   "wallarm_rule_enum",
	"experimental_regex":     "wallarm_rule_regex",
	"file_upload_size_limit": "wallarm_rule_file_upload_size_limit",
	"forced_browsing":        "wallarm_rule_forced_browsing",
	"graphql_detection":      "wallarm_rule_graphql_detection",
	"overlimit_res_settings": "wallarm_rule_overlimit_res_settings",
	"parser_state":           "wallarm_rule_parser_state",
	"rate_limit":             "wallarm_rule_rate_limit",
	"rate_limit_enum":        "wallarm_rule_rate_limit_enum",
	"regex":                  "wallarm_rule_regex",
	"sensitive_data":         "wallarm_rule_masking",
	"set_response_header":    "wallarm_rule_set_response_header",
	"uploads":                "wallarm_rule_uploads",
	"vpatch":                 "wallarm_rule_vpatch",
	"wallarm_mode":           "wallarm_rule_mode",
}

// FourPartIDTypes are rule types whose import ID requires a 4th segment (the API type).
var FourPartIDTypes = map[string]bool{
	"regex":              true,
	"experimental_regex": true,
	"wallarm_mode":       true,
}

// RuleExportEntry represents a single rule with all its details in the export format.
type RuleExportEntry struct {
	RuleID               int                           `json:"rule_id"`
	ActionID             int                           `json:"action_id"`
	ClientID             int                           `json:"client_id"`
	APIType              string                        `json:"api_type"`
	TerraformResource    string                        `json:"terraform_resource"`
	ImportID             string                        `json:"import_id"`
	Path                 string                        `json:"path"`
	Domain               string                        `json:"domain"`
	Instance             string                        `json:"instance"`
	Method               string                        `json:"method"`
	Scheme               string                        `json:"scheme"`
	Proto                string                        `json:"proto"`
	Query                []QueryParam                  `json:"query,omitempty"`
	Headers              []HeaderParam                 `json:"headers,omitempty"`
	Action               []wallarm.ActionDetails       `json:"action"`
	Point                []interface{}                 `json:"point,omitempty"`
	VariativityDisabled  bool                          `json:"variativity_disabled"`
	Comment              string                        `json:"comment"`
	AttackType           string                        `json:"attack_type,omitempty"`
	Stamp                int                           `json:"stamp,omitempty"`
	Mode                 string                        `json:"mode,omitempty"`
	Regex                string                        `json:"regex,omitempty"`
	RegexID              int                           `json:"regex_id,omitempty"`
	Experimental         bool                          `json:"experimental,omitempty"`
	Parser               string                        `json:"parser,omitempty"`
	State                string                        `json:"state,omitempty"`
	FileType             string                        `json:"file_type,omitempty"`
	Delay                int                           `json:"delay,omitempty"`
	Burst                int                           `json:"burst,omitempty"`
	Rate                 int                           `json:"rate,omitempty"`
	RspStatus            int                           `json:"rsp_status,omitempty"`
	TimeUnit             string                        `json:"time_unit,omitempty"`
	OverlimitTime        int                           `json:"overlimit_time,omitempty"`
	Size                 int                           `json:"size,omitempty"`
	SizeUnit             string                        `json:"size_unit,omitempty"`
	MaxDepth             int                           `json:"max_depth,omitempty"`
	MaxValueSizeKb       int                           `json:"max_value_size_kb,omitempty"`
	MaxDocSizeKb         int                           `json:"max_doc_size_kb,omitempty"`
	MaxAliasesSizeKb     int                           `json:"max_aliases_size_kb,omitempty"`
	MaxDocPerBatch       int                           `json:"max_doc_per_batch,omitempty"`
	Introspection        bool                          `json:"introspection,omitempty"`
	DebugEnabled         bool                          `json:"debug_enabled,omitempty"`
	HeaderName           string                        `json:"header_name,omitempty"`
	HeaderValues         []string                      `json:"header_values,omitempty"`
	LoginPoint           []interface{}                 `json:"login_point,omitempty"`
	LoginRegex           string                        `json:"login_regex,omitempty"`
	CaseSensitive        bool                          `json:"case_sensitive,omitempty"`
	CredStuffType        string                        `json:"cred_stuff_type,omitempty"`
	Threshold            *wallarm.Threshold            `json:"threshold,omitempty"`
	Reaction             *wallarm.Reaction             `json:"reaction,omitempty"`
	EnumeratedParameters *wallarm.EnumeratedParameters `json:"enumerated_parameters,omitempty"`
}

// ExportRules converts a list of ActionBody entries to fully-detailed RuleExportEntry items,
// including reverse-mapped path/domain/etc. from action conditions.
func ExportRules(rules []wallarm.ActionBody, clientID int) []RuleExportEntry {
	result := make([]RuleExportEntry, 0, len(rules))

	for _, rule := range rules {
		tfResource, known := APITypeToTerraformResource[rule.Type]
		if !known {
			continue
		}

		var importID string
		if FourPartIDTypes[rule.Type] {
			importID = fmt.Sprintf("%d/%d/%d/%s", clientID, rule.ActionID, rule.ID, rule.Type)
		} else {
			importID = fmt.Sprintf("%d/%d/%d", clientID, rule.ActionID, rule.ID)
		}

		revMap := ReverseMapActions(rule.Action)

		var regexID int
		if rule.RegexID != nil {
			switch v := rule.RegexID.(type) {
			case float64:
				regexID = int(v)
			case int:
				regexID = v
			}
		}

		var headerValues []string
		for _, v := range rule.Values {
			if s, ok := v.(string); ok {
				headerValues = append(headerValues, s)
			}
		}

		entry := RuleExportEntry{
			RuleID: rule.ID, ActionID: rule.ActionID, ClientID: clientID,
			APIType: rule.Type, TerraformResource: tfResource, ImportID: importID,
			VariativityDisabled: rule.VariativityDisabled,
			Path:                revMap.Path, Domain: revMap.Domain, Instance: revMap.Instance,
			Method: revMap.Method, Scheme: revMap.Scheme, Proto: revMap.Proto,
			Query: revMap.Query, Headers: revMap.Headers,
			Action: rule.Action, Point: rule.Point, Comment: rule.Comment,
			AttackType: rule.AttackType, Stamp: rule.Stamp, Mode: rule.Mode,
			Regex: rule.Regex, RegexID: regexID,
			Parser: rule.Parser, State: rule.State, FileType: rule.FileType,
			Delay: rule.Delay, Burst: rule.Burst, Rate: rule.Rate,
			RspStatus: rule.RspStatus, TimeUnit: rule.TimeUnit,
			OverlimitTime: rule.OverlimitTime, Size: rule.Size, SizeUnit: rule.SizeUnit,
			MaxDepth: rule.MaxDepth, MaxValueSizeKb: rule.MaxValueSizeKb,
			MaxDocSizeKb: rule.MaxDocSizeKb, MaxAliasesSizeKb: rule.MaxAliasesSizeKb,
			MaxDocPerBatch: rule.MaxDocPerBatch,
			HeaderName:     rule.Name, HeaderValues: headerValues,
			LoginPoint: rule.LoginPoint, LoginRegex: rule.LoginRegex,
			CredStuffType: rule.CredStuffType,
			Threshold:     rule.Threshold, Reaction: rule.Reaction,
			EnumeratedParameters: rule.EnumeratedParameters,
		}

		if rule.Introspection != nil {
			entry.Introspection = *rule.Introspection
		}
		if rule.DebugEnabled != nil {
			entry.DebugEnabled = *rule.DebugEnabled
		}
		if rule.CaseSensitive != nil {
			entry.CaseSensitive = *rule.CaseSensitive
		}
		entry.Experimental = rule.Type == "experimental_regex"

		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].RuleID < result[j].RuleID
	})

	return result
}
