package wallarm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/wallarm/wallarm-go"
)

func dataSourceWallarmRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWallarmRulesRead,
		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
			"types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"action_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"client_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"regex": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"attack_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"point": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"action": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWallarmRulesRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	var typeFilter []string
	if v, ok := d.GetOk("types"); ok {
		for _, t := range v.([]interface{}) {
			typeFilter = append(typeFilter, t.(string))
		}
	}

	var allRules []wallarm.ActionBody
	const limit = 1000
	offset := 0

	for {
		filter := &wallarm.HintFilter{
			Clientid: []int{clientID},
		}
		if len(typeFilter) > 0 {
			filter.Type = typeFilter
		}

		resp, err := client.HintRead(&wallarm.HintRead{
			Filter:    filter,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil {
			return fmt.Errorf("error reading rules: %w", err)
		}

		if resp == nil || resp.Body == nil || len(*resp.Body) == 0 {
			break
		}

		allRules = append(allRules, *resp.Body...)

		if len(*resp.Body) < limit {
			break
		}
		offset += limit
	}

	rules := make([]interface{}, 0, len(allRules))
	for _, rule := range allRules {
		pointJSON := ""
		if len(rule.Point) > 0 {
			if b, err := json.Marshal(rule.Point); err == nil {
				pointJSON = string(b)
			}
		}

		actionJSON := ""
		if len(rule.Action) > 0 {
			if b, err := json.Marshal(rule.Action); err == nil {
				actionJSON = string(b)
			}
		}

		rules = append(rules, map[string]interface{}{
			"rule_id":     rule.ID,
			"action_id":   rule.ActionID,
			"client_id":   rule.Clientid,
			"type":        rule.Type,
			"enabled":     rule.Enabled,
			"mode":        rule.Mode,
			"regex":       rule.Regex,
			"attack_type": rule.AttackType,
			"name":        rule.Name,
			"point":       pointJSON,
			"action":      actionJSON,
			"create_time": rule.CreateTime,
			"updated_at":  rule.UpdatedAt,
		})
	}

	if err := d.Set("rules", rules); err != nil {
		return fmt.Errorf("error setting rules: %w", err)
	}

	d.SetId(fmt.Sprintf("Rules_%s", time.Now().UTC().String()))
	return nil
}
