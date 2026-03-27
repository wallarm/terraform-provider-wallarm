package wallarm

import (
	"context"
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmEmail() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmEmailCreate,
		ReadContext:   resourceWallarmEmailRead,
		UpdateContext: resourceWallarmEmailUpdate,
		DeleteContext: resourceWallarmEmailDelete,

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
				Default:  "Email integration managed by Terraform",
			},

			"emails": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"event": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 7,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{"system", "aasm_report",
								"api_discovery_hourly_changes_report", "api_discovery_daily_changes_report", "report_daily", "report_weekly", "report_monthly"}, false),
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

func resourceWallarmEmailCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	active := d.Get("active").(bool)
	emails := expandInterfaceToStringList(d.Get("emails"))
	events := expandWallarmEventToIntEvents(d.Get("event"), "email")

	emailBody := wallarm.EmailIntegrationCreate{
		Name:     name,
		Active:   active,
		Target:   emails,
		Clientid: clientID,
		Type:     "email",
		Events:   events,
	}

	createRes, err := client.EmailIntegrationCreate(&emailBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmEmailRead(ctx, d, m)
}

func resourceWallarmEmailRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	email, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", email.ID)
	d.Set("is_active", email.Active)
	d.Set("name", email.Name)
	d.Set("created_by", email.CreatedBy)
	d.Set("type", email.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmEmailUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	active := d.Get("active").(bool)
	emails := expandInterfaceToStringList(d.Get("emails"))
	events := expandWallarmEventToIntEvents(d.Get("event"), "email")

	email, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var updateRes *wallarm.IntegrationCreateResp

	if d.HasChange("active") {
		emailBody := wallarm.EmailIntegrationCreate{
			Active: active,
		}

		updateRes, err = client.EmailIntegrationUpdate(&emailBody, email.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") || d.HasChange("emails") || d.HasChange("event") {
		emailBody := wallarm.EmailIntegrationCreate{
			Name:   name,
			Active: active,
			Target: emails,
			Type:   "email",
			Events: events,
		}

		updateRes, err = client.EmailIntegrationUpdate(&emailBody, email.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if updateRes == nil {
		return nil
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmEmailRead(ctx, d, m)
}

func resourceWallarmEmailDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
