package resourcerule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/mapper/apitotf"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/mapper/tftoapi"
	"github.com/wallarm/wallarm-go"
)

// schemaHasKey checks whether the resource schema includes the given attribute
// by examining the cty type of the raw state. Returns true if undetermined.
func schemaHasKey(d *schema.ResourceData, key string) (exists bool) {
	exists = true // default to true if we can't determine
	defer func() { recover() }()
	return d.GetRawState().Type().HasAttribute(key)
}

// setIfExists calls d.Set for the given key only if the key is present in the
// resource schema. This is needed because ResourceRuleWallarmRead is shared
// across many resources, each of which defines only a subset of the fields.
// In SDK v2 d.Set() logs an [ERROR] and panics (in test mode) for keys not in
// the schema; checking beforehand avoids both.
func setIfExists(d *schema.ResourceData, key string, value interface{}) {
	if !schemaHasKey(d, key) {
		return
	}
	if err := d.Set(key, value); err != nil {
		log.Printf("[ERROR] error setting %s: %s", key, err)
	}
}

// String hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func HashString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// nolint
func ResourceRuleWallarmRead(d *schema.ResourceData, clientID int, cli wallarm.API, opts ...common.ReadOption) error {
	var ruleID = d.Get("rule_id").(int)
	hint := &wallarm.HintRead{
		Limit:     1000,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ID:       []int{ruleID},
		},
	}
	actionHints, err := cli.HintRead(hint)
	if err != nil {
		return err
	}

	var updatedRule *wallarm.ActionBody
	for _, rule := range *actionHints.Body {
		if ruleID == rule.ID {
			updatedRule = &rule
			break
		}
	}

	if updatedRule == nil {
		d.SetId("")
		return nil
	}

	// Fields from commonResourceRuleFields — always present in every resource
	d.Set("rule_id", updatedRule.ID)
	d.Set("client_id", clientID)
	d.Set("active", updatedRule.Active)
	d.Set("title", updatedRule.Title)
	d.Set("mitigation", updatedRule.Mitigation)
	d.Set("set", updatedRule.Set)
	d.Set("variativity_disabled", updatedRule.VariativityDisabled)
	d.Set("comment", updatedRule.Comment)

	// Resource-specific fields — use setIfExists because each resource
	// only defines a subset of these in its schema. In SDK v2, d.Set
	// panics for keys not in the schema.
	setIfExists(d, "point", WrapPointElements(updatedRule.Point))
	setIfExists(d, "threshold", apitotf.Threshold(updatedRule.Threshold))
	setIfExists(d, "reaction", apitotf.Reaction(updatedRule.Reaction))
	setIfExists(d, "mode", updatedRule.Mode)
	setIfExists(d, "enumerated_parameters", apitotf.EnumeratedParameters(updatedRule.EnumeratedParameters))
	setIfExists(d, "advanced_conditions", apitotf.AdvancedConditions(updatedRule.AdvancedConditions))
	setIfExists(d, "arbitrary_conditions", apitotf.ArbitraryConditions(updatedRule.ArbitraryConditions))
	setIfExists(d, "counter", updatedRule.Counter)
	setIfExists(d, "size", updatedRule.Size)
	setIfExists(d, "size_unit", updatedRule.SizeUnit)
	setIfExists(d, "debug_enabled", updatedRule.DebugEnabled)
	setIfExists(d, "introspection", updatedRule.Introspection)
	setIfExists(d, "max_depth", updatedRule.MaxDepth)
	setIfExists(d, "max_value_size_kb", updatedRule.MaxValueSizeKb)
	setIfExists(d, "max_doc_size_kb", updatedRule.MaxDocSizeKb)
	setIfExists(d, "max_alias_size_kb", updatedRule.MaxAliasesSizeKb)
	setIfExists(d, "max_doc_per_batch", updatedRule.MaxDocPerBatch)
	setIfExists(d, "attack_type", updatedRule.AttackType)
	setIfExists(d, "stamp", updatedRule.Stamp)
	setIfExists(d, "file_type", updatedRule.FileType)
	setIfExists(d, "name", updatedRule.Name)
	setIfExists(d, "burst", updatedRule.Burst)
	setIfExists(d, "delay", updatedRule.Delay)
	setIfExists(d, "rate", updatedRule.Rate)
	setIfExists(d, "rsp_status", updatedRule.RspStatus)
	setIfExists(d, "time_unit", updatedRule.TimeUnit)
	setIfExists(d, "parser", updatedRule.Parser)
	setIfExists(d, "state", updatedRule.State)
	setIfExists(d, "overlimit_time", updatedRule.OverlimitTime)
	setIfExists(d, "values", updatedRule.Values)
	setIfExists(d, "regex", updatedRule.Regex)
	setIfExists(d, "regex_id", updatedRule.RegexID)

	actionsSet := schema.Set{F: HashResponseActionDetails}
	for _, a := range updatedRule.Action {
		acts, err := ActionDetailsToMap(a)
		if err != nil {
			return fmt.Errorf("failed to map action details: %w", err)
		}
		actionsSet.Add(acts)
	}
	setIfExists(d, "action", &actionsSet)

	return nil
}

// nolint
func ResourceRuleWallarmCreate(
	ctx context.Context,
	d *schema.ResourceData,
	cli wallarm.API,
	clientID int,
	ruleType, attackType string,
	readMethod func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics,
	m interface{},
) diag.Diagnostics {
	actionsFromState := d.Get("action").(*schema.Set)
	action, err := ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return diag.FromErr(errors.WithMessage(err, "on ExpandSetToActionDetailsList"))
	}

	enumeratedParametersFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "enumerated_parameters")
	enumeratedParameters, err := tftoapi.EnumeratedParameters(enumeratedParametersFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	reactionFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "reaction")
	reaction, err := tftoapi.Reaction(reactionFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	thresholdFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "threshold")
	threshold, err := tftoapi.Threshold(thresholdFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	advancedConditionsFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "advanced_conditions")
	advancedConditions, err := tftoapi.AdvancedConditions(advancedConditionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	arbitraryConditionsFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "arbitrary_conditions")
	arbitraryConditions, err := tftoapi.ArbitraryConditionsReq(arbitraryConditionsFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	pointFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "point")
	points, err := ExpandPointsToTwoDimensionalArray(pointFromState)
	if err != nil {
		return diag.FromErr(err)
	}

	wm := &wallarm.ActionCreate{
		Type:                 ruleType,
		Clientid:             clientID,
		Action:               &action,
		Point:                points,
		Validated:            false,
		Comment:              GetValueWithTypeCastingOrDefault[string](d, "comment"),
		VariativityDisabled:  true,
		Set:                  GetValueWithTypeCastingOrDefault[string](d, "set"),
		Active:               GetValueWithTypeCastingOrDefault[bool](d, "active"),
		Title:                GetValueWithTypeCastingOrDefault[string](d, "title"),
		AttackType:           attackType,
		Reaction:             reaction,
		Threshold:            threshold,
		EnumeratedParameters: enumeratedParameters,
		AdvancedConditions:   advancedConditions,
		ArbitraryConditions:  arbitraryConditions,
		Mode:                 GetValueWithTypeCastingOrDefault[string](d, "mode"),
		MaxDepth:             GetValueWithTypeCastingOrDefault[int](d, "max_depth"),
		MaxValueSizeKb:       GetValueWithTypeCastingOrDefault[int](d, "max_value_size_kb"),
		MaxDocSizeKb:         GetValueWithTypeCastingOrDefault[int](d, "max_doc_size_kb"),
		MaxAliasesSizeKb:     GetValueWithTypeCastingOrDefault[int](d, "max_alias_size_kb"),
		MaxDocPerBatch:       GetValueWithTypeCastingOrDefault[int](d, "max_doc_per_batch"),
		Introspection:        GetPointerWithTypeCastingOrDefault[bool](d, "introspection"),
		DebugEnabled:         GetPointerWithTypeCastingOrDefault[bool](d, "debug_enabled"),
		Size:                 GetValueWithTypeCastingOrDefault[int](d, "size"),
		SizeUnit:             GetValueWithTypeCastingOrDefault[string](d, "size_unit"),
	}

	actionResp, err := cli.HintCreate(wm)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return readMethod(ctx, d, m)
}

// nolint
func ExpandSetToActionDetailsList(action *schema.Set) ([]wallarm.ActionDetails, error) {
	var as []wallarm.ActionDetails
	for _, actionMap := range action.List() {
		// Derive maps consecutively from a Set List
		actionMap := actionMap.(map[string]interface{})

		// Make keys of map sorted to
		// then iterate over a map in order
		keys := make([]string, 0, len(actionMap))
		for k := range actionMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		a := wallarm.ActionDetails{}
		for _, k := range keys {
			switch k {
			case "point":
				point := actionMap[k].(map[string]interface{})
				for pointKey, pointValue := range point {
					switch pointKey {
					case common.Path:
						// Marshalling of the number leads to float64 even though it was int initially
						// Therefore, we parse string into float64 to compare structs properly afterwards
						pointValue, err := strconv.ParseFloat(pointValue.(string), 64)
						if err != nil {
							return nil, err
						}
						a.Point = []interface{}{pointKey, pointValue}
					case "action_name", "action_ext", "method",
						"proto", "scheme", "uri":
						a.Point = []interface{}{pointKey}
						// This is required by the API when case is insensitive
						switch {
						case actionMap["type"] == common.Iequal:
							a.Value = strings.ToLower(pointValue.(string))
						case actionMap["type"] == "absent":
							a.Value = nil
						default:
							a.Value = pointValue.(string)
						}
					case "instance":
						a.Point = []interface{}{pointKey}
						a.Value = pointValue.(string)
						a.Type = "equal"
					case common.Header:
						// This is required by the API when a header field is specified
						a.Point = []interface{}{pointKey, strings.ToUpper(pointValue.(string))}
					case "query":
						// This is required by the API when case is insensitive
						if actionMap["type"] == common.Iequal {
							a.Point = []interface{}{"get", strings.ToLower(pointValue.(string))}
						} else {
							a.Point = []interface{}{"get", pointValue.(string)}
						}
					default:
						// This is required by the API when case is insensitive
						if actionMap["type"] == "iequal" {
							a.Point = []interface{}{pointKey, strings.ToLower(pointValue.(string))}
						} else {
							a.Point = []interface{}{pointKey, pointValue.(string)}
						}
					}
				}

			case "type":
				// Fill out only when it is presented
				// Then default values will be omitted in the JSON request body
				// Otherwise, the API returns 4xx back due to the incorrect schema
				if actionMap[k].(string) != "" {
					a.Type = actionMap[k].(string)
				}
			case "value":
				if actionMap[k].(string) != "" {
					if actionMap["type"] == "iequal" {
						a.Value = strings.ToLower(actionMap[k].(string))
					} else {
						a.Value = actionMap[k].(string)
					}
				}
			}
		}

		// Check if there is anything to append, ensure it's not a default branch
		if a.Type != "" {
			as = append(as, a)
		}
	}
	// Check if this is for a default branch
	if len(as) == 0 {
		as = []wallarm.ActionDetails{}
	}
	return as, nil
}

func GetValueWithTypeCastingOrDefault[T any](d *schema.ResourceData, name string) T {
	var defaultValue T
	resourceValue := d.Get(name)
	if resourceValue == nil {
		return defaultValue
	}
	v, ok := resourceValue.(T)
	if !ok {
		return defaultValue
	}
	return v
}

func GetPointerWithTypeCastingOrDefault[T any](d *schema.ResourceData, name string) *T {
	resourceValue := d.Get(name)
	if resourceValue == nil {
		return nil
	}
	v, ok := d.Get(name).(T)
	if !ok {
		return nil
	}
	return &v
}

// WrapPointElements converts a flat API point array into a 2D string slice
// for the Terraform point schema. 2-part elements (hash, header, get, form_urlencoded, etc.)
// consume the next element as their value; 1-part elements (post, json_doc, uri, etc.) stand alone.
func WrapPointElements(input []interface{}) [][]string {
	var result [][]string // This will store the final result as a 2D slice of strings
	i := 0

	for i < len(input) {
		switch input[i] {
		// Paired point types — consume the next element as key/index.
		// Keep in sync with TYPES_INFO in proton/types.rb (all except simple:true).
		case
			// Core
			"hash", "array", "json", "json_obj", "json_array",
			// HTTP
			"header", "cookie", "get", "path", "multipart",
			"form_urlencoded", "content_disp", "response_header",
			// XML
			"xml_pi", "xml_dtd_entity", "xml_tag_array", "xml_tag",
			"xml_attr", "xml_comment",
			// JWT / Protobuf / gRPC
			"jwt", "grpc", "protobuf",
			// ViewState
			"viewstate_array", "viewstate_pair", "viewstate_triplet",
			"viewstate_dict", "viewstate_sparse_array",
			// GraphQL
			"gql_query", "gql_mutation", "gql_subscription", "gql_fragment",
			"gql_dir", "gql_spread", "gql_type", "gql_var":
			// Check if there is a next element to include
			if i+1 < len(input) {
				// Convert both elements to strings and wrap them in a slice of strings
				result = append(result, []string{
					fmt.Sprintf("%v", input[i]),
					fmt.Sprintf("%v", input[i+1]),
				})
				i++ // Skip the next element as it's already included
			} else {
				// If no next element, still wrap the special case string alone
				result = append(result, []string{fmt.Sprintf("%v", input[i])})
			}
		default:
			// For regular elements, convert to string and wrap it in a slice of strings
			result = append(result, []string{fmt.Sprintf("%v", input[i])})
		}
		i++ // Move to the next element
	}

	return result
}

// HashResponseActionDetails is the hash function for the action TypeSet.
// It transforms API point arrays into point maps as a side effect (e.g.,
// ["header","HOST"] → {header: "HOST"}, ["get","key"] → {query: "key"}).
func HashResponseActionDetails(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	var p []interface{}
	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))
	if val, ok := m["point"]; ok {
		p = val.([]interface{})
		switch p[0].(string) {
		case "action_name":
			pointMap := make(map[string]string)
			pointMap["action_name"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "action_ext":
			pointMap := make(map[string]string)
			pointMap["action_ext"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "scheme":
			pointMap := make(map[string]string)
			pointMap["scheme"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "uri":
			pointMap := make(map[string]string)
			pointMap["uri"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "proto":
			pointMap := make(map[string]string)
			pointMap["proto"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case "method":
			pointMap := make(map[string]string)
			pointMap["method"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
		case common.Path:
			pointMap := make(map[string]string)
			pointMap[common.Path] = fmt.Sprintf("%d", int(p[1].(float64)))
			m["point"] = pointMap
		case "instance":
			pointMap := make(map[string]string)
			pointMap["instance"] = m["value"].(string)
			m["point"] = pointMap
			m["value"] = ""
			m["type"] = ""
		case "header":
			pointMap := make(map[string]string)
			pointMap["header"] = p[1].(string)
			m["point"] = pointMap
		case "get":
			pointMap := make(map[string]string)
			pointMap["query"] = p[1].(string)
			m["point"] = pointMap
		}

		buf.WriteString(fmt.Sprintf("%v-", m["point"]))
	}
	return HashString(buf.String())
}

// ActionDetailsToMap converts an API ActionDetails struct to a Terraform-compatible map
// via JSON marshal/unmarshal. Ensures "value" key is always present.
func ActionDetailsToMap(actionDetails wallarm.ActionDetails) (map[string]interface{}, error) {
	jsonActions, err := json.Marshal(actionDetails)
	if err != nil {
		return nil, err
	}
	var mapActions map[string]interface{}
	if err = json.Unmarshal(jsonActions, &mapActions); err != nil {
		return nil, err
	}
	if v, ok := mapActions["value"]; !ok || v == nil {
		mapActions["value"] = ""
	}
	return mapActions, nil
}

// ExpandPointsToTwoDimensionalArray converts the Terraform point schema (list of lists of strings)
// to the API TwoDimensionalSlice format. Numeric-value point types (path, array, etc.) are
// converted from string to float64.
func ExpandPointsToTwoDimensionalArray(ps []interface{}) (wallarm.TwoDimensionalSlice, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	points := make(wallarm.TwoDimensionalSlice, len(ps))
	for i, point := range ps {
		pointSlice := point.([]interface{})
		switch pointSlice[0] {
		case "path", "array", "grpc", "json_array", "xml_comment",
			"xml_dtd_entity", "xml_pi", "xml_tag_array":
			// Align to the []string{} schema, float is used since marshalling considers numbers as float64
			if len(pointSlice) > 1 {
				number, err := strconv.ParseFloat(pointSlice[1].(string), 64)
				if err != nil {
					return nil, err
				}
				pointSlice[1] = number
				points[i] = pointSlice
			}
		default:
			points[i] = pointSlice
		}
	}
	return points, nil
}
