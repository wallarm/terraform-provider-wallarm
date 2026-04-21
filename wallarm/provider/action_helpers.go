package wallarm

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"
)

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
//     formatting when needed)
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

	wantHash := resourcerule.ConditionsHash(action)
	var matchedAction *wallarm.ActionEntry
	for i, entry := range listResp.Body {
		if len(entry.Conditions) != len(action) {
			continue
		}
		if resourcerule.ConditionsHash(entry.Conditions) == wantHash {
			matchedAction = &listResp.Body[i]
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

	if body := *hintResp.Body; len(body) > 0 {
		return matchedAction.ID, &body[0], true, nil
	}
	return 0, nil, false, nil
}

// ImportAsExistsError returns an error when a resource already exists in the API
// and should be imported into Terraform state first.
func ImportAsExistsError(resourceName, id string) error {
	return fmt.Errorf("the resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for %q for more information", id, resourceName)
}
