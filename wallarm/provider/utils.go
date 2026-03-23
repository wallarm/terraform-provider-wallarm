package wallarm

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"unicode"

	stderrors "errors"
	"net/http"

	"github.com/pkg/errors"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const eventTypeSIEM = "siem"

// isNotFoundError checks if the error is a Wallarm API 404 response.
func isNotFoundError(err error) bool {
	var apiErr *wallarm.APIError
	return stderrors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// validateWithHeadersOnlySiem returns a CustomizeDiffFunc that ensures
// with_headers is only set to true on events of type "siem".
func validateWithHeadersOnlySiem() schema.CustomizeDiffFunc {
	return customdiff.All(
		func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			events, ok := d.GetOk("event")
			if !ok {
				return nil
			}
			for _, e := range events.(*schema.Set).List() {
				m := e.(map[string]interface{})
				eventType, _ := m["event_type"].(string)
				withHeaders, _ := m["with_headers"].(bool)
				if withHeaders && eventType != eventTypeSIEM {
					return fmt.Errorf("with_headers can only be set for the 'siem' event type, got event_type=%q", eventType)
				}
			}
			return nil
		},
	)
}

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

func interfaceToString(i interface{}) string {
	r, _ := i.(string)
	return r
}

func interfaceToInt(i interface{}) int {
	r, _ := i.(int)
	return r
}

// retrieveClientID extracts client_id from a resource or falls back to the provider default.
func retrieveClientID(d *schema.ResourceData, m interface{}) (int, error) {
	meta := m.(*ProviderMeta)
	return meta.RetrieveClientID(d)
}

// apiClient extracts the wallarm.API client from the provider meta.
func apiClient(m interface{}) wallarm.API {
	return m.(*ProviderMeta).Client
}

func passwordGenerate(length int) (string, error) {
	digits := "0123456789"
	specials := "~=+%^*()_[]{}!@#$?"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	var err error
	if buf[0], err = cryptoRandByte(digits); err != nil {
		return "", err
	}
	if buf[1], err = cryptoRandByte(specials); err != nil {
		return "", err
	}
	for i := 2; i < length; i++ {
		if buf[i], err = cryptoRandByte(all); err != nil {
			return "", err
		}
	}
	// Fisher-Yates shuffle using crypto/rand
	for i := len(buf) - 1; i > 0; i-- {
		j, err := cryptoRandIntn(i + 1)
		if err != nil {
			return "", err
		}
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf), nil
}

func cryptoRandByte(charset string) (byte, error) {
	idx, err := cryptoRandIntn(len(charset))
	if err != nil {
		return 0, err
	}
	return charset[idx], nil
}

func cryptoRandIntn(n int) (int, error) {
	maxN := big.NewInt(int64(n))
	v, err := crand.Int(crand.Reader, maxN)
	if err != nil {
		return 0, fmt.Errorf("crypto/rand failed: %w", err)
	}
	return int(v.Int64()), nil
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
					"event_type": "system",
					"active":     false},
				{
					"event_type": "aasm_report",
					"active":     false},
				{
					"event_type": "api_discovery_hourly_changes_report",
					"active":     false},
				{
					"event_type": "api_discovery_daily_changes_report",
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
		case "data_dog", "insight_connect", "splunk", "sumo_logic", "web_hooks":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": eventTypeSIEM,
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "number_of_requests_per_hour",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
					"active":     false},
				{
					"event_type": "system",
					"active":     false},
			}
		case "telegram":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
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
		case "ms_teams", "opsgenie", "pager_duty", "slack":
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
					"active":     false},
			}
		default:
			defaultEvents = []map[string]interface{}{
				{
					"event_type": "system",
					"active":     false},
				{
					"event_type": "rules_and_triggers",
					"active":     false},
				{
					"event_type": "security_issue_critical",
					"active":     false},
				{
					"event_type": "security_issue_high",
					"active":     false},
				{
					"event_type": "security_issue_medium",
					"active":     false},
				{
					"event_type": "security_issue_low",
					"active":     false},
				{
					"event_type": "security_issue_info",
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
				e.Event = eventTypeSIEM
			} else {
				e.Event = event.(string)
			}
		}

		active, ok := m["active"]
		if ok {
			e.Active = active.(bool)
		}
		// with_headers is only applicable to the siem event type
		if e.Event == eventTypeSIEM {
			if wh, ok := m["with_headers"]; ok {
				whBool := wh.(bool)
				e.WithHeaders = &whBool
			}
		}
		events = append(events, e)
	}
	return &events
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
		// Convert Point to String for Comparison
		pointStrI := strings.Join(convertToStringSlice(conditionsA[i].Point), "/")
		pointStrJ := strings.Join(convertToStringSlice(conditionsA[j].Point), "/")
		return pointStrI < pointStrJ
	})

	sort.Slice(conditionsB, func(i, j int) bool {
		// Convert Point to String for Comparison
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
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return "", false, err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return "", false, err
	}

	params := &wallarm.ActionListParams{
		Filter: &wallarm.ActionListFilter{
			HintType: []string{hintType},
			Clientid: []int{clientID},
		},
		Limit:  APIListLimit,
		Offset: 0,
	}

	resp, err := client.ActionList(params)
	if err != nil {
		return "", false, err
	}

	for _, entry := range resp.Body {
		if equalWithoutOrder(action, entry.Conditions) {
			return existsHint(d, m, entry.ID, hintType)
		}
	}
	return "", false, err
}

func existsHint(d *schema.ResourceData, m interface{}, actionID int, hintType string) (string, bool, error) {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return "", false, err
	}

	var points wallarm.TwoDimensionalSlice

	if ps, ok := d.GetOk("point"); ok {
		var err error
		points, err = resourcerule.ExpandPointsToTwoDimensionalArray(ps.([]interface{}))
		if err != nil {
			return "", false, err
		}
	}

	hint := &wallarm.HintRead{
		Limit:     APIListLimit,
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
	return fmt.Errorf("the resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for %q for more information", id, resourceName)
}

// findCredentialStuffingRule fetches credential stuffing configs via the v4 API
// and returns the one matching ruleID.
// API: GET /v4/clients/{clientID}/credential_stuffing/configs
func findCredentialStuffingRule(client wallarm.API, clientID, ruleID int) (*wallarm.ActionBody, error) {
	configs, err := client.CredentialStuffingConfigsRead(clientID)
	if err != nil {
		return nil, errors.WithMessagef(err, "on CredentialStuffingConfigsRead, client ID %d", clientID)
	}
	for i := range configs {
		if configs[i].ID == ruleID {
			return &configs[i], nil
		}
	}
	return nil, &ruleNotFoundError{clientID: clientID, ruleID: ruleID}
}
