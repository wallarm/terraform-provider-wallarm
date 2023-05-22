package wallarm

import (
	"fmt"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmEmail() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmIntegrationCreate,
		Read:   resourceWallarmEmailRead,
		Update: resourceWallarmEmailUpdate,
		Delete: resourceWallarmEmailDelete,

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
				Default:  "Email integration managed by Terraform",
			},

			"emails": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"event": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 8,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{"report_daily", "report_weekly",
								"report_monthly", "system", "vuln_high", "vuln_medium", "vuln_low", "scope"}, false),
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

func resourceWallarmIntegrationCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	active := d.Get("active").(bool)
	emails := expandInterfaceToStringList(d.Get("emails").(interface{}))
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "email")
	if err != nil {
		return err
	}

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
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmEmailRead(d, m)
}

func resourceWallarmEmailRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	email, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	d.Set("integration_id", email.ID)
	d.Set("is_active", email.Active)
	d.Set("name", email.Name)
	d.Set("created_by", email.CreatedBy)
	d.Set("type", email.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmEmailUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	active := d.Get("active").(bool)
	emails := expandInterfaceToStringList(d.Get("emails").(interface{}))
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "email")
	if err != nil {
		return err
	}

	email, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	var updateRes *wallarm.IntegrationCreateResp

	if d.HasChange("active") {
		emailBody := wallarm.EmailIntegrationCreate{
			Active: active,
		}

		updateRes, err = client.EmailIntegrationUpdate(&emailBody, email.ID)
		if err != nil {
			return err
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
			return err
		}
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmEmailRead(d, m)
}

func resourceWallarmEmailDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	email, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}
	if err := client.IntegrationDelete(email.ID); err != nil {
		return err
	}
	return nil
}
