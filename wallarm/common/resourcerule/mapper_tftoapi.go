package resourcerule

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common"
	"github.com/wallarm/wallarm-go"
)

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
	case "exact":
		return mapEnumeratedParameterExactToAPI(enumeratedParameterObj)
	default:
		return mapEnumeratedParameterRegexpToAPI(enumeratedParameterObj)
	}
}

func mapEnumeratedParameterRegexpToAPI(enumeratedParameter map[string]interface{}) (*wallarm.EnumeratedParameters, error) {
	nameRegexpsRaw, ok := enumeratedParameter["name_regexps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("enumerated_parameters.name_regexps: expected list, got %T", enumeratedParameter["name_regexps"])
	}
	nameRegexps := common.ConvertToStringSlice(nameRegexpsRaw)
	if len(nameRegexps) == 0 {
		nameRegexps = []string{""}
	}

	valueRegexpsRaw, ok := enumeratedParameter["value_regexps"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("enumerated_parameters.value_regexps: expected list, got %T", enumeratedParameter["value_regexps"])
	}
	valueRegexps := common.ConvertToStringSlice(valueRegexpsRaw)
	if len(valueRegexps) == 0 {
		valueRegexps = []string{""}
	}

	plainParameters, _ := enumeratedParameter["plain_parameters"].(bool)
	additionalParameters, _ := enumeratedParameter["additional_parameters"].(bool)

	return &wallarm.EnumeratedParameters{
		Mode:                 "regexp",
		NameRegexps:          nameRegexps,
		ValueRegexp:          valueRegexps,
		PlainParameters:      lo.ToPtr(plainParameters),
		AdditionalParameters: lo.ToPtr(additionalParameters),
	}, nil
}

func mapEnumeratedParameterExactToAPI(enumeratedParameter map[string]interface{}) (*wallarm.EnumeratedParameters, error) {
	result := &wallarm.EnumeratedParameters{
		Mode: "exact",
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
			Value:    common.ConvertToStringSlice(valueRaw),
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
			Value:    common.ConvertToStringSlice(valueRaw),
			Operator: operator,
		})
	}

	return response, nil
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
