package wallarm

import (
	"fmt"
	"strings"

	"github.com/wallarm/wallarm-go"

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
				MaxItems: 7,
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
		},
	}
}

func resourceWallarmPagerDutyCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiToken := d.Get("integration_key").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event"), "pager_duty")
	if err != nil {
		d.SetId("")
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

	if err = d.Set("integration_id", createRes.Body.ID); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmPagerDutyRead(d, m)
}

func resourceWallarmPagerDutyRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	pagerduty, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not found.") {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	if err = d.Set("integration_id", pagerduty.ID); err != nil {
		return err
	}
	if err = d.Set("is_active", pagerduty.Active); err != nil {
		return err
	}
	if err = d.Set("name", pagerduty.Name); err != nil {
		return err
	}
	if err = d.Set("created_by", pagerduty.CreatedBy); err != nil {
		return err
	}
	if err = d.Set("type", pagerduty.Type); err != nil {
		return err
	}
	if err = d.Set("client_id", clientID); err != nil {
		return err
	}

	return nil
}

func resourceWallarmPagerDutyUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	integrationKey := d.Get("integration_key").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event"), "pager_duty")
	if err != nil {
		return err
	}

	pagerDuty, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not found.") {
			d.SetId("")
			return nil
		} else {
			return err
		}
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

	if err = d.Set("integration_id", updateRes.Body.ID); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmPagerDutyRead(d, m)
}

func resourceWallarmPagerDutyDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return err
	}

	return nil
}
