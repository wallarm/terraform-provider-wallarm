package wallarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func expandInterfaceToStringList(list interface{}) []string {
	ifaceList := list.([]interface{})
	vs := []string{}
	for _, v := range ifaceList {
		vs = append(vs, v.(string))
	}
	return vs
}

func expandInterfaceToIntList(list interface{}) []int {
	ifaceList := list.([]interface{})
	vs := []int{}
	for _, v := range ifaceList {
		vs = append(vs, v.(int))
	}
	return vs
}

func expandSetToActionDetailsList(action *schema.Set) ([]wallarm.ActionDetails, error) {
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
					case "path":
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
						if actionMap["type"] == "iequal" {
							a.Value = strings.ToLower(pointValue.(string))
						} else if actionMap["type"] == "absent" {
							a.Value = nil
						} else {
							a.Value = pointValue.(string)
						}
					case "instance":
						a.Point = []interface{}{pointKey}
						a.Value = pointValue.(string)
						a.Type = "equal"
					case "header":
						// This is required by the API when a header field is specified
						a.Point = []interface{}{pointKey, strings.ToUpper(pointValue.(string))}
					case "query":
						// This is required by the API when case is insensitive
						if actionMap["type"] == "iequal" {
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

func actionDetailsToMap(actionDetails wallarm.ActionDetails) (mapActions map[string]interface{}, err error) {
	jsonActions, err := json.Marshal(actionDetails)
	if err != nil {
		return
	}
	json.Unmarshal(jsonActions, &mapActions)
	if _, ok := mapActions["value"]; !ok {
		mapActions["value"] = ""
	}
	return
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
			pointMap["proto"] = m["value"].(string)
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
		case "path":
			pointMap := make(map[string]string)
			pointMap["path"] = fmt.Sprintf("%d", int(p[1].(float64)))
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
	return hashcode.String(buf.String())
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
			if wallarm.Contains(numericPoints, (rulePoint)[i-1]) {
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
	switch i.(type) {
	case string:
		return i.(string)
	default:
		return ""
	}
}

func interfaceToInt(i interface{}) int {
	switch i.(type) {
	case int:
		return i.(int)
	default:
		return 0
	}
}

func appendMap(united, b map[string]int) map[string]int {
	for k, v := range b {
		united[k] = v
	}
	return united
}

func reverseMap(m map[string]int) map[int]string {
	n := make(map[int]string)
	for k, v := range m {
		n[v] = k
	}
	return n
}

func retrieveClientID(d *schema.ResourceData, client wallarm.API) int {
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

func passwordGenerate(length int) string {
	rand.Seed(time.Now().UnixNano())
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

func expandWallarmEventToIntEvents(d interface{}, resourceType string) (*[]wallarm.IntegrationEvents, error) {
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
		return &events, nil
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
	return &events, nil
}

func fillInDefaultValues(action *[]wallarm.ActionDetails) {
	var acts []wallarm.ActionDetails
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
func equalWithoutOrder(a, b []wallarm.ActionDetails) bool {
	if len(a) != len(b) {
		return false
	}

	// To embrace the default branch without conditions
	if len(a) == 0 && len(b) == 0 {
		return true
	}

	for _, outer := range a {
		flag := false
		for _, inner := range b {
			if outer.Type == inner.Type {
				flag = true
			}
		}
		if flag {
			flag = false
			for _, inner := range b {
				if outer.Value == inner.Value {
					flag = true
				}
			}
		}
		if flag {
			for _, inner := range b {
				if actionPointsEqual(outer.Point, inner.Point) {
					return true
				}

			}
		}
	}
	return false
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
	clientID := retrieveClientID(d, client)

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := expandSetToActionDetailsList(actionsFromState)
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
		var apiActions []wallarm.ActionDetails = nil
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
	clientID := retrieveClientID(d, client)

	hint := &wallarm.HintRead{
		Limit:     1000,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ActionID: []int{actionID},
			Type:     []string{hintType},
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
	return fmt.Errorf("the resource with the ID %q already exists "+
		"- to be managed via Terraform this resource needs to be imported into the State. "+
		"Please see the resource documentation for %q for more information.", id, resourceName)
}

func isNotFoundError(err error) (bool, error) {
	matched, matchErr := regexp.MatchString("HTTP Status: 404", err.Error())
	if matchErr != nil {
		return false, matchErr
	}

	return matched, nil
}
