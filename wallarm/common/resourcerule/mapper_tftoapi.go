package resourcerule

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
	"github.com/wallarm/wallarm-go"
)

// EnumeratedParametersFromResourceData wraps EnumeratedParametersToAPI with
// raw-config awareness for the boolean fields. The schema declares both
// `additional_parameters` and `plain_parameters` Optional+Computed, so SDK
// fills `false` from the type zero when the user omits them. If we
// unconditionally sent that to the API in regexp mode, we'd override the
// API's `true` default. By stripping keys that are null in the user's raw
// config, the mapper writes a nil pointer for those fields and `omitempty`
// on the JSON tag drops them — the API default wins.
//
// Use this helper from Create/Update paths for resources that include the
// `enumerated_parameters` block (`wallarm_rule_brute`, `_bola`, `_enum`).
func EnumeratedParametersFromResourceData(d *schema.ResourceData) (*wallarm.EnumeratedParameters, error) {
	list, ok := d.Get("enumerated_parameters").([]interface{})
	if !ok || len(list) == 0 {
		return nil, nil
	}
	cloned := cloneEnumeratedParameters(list)
	stripUnsetEnumeratedBools(cloned, d.GetRawConfig())
	return EnumeratedParametersToAPI(cloned)
}

// cloneEnumeratedParameters shallow-copies the outer list and the inner
// block map so we can mutate without affecting d's cached value.
func cloneEnumeratedParameters(list []interface{}) []interface{} {
	if len(list) == 0 {
		return list
	}
	block, ok := list[0].(map[string]interface{})
	if !ok {
		return list
	}
	dup := make(map[string]interface{}, len(block))
	for k, v := range block {
		dup[k] = v
	}
	return []interface{}{dup}
}

// stripUnsetEnumeratedBools removes `additional_parameters` and
// `plain_parameters` from the inner block map when the user did not write
// them in HCL (raw config has cty.NullVal at the path). The mapper's
// `, ok` type-assert form then yields no value → nil pointer → omitempty.
func stripUnsetEnumeratedBools(list []interface{}, rawCfg cty.Value) {
	if len(list) == 0 {
		return
	}
	block, ok := list[0].(map[string]interface{})
	if !ok {
		return
	}
	if rawCfg == cty.NilVal || rawCfg.IsNull() || !rawCfg.IsKnown() {
		return
	}
	if !rawCfg.Type().IsObjectType() || !rawCfg.Type().HasAttribute("enumerated_parameters") {
		return
	}
	ep := rawCfg.GetAttr("enumerated_parameters")
	if ep.IsNull() || !ep.IsKnown() {
		return
	}
	t := ep.Type()
	if !t.IsListType() && !t.IsSetType() && !t.IsTupleType() {
		return
	}
	elems := ep.AsValueSlice()
	if len(elems) == 0 {
		return
	}
	first := elems[0]
	if first.IsNull() || !first.IsKnown() || !first.Type().IsObjectType() {
		return
	}
	for _, key := range []string{"additional_parameters", "plain_parameters"} {
		if !first.Type().HasAttribute(key) {
			delete(block, key)
			continue
		}
		v := first.GetAttr(key)
		if v.IsNull() || !v.IsKnown() {
			delete(block, key)
		}
	}
}

func EnumeratedParametersToAPI(enumeratedParameters []interface{}) (*wallarm.EnumeratedParameters, error) {
	if len(enumeratedParameters) == 0 {
		return nil, nil
	}
	enumeratedParameterObj, ok := enumeratedParameters[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("enumerated_parameters[0]: expected map, got %T", enumeratedParameters[0])
	}
	mode, _ := enumeratedParameterObj["mode"].(string)
	switch mode {
	case modeExact:
		return mapEnumeratedParameterExactToAPI(enumeratedParameterObj)
	default:
		return mapEnumeratedParameterRegexpToAPI(enumeratedParameterObj)
	}
}

func mapEnumeratedParameterRegexpToAPI(enumeratedParameter map[string]interface{}) (*wallarm.EnumeratedParameters, error) {
	// EnumeratedParamsCustomizeDiff guarantees both lists are non-empty in
	// regexp mode at plan time. We use convertRegexpList (not the shared
	// ConvertToStringSlice, which skips nils) so an HCL `[""]` survives —
	// SDKv2 normalizes empty strings in TypeString lists to cty.NullVal,
	// which arrives at d.Get as a nil entry. Skipping the nil would produce
	// an empty slice, which `omitempty` strips from JSON, and the API
	// rejects regexp mode without name_regexps/value_regexps keys.
	nameRegexpsRaw, ok := enumeratedParameter["name_regexps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("enumerated_parameters.name_regexps: expected list, got %T", enumeratedParameter["name_regexps"])
	}
	nameRegexps := convertRegexpList(nameRegexpsRaw)

	valueRegexpsRaw, ok := enumeratedParameter["value_regexps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("enumerated_parameters.value_regexps: expected list, got %T", enumeratedParameter["value_regexps"])
	}
	valueRegexps := convertRegexpList(valueRegexpsRaw)

	// Pointer fields are populated only when the key is present in the
	// block map. Callers building this map from `*schema.ResourceData`
	// should remove unset keys (via `EnumeratedParametersFromResourceData`)
	// so the wire payload omits the field when the user didn't write it —
	// letting the API default win (`true` for both fields in regexp mode).
	result := &wallarm.EnumeratedParameters{
		Mode:        modeRegexp,
		NameRegexps: nameRegexps,
		ValueRegexp: valueRegexps,
	}
	if v, ok := enumeratedParameter["plain_parameters"].(bool); ok {
		result.PlainParameters = lo.ToPtr(v)
	}
	if v, ok := enumeratedParameter["additional_parameters"].(bool); ok {
		result.AdditionalParameters = lo.ToPtr(v)
	}
	return result, nil
}

func mapEnumeratedParameterExactToAPI(enumeratedParameter map[string]interface{}) (*wallarm.EnumeratedParameters, error) {
	result := &wallarm.EnumeratedParameters{
		Mode: modeExact,
	}

	pointsList, ok := enumeratedParameter["points"].([]interface{})
	if !ok || len(pointsList) == 0 {
		return result, nil
	}

	points := make([]*wallarm.Points, 0, len(pointsList))
	for _, item := range pointsList {
		pointsObj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		point, _ := pointsObj["point"].([]interface{})
		sensitive, _ := pointsObj["sensitive"].(bool)

		points = append(points, &wallarm.Points{
			Point:     point,
			Sensitive: sensitive,
		})
	}

	result.Points = points
	return result, nil
}

func ReactionToAPI(reaction []interface{}) (*wallarm.Reaction, error) {
	if len(reaction) == 0 {
		return nil, nil
	}

	reactionObj, ok := reaction[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("reaction[0]: expected map, got %T", reaction[0])
	}

	var blockBySession, blockByIP, graylistByIP *int
	if bbs, _ := reactionObj["block_by_session"].(int); bbs != 0 {
		blockBySession = &bbs
	}
	if bbip, _ := reactionObj["block_by_ip"].(int); bbip != 0 {
		blockByIP = &bbip
	}
	if gbip, _ := reactionObj["graylist_by_ip"].(int); gbip != 0 {
		graylistByIP = &gbip
	}

	return &wallarm.Reaction{
		BlockBySession: blockBySession,
		BlockByIP:      blockByIP,
		GraylistByIP:   graylistByIP,
	}, nil
}

func ThresholdToAPI(threshold []interface{}) (*wallarm.Threshold, error) {
	if len(threshold) == 0 {
		return nil, nil
	}

	thresholdObj, ok := threshold[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("threshold[0]: expected map, got %T", threshold[0])
	}

	count, _ := thresholdObj["count"].(int)
	period, _ := thresholdObj["period"].(int)

	return &wallarm.Threshold{
		Count:  count,
		Period: period,
	}, nil
}

func AdvancedConditionsToAPI(advancedConditions []interface{}) ([]wallarm.AdvancedCondition, error) {
	if len(advancedConditions) == 0 {
		return nil, nil
	}

	response := make([]wallarm.AdvancedCondition, 0, len(advancedConditions))
	for i, advancedCondition := range advancedConditions {
		advancedConditionObj, ok := advancedCondition.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("advanced_condition[%d]: expected map, got %T", i, advancedCondition)
		}
		valueRaw, ok := advancedConditionObj["value"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("advanced_condition[%d].value: expected list, got %T", i, advancedConditionObj["value"])
		}
		field, _ := advancedConditionObj["field"].(string)
		operator, _ := advancedConditionObj["operator"].(string)
		response = append(response, wallarm.AdvancedCondition{
			Field:    field,
			Value:    ConvertToStringSlice(valueRaw),
			Operator: operator,
		})
	}

	return response, nil
}

func ArbitraryConditionsToAPI(arbitraryConditions []interface{}) ([]wallarm.ArbitraryConditionReq, error) {
	if len(arbitraryConditions) == 0 {
		return nil, nil
	}

	response := make([]wallarm.ArbitraryConditionReq, 0, len(arbitraryConditions))
	for i, arbitraryCondition := range arbitraryConditions {
		arbitraryConditionObj, ok := arbitraryCondition.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("arbitrary_condition[%d]: expected map, got %T", i, arbitraryCondition)
		}
		pointRaw, ok := arbitraryConditionObj["point"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("arbitrary_condition[%d].point: expected list, got %T", i, arbitraryConditionObj["point"])
		}
		valueRaw, ok := arbitraryConditionObj["value"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("arbitrary_condition[%d].value: expected list, got %T", i, arbitraryConditionObj["value"])
		}
		operator, _ := arbitraryConditionObj["operator"].(string)
		response = append(response, wallarm.ArbitraryConditionReq{
			Point:    mapPointToAPI(pointRaw),
			Value:    ConvertToStringSlice(valueRaw),
			Operator: operator,
		})
	}

	return response, nil
}

// convertRegexpList converts a TF []interface{} regexp list into []string,
// preserving nil entries as "". The shared ConvertToStringSlice skips nils;
// here that would silently drop the user's `[""]` (which SDKv2 normalizes to
// cty.NullVal at d.Get) and the API would reject the regexp-mode payload.
func convertRegexpList(input []interface{}) []string {
	out := make([]string, 0, len(input))
	for _, v := range input {
		if v == nil {
			out = append(out, "")
			continue
		}
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprintf("%v", v)
		}
		out = append(out, s)
	}
	return out
}

func mapPointToAPI(point []interface{}) wallarm.TwoDimensionalSlice {
	response := make(wallarm.TwoDimensionalSlice, 0, len(point))
	if len(point) == 0 {
		return response
	}

	for _, p1 := range point {
		p1Slice, ok := p1.([]interface{})
		if !ok {
			continue
		}
		p1ToResp := make([]interface{}, 0, len(p1Slice))
		p1ToResp = append(p1ToResp, p1Slice...)
		response = append(response, p1ToResp)
	}

	return response
}
