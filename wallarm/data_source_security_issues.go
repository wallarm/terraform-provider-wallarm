package wallarm

import (
	"fmt"
	"time"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceWallarmSecurityIssues() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWallarmSecurityIssuesRead,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1000,
				ValidateFunc: validation.IntBetween(0, 1000),
			},
			"offset": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"unlimited": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"issues": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"client_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"severity": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"volume": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"discovered_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"discovered_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"discovered_by_display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parameter_display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parameter_position": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parameter_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"aasm_template": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mitigations": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vpatch": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rule_id": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"issue_type": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"owasp": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"full_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"link": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"tags": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"slug": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"verified": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWallarmSecurityIssuesRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)

	req := &wallarm.GetSecurityIssuesRead{
		ClientID:  d.Get("client_id").(int),
		Limit:     GetValueWithTypeCastingOrOverridedDefault[int](d, "limit", 1000),
		Offset:    GetValueWithTypeCastingOrOverridedDefault[int](d, "offset", 0),
		Unlimited: GetValueWithTypeCastingOrOverridedDefault[bool](d, "unlimited", false),
	}

	resp, err := client.GetSecurityIssuesRead(req)
	if err != nil {
		return err
	}

	issues := make([]interface{}, 0)
	for _, v := range resp {
		issue := map[string]interface{}{
			"id":                         v.Id,
			"client_id":                  v.ClientId,
			"severity":                   v.Severity,
			"state":                      v.State,
			"volume":                     v.Volume,
			"name":                       v.Name,
			"created_at":                 v.CreatedAt,
			"discovered_at":              v.DiscoveredAt,
			"discovered_by":              v.DiscoveredBy,
			"discovered_by_display_name": v.DiscoveredByDisplayName,
			"url":                        v.Url,
			"host":                       v.Host,
			"path":                       v.Path,
			"parameter_display_name":     v.ParameterDisplayName,
			"parameter_position":         v.ParameterPosition,
			"parameter_name":             v.ParameterName,
			"http_method":                v.HttpMethod,
			"aasm_template":              v.AasmTemplate,
			"verified":                   v.Verified,
		}

		// Mitigations
		mitigations := map[string]interface{}{
			"vpatch": []interface{}{
				map[string]interface{}{
					"rule_id": v.Mitigations.Vpatch.RuleId,
				},
			},
		}
		issue["mitigations"] = []interface{}{mitigations}

		// Issue Type
		issueType := map[string]interface{}{
			"id":   v.IssueType.Id,
			"name": v.IssueType.Name,
		}
		issue["issue_type"] = []interface{}{issueType}

		// OWASP
		owasp := make([]interface{}, 0, len(v.Owasp))
		for _, o := range v.Owasp {
			owasp = append(owasp, map[string]interface{}{
				"id":        o.Id,
				"name":      o.Name,
				"full_name": o.FullName,
				"link":      o.Link,
			})
		}
		issue["owasp"] = owasp

		// Tags
		tags := make([]interface{}, 0, len(v.Tags))
		for _, t := range v.Tags {
			tags = append(tags, map[string]interface{}{
				"id":   t.Id,
				"name": t.Name,
				"slug": t.Slug,
			})
		}
		issue["tags"] = tags

		issues = append(issues, issue)
	}

	d.SetId(fmt.Sprintf("Issues_%s", time.Now().UTC().String()))

	if err = d.Set("issues", issues); err != nil {
		return fmt.Errorf("error setting Issues: %s", err)
	}
	return nil
}
