package wallarm

import (
	"fmt"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmDataDog() *schema.Resource {
	return &schema.Resource{
		Create:        resourceWallarmDataDogCreate,
		Read:          resourceWallarmDataDogRead,
		Update:        resourceWallarmDataDogUpdate,
		Delete:        resourceWallarmDataDogDelete,
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
				Default:  "DataDog integration managed by Terraform",
			},

			"token": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(32, 40),
				Description:  "DataDog API key.",
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
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
							Default:  false,
						},
						"with_headers": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Send requests with headers. Only applicable to the 'siem' event type.",
						},
					},
				},
			},
		},
	}
}

func resourceWallarmDataDogCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	name := d.Get("name").(string)
	token := d.Get("token").(string)
	region := d.Get("region").(string)
	active := d.Get("active").(bool)
	events := expandWallarmEventToIntEvents(d.Get("event"), "data_dog")

	ddBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: &wallarm.DatadogTarget{
			Token:  token,
			Region: region,
		},
		Clientid: clientID,
		Type:     "data_dog",
		Events:   events,
	}

	createRes, err := client.IntegrationCreate(&ddBody)
	if err != nil {
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmDataDogRead(d, m)
}

func resourceWallarmDataDogRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	dd, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not found.") {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("integration_id", dd.ID)
	d.Set("is_active", dd.Active)
	d.Set("name", dd.Name)
	d.Set("created_by", dd.CreatedBy)
	d.Set("type", dd.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmDataDogUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	dd, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not found.") {
			d.SetId("")
			return nil
		}
		return err
	}

	if d.HasChange("event") {
		// When events change, API requires the full configuration
		fullBody := wallarm.IntegrationCreate{
			Name:   d.Get("name").(string),
			Active: d.Get("active").(bool),
			Target: &wallarm.DatadogTarget{
				Token:  d.Get("token").(string),
				Region: d.Get("region").(string),
			},
			Events: expandWallarmEventToIntEvents(d.Get("event"), "data_dog"),
			Type:   "data_dog",
		}
		updateRes, err := client.IntegrationUpdate(&fullBody, dd.ID)
		if err != nil {
			return err
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
		if d.HasChange("token") || d.HasChange("region") {
			updateBody["target"] = &wallarm.DatadogTarget{
				Token:  d.Get("token").(string),
				Region: d.Get("region").(string),
			}
		}
		if len(updateBody) > 0 {
			updateRes, err := client.IntegrationPartialUpdate(dd.ID, updateBody)
			if err != nil {
				return err
			}
			d.Set("integration_id", updateRes.Body.ID)
			resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
			d.SetId(resID)
		}
	}

	return resourceWallarmDataDogRead(d, m)
}

func resourceWallarmDataDogDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return err
	}

	return nil
}
