package wallarm

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"

	resourcerule "github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
)

func dataSourceWallarmActions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWallarmActionsRead,
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Client/tenant ID.",
			},
			"actions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"conditions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"point": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"conditions_hash": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dir_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_instance": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Description: "List of all actions for the client.",
			},
		},
	}
}

func dataSourceWallarmActionsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("client_id", clientID)

	// Fetch all non-empty actions (actions that have at least one rule).
	nonEmpty := false
	params := &wallarm.ActionListParams{
		Filter: &wallarm.ActionListFilter{
			Clientid: []int{clientID},
			Empty:    &nonEmpty,
		},
		Limit:  1000,
		Offset: 0,
	}

	var allEntries []wallarm.ActionEntry
	for {
		resp, err := client.ActionList(params)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to list actions: %w", err))
		}
		allEntries = append(allEntries, resp.Body...)
		if len(resp.Body) < params.Limit {
			break
		}
		params.Offset += params.Limit
	}

	log.Printf("[DEBUG] data.wallarm_actions: fetched %d actions for client_id=%d", len(allEntries), clientID)

	actions := make([]map[string]interface{}, len(allEntries))
	for i, entry := range allEntries {
		conditions := flattenActionConditions(entry.Conditions)
		condHash := resourcerule.ConditionsHash(entry.Conditions)
		dirName := resourcerule.ActionDirName(entry.Conditions)

		action := map[string]interface{}{
			"action_id":         entry.ID,
			"conditions":        conditions,
			"conditions_hash":   condHash,
			"dir_name":          dirName,
			"updated_at":        entry.UpdatedAt,
			"endpoint_path":     ptrStringValue(entry.EndpointPath),
			"endpoint_domain":   ptrStringValue(entry.EndpointDomain),
			"endpoint_instance": ptrStringValue(entry.EndpointInstance),
		}
		actions[i] = action
	}

	if err := d.Set("actions", actions); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func ptrStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
