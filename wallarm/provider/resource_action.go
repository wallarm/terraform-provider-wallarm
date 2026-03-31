package wallarm

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	wallarm "github.com/wallarm/wallarm-go"

	resourcerule "github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
)

func resourceWallarmAction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmActionCreate,
		ReadContext:   resourceWallarmActionRead,
		DeleteContext: resourceWallarmActionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmActionImport,
		},
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Client/tenant ID. Defaults to provider's default client ID.",
			},
			"action_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The API-assigned action ID. Null until the first rule under this action is created.",
			},
			"conditions": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"equal", "iequal", "regex", "absent", ""}, false),
						},
						"point": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
				Description: "Action conditions. Can be empty [] for the default action.",
			},
			"conditions_hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA256 hash of conditions (Ruby-compatible).",
			},
			"dir_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Computed directory name for organizing rule files.",
			},
			"endpoint_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint path from the API (lossy, use conditions for accuracy).",
			},
			"endpoint_domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint domain from the API.",
			},
			"endpoint_instance": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint instance from the API.",
			},
		},
	}
}

// resourceWallarmActionCreate looks up an existing action in the API by conditions match.
// If found, stores action_id. If not found (rules not yet created), stores null action_id.
func resourceWallarmActionCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	conditions := expandActionConditions(d.Get("conditions").([]interface{}))
	condHash := resourcerule.ConditionsHash(conditions)
	dirName := resourcerule.ActionDirName(conditions)

	// Set the resource ID to conditions_hash (stable identifier).
	d.SetId(fmt.Sprintf("%d/%s", clientID, condHash))
	d.Set("client_id", clientID)
	d.Set("conditions_hash", condHash)
	d.Set("dir_name", dirName)

	// Try to find the action in the API.
	actionID, entry, found, err := findActionByConditions(client, clientID, conditions)
	if err != nil {
		return diag.FromErr(err)
	}
	if found {
		d.Set("action_id", actionID)
		if entry != nil {
			setEndpointFields(d, entry)
		}
		log.Printf("[DEBUG] wallarm_action: found existing action_id=%d for hash=%s", actionID, condHash[:8])
	} else {
		log.Printf("[DEBUG] wallarm_action: no matching action in API for hash=%s (rules not yet created)", condHash[:8])
	}

	return nil
}

// resourceWallarmActionRead refreshes the action state from the API.
// Only makes API calls when action_id is not yet known. Once action_id is
// populated, the action is considered stable — no re-reads on subsequent plans.
func resourceWallarmActionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	conditions := expandActionConditions(d.Get("conditions").([]interface{}))
	condHash := resourcerule.ConditionsHash(conditions)
	dirName := resourcerule.ActionDirName(conditions)

	d.Set("conditions_hash", condHash)
	d.Set("dir_name", dirName)

	// If action_id is already known, no API call needed.
	// Conditions are ForceNew — they can't change without recreate.
	if v, ok := d.GetOk("action_id"); ok && v.(int) != 0 {
		log.Printf("[DEBUG] wallarm_action: action_id=%d already known for hash=%s, skipping API read", v.(int), condHash[:8])
		return nil
	}

	// No action_id — try to find by conditions. This is the only case that
	// makes API calls (after Create when rules haven't been applied yet).
	actionID, entry, found, findErr := findActionByConditions(client, clientID, conditions)
	if findErr != nil {
		log.Printf("[WARN] wallarm_action: %v", findErr)
		return nil
	}
	if found {
		d.Set("action_id", actionID)
		if entry != nil {
			setEndpointFields(d, entry)
		}
		log.Printf("[DEBUG] wallarm_action: discovered action_id=%d for hash=%s", actionID, condHash[:8])
	}

	return nil
}

// resourceWallarmActionDelete removes the action from state only. No API call.
func resourceWallarmActionDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

// resourceWallarmActionImport imports an action by action_id.
// Format: {action_id} or {client_id}/{action_id}
func resourceWallarmActionImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := apiClient(m)

	var clientID, actionID int
	var err error

	parts := strings.SplitN(d.Id(), "/", 2)
	switch len(parts) {
	case 1:
		actionID, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid action_id: %s", parts[0])
		}
		clientID, err = retrieveClientID(d, m)
		if err != nil {
			return nil, err
		}
	case 2:
		clientID, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid client_id: %s", parts[0])
		}
		actionID, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid action_id: %s", parts[1])
		}
	default:
		return nil, fmt.Errorf("import format: {action_id} or {client_id}/{action_id}")
	}

	entry, err := client.ActionReadByID(actionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch action %d: %w", actionID, err)
	}

	conditions := flattenActionConditions(entry.Conditions)
	condHash := resourcerule.ConditionsHash(entry.Conditions)
	dirName := resourcerule.ActionDirName(entry.Conditions)

	d.SetId(fmt.Sprintf("%d/%s", clientID, condHash))
	d.Set("client_id", clientID)
	d.Set("action_id", actionID)
	if err := d.Set("conditions", conditions); err != nil {
		return nil, fmt.Errorf("error setting conditions: %w", err)
	}
	d.Set("conditions_hash", condHash)
	d.Set("dir_name", dirName)
	setEndpointFields(d, entry)

	return []*schema.ResourceData{d}, nil
}

// --- Helpers ---

// expandActionConditions converts the schema conditions list to []ActionDetails.
func expandActionConditions(raw []interface{}) []wallarm.ActionDetails {
	conditions := make([]wallarm.ActionDetails, 0, len(raw))
	for _, item := range raw {
		m := item.(map[string]interface{})
		condType := m["type"].(string)

		pointRaw := m["point"].([]interface{})
		point := make([]interface{}, len(pointRaw))
		for i, p := range pointRaw {
			s := p.(string)
			// Try to parse as integer (path indices come as strings from schema).
			if n, err := strconv.Atoi(s); err == nil {
				point[i] = float64(n)
			} else {
				point[i] = s
			}
		}

		var value interface{}
		if v, ok := m["value"]; ok && v.(string) != "" {
			value = v.(string)
		} else if condType != hitsCondTypeAbsent && condType != "" {
			value = ""
		}
		// For absent type, value stays nil.

		conditions = append(conditions, wallarm.ActionDetails{
			Type:  condType,
			Point: point,
			Value: value,
		})
	}
	return conditions
}

// flattenActionConditions converts []ActionDetails to the schema format.
func flattenActionConditions(conditions []wallarm.ActionDetails) []map[string]interface{} {
	result := make([]map[string]interface{}, len(conditions))
	for i, c := range conditions {
		point := make([]string, len(c.Point))
		for j, p := range c.Point {
			switch v := p.(type) {
			case string:
				point[j] = v
			case float64:
				point[j] = strconv.Itoa(int(v))
			default:
				point[j] = fmt.Sprintf("%v", v)
			}
		}

		value := ""
		if c.Value != nil {
			switch v := c.Value.(type) {
			case string:
				value = v
			default:
				value = fmt.Sprintf("%v", v)
			}
		}

		result[i] = map[string]interface{}{
			"type":  c.Type,
			"point": point,
			"value": value,
		}
	}
	return result
}

// findActionByConditions searches the API for an action matching the given conditions.
// Returns (action_id, entry, found, error). Error is non-nil on API failures.
func findActionByConditions(client wallarm.API, clientID int, conditions []wallarm.ActionDetails) (int, *wallarm.ActionEntry, bool, error) {
	targetHash := resourcerule.ConditionsHash(conditions)

	empty := len(conditions) == 0
	params := &wallarm.ActionListParams{
		Filter: &wallarm.ActionListFilter{
			Clientid: []int{clientID},
			Empty:    &empty,
		},
		Limit:  1000,
		Offset: 0,
	}

	for {
		resp, err := client.ActionList(params)
		if err != nil {
			return 0, nil, false, fmt.Errorf("failed to list actions: %w", err)
		}

		for i, entry := range resp.Body {
			if resourcerule.ConditionsHash(entry.Conditions) == targetHash {
				return entry.ID, &resp.Body[i], true, nil
			}
		}

		if len(resp.Body) < params.Limit {
			break
		}
		params.Offset += params.Limit
	}

	return 0, nil, false, nil
}

// setEndpointFields sets the endpoint_* computed fields from an ActionEntry.
func setEndpointFields(d *schema.ResourceData, entry *wallarm.ActionEntry) {
	if entry.EndpointPath != nil {
		d.Set("endpoint_path", *entry.EndpointPath)
	}
	if entry.EndpointDomain != nil {
		d.Set("endpoint_domain", *entry.EndpointDomain)
	}
	if entry.EndpointInstance != nil {
		d.Set("endpoint_instance", *entry.EndpointInstance)
	}
}
