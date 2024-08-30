package wallarm

import (
	"fmt"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmSplunk() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmSplunkCreate,
		Read:   resourceWallarmSplunkRead,
		Update: resourceWallarmSplunkUpdate,
		Delete: resourceWallarmSplunkDelete,

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
				Default:  "Splunk integration managed by Terraform",
			},

			"api_token": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.IsUUID,
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

func resourceWallarmSplunkCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiURL := d.Get("api_url").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "splunk")
	if err != nil {
		d.SetId("")
		return err
	}

	splunkBody := wallarm.IntegrationWithAPICreate{
		Name:   name,
		Active: active,
		Target: &wallarm.IntegrationWithAPITarget{
			Token: apiToken,
			API:   apiURL,
		},
		Clientid: clientID,
		Type:     "splunk",
		Events:   events,
	}

	createRes, err := client.IntegrationWithAPICreate(&splunkBody)
	if err != nil {
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmSplunkRead(d, m)
}

func resourceWallarmSplunkRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	splunk, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		if strings.HasPrefix(err.Error(), "Not found.") {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}
	d.Set("integration_id", splunk.ID)
	d.Set("is_active", splunk.Active)
	d.Set("name", splunk.Name)
	d.Set("created_by", splunk.CreatedBy)
	d.Set("type", splunk.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmSplunkUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	name := d.Get("name").(string)
	apiURL := d.Get("api_url").(string)
	apiToken := d.Get("api_token").(string)
	active := d.Get("active").(bool)
	events, err := expandWallarmEventToIntEvents(d.Get("event").(interface{}), "splunk")
	if err != nil {
		return err
	}

	splunk, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	splunkBody := wallarm.IntegrationWithAPICreate{
		Name:   name,
		Active: active,
		Target: &wallarm.IntegrationWithAPITarget{
			Token: apiToken,
			API:   apiURL,
		},
		Type:   "splunk",
		Events: events,
	}

	updateRes, err := client.IntegrationWithAPIUpdate(&splunkBody, splunk.ID)
	if err != nil {
		return err
	}

	d.Set("integration_id", updateRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, updateRes.Body.Type, updateRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmSplunkRead(d, m)
}

func resourceWallarmSplunkDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return err
	}

	return nil
}
