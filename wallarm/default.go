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
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		Description: "The Client ID to perform changes",
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			v := val.(int)
			if v <= 0 {
				errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
			}
			return
		},
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
)
