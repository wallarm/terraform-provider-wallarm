package wallarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ruleNotFoundError struct {
	clientID int
	ruleID   int
}

func (e *ruleNotFoundError) Error() string {
	return fmt.Sprintf("rule %d for client %d not found", e.ruleID, e.clientID)
}

func expandInterfaceToStringList(list interface{}) []string {
	ifaceList := list.([]interface{})
	vs := []string{}
	for _, v := range ifaceList {
		vs = append(vs, v.(string))
	}
	return vs
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
		case path:
			pointMap := make(map[string]string)
			pointMap[path] = fmt.Sprintf("%d", int(p[1].(float64)))
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

func expandPointsToTwoDimensionalArray(ps []interface{}) (wallarm.TwoDimensionalSlice, error) {
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

func alignPointScheme(rulePoint []interface{}) []interface{} {
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

func interfaceToString(i interface{}) string {
	r, _ := i.(string)
	return r
}

func interfaceToInt(i interface{}) int {
	r, _ := i.(int)
	return r
}

func appendMap(united, b map[string]int) map[string]int {
	for k, v := range b {
		united[k] = v
	}
	return united
}

func retrieveClientID(d *schema.ResourceData) int {
	if v, ok := d.GetOk("client_id"); ok {
		return v.(int)
	}
	return ClientID
}

func diffStringSlice(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// nolint:gosec
func passwordGenerate(length int) string {
	digits := "0123456789"
	specials := "~=+%^*()_[]{}!@#$?"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	str := string(buf)
	return str
}

func isPasswordValid(s string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 7 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func expandWallarmEventToIntEvents(d interface{}, resourceType string) *[]wallarm.IntegrationEvents {
	cfg := d.(*schema.Set).List()
	events := []wallarm.IntegrationEvents{}
	if len(cfg) == 0 || cfg[0] == nil {
		var defaultEvents []map[string]interface{}
		switch resourceType {
		case "email":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "vuln_high",
					"active":     false},
				{
					"event_type": "vuln_medium",
					"active":     false},
				{
					"event_type": "vuln_low",
					"active":     false},
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "scope",
					"active":     false},
				{
					"event_type": "report_daily",
					"active":     false},
				{
					"event_type": "report_weekly",
					"active":     false},
				{
					"event_type": "report_monthly",
					"active":     false},
			}
		case "opsgenie":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "vuln_high",
					"active":     false},
				{
					"event_type": "vuln_medium",
					"active":     false},
				{
					"event_type": "vuln_low",
					"active":     false},
				{
					"event_type": "siem",
					"active":     false},
			}
		default:
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "vuln_high",
					"active":     false},
				{
					"event_type": "vuln_medium",
					"active":     false},
				{
					"event_type": "vuln_low",
					"active":     false},
				{
					"event_type": "siem",
					"active":     false},
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "scope",
					"active":     false},
			}
		}
		for _, v := range defaultEvents {
			event := wallarm.IntegrationEvents{
				Event:  v["event_type"].(string),
				Active: v["active"].(bool),
			}
			events = append(events, event)
		}
		return &events
	}

	for _, conf := range cfg {

		m := conf.(map[string]interface{})
		e := wallarm.IntegrationEvents{}
		event, ok := m["event_type"]
		if ok {
			if event.(string) == "hit" {
				e.Event = "siem"
			} else {
				e.Event = event.(string)
			}
		}

		active, ok := m["active"]
		if ok {
			e.Active = active.(bool)
		}
		events = append(events, e)
	}
	return &events
}

func fillInDefaultValues(action *[]wallarm.ActionDetails) {
	acts := make([]wallarm.ActionDetails, 0, len(*action))
	for _, a := range *action {
		if a.Type == "absent" {
			a.Value = nil
		}
		acts = append(acts, a)
	}
	*action = acts
}

// equalWithoutOrder tells whether a and b contain
// the same elements regardless the order.
// Applicable only for []wallarm.ActionDetails
func equalWithoutOrder(conditionsA, conditionsB []wallarm.ActionDetails) bool {
	if len(conditionsA) != len(conditionsB) {
		return false
	}

	// To embrace the default branch without conditions
	if len(conditionsA) == 0 && len(conditionsB) == 0 {
		return true
	}

	sort.Slice(conditionsA, func(i, j int) bool {
		// Преобразуем Point в строку для сравнения
		pointStrI := strings.Join(convertToStringSlice(conditionsA[i].Point), "/")
		pointStrJ := strings.Join(convertToStringSlice(conditionsA[j].Point), "/")
		return pointStrI < pointStrJ
	})

	sort.Slice(conditionsB, func(i, j int) bool {
		// Преобразуем Point в строку для сравнения
		pointStrI := strings.Join(convertToStringSlice(conditionsB[i].Point), "/")
		pointStrJ := strings.Join(convertToStringSlice(conditionsB[j].Point), "/")
		return pointStrI < pointStrJ
	})

	for i := range conditionsA {
		if !compareActionDetails(conditionsA[i], conditionsB[i]) {
			return false
		}
	}

	return true
}

func convertToStringSlice(input []interface{}) []string {
	result := make([]string, 0, len(input))
	for _, v := range input {
		result = append(result, fmt.Sprintf("%v", v))
	}
	return result
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

func existsAction(d *schema.ResourceData, m interface{}, hintType string) (string, bool, error) {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return "", false, err
	}

	rule := &wallarm.ActionRead{
		Filter: &wallarm.ActionFilter{
			HintType: []string{hintType},
			Clientid: []int{clientID},
		},
		Limit:  1000,
		Offset: 0,
	}

	respRules, err := client.RuleRead(rule)
	if err != nil {
		return "", false, err
	}

	for _, body := range respRules.Body {

		var apiActions []wallarm.ActionDetails

		for _, condition := range body.Conditions {
			apiAct := condition.(map[string]interface{})
			result, err := json.Marshal(apiAct)
			if err != nil {
				return "", false, err
			}
			var apiAction wallarm.ActionDetails
			err = json.Unmarshal(result, &apiAction)
			if err != nil {
				return "", false, err
			}
			apiActions = append(apiActions, apiAction)
		}
		if equalWithoutOrder(action, apiActions) {
			actionID := body.ID
			return existsHint(d, m, actionID, hintType)
		}
	}
	return "", false, err
}

func existsHint(d *schema.ResourceData, m interface{}, actionID int, hintType string) (string, bool, error) {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	var points wallarm.TwoDimensionalSlice

	if ps, ok := d.GetOk("point"); ok {
		var err error
		points, err = expandPointsToTwoDimensionalArray(ps.([]interface{}))
		if err != nil {
			return "", false, err
		}
	}

	hint := &wallarm.HintRead{
		Limit:     1000,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ActionID: []int{actionID},
			Type:     []string{hintType},
			Point:    points,
		},
	}
	actionHints, err := client.HintRead(hint)
	if err != nil {
		return "", false, err
	}
	for _, rule := range *actionHints.Body {
		existingID := fmt.Sprintf("%d/%d/%d/%s", clientID, actionID, rule.ID, hintType)
		return existingID, true, nil
	}
	return "", false, nil
}

// ImportAsExistsError returns an error when a resource exists in API
// on time of executing a module and should be
// imported beforehand to work within Terraform. It
// accepts resource name with its resource identificator.
// Generally, ID is something like `/6039/4123/93830`
func ImportAsExistsError(resourceName, id string) error {
	return fmt.Errorf(`the resource with the ID %q already exists -
		to be managed via Terraform this resource needs to be imported into the State. 
		Please see the resource documentation for %q for more information`, id, resourceName)
}

func isNotFoundError(err error) (bool, error) {
	matched, matchErr := regexp.MatchString("HTTP Status: 404", err.Error())
	if matchErr != nil {
		return false, matchErr
	}

	return matched, nil
}

func findRule(client wallarm.API, clientID, ruleID int) (*wallarm.ActionBody, error) {
	resp, err := client.HintRead(&wallarm.HintRead{
		Limit:   1,
		OrderBy: "updated_at",
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ID:       []int{ruleID},
		},
	})
	if err != nil {
		return nil, errors.WithMessagef(err, "on client.HintRead, client ID %d, rule ID %d", clientID, ruleID)
	}
	if resp == nil || resp.Body == nil || len(*resp.Body) == 0 {
		return nil, &ruleNotFoundError{clientID: clientID, ruleID: ruleID}
	}

	return &(*resp.Body)[0], nil
}
