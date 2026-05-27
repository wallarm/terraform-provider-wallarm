package wallarm

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
func existingHintForAction(d *schema.ResourceData, m any, hintType string) (actionID int, rule *wallarm.ActionBody, exists bool, err error) {
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

	wantHash := resourcerule.ConditionsHash(action)
	matchedAction, err := findActionByConditionsHash(client, clientID, hintType, wantHash, len(action))
	if err != nil {
		return 0, nil, false, err
	}
	if matchedAction == nil {
		return 0, nil, false, nil
	}

	// Now look up the hint on that action. Preserve the existing Point filter:
	// if the resource's schema has a `point` field set, narrow the search by
	// point; otherwise pass empty (existing filter behavior).
	var points wallarm.TwoDimensionalSlice
	if ps, ok := d.GetOk("point"); ok {
		expanded, err := resourcerule.ExpandPointsToTwoDimensionalArray(ps.([]any))
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

	if hintResp.Body != nil && len(*hintResp.Body) > 0 {
		body := *hintResp.Body
		return matchedAction.ID, &body[0], true, nil
	}
	return 0, nil, false, nil
}

// findActionByConditionsHashPageCap caps pagination at 200 pages × APIListLimit
// (100k actions per hint type) — guards against a misbehaving API that keeps
// returning full pages without progress. Cap is far above the largest tenant
// we expect; hitting it indicates an API bug, so we surface an error instead
// of hanging the apply.
const findActionByConditionsHashPageCap = 200

// findActionByConditionsHash paginates ActionList for (clientID, hintType) and
// returns the first ActionEntry whose conditions hash matches wantHash, or nil
// if no match. Pagination terminates on a short page, on match, or on the
// page cap (with error). Common case (match on page 1) makes a single API call.
//
// expectConditionCount is a cheap length pre-filter to skip the hash compute
// for actions that can't possibly match. Pass len(action).
func findActionByConditionsHash(client wallarm.API, clientID int, hintType, wantHash string, expectConditionCount int) (*wallarm.ActionEntry, error) {
	const pageSize = APIListLimit
	for page := 0; page < findActionByConditionsHashPageCap; page++ {
		offset := page * pageSize
		listResp, err := client.ActionList(&wallarm.ActionListParams{
			Filter: &wallarm.ActionListFilter{
				HintType: []string{hintType},
				Clientid: []int{clientID},
			},
			Limit:  pageSize,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		for i, entry := range listResp.Body {
			if len(entry.Conditions) != expectConditionCount {
				continue
			}
			if resourcerule.ConditionsHash(entry.Conditions) == wantHash {
				return &listResp.Body[i], nil
			}
		}
		// Short page (including empty) → no more results.
		if len(listResp.Body) < pageSize {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("findActionByConditionsHash: pagination cap (%d pages × %d) exceeded for hint type %q without short page — possible API bug", findActionByConditionsHashPageCap, APIListLimit, hintType)
}

// ImportAsExistsError returns an error when a resource already exists in the API
// and should be imported into Terraform state first.
func ImportAsExistsError(resourceName, id string) error {
	return fmt.Errorf("the resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for %q for more information", id, resourceName)
}

// guardExistingHint runs the existingHintForAction collision check on Create
// and returns ImportAsExistsError diagnostics if a hint of `hintType` is
// already attached to a matching action scope. idFmt formats the import-id
// from (clientID, actionID, matched rule); pass nil for the default 3-part
// `{clientID}/{actionID}/{ruleID}` format. Resources with a 4-part import ID
// (e.g., wallarm_rule_mode) pass an idFmt that appends their suffix.
func guardExistingHint(
	d *schema.ResourceData, m any,
	hintType, resourceName string,
	idFmt func(clientID, actionID int, r *wallarm.ActionBody) string,
) diag.Diagnostics {
	if !d.IsNewResource() {
		return nil
	}
	actionID, rule, exists, err := existingHintForAction(d, m, hintType)
	if err != nil {
		return diag.FromErr(err)
	}
	if !exists {
		return nil
	}
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	var existingID string
	if idFmt != nil {
		existingID = idFmt(clientID, actionID, rule)
	} else {
		existingID = fmt.Sprintf("%d/%d/%d", clientID, actionID, rule.ID)
	}
	return diag.FromErr(ImportAsExistsError(resourceName, existingID))
}
