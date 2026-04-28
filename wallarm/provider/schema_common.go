package wallarm

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	experimentalRegex = "experimental_regex"
)

var (
	// defaultPointSchema is the standard required point schema used across rule resources.
	// The point field is a list of lists of strings, representing a 2D point structure
	// (e.g., [["get", "query"], ["header", "HOST"]]).
	defaultPointSchema = &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		ForceNew: true,
		Elem: &schema.Schema{
			Type: schema.TypeList,
			Elem: &schema.Schema{Type: schema.TypeString},
		},
	}

	defaultClientIDWithValidationSchema = &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		Description:  "The Client ID to perform changes",
		ValidateFunc: validation.IntAtLeast(1),
	}

	commonResourceRuleFields = map[string]*schema.Schema{
		"rule_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The numeric ID of the rule (hint) in the Wallarm Cloud.",
		},
		"action_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the action (rule branch) this rule belongs to.",
		},
		"rule_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The type identifier of the rule (e.g. wallarm_mode, brute, bola).",
		},
		"client_id": defaultClientIDWithValidationSchema,
		"comment": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Managed by Terraform",
			Description: "A human-readable comment for the rule.",
		},
		"set": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The rule set name. Used to group related rules together.",
		},
		"active": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Whether the rule is active.",
		},
		"title": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "A short title for the rule.",
		},
		"mitigation": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Read-only mitigation type assigned by the API. Accepted in config but not sent to the API.",
		},
		"variativity_disabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether variativity is disabled for this rule. Defaults to true.",
		},
	}

	// counterFieldOverrides makes the user-mutable common fields read-only
	// for counter resources (bola_counter, bruteforce_counter, dirbust_counter).
	// Counters have no UpdateContext (state-only Delete, no Update path), so
	// every common field that v2.3.7 made mutable must be overridden to
	// Computed-only here — otherwise Terraform plans an update-in-place and
	// the SDK invokes a nil UpdateContext at apply time.
	// Merge after commonResourceRuleFields via lo.Assign to override.
	counterFieldOverrides = map[string]*schema.Schema{
		"comment":              {Type: schema.TypeString, Computed: true},
		"variativity_disabled": {Type: schema.TypeBool, Computed: true},
		"title":                {Type: schema.TypeString, Computed: true},
		"active":               {Type: schema.TypeBool, Computed: true},
		"set":                  {Type: schema.TypeString, Computed: true},
	}

	thresholdSchema = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"period": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"count": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		},
	}

	reactionSchema = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"block_by_session": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"block_by_ip": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"graylist_by_ip": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
			},
		},
	}

	enumeratedParametersSchema = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"mode": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"regexp", "exact"}, false),
				},
				"points": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"point": {
								Type:     schema.TypeList,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"sensitive": {
								Type:     schema.TypeBool,
								Default:  false,
								Optional: true,
							},
						},
					},
				},
				"name_regexps": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"value_regexps": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"additional_parameters": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
				"plain_parameters": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
			},
		},
	}

	advancedConditionsSchema = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"ip", "method", "user_agent", "domain", "uri", "status_code", "request_time", "request_size", "response_size", "attack_type", "blocked"}, false),
					ForceNew:     true,
				},
				"value": {
					Type:     schema.TypeList,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Required: true,
					ForceNew: true,
				},
				"operator": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"eq", "ne", "imatch", "notimatch", "match", "notmatch", "lt", "gt", "le", "ge"}, false),
					ForceNew:     true,
				},
			},
		},
	}

	arbitraryConditionsSchema = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"point": defaultPointSchema,
				"value": {
					Type:     schema.TypeList,
					Elem:     &schema.Schema{Type: schema.TypeString},
					Required: true,
					ForceNew: true,
				},
				"operator": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"eq", "ne", "imatch", "notimatch", "match", "notmatch", "lt", "gt", "le", "ge"}, false),
					ForceNew:     true,
				},
			},
		},
	}
)

type CommonResourceRuleFieldsDTO struct {
	Comment string
	Set     string
	Active  bool
	Title   string
}

func getCommonResourceRuleFieldsDTOFromResourceData(d *schema.ResourceData) CommonResourceRuleFieldsDTO {
	if d == nil {
		return CommonResourceRuleFieldsDTO{}
	}
	comment, _ := d.Get("comment").(string)
	set, _ := d.Get("set").(string)
	title, _ := d.Get("title").(string)

	// Default to true when not explicitly set (replaced schema Default which can't coexist with Computed).
	active := true
	if v, ok := d.GetOkExists("active"); ok { //nolint:staticcheck
		active = v.(bool)
	}

	return CommonResourceRuleFieldsDTO{
		Comment: comment,
		Set:     set,
		Active:  active,
		Title:   title,
	}
}
