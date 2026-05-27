package resourcerule

import (
	"fmt"

	"github.com/wallarm/wallarm-go"
)

func ThresholdToTF(threshold *wallarm.Threshold) []any {
	if threshold == nil {
		return nil
	}

	return []any{
		map[string]any{
			"count":  threshold.Count,
			"period": threshold.Period,
		},
	}
}

func ReactionToTF(reaction *wallarm.Reaction) []any {
	if reaction == nil {
		return nil
	}

	// Only emit keys the API actually returned. The wallarm-go Reaction fields
	// are *int + omitempty, so absent keys arrive as nil. Writing nil into the
	// state map keeps the slot at its type-zero default in SDKv2's flat-state
	// model — so a partially-set reaction (e.g. only block_by_ip) doesn't pull
	// stray block_by_session=0 / graylist_by_ip=0 values into terraform import
	// + -generate-config-out output, which the IntBetween(600, 315569520)
	// validator would then reject at plan time.
	m := map[string]any{}
	if reaction.BlockBySession != nil {
		m["block_by_session"] = *reaction.BlockBySession
	}
	if reaction.BlockByIP != nil {
		m["block_by_ip"] = *reaction.BlockByIP
	}
	if reaction.GraylistByIP != nil {
		m["graylist_by_ip"] = *reaction.GraylistByIP
	}
	return []any{m}
}

func EnumeratedParametersToTF(enumeratedParameters *wallarm.EnumeratedParameters) []any {
	if enumeratedParameters == nil {
		return nil
	}

	result := map[string]any{
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

	return []any{result}
}

func mapPointsToTF(points []*wallarm.Points) []any {
	if len(points) == 0 {
		return nil
	}

	result := make([]any, 0, len(points))
	for _, pts := range points {
		if pts == nil {
			continue
		}
		point := make([]any, 0, len(pts.Point))
		point = append(point, pts.Point...)
		result = append(result, map[string]any{
			"point":     point,
			"sensitive": pts.Sensitive,
		})
	}

	return result
}

func AdvancedConditionsToTF(advancedConditions []wallarm.AdvancedCondition) []any {
	if advancedConditions == nil {
		return nil
	}

	result := make([]any, 0, len(advancedConditions))
	for _, advancedCondition := range advancedConditions {
		result = append(result, map[string]any{
			"field":    advancedCondition.Field,
			"operator": advancedCondition.Operator,
			"value":    advancedCondition.Value,
		})
	}

	return result
}

func ArbitraryConditionsToTF(arbitraryConditions []wallarm.ArbitraryConditionResp) []any {
	if arbitraryConditions == nil {
		return nil
	}

	result := make([]any, 0, len(arbitraryConditions))
	for _, arbitraryCondition := range arbitraryConditions {
		// API returns `point` as a flat array (e.g.
		// ["post", "json_doc", "hash", "user_id"]). The Terraform schema
		// expects a 2D representation grouping paired elements with their
		// value (e.g. [["post"], ["json_doc"], ["hash", "user_id"]]).
		// WrapPointElements consults the same paired/simple element table
		// used by the rule-level `point` field so the round-trip matches
		// what the user wrote in HCL.
		result = append(result, map[string]any{
			"point":    wrappedPointToInterface(WrapPointElements(arbitraryCondition.Point)),
			"operator": arbitraryCondition.Operator,
			"value":    arbitraryCondition.Value,
		})
	}

	return result
}

// wrappedPointToInterface converts the [][]string output of WrapPointElements
// to the []any of []any shape SDKv2 expects when setting a
// nested `TypeList` of `TypeList` of `TypeString`.
func wrappedPointToInterface(in [][]string) []any {
	out := make([]any, 0, len(in))
	for _, sub := range in {
		inner := make([]any, 0, len(sub))
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
