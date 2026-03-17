package wallarm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	wallarm "github.com/wallarm/wallarm-go"
)

func dataSourceWallarmIPLists() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWallarmIPListsRead,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"list_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allowlist", "denylist", "graylist"}, false),
				Description:  "IP list type: allowlist, denylist, or graylist",
			},

			"entries": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "All IP list groups for the specified list type",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"rule_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"values": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expired_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"application_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWallarmIPListsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	listTypeStr := d.Get("list_type").(string)
	listType := mapListType(listTypeStr)

	groups, err := client.IPListRead(listType, clientID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading IP lists (%s) for client %d: %w", listTypeStr, clientID, err))
	}

	// Filter out expired entries.
	now := int(time.Now().Unix())
	entries := make([]interface{}, 0, len(groups))
	for _, g := range groups {
		if g.ExpiredAt > 0 && g.ExpiredAt < now {
			continue
		}

		appIDs := make([]interface{}, len(g.ApplicationIDs))
		for i, id := range g.ApplicationIDs {
			appIDs[i] = id
		}

		values := make([]interface{}, len(g.Values))
		for i, v := range g.Values {
			values[i] = v
		}

		entries = append(entries, map[string]interface{}{
			"id":              g.ID,
			"rule_type":       g.RuleType,
			"values":          values,
			"reason":          g.Reason,
			"expired_at":      g.ExpiredAt,
			"created_at":      g.CreatedAt,
			"application_ids": appIDs,
			"status":          g.Status,
		})
	}

	d.SetId(fmt.Sprintf("ip_lists_%d_%s", clientID, listTypeStr))

	if err := d.Set("entries", entries); err != nil {
		return diag.FromErr(fmt.Errorf("error setting entries: %w", err))
	}

	return nil
}

func mapListType(s string) wallarm.IPListType {
	switch s {
	case "allowlist":
		return wallarm.AllowlistType
	case "denylist":
		return wallarm.DenylistType
	case "graylist":
		return wallarm.GraylistType
	default:
		return wallarm.IPListType(s)
	}
}
