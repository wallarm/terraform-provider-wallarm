package wallarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/terraform-provider-wallarm/wallarm/common/resourcerule"
	"github.com/wallarm/wallarm-go"
)

func dataSourceWallarmRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWallarmRulesRead,

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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Basic rule list with IDs and import info (backward-compatible).",
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

			"rules_export": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Full rule details with reverse-mapped path/domain/etc. for YAML config generation.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_id":            {Type: schema.TypeInt, Computed: true},
						"action_id":          {Type: schema.TypeInt, Computed: true},
						"client_id":          {Type: schema.TypeInt, Computed: true},
						"api_type":           {Type: schema.TypeString, Computed: true, Description: "API rule type"},
						"terraform_resource": {Type: schema.TypeString, Computed: true},
						"import_id":          {Type: schema.TypeString, Computed: true},

						// Reverse-mapped scope
						"path":     {Type: schema.TypeString, Computed: true, Description: "Reverse-mapped path from action conditions"},
						"domain":   {Type: schema.TypeString, Computed: true},
						"instance": {Type: schema.TypeString, Computed: true},
						"method":   {Type: schema.TypeString, Computed: true},
						"scheme":   {Type: schema.TypeString, Computed: true},
						"proto":    {Type: schema.TypeString, Computed: true},

						// Action grouping
						"conditions_hash": {Type: schema.TypeString, Computed: true, Description: "SHA256 hash of action conditions (Ruby-compatible)"},
						"action_dir_name": {Type: schema.TypeString, Computed: true, Description: "Computed directory name for this action scope"},

						// Query & headers as JSON (easier to pass through to HCL)
						"query_json":   {Type: schema.TypeString, Computed: true, Description: "Query params as JSON array"},
						"headers_json": {Type: schema.TypeString, Computed: true, Description: "Custom headers as JSON array"},

						// Action conditions as JSON (for reference HCL)
						"action_json": {Type: schema.TypeString, Computed: true, Description: "Raw action conditions as JSON"},

						// Detection point as JSON
						"point_json": {Type: schema.TypeString, Computed: true, Description: "Detection point as JSON"},

						// Rule-specific fields
						"comment":        {Type: schema.TypeString, Computed: true},
						"attack_type":    {Type: schema.TypeString, Computed: true},
						"stamp":          {Type: schema.TypeInt, Computed: true},
						"mode":           {Type: schema.TypeString, Computed: true},
						"regex":          {Type: schema.TypeString, Computed: true},
						"regex_id":       {Type: schema.TypeInt, Computed: true},
						"experimental":   {Type: schema.TypeBool, Computed: true},
						"parser":         {Type: schema.TypeString, Computed: true},
						"state":          {Type: schema.TypeString, Computed: true},
						"file_type":      {Type: schema.TypeString, Computed: true},
						"delay":          {Type: schema.TypeInt, Computed: true},
						"burst":          {Type: schema.TypeInt, Computed: true},
						"rate":           {Type: schema.TypeInt, Computed: true},
						"rsp_status":     {Type: schema.TypeInt, Computed: true},
						"time_unit":      {Type: schema.TypeString, Computed: true},
						"overlimit_time": {Type: schema.TypeInt, Computed: true},
						"size":           {Type: schema.TypeInt, Computed: true},
						"size_unit":      {Type: schema.TypeString, Computed: true},

						// GraphQL
						"max_depth":           {Type: schema.TypeInt, Computed: true},
						"max_value_size_kb":   {Type: schema.TypeInt, Computed: true},
						"max_doc_size_kb":     {Type: schema.TypeInt, Computed: true},
						"max_aliases_size_kb": {Type: schema.TypeInt, Computed: true},
						"max_doc_per_batch":   {Type: schema.TypeInt, Computed: true},
						"introspection":       {Type: schema.TypeBool, Computed: true},
						"debug_enabled":       {Type: schema.TypeBool, Computed: true},

						// Response header
						"header_name":        {Type: schema.TypeString, Computed: true},
						"header_values_json": {Type: schema.TypeString, Computed: true},

						// Credential stuffing
						"login_point_json": {Type: schema.TypeString, Computed: true},
						"login_regex":      {Type: schema.TypeString, Computed: true},
						"case_sensitive":   {Type: schema.TypeBool, Computed: true},
						"cred_stuff_type":  {Type: schema.TypeString, Computed: true},

						// Threshold/reaction/enum as JSON (complex nested objects)
						"threshold_json":             {Type: schema.TypeString, Computed: true},
						"reaction_json":              {Type: schema.TypeString, Computed: true},
						"enumerated_parameters_json": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceWallarmRulesRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

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
			return diag.FromErr(fmt.Errorf("error reading rules from cache: %w", err))
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
				return diag.FromErr(fmt.Errorf("error reading rules at offset %d: %w", offset, err))
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

	// Fetch credential stuffing configs from the v4 API — they are not
	// returned by HintRead, so we merge them into the common list.
	credConfigs, err := client.CredentialStuffingConfigsRead(clientID)
	if err != nil {
		log.Printf("[WARN] wallarm_rules data source: failed to read credential stuffing configs: %s", err)
	} else {
		allRules = append(allRules, credConfigs...)
		log.Printf("[INFO] wallarm_rules data source: added %d credential stuffing configs for client %d", len(credConfigs), clientID)
	}

	// Filter rules by type.
	var filteredRules []wallarm.ActionBody
	for _, rule := range allRules {
		if _, known := resourcerule.APITypeToTerraformResource[rule.Type]; !known {
			continue
		}
		if len(typeFilter) > 0 && !typeFilter[rule.Type] {
			continue
		}
		filteredRules = append(filteredRules, rule)
	}

	// Build basic rules list (backward-compatible).
	basicRules := make([]interface{}, 0, len(filteredRules))
	for _, rule := range filteredRules {
		var importID string
		if resourcerule.FourPartIDTypes[rule.Type] {
			importID = fmt.Sprintf("%d/%d/%d/%s", clientID, rule.ActionID, rule.ID, rule.Type)
		} else {
			importID = fmt.Sprintf("%d/%d/%d", clientID, rule.ActionID, rule.ID)
		}

		basicRules = append(basicRules, map[string]interface{}{
			"rule_id":            rule.ID,
			"action_id":          rule.ActionID,
			"client_id":          clientID,
			"type":               rule.Type,
			"terraform_resource": resourcerule.APITypeToTerraformResource[rule.Type],
			"import_id":          importID,
		})
	}

	// Build full export list with reverse-mapped fields.
	exported := resourcerule.ExportRules(filteredRules, clientID)
	exportList := make([]interface{}, 0, len(exported))
	for _, e := range exported {
		entry := map[string]interface{}{
			"rule_id":            e.RuleID,
			"action_id":          e.ActionID,
			"client_id":          e.ClientID,
			"api_type":           e.APIType,
			"terraform_resource": e.TerraformResource,
			"import_id":          e.ImportID,

			"path":     e.Path,
			"domain":   e.Domain,
			"instance": e.Instance,
			"method":   e.Method,
			"scheme":   e.Scheme,
			"proto":    e.Proto,

			"conditions_hash": resourcerule.ConditionsHash(e.Action),
			"action_dir_name": resourcerule.ActionDirName(e.Action),

			"comment":        e.Comment,
			"attack_type":    e.AttackType,
			"stamp":          e.Stamp,
			"mode":           e.Mode,
			"regex":          e.Regex,
			"regex_id":       e.RegexID,
			"experimental":   e.Experimental,
			"parser":         e.Parser,
			"state":          e.State,
			"file_type":      e.FileType,
			"delay":          e.Delay,
			"burst":          e.Burst,
			"rate":           e.Rate,
			"rsp_status":     e.RspStatus,
			"time_unit":      e.TimeUnit,
			"overlimit_time": e.OverlimitTime,
			"size":           e.Size,
			"size_unit":      e.SizeUnit,

			"max_depth":           e.MaxDepth,
			"max_value_size_kb":   e.MaxValueSizeKb,
			"max_doc_size_kb":     e.MaxDocSizeKb,
			"max_aliases_size_kb": e.MaxAliasesSizeKb,
			"max_doc_per_batch":   e.MaxDocPerBatch,
			"introspection":       e.Introspection,
			"debug_enabled":       e.DebugEnabled,

			"header_name":     e.HeaderName,
			"login_regex":     e.LoginRegex,
			"case_sensitive":  e.CaseSensitive,
			"cred_stuff_type": e.CredStuffType,
		}

		// Serialize complex fields as JSON strings.
		entry["query_json"] = mustJSON(e.Query)
		entry["headers_json"] = mustJSON(e.Headers)
		// Convert action to TF format (point as map) before serializing.
		tfActions := make([]map[string]interface{}, 0, len(e.Action))
		for _, a := range e.Action {
			m, err := resourcerule.ActionDetailsToMap(a)
			if err != nil {
				continue
			}
			resourcerule.HashResponseActionDetails(m) // side effect: converts point array → map
			tfActions = append(tfActions, m)
		}
		entry["action_json"] = mustJSON(tfActions)
		entry["point_json"] = mustJSON(resourcerule.WrapPointElements(e.Point))
		entry["header_values_json"] = mustJSON(e.HeaderValues)
		entry["login_point_json"] = mustJSON(resourcerule.WrapPointElements(e.LoginPoint))
		entry["threshold_json"] = mustJSON(e.Threshold)
		entry["reaction_json"] = mustJSON(e.Reaction)
		entry["enumerated_parameters_json"] = mustJSON(e.EnumeratedParameters)

		exportList = append(exportList, entry)
	}

	d.SetId(fmt.Sprintf("rules_%d", clientID))
	if err := d.Set("rules", basicRules); err != nil {
		return diag.FromErr(fmt.Errorf("error setting rules: %w", err))
	}
	if err := d.Set("rules_export", exportList); err != nil {
		return diag.FromErr(fmt.Errorf("error setting rules_export: %w", err))
	}

	return nil
}

// mustJSON serializes a value to JSON string. Returns "null" on nil, "[]" on empty slices.
func mustJSON(v interface{}) string {
	if v == nil {
		return "null"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}
