package wallarm

import (
	"fmt"
	"log"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// apiTypeToTerraformResource maps Wallarm API rule types to Terraform resource names.
var apiTypeToTerraformResource = map[string]string{
	"binary_data":            "wallarm_rule_binary_data",
	"bola":                   "wallarm_rule_bola",
	"bola_counter":           "wallarm_rule_bola_counter",
	"brute":                  "wallarm_rule_brute",
	"brute_counter":          "wallarm_rule_bruteforce_counter",
	"credentials_point":      "wallarm_rule_credential_stuffing_point",
	"credentials_regex":      "wallarm_rule_credential_stuffing_regex",
	"dirbust_counter":        "wallarm_rule_dirbust_counter",
	"disable_attack_type":    "wallarm_rule_disable_attack_type",
	"disable_regex":          "wallarm_rule_ignore_regex",
	"disable_stamp":          "wallarm_rule_disable_stamp",
	"enum":                   "wallarm_rule_enum",
	"experimental_regex":     "wallarm_rule_regex",
	"file_upload_size_limit": "wallarm_rule_file_upload_size_limit",
	"forced_browsing":        "wallarm_rule_forced_browsing",
	"graphql_detection":      "wallarm_rule_graphql_detection",
	"overlimit_res_settings": "wallarm_rule_overlimit_res_settings",
	"parser_state":           "wallarm_rule_parser_state",
	"rate_limit":             "wallarm_rule_rate_limit",
	"rate_limit_enum":        "wallarm_rule_rate_limit_enum",
	"regex":                  "wallarm_rule_regex",
	"sensitive_data":         "wallarm_rule_masking",
	"set_response_header":    "wallarm_rule_set_response_header",
	"uploads":                "wallarm_rule_uploads",
	"vpatch":                 "wallarm_rule_vpatch",
	"wallarm_mode":           "wallarm_rule_mode",
}

// fourPartIDTypes are rule types whose import ID requires a 4th segment (the API type).
var fourPartIDTypes = map[string]bool{
	"regex":              true,
	"experimental_regex": true,
	"wallarm_mode":       true,
}

func dataSourceWallarmRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWallarmRulesRead,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"type": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Filter by API rule type(s), e.g. [\"rate_limit\", \"wallarm_mode\"]. Returns all types if omitted.",
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
							Type:        schema.TypeString,
							Computed:    true,
							Description: "API rule type, e.g. rate_limit, wallarm_mode",
						},
						"terraform_resource": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Terraform resource name, e.g. wallarm_rule_rate_limit",
						},
						"import_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Pre-computed import ID for use in import blocks",
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

	// Build optional type filter set.
	typeFilter := make(map[string]bool)
	if v, ok := d.GetOk("type"); ok {
		for _, t := range v.([]interface{}) {
			typeFilter[t.(string)] = true
		}
	}

	// Fetch all non-system rules, using the hint cache when available.
	var allRules []wallarm.ActionBody

	if cached, ok := client.(*CachedClient); ok {
		rules, err := cached.AllRules(clientID)
		if err != nil {
			return fmt.Errorf("error reading rules from cache: %w", err)
		}
		allRules = rules
		log.Printf("[INFO] wallarm_rules data source: got %d rules from cache for client %d", len(allRules), clientID)
	} else {
		// Fallback: paginate directly when caching is disabled.
		const batchSize = 200
		const maxPages = 500
		systemFalse := false

		for page, offset := 0, 0; page < maxPages; page++ {
			resp, err := client.HintRead(&wallarm.HintRead{
				Limit:     batchSize,
				Offset:    offset,
				OrderBy:   "id",
				OrderDesc: true,
				Filter: &wallarm.HintFilter{
					Clientid: []int{clientID},
					System:   &systemFalse,
				},
			})
			if err != nil {
				return fmt.Errorf("error reading rules at offset %d: %w", offset, err)
			}

			if resp.Body == nil || len(*resp.Body) == 0 {
				break
			}

			batch := *resp.Body
			allRules = append(allRules, batch...)

			if len(batch) < batchSize {
				break
			}
			offset += batchSize
		}

		log.Printf("[INFO] wallarm_rules data source: fetched %d rules for client %d (direct API)", len(allRules), clientID)
	}

	// Flatten results.
	rules := make([]interface{}, 0)
	for _, rule := range allRules {
		tfResource, known := apiTypeToTerraformResource[rule.Type]
		if !known {
			continue // skip types not managed by this provider
		}

		// Apply optional type filter.
		if len(typeFilter) > 0 && !typeFilter[rule.Type] {
			continue
		}

		// Compute import ID: 4-part for regex/mode, 3-part for all others.
		var importID string
		if fourPartIDTypes[rule.Type] {
			importID = fmt.Sprintf("%d/%d/%d/%s", clientID, rule.ActionID, rule.ID, rule.Type)
		} else {
			importID = fmt.Sprintf("%d/%d/%d", clientID, rule.ActionID, rule.ID)
		}

		rules = append(rules, map[string]interface{}{
			"rule_id":            rule.ID,
			"action_id":          rule.ActionID,
			"client_id":          clientID,
			"type":               rule.Type,
			"terraform_resource": tfResource,
			"import_id":          importID,
		})
	}

	d.SetId(fmt.Sprintf("rules_%d", clientID))
	if err := d.Set("rules", rules); err != nil {
		return fmt.Errorf("error setting rules: %w", err)
	}

	return nil
}
