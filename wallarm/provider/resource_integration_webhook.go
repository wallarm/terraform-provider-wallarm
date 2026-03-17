package wallarm

import (
	"context"
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmWebhookCreate,
		ReadContext:   resourceWallarmWebhookRead,
		UpdateContext: resourceWallarmWebhookUpdate,
		DeleteContext: resourceWallarmWebhookDelete,

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
				Default:  "Webhook integration managed by Terraform",
			},

			"http_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"PUT", "POST"}, false),
				Default:      "POST",
			},

			"format": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"json", "jsonl"}, false),
				Default:      "json",
			},

			"webhook_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
				Sensitive:    true,
			},

			"ca_file": {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},

			"ca_verify": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  15,
			},

			"open_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  20,
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

			"headers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceWallarmWebhookCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	active := d.Get("active").(bool)
	webhookURL := d.Get("webhook_url").(string)
	method := d.Get("http_method").(string)
	caFile := d.Get("ca_file").(string)
	caVerify := d.Get("ca_verify").(bool)
	timeout := d.Get("timeout").(int)
	openTimeout := d.Get("open_timeout").(int)
	headers := d.Get("headers").(map[string]interface{})
	format := d.Get("format").(string)

	events := expandWallarmEventToIntEvents(d.Get("event"), "web_hooks")

	webhookBody := wallarm.IntegrationWithAPICreate{
		Name:   name,
		Active: active,
		Target: &wallarm.IntegrationWithAPITarget{
			URL:         webhookURL,
			HTTPMethod:  method,
			Timeout:     timeout,
			OpenTimeout: openTimeout,
			CaFile:      caFile,
			CaVerify:    caVerify,
			Headers:     headers,
			Format:      format,
			Subtype:     "web_hooks",
		},
		Clientid: clientID,
		Type:     "web_hooks",
		Events:   events,
	}

	createRes, err := client.IntegrationWithAPICreate(&webhookBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmWebhookRead(context.TODO(), d, m)
}

func resourceWallarmWebhookRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	webhook, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", webhook.ID)
	d.Set("is_active", webhook.Active)
	d.Set("name", webhook.Name)
	d.Set("created_by", webhook.CreatedBy)
	d.Set("type", webhook.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmWebhookUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	webhook, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if d.HasChange("event") {
		// When events change, API requires the full configuration
		fullBody := wallarm.IntegrationWithAPICreate{
			Name:   d.Get("name").(string),
			Active: d.Get("active").(bool),
			Target: &wallarm.IntegrationWithAPITarget{
				URL:         d.Get("webhook_url").(string),
				HTTPMethod:  d.Get("http_method").(string),
				Timeout:     d.Get("timeout").(int),
				OpenTimeout: d.Get("open_timeout").(int),
				CaFile:      d.Get("ca_file").(string),
				CaVerify:    d.Get("ca_verify").(bool),
				Headers:     d.Get("headers").(map[string]interface{}),
				Format:      d.Get("format").(string),
				Subtype:     "web_hooks",
			},
			Type:   "web_hooks",
			Events: expandWallarmEventToIntEvents(d.Get("event"), "web_hooks"),
		}
		updateRes, err := client.IntegrationWithAPIUpdate(&fullBody, webhook.ID)
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
		if d.HasChanges("webhook_url", "http_method", "ca_file", "ca_verify", "timeout", "open_timeout", "headers", "format") {
			updateBody["target"] = &wallarm.IntegrationWithAPITarget{
				URL:         d.Get("webhook_url").(string),
				HTTPMethod:  d.Get("http_method").(string),
				Timeout:     d.Get("timeout").(int),
				OpenTimeout: d.Get("open_timeout").(int),
				CaFile:      d.Get("ca_file").(string),
				CaVerify:    d.Get("ca_verify").(bool),
				Headers:     d.Get("headers").(map[string]interface{}),
				Format:      d.Get("format").(string),
				Subtype:     "web_hooks",
			}
		}
		if len(updateBody) > 0 {
			updateRes, err := client.IntegrationPartialUpdate(webhook.ID, updateBody)
			if err != nil {
				return diag.FromErr(err)
			}
			d.Set("integration_id", updateRes.Body.ID)
			resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
			d.SetId(resID)
		}
	}

	return resourceWallarmWebhookRead(context.TODO(), d, m)
}

func resourceWallarmWebhookDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
