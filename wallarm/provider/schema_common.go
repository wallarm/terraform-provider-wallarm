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
		// Optional only (NOT Computed). Optional+Computed on a TypeString
		// has a real bug: SDKv2 normalises an explicit empty string in HCL
		// to cty.NullVal, after which Computed semantics preserve state and
		// `set = ""` silently fails to clear the value. The cost of dropping
		// Computed: post-import-CLI workflows without -generate-config-out
		// see HCL-omitted as "" rather than the API-echoed value — but the
		// modern import{}+generate-config-out flow generates HCL with the
		// value populated, so there's no real-world hit. (Test:
		// TestAccRuleParserState_UpdateSetToEmpty.)
		"set": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The rule set name. Used to group related rules together.",
		},
		"active": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether the rule is active. Defaults to true.",
		},
		// See `set` above for why this is Optional only (not Optional+Computed):
		// SDKv2 string-normalisation breaks `title = ""` clears with Computed.
		"title": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A short title for the rule.",
		},
		"mitigation": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Read-only mitigation type assigned by the API.",
		},
		"variativity_disabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
			Description: "Whether variativity is disabled for this rule. Defaults to true. " +
				"Provider locks this to true on Create regardless of user input — " +
				"keeps Terraform state synchronized with the API by preventing " +
				"server-side variative-rule mutations from drifting state. " +
				"The API default varies by rule type, but the provider always sends true.",
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

	// Both fields Required; API enforces count >= 1 (HTTP 400
	// "must be greater than or equal to 1") and a 0-second period is
	// nonsensical. IntAtLeast(1) on each surfaces the constraint at plan time
	// instead of as an HTTP 400 on apply.
	thresholdSchema = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"period": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},
				"count": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},
			},
		},
	}

	// Reaction values are session-/IP-block durations in seconds. API range is
	// 600..315569520 (10 minutes to 10 years). The mode↔reaction whitelist
	// (block_by_session/block_by_ip for mode=block, graylist_by_ip for
	// mode=monitoring) is API-enforced — keep that distinction at runtime.
	//
	// Validator allows 0 alongside the valid range because SDKv2's legacy
	// flat state model has no NullVal slot for TypeInt inside a nested Resource
	// block: when the API omits a reaction key, state still materialises it as
	// 0, and `terraform import` + -generate-config-out then emits literal `= 0`
	// lines that a strict IntBetween validator would reject at plan time. The
	// mapper drops 0 on the wire (mapper_tftoapi.go), so 0-in-HCL means "unset"
	// — round-trip safe.
	reactionRangeOrZero = validation.Any(
		validation.IntInSlice([]int{0}),
		validation.IntBetween(600, 315569520),
	)
	reactionSchema = &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"block_by_session": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: reactionRangeOrZero,
				},
				"block_by_ip": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: reactionRangeOrZero,
				},
				"graylist_by_ip": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: reactionRangeOrZero,
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
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"value_regexps": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				// Optional+Default:false. Removing either line from HCL plans
				// `current → false` (symmetric: adding `= true` plans
				// `false → true`). For legacy `terraform import` (CLI command
				// without `-generate-config-out`), users must explicitly set
				// the value in HCL to match the API-echoed value, otherwise
				// applying without HCL would silently overwrite to false.
				// The validator (EnumeratedParamsCustomizeDiff) rejects
				// `=true` in exact mode regardless of source.
				"additional_parameters": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"plain_parameters": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
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
	active, _ := d.Get("active").(bool)
	return CommonResourceRuleFieldsDTO{
		Comment: comment,
		Set:     set,
		Active:  active,
		Title:   title,
	}
}
