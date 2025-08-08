package resourcerule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

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
	var (
		//actionID                 = d.Get("action_id").(int)
		ruleID = d.Get("rule_id").(int)
		//withPoint                = slices.Contains(opts, common.ReadOptionWithPoint)
		//withAction               = slices.Contains(opts, common.ReadOptionWithAction)
		//withRegexID              = slices.Contains(opts, common.ReadOptionWithRegexID)
		//withMode                 = slices.Contains(opts, common.ReadOptionWithMode)
		//withName                 = slices.Contains(opts, common.ReadOptionWithName)
		//withValues               = slices.Contains(opts, common.ReadOptionWithValues)
		//withThreshold            = slices.Contains(opts, common.ReadOptionWithThreshold)
		//withReaction             = slices.Contains(opts, common.ReadOptionWithReaction)
		//withEnumeratedParameters = slices.Contains(opts, common.ReadOptionWithEnumeratedParameters)
		// withArbitraryConditions  = slices.Contains(opts, common.ReadOptionWithArbitraryConditions)
	)

	//actionsFromState := d.Get("action").(*schema.Set)
	//action, err := ExpandSetToActionDetailsList(actionsFromState)
	//if err != nil {
	//	return err
	//}
	//
	//actsSlice := make([]interface{}, 0, len(action))
	//for _, a := range action {
	//	acts, err := ActionDetailsToMap(a)
	//	if err != nil {
	//		return err
	//	}
	//	actsSlice = append(actsSlice, acts)
	//}
	//
	//actionsSet := schema.NewSet(HashResponseActionDetails, actsSlice)

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
		log.Println("hihihi1", err)
		return err
	}

	//// This is mandatory to fill in the default values in order to compare them deeply.
	//// Assign new values to the old struct slice.
	//FillInDefaultValues(&action)

	//expectedRule := wallarm.ActionBody{ActionID: actionID}
	//if withPoint {
	//	expectedRule.Point = pointsFromResource(d)
	//}
	//if withAction {
	//	expectedRule.Action = action
	//}
	//if withRegexID {
	//	expectedRule.RegexID = d.Get("regex_id").(int)
	//}
	//if withMode {
	//	expectedRule.Mode = d.Get("mode").(string)
	//}
	//if withName {
	//	expectedRule.Name = d.Get("name").(string)
	//}
	//if withValues {
	//	expectedRule.Values = d.Get("values").([]interface{})
	//}
	//if withThreshold {
	//	expectedRule.Threshold = tftoapi.Threshold(d.Get("threshold").([]interface{}))
	//}
	//if withReaction {
	//	expectedRule.Reaction = tftoapi.Reaction(d.Get("reaction").([]interface{}))
	//}
	//if withEnumeratedParameters {
	//	expectedRule.EnumeratedParameters = tftoapi.EnumeratedParameters(d.Get("enumerated_parameters").([]interface{}))
	//}
	// if withArbitraryConditions {
	//	expectedRule.ArbitraryConditions = tf_to_api.ArbitraryConditionsReq(d.Get("arbitrary_conditions").([]interface{}))
	// }
	var updatedRule *wallarm.ActionBody
	for _, rule := range *actionHints.Body {
		if ruleID == rule.ID {
			updatedRule = &rule
			log.Println("hihihi3 point from api", updatedRule.Point)
			break
		}

		//actualRule := &wallarm.ActionBody{ActionID: rule.ActionID}
		//if withPoint {
		//	// The response has a different structure so we have to align them
		//	// to uniform view then to compare.
		//	actualRule.Point = AlignPointScheme(rule.Point)
		//}
		//if withAction {
		//	actualRule.Action = action
		//}
		//if withRegexID {
		//	actualRule.RegexID = rule.RegexID
		//}
		//if withMode {
		//	actualRule.Mode = rule.Mode
		//}
		//if withName {
		//	actualRule.Name = rule.Name
		//}
		//if withValues {
		//	actualRule.Values = rule.Values
		//}
		//if withThreshold {
		//	actualRule.Threshold = rule.Threshold
		//}
		//if withEnumeratedParameters {
		//	actualRule.EnumeratedParameters = rule.EnumeratedParameters
		//}
		//
		//if cmp.Equal(expectedRule, *actualRule) && EqualWithoutOrder(action, rule.Action) {
		//	updatedRule = &rule
		//	break
		//}
	}

	if updatedRule == nil {
		log.Println("hihihi2 not found in API")
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
	d.Set("enumerated_parameters", apitotf.EnumeratedParameters(updatedRule.EnumeratedParameters))
	d.Set("arbitrary_conditions", apitotf.ArbitraryConditions(updatedRule.ArbitraryConditions))
	d.Set("counter", updatedRule.Counter)

	actionsSet := schema.Set{F: hashResponseActionDetails}
	for _, a := range updatedRule.Action {
		acts, err := actionDetailsToMap(a)
		if err != nil {
			log.Println("hihihi4 on actionDetailsToMap", err)
		} else {
			actionsSet.Add(acts)
		}
	}
	d.Set("action", &actionsSet)

	log.Println("hihihi3 found in API, no errors")

	log.Println("hihihi3 point", d.Get("point"))
	log.Println("hihihi3 action", d.Get("action"))
	return nil
}

// nolint
func ResourceRuleWallarmCreate(
	d *schema.ResourceData,
	cli wallarm.API,
	clientID int,
	ruleType, attackType string,
	readMethod func(*schema.ResourceData, interface{}) error,
	opts ...common.CreateOption) error {
	actionsFromState := d.Get("action").(*schema.Set)
	action, err := ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return errors.WithMessage(err, "on ExpandSetToActionDetailsList")
	}

	enumeratedParametersFromState := d.Get("enumerated_parameters").([]interface{})
	enumeratedParameters := tftoapi.EnumeratedParameters(enumeratedParametersFromState)

	reactionFromState := d.Get("reaction").([]interface{})
	reaction := tftoapi.Reaction(reactionFromState)

	thresholdFromState := d.Get("threshold").([]interface{})
	threshold := tftoapi.Threshold(thresholdFromState)

	advancedConditionsFromState := d.Get("advanced_conditions").([]interface{})
	advancedConditions := tftoapi.AdvancedConditions(advancedConditionsFromState)

	arbitraryConditionsFromState := d.Get("arbitrary_conditions").([]interface{})
	arbitraryConditions := tftoapi.ArbitraryConditionsReq(arbitraryConditionsFromState)

	wm := &wallarm.ActionCreate{
		Type:                 ruleType,
		Clientid:             clientID,
		Action:               &action,
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

func FillInDefaultValues(action *[]wallarm.ActionDetails) {
	acts := make([]wallarm.ActionDetails, 0, len(*action))
	for _, a := range *action {
		if a.Type == "absent" {
			a.Value = nil
		}
		acts = append(acts, a)
	}
	*action = acts
}

func ActionDetailsToMap(actionDetails wallarm.ActionDetails) (map[string]interface{}, error) {
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
	return hashcode.String(buf.String()) // nolint:staticcheck
}

// equalWithoutOrder tells whether a and b contain
// the same elements regardless the order.
// Applicable only for []wallarm.ActionDetails
func EqualWithoutOrder(conditionsA, conditionsB []wallarm.ActionDetails) bool {
	if len(conditionsA) != len(conditionsB) {
		return false
	}

	// To embrace the default branch without conditions
	if len(conditionsA) == 0 && len(conditionsB) == 0 {
		return true
	}

	sort.Slice(conditionsA, func(i, j int) bool {
		// Преобразуем Point в строку для сравнения
		pointStrI := strings.Join(common.ConvertToStringSlice(conditionsA[i].Point), "/")
		pointStrJ := strings.Join(common.ConvertToStringSlice(conditionsA[j].Point), "/")
		return pointStrI < pointStrJ
	})

	sort.Slice(conditionsB, func(i, j int) bool {
		// Преобразуем Point в строку для сравнения
		pointStrI := strings.Join(common.ConvertToStringSlice(conditionsB[i].Point), "/")
		pointStrJ := strings.Join(common.ConvertToStringSlice(conditionsB[j].Point), "/")
		return pointStrI < pointStrJ
	})

	for i := range conditionsA {
		if !compareActionDetails(conditionsA[i], conditionsB[i]) {
			return false
		}
	}

	return true
}

func AlignPointScheme(rulePoint []interface{}) []interface{} {
	// Check by comparing the defined struct.
	// This is needed when a new rule was overwritten by the old one, but API responds with rule ID
	// which will be immediately removed as duplicate in favor of the first once we post it.
	numericPoints := []string{"path", "array", "grpc", "json_array", "xml_comment",
		"xml_dtd_entity", "xml_pi", "xml_tag_array"}
	var points []interface{}
	for i, point := range rulePoint {
		if i == 0 {
			points = append(points, point)
		} else {
			if wallarm.Contains(numericPoints, rulePoint[i-1]) {
				number := fmt.Sprintf("%d", int(point.(float64)))
				points = append(points, number)
			} else {
				points = append(points, point)
			}
		}
	}
	return points
}

func AlignPointScheme2(rulePoint []interface{}) []interface{} {
	// Check by comparing the defined struct.
	// This is needed when a new rule was overwritten by the old one, but API responds with rule ID
	// which will be immediately removed as duplicate in favor of the first once we post it.
	numericPoints := []string{"path", "array", "grpc", "json_array", "xml_comment",
		"xml_dtd_entity", "xml_pi", "xml_tag_array"}
	var points []interface{}
	for i, point := range rulePoint {
		if i == 0 {
			points = append(points, point)
		} else {
			if wallarm.Contains(numericPoints, rulePoint[i-1]) {
				number := fmt.Sprintf("%d", int(point.(float64)))
				points = append(points, number)
			} else {
				points = append(points, point)
			}
		}
	}

	return []interface{}{points}
}

// compare for action condition
func compareActionDetails(condition1, condition2 wallarm.ActionDetails) bool {
	return condition1.Type == condition2.Type &&
		actionPointsEqual(condition1.Point, condition2.Point) &&
		reflect.DeepEqual(condition1.Value, condition2.Value)
}

func actionPointsEqual(listA, listB []interface{}) bool {
	aLen := len(listA)
	bLen := len(listB)

	if aLen != bLen {
		return false
	}

	visited := make([]bool, bLen)

	for i := 0; i < aLen; i++ {
		found := false
		element := listA[i]
		for j := 0; j < bLen; j++ {
			if visited[j] {
				continue
			}
			if element == listB[j] {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func pointsFromResource(d *schema.ResourceData) []interface{} {
	var points []interface{}
	ps := d.Get("point").([]interface{})
	for _, point := range ps {
		p := point.([]interface{})
		points = append(points, p...)
	}
	return points
}

func RetrieveClientID(d *schema.ResourceData, defaultClientID int) int {
	if v, ok := d.GetOk("client_id"); ok {
		return v.(int)
	}
	return defaultClientID
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
