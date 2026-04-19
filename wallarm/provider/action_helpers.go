package wallarm

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"
)

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
		pointStrI := strings.Join(convertToStringSlice(conditionsA[i].Point), "/")
		pointStrJ := strings.Join(convertToStringSlice(conditionsA[j].Point), "/")
		return pointStrI < pointStrJ
	})

	sort.Slice(conditionsB, func(i, j int) bool {
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

// compareActionDetails compares two action conditions for equality.
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

// TODO: add test — needs mock API + schema.ResourceData, verify action lookup and existsHint delegation
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

// TODO: add test — needs mock API + schema.ResourceData, verify hint lookup returns correct ID
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

// ImportAsExistsError returns an error when a resource already exists in the API
// and should be imported into Terraform state first.
func ImportAsExistsError(resourceName, id string) error {
	return fmt.Errorf("the resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for %q for more information", id, resourceName)
}
