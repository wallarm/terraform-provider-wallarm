package wallarm

import (
	"fmt"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmWebhookCreate,
		Read:   resourceWallarmWebhookRead,
		Update: resourceWallarmWebhookUpdate,
		Delete: resourceWallarmWebhookDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
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
			},

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
				MaxItems: 6,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"hit", "vuln_high", "vuln_medium", "vuln_low", "vuln_low", "system", "scope"}, false),
						},
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
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

func resourceWallarmWebhookCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
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

	events, err := expandWallarmEventToIntEvents(d.Get("event"), "web_hooks")
	if err != nil {
		d.SetId("")
		return err
	}

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
		},
		Clientid: clientID,
		Type:     "web_hooks",
		Events:   events,
	}

	createRes, err := client.IntegrationWithAPICreate(&webhookBody)
	if err != nil {
		return err
	}

	if err = d.Set("integration_id", createRes.Body.ID); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmWebhookRead(d, m)
}

func resourceWallarmWebhookRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	webhook, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not found.") {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	if err = d.Set("integration_id", webhook.ID); err != nil {
		return err
	}
	if err = d.Set("is_active", webhook.Active); err != nil {
		return err
	}
	if err = d.Set("name", webhook.Name); err != nil {
		return err
	}
	if err = d.Set("created_by", webhook.CreatedBy); err != nil {
		return err
	}
	if err = d.Set("type", webhook.Type); err != nil {
		return err
	}
	if err = d.Set("client_id", clientID); err != nil {
		return err
	}

	return nil
}

func resourceWallarmWebhookUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
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

	events, err := expandWallarmEventToIntEvents(d.Get("event"), "web_hooks")
	if err != nil {
		return err
	}

	webhook, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not found.") {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

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
		},
		Type:   "web_hooks",
		Events: events,
	}

	updateRes, err := client.IntegrationWithAPIUpdate(&webhookBody, webhook.ID)
	if err != nil {
		return err
	}

	if err = d.Set("integration_id", updateRes.Body.ID); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmWebhookRead(d, m)
}

func resourceWallarmWebhookDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return err
	}

	return nil
}
