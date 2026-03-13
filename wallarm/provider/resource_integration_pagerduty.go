package wallarm

import (
	"context"
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmPagerDuty() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmPagerDutyCreate,
		ReadContext:   resourceWallarmPagerDutyRead,
		UpdateContext: resourceWallarmPagerDutyUpdate,
		DeleteContext: resourceWallarmPagerDutyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
				Default:  "PagerDuty integration managed by Terraform",
			},

			"integration_key": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(32, 32),
			},

			"event": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 7,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"system",
								"rules_and_triggers",
								"security_issue_critical",
								"security_issue_high",
								"security_issue_medium",
								"security_issue_low",
								"security_issue_info",
							}, false),
						},
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
		},
	}
}

func resourceWallarmPagerDutyCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	apiToken := d.Get("integration_key").(string)
	active := d.Get("active").(bool)
	events := expandWallarmEventToIntEvents(d.Get("event"), "pager_duty")

	pagerdutyBody := wallarm.IntegrationCreate{
		Name:     name,
		Active:   active,
		Target:   apiToken,
		Clientid: clientID,
		Type:     "pager_duty",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&pagerdutyBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmPagerDutyRead(context.TODO(), d, m)
}

func resourceWallarmPagerDutyRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	pagerduty, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", pagerduty.ID)
	d.Set("is_active", pagerduty.Active)
	d.Set("name", pagerduty.Name)
	d.Set("created_by", pagerduty.CreatedBy)
	d.Set("type", pagerduty.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmPagerDutyUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	pagerDuty, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
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
			Target: d.Get("integration_key").(string),
			Events: expandWallarmEventToIntEvents(d.Get("event"), "pager_duty"),
			Type:   "pager_duty",
		}
		updateRes, err := client.IntegrationUpdate(&fullBody, pagerDuty.ID)
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
		if d.HasChange("integration_key") {
			updateBody["target"] = d.Get("integration_key").(string)
		}
		if len(updateBody) > 0 {
			updateRes, err := client.IntegrationPartialUpdate(pagerDuty.ID, updateBody)
			if err != nil {
				return diag.FromErr(err)
			}
			d.Set("integration_id", updateRes.Body.ID)
			resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
			d.SetId(resID)
		}
	}

	return resourceWallarmPagerDutyRead(context.TODO(), d, m)
}

func resourceWallarmPagerDutyDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
