package wallarm

import (
	"context"
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmInsightConnect() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmInsightConnectCreate,
		ReadContext:   resourceWallarmInsightConnectRead,
		UpdateContext: resourceWallarmInsightConnectUpdate,
		DeleteContext: resourceWallarmInsightConnectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: importIntegration("insight_connect"),
		},

		CustomizeDiff: validateWithHeadersOnlySiem(),

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"integration_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"is_active": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "InsightConnect integration managed by Terraform",
			},

			"api_token": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"api_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Sensitive:    true,
			},

			"event": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 9,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"siem",
								"rules_and_triggers",
								"number_of_requests_per_hour",
								"security_issue_critical",
								"security_issue_high",
								"security_issue_medium",
								"security_issue_low",
								"security_issue_info",
								"system",
							}, false),
						},
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"with_headers": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Send requests with headers. Only applicable to the 'siem' event type.",
						},
					},
				},
			},
		},
	}
}

func resourceWallarmInsightConnectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	apiURL := d.Get("api_url").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events := expandWallarmEventToIntEvents(d.Get("event"), "insight_connect")

	insightBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: &wallarm.IntegrationTokenAPITarget{
			Token: apiToken,
			API:   apiURL,
		},
		Clientid: clientID,
		Type:     "insight_connect",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&insightBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmInsightConnectRead(ctx, d, m)
}

func resourceWallarmInsightConnectRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	insight, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", insight.ID)
	d.Set("is_active", insight.Active)
	d.Set("name", insight.Name)
	d.Set("created_by", insight.CreatedBy)
	d.Set("type", insight.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmInsightConnectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	insight, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if d.HasChange("event") {
		// When events change, API requires the full configuration
		fullBody := wallarm.IntegrationCreate{
			Name:   d.Get("name").(string),
			Active: d.Get("active").(bool),
			Target: &wallarm.IntegrationTokenAPITarget{
				Token: d.Get("api_token").(string),
				API:   d.Get("api_url").(string),
			},
			Events: expandWallarmEventToIntEvents(d.Get("event"), "insight_connect"),
			Type:   "insight_connect",
		}
		updateRes, err := client.IntegrationUpdate(&fullBody, insight.ID)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("integration_id", updateRes.Body.ID)
		resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
		d.SetId(resID)
	} else {
		updateBody := make(map[string]interface{})
		if d.HasChange("name") {
			updateBody["name"] = d.Get("name").(string)
		}
		if d.HasChange("active") {
			updateBody["active"] = d.Get("active").(bool)
		}
		if d.HasChange("api_token") || d.HasChange("api_url") {
			updateBody["target"] = &wallarm.IntegrationTokenAPITarget{
				Token: d.Get("api_token").(string),
				API:   d.Get("api_url").(string),
			}
		}
		if len(updateBody) > 0 {
			updateRes, err := client.IntegrationPartialUpdate(insight.ID, updateBody)
			if err != nil {
				return diag.FromErr(err)
			}
			d.Set("integration_id", updateRes.Body.ID)
			resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
			d.SetId(resID)
		}
	}

	return resourceWallarmInsightConnectRead(ctx, d, m)
}

func resourceWallarmInsightConnectDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
