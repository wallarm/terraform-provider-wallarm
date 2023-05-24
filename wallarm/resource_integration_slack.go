package wallarm

import (
	"fmt"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmSlack() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmSlackCreate,
		Read:   resourceWallarmSlackRead,
		Update: resourceWallarmSlackUpdate,
		Delete: resourceWallarmSlackDelete,

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
				Default:  "Slack integration managed by Terraform",
			},

			"webhook_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Sensitive:    true,
			},

			"event": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"system", "vuln_high", "vuln_medium", "vuln_low", "scope"}, false),
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

func resourceWallarmSlackCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	webhookURL := d.Get("webhook_url").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "slack")
	if err != nil {
		return err
	}

	slackBody := wallarm.IntegrationCreate{
		Name:     name,
		Active:   active,
		Target:   webhookURL,
		Clientid: clientID,
		Type:     "slack",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&slackBody)
	if err != nil {
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmSlackRead(d, m)
}

func resourceWallarmSlackRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	slack, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}
	d.Set("integration_id", slack.ID)
	d.Set("is_active", slack.Active)
	d.Set("name", slack.Name)
	d.Set("created_by", slack.CreatedBy)
	d.Set("type", slack.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmSlackUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	webhookURL := d.Get("webhook_url").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "slack")
	if err != nil {
		return err
	}

	slack, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	slackBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: webhookURL,
		Type:   "slack",
		Events: events,
	}

	updateRes, err := client.IntegrationUpdate(&slackBody, slack.ID)
	if err != nil {
		return err
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmSlackRead(d, m)
}

func resourceWallarmSlackDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	slack, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}
	if err := client.IntegrationDelete(slack.ID); err != nil {
		return err
	}

	return nil
}
