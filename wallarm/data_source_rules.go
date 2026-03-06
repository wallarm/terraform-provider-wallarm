package wallarm

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

// apiTypeToTerraformResource maps the API's hint type string to the
// corresponding Terraform resource type name. Unknown types are exposed
// as-is in the "api_type" field but get an empty "terraform_resource_type".
var apiTypeToTerraformResource = map[string]string{
	"binary_data":           "wallarm_rule_binary_data",
	"bola":                  "wallarm_rule_bola",
	"bola_counter":          "wallarm_rule_bola_counter",
	"brute":                 "wallarm_rule_brute",
	"brute_counter":         "wallarm_rule_bruteforce_counter",
	"credentials_point":     "wallarm_rule_credential_stuffing_point",
	"credentials_regex":     "wallarm_rule_credential_stuffing_regex",
	"dirbust_counter":       "wallarm_rule_dirbust_counter",
	"disable_attack_type":   "wallarm_rule_disable_attack_type",
	"disable_stamp":         "wallarm_rule_disable_stamp",
	"disable_regex":         "wallarm_rule_ignore_regex",
	"enum":                  "wallarm_rule_enum",
	"file_upload_size_limit": "wallarm_rule_file_upload_size_limit",
	"forced_browsing":       "wallarm_rule_forced_browsing",
	"graphql_detection":     "wallarm_rule_graphql_detection",
	"overlimit_res_settings": "wallarm_rule_overlimit_res_settings",
	"parser_state":          "wallarm_rule_parser_state",
	"rate_limit":            "wallarm_rule_rate_limit",
	"rate_limit_enum":       "wallarm_rule_rate_limit_enum",
	"regex":                 "wallarm_rule_regex",
	"experimental_regex":    "wallarm_rule_regex",
	"sensitive_data":        "wallarm_rule_masking",
	"set_response_header":   "wallarm_rule_set_response_header",
	"uploads":               "wallarm_rule_uploads",
	"variative_keys":        "wallarm_rule_variative_keys",
	"variative_values":      "wallarm_rule_variative_values",
	"vpatch":                "wallarm_rule_vpatch",
	"wallarm_mode":          "wallarm_rule_mode",
}

func dataSourceWallarmRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWallarmRulesRead,
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Client ID to fetch rules for.",
			},
			"type_filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "If set, only return rules matching these API type strings (e.g. \"disable_stamp\", \"parser_state\").",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "All non-system rules fetched from the API.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Hint ID (the rule instance ID).",
						},
						"action_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Action ID (groups hints that share the same action/condition).",
						},
						"client_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"api_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Rule type as returned by the API (e.g. \"disable_stamp\", \"parser_state\").",
						},
						"terraform_resource_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Corresponding Terraform resource type (e.g. \"wallarm_rule_disable_stamp\"). Empty if the API type has no provider mapping.",
						},
						"import_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Import ID in the format \"{client_id}/{action_id}/{rule_id}\" — ready to use in import blocks.",
						},
						"attack_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"stamp": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"point": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Point as JSON string.",
						},
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Action conditions as JSON string.",
						},
						"comment": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parser": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"created_at": {
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
	clientID := d.Get("client_id").(int)

	// Build type filter set if specified
	typeFilter := make(map[string]bool)
	if v, ok := d.GetOk("type_filter"); ok {
		for _, t := range v.([]interface{}) {
			typeFilter[t.(string)] = true
		}
	}

	var allHints []wallarm.ActionBody

	// Reuse the hint cache if it's already loaded (avoids duplicate API calls)
	if cachedClient, ok := m.(*CachedClient); ok {
		if cached := cachedClient.hintCache.All(); cached != nil {
			allHints = cached
			log.Printf("[INFO] data.wallarm_rules: reusing %d hints from prefetch cache", len(allHints))
		}
	}

	// Fall back to fetching directly if cache wasn't available
	if allHints == nil {
		systemFalse := false
		offset := 0
		startTime := time.Now()

		for page := 0; page < maxBulkFetchPages; page++ {
			resp, err := client.HintRead(&wallarm.HintRead{
				Limit:     defaultBulkFetchLimit,
				Offset:    offset,
				OrderBy:   "updated_at",
				OrderDesc: true,
				Filter: &wallarm.HintFilter{
					Clientid: []int{clientID},
					System:   &systemFalse,
				},
			})
			if err != nil {
				return fmt.Errorf("data.wallarm_rules: failed to fetch hints at offset %d: %w", offset, err)
			}

			if resp.Body == nil || len(*resp.Body) == 0 {
				break
			}

			batch := *resp.Body
			allHints = append(allHints, batch...)

			if len(batch) < defaultBulkFetchLimit {
				break
			}
			offset += defaultBulkFetchLimit
		}

		log.Printf("[INFO] data.wallarm_rules: fetched %d non-system hints for client %d in %s",
			len(allHints), clientID, time.Since(startTime).Round(time.Millisecond))
	}

	// Build output list
	rules := make([]map[string]interface{}, 0, len(allHints))
	for _, h := range allHints {
		// Apply type filter if set
		if len(typeFilter) > 0 && !typeFilter[h.Type] {
			continue
		}

		tfResourceType := apiTypeToTerraformResource[h.Type]

		pointJSON, _ := json.Marshal(h.Point)
		actionJSON, _ := json.Marshal(h.Action)

		// Build import ID — most types use {client_id}/{action_id}/{rule_id}
		// but some need a 4th segment
		importID := fmt.Sprintf("%d/%d/%d", h.Clientid, h.ActionID, h.ID)
		switch h.Type {
		case "wallarm_mode":
			// wallarm_rule_mode import: {clientID}/{actionID}/{ruleID}/{mode}
			importID = fmt.Sprintf("%d/%d/%d/%s", h.Clientid, h.ActionID, h.ID, h.Mode)
		case "regex", "experimental_regex":
			// wallarm_rule_regex import: {clientID}/{actionID}/{ruleID}/{regex|experimental_regex}
			importID = fmt.Sprintf("%d/%d/%d/%s", h.Clientid, h.ActionID, h.ID, h.Type)
		}

		rules = append(rules, map[string]interface{}{
			"rule_id":                 h.ID,
			"action_id":              h.ActionID,
			"client_id":              h.Clientid,
			"api_type":               h.Type,
			"terraform_resource_type": tfResourceType,
			"import_id":              importID,
			"attack_type":            h.AttackType,
			"stamp":                  h.Stamp,
			"point":                  string(pointJSON),
			"action":                 string(actionJSON),
			"comment":                h.Comment,
			"parser":                 h.Parser,
			"state":                  h.State,
			"mode":                   h.Mode,
			"active":                 h.Active,
			"enabled":                h.Enabled,
			"updated_at":             h.UpdatedAt,
			"created_at":             h.CreateTime,
		})
	}

	d.SetId(fmt.Sprintf("wallarm_rules_%d_%d", clientID, time.Now().Unix()))
	d.Set("rules", rules)

	return nil
}
