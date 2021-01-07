package wallarm

import (
	"fmt"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmOpsGenie() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmOpsGenieCreate,
		Read:   resourceWallarmOpsGenieRead,
		Update: resourceWallarmOpsGenieUpdate,
		Delete: resourceWallarmOpsGenieDelete,

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
				Default:  "OpsGenie integration managed by Terraform",
			},

			"api_token": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"event": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"hit", "vuln"}, false),
							Default:      "vuln",
						},
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func resourceWallarmOpsGenieCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "opsgenie")
	if err != nil {
		return err
	}

	opsGenieBody := wallarm.IntegrationCreate{
		Name:     name,
		Active:   active,
		Target:   apiToken,
		Clientid: clientID,
		Type:     "opsgenie",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&opsGenieBody)
	if err != nil {
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmOpsGenieRead(d, m)
}

func resourceWallarmOpsGenieRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	opsGenie, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	d.Set("integration_id", opsGenie.ID)
	d.Set("is_active", opsGenie.Active)
	d.Set("name", opsGenie.Name)
	d.Set("created_by", opsGenie.CreatedBy)
	d.Set("type", opsGenie.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmOpsGenieUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "opsgenie")
	if err != nil {
		return err
	}

	opsgenie, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	opsgenieBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: apiToken,
		Type:   "opsgenie",
		Events: events,
	}

	updateRes, err := client.IntegrationUpdate(&opsgenieBody, opsgenie.ID)
	if err != nil {
		return err
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmOpsGenieRead(d, m)
}

func resourceWallarmOpsGenieDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	opsGenie, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}
	if err := client.IntegrationDelete(opsGenie.ID); err != nil {
		return err
	}

	return nil
}
