package wallarm

import (
	"fmt"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmPagerDuty() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmPagerDutyCreate,
		Read:   resourceWallarmPagerDutyRead,
		Update: resourceWallarmPagerDutyUpdate,
		Delete: resourceWallarmPagerDutyDelete,

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
				Default:  "PagerDuty integration managed by Terraform",
			},

			"integration_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if len(v) != 32 {
						errs = append(errs, fmt.Errorf("length of %q must be equal to 32, got: %d", key, len(v)))
					}
					return
				},
			},

			"event": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"hit", "vuln", "system", "scope"}, false),
							Default:      "vuln",
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

func resourceWallarmPagerDutyCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiToken := d.Get("integration_key").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "pager_duty")
	if err != nil {
		return err
	}

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
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmPagerDutyRead(d, m)
}

func resourceWallarmPagerDutyRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	pagerduty, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	d.Set("integration_id", pagerduty.ID)
	d.Set("is_active", pagerduty.Active)
	d.Set("name", pagerduty.Name)
	d.Set("created_by", pagerduty.CreatedBy)
	d.Set("type", pagerduty.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmPagerDutyUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	integrationKey := d.Get("integration_key").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "pager_duty")
	if err != nil {
		return err
	}

	pagerDuty, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	pagerBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: integrationKey,
		Type:   "pager_duty",
		Events: events,
	}

	updateRes, err := client.IntegrationUpdate(&pagerBody, pagerDuty.ID)
	if err != nil {
		return err
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmPagerDutyRead(d, m)
}

func resourceWallarmPagerDutyDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	pagerduty, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}
	if err := client.IntegrationDelete(pagerduty.ID); err != nil {
		return err
	}

	return nil
}
