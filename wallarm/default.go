package wallarm

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	Minutes           = "Minutes"
	header            = "header"
	path              = "path"
	experimentalRegex = "experimental_regex"
	iequal            = "iequal"
)

var (
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
					Elem: &schema.Resource{
						Schema: defaultResourceRuleActionPointElemSchemaMap,
					},
				},
			},
		},
	}

	defaultResourceRuleActionPointElemSchemaMap = map[string]*schema.Schema{
		"header": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"method": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			ValidateFunc: validation.StringInSlice([]string{"GET", "HEAD", "POST",
				"PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}, false),
		},

		"path": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(int)
				if v < 0 || v > 60 {
					errs = append(errs, fmt.Errorf("%q must be between 0 and 60 inclusive, got: %d", key, v))
				}
				return
			},
		},

		"action_name": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},

		"action_ext": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},

		"query": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},

		"proto": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"1.0", "1.1", "2.0", "3.0"}, false),
		},

		"scheme": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"http", "https"}, true),
		},

		"uri": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},

		"instance": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(int)
				if v < -1 {
					errs = append(errs, fmt.Errorf("%q must be be greater than -1 inclusive, got: %d", key, v))
				}
				return
			},
		},
	}

	defaultResourceLimitActionSchema = &schema.Schema{
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
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"header": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"method": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								ValidateFunc: validation.StringInSlice([]string{"GET", "HEAD", "POST",
									"PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}, false),
							},

							"path": {
								Type:     schema.TypeInt,
								Optional: true,
								ForceNew: true,
								ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
									v := val.(int)
									if v < 0 || v > 60 {
										errs = append(errs, fmt.Errorf("%q must be between 0 and 60 inclusive, got: %d", key, v))
									}
									return
								},
							},

							"action_name": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},

							"action_ext": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},

							"query": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},

							"proto": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},

							"scheme": {
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								Computed:     true,
								ValidateFunc: validation.StringInSlice([]string{"http", "https"}, true),
							},

							"uri": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},

							"instance": {
								Type:     schema.TypeInt,
								Optional: true,
								ForceNew: true,
								ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
									v := val.(int)
									if v < -1 {
										errs = append(errs, fmt.Errorf("%q must be greater than -1 inclusive, got: %d", key, v))
									}
									return
								},
							},
						},
					},
				},
			},
		},
	}

	commonResourceRuleFields = map[string]*schema.Schema{
		"rule_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"action_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"rule_type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"client_id": defaultClientIDWithValidationSchema,
		"comment": {
			Type:     schema.TypeString,
			Default:  "Managed by Terraform",
			Optional: true,
		},
		"set": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"active": {
			Type:     schema.TypeBool,
			Default:  true,
			Optional: true,
			ForceNew: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"mitigation": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"variativity_disabled": {
			Type:     schema.TypeBool,
			Default:  true,
			Optional: true,
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
					Default:  false,
					ForceNew: true,
				},
				"plain_parameters": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
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
				"point": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeList,
						Elem: &schema.Schema{Type: schema.TypeString},
					},
					ForceNew: true,
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
)

type CommonResourceRuleFieldsDTO struct {
	Comment    string
	Set        string
	Active     bool
	Title      string
	Mitigation string
}

func getCommonResourceRuleFieldsDTOFromResourceData(d *schema.ResourceData) CommonResourceRuleFieldsDTO {
	if d == nil {
		return CommonResourceRuleFieldsDTO{}
	}
	comment, _ := d.Get("comment").(string)
	set, _ := d.Get("set").(string)
	active, _ := d.Get("active").(bool)
	title, _ := d.Get("title").(string)
	mitigation, _ := d.Get("mitigation").(string)
	return CommonResourceRuleFieldsDTO{
		Comment:    comment,
		Set:        set,
		Active:     active,
		Title:      title,
		Mitigation: mitigation,
	}
}
