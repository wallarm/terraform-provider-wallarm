package wallarm

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	Minutes           = "Minutes"
	header            = "header"
	path              = "path"
	experimentalRegex = "experimental_regex"
	iequal            = "iequal"

	// DefaultAPIListLimit is the default limit for API list/read requests.
	DefaultAPIListLimit = 500
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

	defaultResourceRuleActionSchema = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice([]string{"equal", "iequal", "regex", "absent"}, false),
					ForceNew:     true,
				},

				"value": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Computed: true,
				},

				"point": {
					Type:     schema.TypeMap,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
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
			Default:     "Managed by Terraform",
			Optional:    true,
			Description: "A human-readable comment for the rule.",
		},
		"set": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
			Description: "The rule set name. Used to group related rules together.",
		},
		"active": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			Description: "Whether the rule is active.",
		},
		"title": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
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
			Computed:    true,
			Description: "Whether variativity is disabled. Always set to true by the provider.",
		},
	}

	thresholdSchema = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		ForceNew: true,
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
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"block_by_session": {
					Type:     schema.TypeInt,
					Optional: true,
					ForceNew: true,
				},
				"block_by_ip": {
					Type:     schema.TypeInt,
					Optional: true,
					ForceNew: true,
				},
				"graylist_by_ip": {
					Type:     schema.TypeInt,
					Optional: true,
					ForceNew: true,
				},
			},
		},
	}

	enumeratedParametersSchema = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"mode": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
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
								ForceNew: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"sensitive": {
								Type:     schema.TypeBool,
								Default:  false,
								Optional: true,
								ForceNew: true,
							},
						},
					},
				},
				"name_regexps": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
					ForceNew: true,
				},
				"value_regexps": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
					ForceNew: true,
				},
				"additional_parameters": {
					Type:     schema.TypeBool,
					Optional: true,
					ForceNew: true,
				},
				"plain_parameters": {
					Type:     schema.TypeBool,
					Optional: true,
					ForceNew: true,
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
