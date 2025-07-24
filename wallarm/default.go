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
			Optional: true,
			ForceNew: true,
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
