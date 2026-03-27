package wallarm

import (
	"context"
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmSumologic() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmSumologicCreate,
		ReadContext:   resourceWallarmSumologicRead,
		UpdateContext: resourceWallarmSumologicUpdate,
		DeleteContext: resourceWallarmSumologicDelete,

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
				Default:  "Sumologic integration managed by Terraform",
			},

			"sumologic_url": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
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

func resourceWallarmSumologicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	apiToken := d.Get("sumologic_url").(string)
	active := d.Get("active").(bool)
	events := expandWallarmEventToIntEvents(d.Get("event"), "sumo_logic")

	sumoBody := wallarm.IntegrationCreate{
		Name:     name,
		Active:   active,
		Target:   apiToken,
		Clientid: clientID,
		Type:     "sumo_logic",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&sumoBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmSumologicRead(ctx, d, m)
}

func resourceWallarmSumologicRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	sumo, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", sumo.ID)
	d.Set("is_active", sumo.Active)
	d.Set("name", sumo.Name)
	d.Set("created_by", sumo.CreatedBy)
	d.Set("type", sumo.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmSumologicUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	sumo, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
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
			Target: d.Get("sumologic_url").(string),
			Events: expandWallarmEventToIntEvents(d.Get("event"), "sumo_logic"),
			Type:   "sumo_logic",
		}
		updateRes, err := client.IntegrationUpdate(&fullBody, sumo.ID)
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
		if d.HasChange("sumologic_url") {
			updateBody["target"] = d.Get("sumologic_url").(string)
		}
		if len(updateBody) > 0 {
			updateRes, err := client.IntegrationPartialUpdate(sumo.ID, updateBody)
			if err != nil {
				return diag.FromErr(err)
			}
			d.Set("integration_id", updateRes.Body.ID)
			resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
			d.SetId(resID)
		}
	}

	return resourceWallarmSumologicRead(ctx, d, m)
}

func resourceWallarmSumologicDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
