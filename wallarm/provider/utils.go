package wallarm

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func containsStr(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
