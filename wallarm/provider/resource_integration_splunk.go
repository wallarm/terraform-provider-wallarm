package wallarm

import (
	"context"
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmSplunk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmSplunkCreate,
		ReadContext:   resourceWallarmSplunkRead,
		UpdateContext: resourceWallarmSplunkUpdate,
		DeleteContext: resourceWallarmSplunkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				Default:  "Splunk integration managed by Terraform",
			},

			"api_token": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.IsUUID,
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

func resourceWallarmSplunkCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	apiURL := d.Get("api_url").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events := expandWallarmEventToIntEvents(d.Get("event"), "splunk")

	splunkBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: &wallarm.IntegrationTokenAPITarget{
			Token: apiToken,
			API:   apiURL,
		},
		Clientid: clientID,
		Type:     "splunk",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&splunkBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmSplunkRead(context.TODO(), d, m)
}

func resourceWallarmSplunkRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	splunk, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.Set("integration_id", splunk.ID)
	d.Set("is_active", splunk.Active)
	d.Set("name", splunk.Name)
	d.Set("created_by", splunk.CreatedBy)
	d.Set("type", splunk.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmSplunkUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	splunk, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
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
			Events: expandWallarmEventToIntEvents(d.Get("event"), "splunk"),
			Type:   "splunk",
		}
		updateRes, err := client.IntegrationUpdate(&fullBody, splunk.ID)
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
			updateRes, err := client.IntegrationPartialUpdate(splunk.ID, updateBody)
			if err != nil {
				return diag.FromErr(err)
			}
			d.Set("integration_id", updateRes.Body.ID)
			resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
			d.SetId(resID)
		}
	}

	return resourceWallarmSplunkRead(context.TODO(), d, m)
}

func resourceWallarmSplunkDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
