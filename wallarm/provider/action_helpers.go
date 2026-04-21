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

// existingHintForAction looks for a rule of the given hintType attached to an
// action whose conditions match the resource's current `action {}` blocks,
// optionally also matching the `point` schema field. The match is
// mode-agnostic — a hit means "there's already a rule of this type on this
// action scope", which is either a duplicate or a contradiction. Callers use
// this on Create to detect conflicts and abort with ImportAsExistsError
// pointing at a resource-specific import ID format.
//
// Returns:
//   - actionID: the ID of the matched action (so callers can build a full ID)
//   - rule:     the existing rule itself (callers read fields like Mode for ID
//               formatting when needed)
//   - exists:   true iff a match was found
func existingHintForAction(d *schema.ResourceData, m interface{}, hintType string) (actionID int, rule *wallarm.ActionBody, exists bool, err error) {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return 0, nil, false, err
	}

	actionsFromState := d.Get("action").(*schema.Set)
	action, err := resourcerule.ExpandSetToActionDetailsList(actionsFromState)
	if err != nil {
		return 0, nil, false, err
	}

	listResp, err := client.ActionList(&wallarm.ActionListParams{
		Filter: &wallarm.ActionListFilter{
			HintType: []string{hintType},
			Clientid: []int{clientID},
		},
		Limit:  APIListLimit,
		Offset: 0,
	})
	if err != nil {
		return 0, nil, false, err
	}

	var matchedAction *wallarm.ActionEntry
	for _, entry := range listResp.Body {
		if equalWithoutOrder(action, entry.Conditions) {
			matchedAction = &entry
			break
		}
	}
	if matchedAction == nil {
		return 0, nil, false, nil
	}

	// Now look up the hint on that action. Preserve the existing Point filter:
	// if the resource's schema has a `point` field set, narrow the search by
	// point; otherwise pass empty (existing filter behavior).
	var points wallarm.TwoDimensionalSlice
	if ps, ok := d.GetOk("point"); ok {
		expanded, err := resourcerule.ExpandPointsToTwoDimensionalArray(ps.([]interface{}))
		if err != nil {
			return 0, nil, false, err
		}
		points = expanded
	}

	hintResp, err := client.HintRead(&wallarm.HintRead{
		Limit:     APIListLimit,
		Offset:    0,
		OrderBy:   "updated_at",
		OrderDesc: true,
		Filter: &wallarm.HintFilter{
			Clientid: []int{clientID},
			ActionID: []int{matchedAction.ID},
			Type:     []string{hintType},
			Point:    points,
		},
	})
	if err != nil {
		return 0, nil, false, err
	}

	for _, r := range *hintResp.Body {
		return matchedAction.ID, &r, true, nil
	}
	return 0, nil, false, nil
}

// ImportAsExistsError returns an error when a resource already exists in the API
// and should be imported into Terraform state first.
func ImportAsExistsError(resourceName, id string) error {
	return fmt.Errorf("the resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for %q for more information", id, resourceName)
}
