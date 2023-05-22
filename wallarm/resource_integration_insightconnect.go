package wallarm

import (
	"fmt"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmInsightConnect() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmInsightConnectCreate,
		Read:   resourceWallarmInsightConnectRead,
		Update: resourceWallarmInsightConnectUpdate,
		Delete: resourceWallarmInsightConnectDelete,

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
				MaxItems: 6,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"hit", "vuln_high", "vuln_medium", "vuln_low", "system", "scope"}, false),
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

func resourceWallarmInsightConnectCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiURL := d.Get("api_url").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "insight_connect")
	if err != nil {
		return err
	}

	insightBody := wallarm.IntegrationWithAPICreate{
		Name:   name,
		Active: active,
		Target: &wallarm.IntegrationWithAPITarget{
			Token: apiToken,
			API:   apiURL,
		},
		Clientid: clientID,
		Type:     "insight_connect",
		Events:   events,
	}

	createRes, err := client.IntegrationWithAPICreate(&insightBody)
	if err != nil {
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmInsightConnectRead(d, m)
}

func resourceWallarmInsightConnectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	insight, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	d.Set("integration_id", insight.ID)
	d.Set("is_active", insight.Active)
	d.Set("name", insight.Name)
	d.Set("created_by", insight.CreatedBy)
	d.Set("type", insight.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmInsightConnectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiURL := d.Get("api_url").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "insight_connect")
	if err != nil {
		return err
	}

	insight, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	insightBody := wallarm.IntegrationWithAPICreate{
		Name:   name,
		Active: active,
		Target: &wallarm.IntegrationWithAPITarget{
			Token: apiToken,
			API:   apiURL,
		},
		// Clientid: clientID,
		Type:   "insight_connect",
		Events: events,
	}

	updateRes, err := client.IntegrationWithAPIUpdate(&insightBody, insight.ID)
	if err != nil {
		return err
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmInsightConnectRead(d, m)
}

func resourceWallarmInsightConnectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	insight, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}
	if err := client.IntegrationDelete(insight.ID); err != nil {
		return err
	}

	return nil
}
