package wallarm

import (
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmDataDog() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmDataDogCreate,
		Read:   resourceWallarmDataDogRead,
		Update: resourceWallarmDataDogUpdate,
		Delete: resourceWallarmDataDogDelete,

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
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
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
							Default:  false,
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
	name := d.Get("name").(string)
	region := d.Get("region").(string)
	token := d.Get("token").(string)
	active := d.Get("active").(bool)
	events := expandWallarmEventToIntEvents(d.Get("event"), "data_dog")

	dd, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	ddBody := wallarm.IntegrationCreate{
		Name:   name,
		Active: active,
		Target: &wallarm.DatadogTarget{
			Token:  token,
			Region: region,
		},
		Type:   "data_dog",
		Events: events,
	}

	updateRes, err := client.IntegrationUpdate(&ddBody, dd.ID)
	if err != nil {
		return err
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

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
