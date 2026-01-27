package resourcerule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/mapper/apitotf"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/mapper/tftoapi"
	"github.com/wallarm/wallarm-go"
)

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

	d.Set("point", wrapPointElements(updatedRule.Point))
	d.Set("rule_id", updatedRule.ID)
	d.Set("client_id", clientID)
	d.Set("active", updatedRule.Active)
	d.Set("title", updatedRule.Title)
	d.Set("mitigation", updatedRule.Mitigation)
	d.Set("set", updatedRule.Set)
	d.Set("threshold", apitotf.Threshold(updatedRule.Threshold))
	d.Set("reaction", apitotf.Reaction(updatedRule.Reaction))
	d.Set("mode", updatedRule.Mode)
	d.Set("enumerated_parameters", apitotf.EnumeratedParameters(updatedRule.EnumeratedParameters))
	d.Set("advanced_conditions", apitotf.AdvancedConditions(updatedRule.AdvancedConditions))
	d.Set("arbitrary_conditions", apitotf.ArbitraryConditions(updatedRule.ArbitraryConditions))
	d.Set("counter", updatedRule.Counter)
	d.Set("size", updatedRule.Size)
	d.Set("size_unit", updatedRule.SizeUnit)
	d.Set("debug_enabled", updatedRule.DebugEnabled)
	d.Set("introspection", updatedRule.Introspection)
	d.Set("max_depth", updatedRule.MaxDepth)
	d.Set("max_value_size_kb", updatedRule.MaxValueSizeKb)
	d.Set("max_doc_size_kb", updatedRule.MaxDocSizeKb)
	d.Set("max_alias_size_kb", updatedRule.MaxAliasesSizeKb)
	d.Set("max_doc_per_batch", updatedRule.MaxDocPerBatch)
	d.Set("attack_type", updatedRule.AttackType)
	d.Set("file_type", updatedRule.FileType)
	d.Set("name", updatedRule.Name)
	d.Set("burst", updatedRule.Burst)
	d.Set("delay", updatedRule.Delay)
	d.Set("rate", updatedRule.Rate)
	d.Set("rsp_status", updatedRule.RspStatus)
	d.Set("time_unit", updatedRule.TimeUnit)
	d.Set("parser", updatedRule.Parser)
	d.Set("state", updatedRule.State)
	d.Set("overlimit_time", updatedRule.OverlimitTime)
	tflog.Debug(context.Background(), "DEBUGGG values", map[string]interface{}{
		"values": updatedRule.Values,
	})
	tflog.Debug(context.Background(), "DEBUGGG values prepared", map[string]interface{}{
		"values": apitotf.SliceAnyToSliceString(updatedRule.Values),
	})
	d.Set("values", apitotf.SliceAnyToSliceString(updatedRule.Values))
	d.Set("regex", updatedRule.Regex)
	d.Set("regex_id", updatedRule.RegexID)

	actionsSet := schema.Set{F: hashResponseActionDetails}
	for _, a := range updatedRule.Action {
		acts, err := actionDetailsToMap(a)
		if err != nil {
		} else {
			actionsSet.Add(acts)
		}
	}
	d.Set("action", &actionsSet)

	return nil
}

// nolint
func ResourceRuleWallarmCreate(
	d *schema.ResourceData,
	cli wallarm.API,
	clientID int,
	ruleType, attackType string,
	readMethod func(*schema.ResourceData, interface{}) error,
) error {
	actionsFromState := d.Get("action").(*schema.Set)
	action, err := ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return errors.WithMessage(err, "on ExpandSetToActionDetailsList")
	}

	enumeratedParametersFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "enumerated_parameters")
	enumeratedParameters := tftoapi.EnumeratedParameters(enumeratedParametersFromState)

	reactionFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "reaction")
	reaction := tftoapi.Reaction(reactionFromState)

	thresholdFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "threshold")
	threshold := tftoapi.Threshold(thresholdFromState)

	advancedConditionsFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "advanced_conditions")
	advancedConditions := tftoapi.AdvancedConditions(advancedConditionsFromState)

	arbitraryConditionsFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "arbitrary_conditions")
	arbitraryConditions := tftoapi.ArbitraryConditionsReq(arbitraryConditionsFromState)

	pointFromState := GetValueWithTypeCastingOrDefault[[]interface{}](d, "point")
	points, err := expandPointsToTwoDimensionalArray(pointFromState)
	if err != nil {
		return err
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
		Mitigation:           GetValueWithTypeCastingOrDefault[string](d, "mitigation"),
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
		return err
	}

	d.Set("rule_id", actionResp.Body.ID)
	d.Set("action_id", actionResp.Body.ActionID)
	d.Set("rule_type", actionResp.Body.Type)

	resID := fmt.Sprintf("%d/%d/%d", clientID, actionResp.Body.ActionID, actionResp.Body.ID)
	d.SetId(resID)

	return readMethod(d, cli)
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

func wrapPointElements(input []interface{}) [][]string {
	var result [][]string // This will store the final result as a 2D slice of strings
	i := 0

	for i < len(input) {
		switch input[i] {
		case "json_array", "xml_pi", "hash", "array", "viewstate_array", "viewstate_pair",
			"viewstate_triplet", "viewstate_dict", "header", "xml_dtd_entity",
			"xml_tag_array", "xml_tag", "xml_attr", "xml_comment", "grpc", "protobuf",
			"json_obj", "json", "jwt", "multipart", "get", "content_disp", "form_urlencoded",
			"path", "cookie", "response_header", "viewstate_sparse_array":
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

func hashResponseActionDetails(v interface{}) int {
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
	return hashcode.String(buf.String()) // nolint:staticcheck
}

func actionDetailsToMap(actionDetails wallarm.ActionDetails) (map[string]interface{}, error) {
	jsonActions, err := json.Marshal(actionDetails)
	if err != nil {
		return nil, err
	}
	var mapActions map[string]interface{}
	if err = json.Unmarshal(jsonActions, &mapActions); err != nil {
		return nil, err
	}
	if _, ok := mapActions["value"]; !ok {
		mapActions["value"] = ""
	}
	return mapActions, nil
}

func expandPointsToTwoDimensionalArray(ps []interface{}) (wallarm.TwoDimensionalSlice, error) {
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
