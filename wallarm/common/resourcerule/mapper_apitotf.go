package resourcerule

import (
	"fmt"

	"github.com/wallarm/wallarm-go"
)

func ThresholdToTF(threshold *wallarm.Threshold) []interface{} {
	if threshold == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"count":  threshold.Count,
			"period": threshold.Period,
		},
	}
}

func ReactionToTF(reaction *wallarm.Reaction) []interface{} {
	if reaction == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"block_by_session": reaction.BlockBySession,
			"block_by_ip":      reaction.BlockByIP,
			"graylist_by_ip":   reaction.GraylistByIP,
		},
	}
}

func EnumeratedParametersToTF(enumeratedParameters *wallarm.EnumeratedParameters) []interface{} {
	if enumeratedParameters == nil {
		return nil
	}

	result := map[string]interface{}{
		"mode": enumeratedParameters.Mode,
	}
	switch enumeratedParameters.Mode {
	case modeExact:
		result["points"] = mapPointsToTF(enumeratedParameters.Points)
	default:
		result["name_regexps"] = enumeratedParameters.NameRegexps
		result["value_regexps"] = enumeratedParameters.ValueRegexp
		if enumeratedParameters.PlainParameters != nil {
			result["plain_parameters"] = *enumeratedParameters.PlainParameters
		}
		if enumeratedParameters.AdditionalParameters != nil {
			result["additional_parameters"] = *enumeratedParameters.AdditionalParameters
		}
	}

	return []interface{}{result}
}

func mapPointsToTF(points []*wallarm.Points) []interface{} {
	if len(points) == 0 {
		return nil
	}

	result := make([]interface{}, 0, len(points))
	for _, pts := range points {
		if pts == nil {
			continue
		}
		point := make([]interface{}, 0, len(pts.Point))
		point = append(point, pts.Point...)
		result = append(result, map[string]interface{}{
			"point":     point,
			"sensitive": pts.Sensitive,
		})
	}

	return result
}

func AdvancedConditionsToTF(advancedConditions []wallarm.AdvancedCondition) []interface{} {
	if advancedConditions == nil {
		return nil
	}

	result := make([]interface{}, 0, len(advancedConditions))
	for _, advancedCondition := range advancedConditions {
		result = append(result, map[string]interface{}{
			"field":    advancedCondition.Field,
			"operator": advancedCondition.Operator,
			"value":    advancedCondition.Value,
		})
	}

	return result
}

func ArbitraryConditionsToTF(arbitraryConditions []wallarm.ArbitraryConditionResp) []interface{} {
	if arbitraryConditions == nil {
		return nil
	}

	result := make([]interface{}, 0, len(arbitraryConditions))
	for _, arbitraryCondition := range arbitraryConditions {
		// API returns `point` as a flat array (e.g.
		// ["post", "json_doc", "hash", "user_id"]). The Terraform schema
		// expects a 2D representation grouping paired elements with their
		// value (e.g. [["post"], ["json_doc"], ["hash", "user_id"]]).
		// WrapPointElements consults the same paired/simple element table
		// used by the rule-level `point` field so the round-trip matches
		// what the user wrote in HCL.
		result = append(result, map[string]interface{}{
			"point":    wrappedPointToInterface(WrapPointElements(arbitraryCondition.Point)),
			"operator": arbitraryCondition.Operator,
			"value":    arbitraryCondition.Value,
		})
	}

	return result
}

// wrappedPointToInterface converts the [][]string output of WrapPointElements
// to the []interface{} of []interface{} shape SDKv2 expects when setting a
// nested `TypeList` of `TypeList` of `TypeString`.
func wrappedPointToInterface(in [][]string) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, sub := range in {
		inner := make([]interface{}, 0, len(sub))
		for _, s := range sub {
			inner = append(inner, s)
		}
		out = append(out, inner)
	}
	return out
}

func SliceAnyToSliceString(in []any) []string {
	if in == nil {
		return nil
	}

	result := make([]string, 0, len(in))
	for _, el := range in {
		result = append(result, fmt.Sprintf("%s", el))
	}

	return result
}
