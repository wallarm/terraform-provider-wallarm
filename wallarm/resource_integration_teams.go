package wallarm

import (
	"fmt"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmTeams() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmTeamsCreate,
		Read:   resourceWallarmTeamsRead,
		Update: resourceWallarmTeamsUpdate,
		Delete: resourceWallarmTeamsDelete,

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
				Default:  "MS Teams integration managed by Terraform",
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

func resourceWallarmTeamsCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	webhookURL := d.Get("webhook_url").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "ms_teams")
	if err != nil {
		return err
	}

	teamsBody := wallarm.IntegrationCreate{
		Name:     name,
		Active:   active,
		Target:   webhookURL,
		Clientid: clientID,
		Type:     "ms_teams",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&teamsBody)
	if err != nil {
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmTeamsRead(d, m)
}

func resourceWallarmTeamsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	teams, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}
	d.Set("integration_id", teams.ID)
	d.Set("is_active", teams.Active)
	d.Set("name", teams.Name)
	d.Set("created_by", teams.CreatedBy)
	d.Set("type", teams.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmTeamsUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	webhookURL := d.Get("webhook_url").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "ms_teams")
	if err != nil {
		return err
	}

	teams, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	teamsBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: webhookURL,
		Type:   "ms_teams",
		Events: events,
	}

	updateRes, err := client.IntegrationUpdate(&teamsBody, teams.ID)
	if err != nil {
		return err
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmTeamsRead(d, m)
}

func resourceWallarmTeamsDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return err
	}

	return nil
}
