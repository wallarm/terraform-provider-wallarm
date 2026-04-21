package wallarm

import (
	"context"
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmTelegram() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmTelegramCreate,
		ReadContext:   resourceWallarmTelegramRead,
		UpdateContext: resourceWallarmTelegramUpdate,
		DeleteContext: resourceWallarmTelegramDelete,

		Importer: &schema.ResourceImporter{
			StateContext: importIntegration("telegram"),
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
				Default:  "Telegram integration managed by Terraform",
			},

			"telegram_username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"chat_data": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},

			"event": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
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
								"report_daily",
								"report_weekly",
								"report_monthly",
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

func resourceWallarmTelegramCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	telegramUsername := d.Get("telegram_username").(string)
	chatData := d.Get("chat_data").(string)

	tgBody := wallarm.TelegramIntegrationCreate{
		Name:     telegramUsername,
		Clientid: clientID,
		ChatData: chatData,
	}

	createRes, err := client.TelegramIntegrationCreate(&tgBody)
	if err != nil {
		return diag.FromErr(err)
	}

	integrationID := createRes.Body.ID
	d.Set("integration_id", integrationID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, integrationID)
	d.SetId(resID)

	// After creation, set integration name, events and active state via update
	updateBody := map[string]interface{}{
		"name":   d.Get("name").(string),
		"active": d.Get("active").(bool),
		"events": expandWallarmEventToIntEvents(d.Get("event"), "telegram"),
	}
	if _, err := client.IntegrationPartialUpdate(integrationID, updateBody); err != nil {
		return diag.FromErr(err)
	}

	return resourceWallarmTelegramRead(ctx, d, m)
}

func resourceWallarmTelegramRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	telegram, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", telegram.ID)
	d.Set("is_active", telegram.Active)
	d.Set("name", telegram.Name)
	d.Set("created_by", telegram.CreatedBy)
	d.Set("type", telegram.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmTelegramUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	telegram, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	updateBody := make(map[string]interface{})
	if d.HasChange("name") {
		updateBody["name"] = d.Get("name").(string)
	}
	if d.HasChange("active") {
		updateBody["active"] = d.Get("active").(bool)
	}
	if d.HasChange("event") {
		updateBody["events"] = expandWallarmEventToIntEvents(d.Get("event"), "telegram")
	}
	if len(updateBody) > 0 {
		updateRes, err := client.IntegrationPartialUpdate(telegram.ID, updateBody)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("integration_id", updateRes.Body.ID)
		resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
		d.SetId(resID)
	}

	return resourceWallarmTelegramRead(ctx, d, m)
}

func resourceWallarmTelegramDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
