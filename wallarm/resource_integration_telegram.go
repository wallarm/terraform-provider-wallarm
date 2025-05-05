package wallarm

import (
	"fmt"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmTelegram() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmTelegramCreate,
		Read:   resourceWallarmTelegramRead,
		Delete: resourceWallarmTelegramDelete,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
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
				ForceNew: true,
			},

			"token": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},

			"chat_data": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
		},
	}
}

func resourceWallarmTelegramCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	name := d.Get("name").(string)
	chatData := d.Get("chat_data").(string)
	token := d.Get("token").(string)

	tgBody := wallarm.TelegramIntegrationCreate{
		Name:     name,
		Clientid: clientID,
		Token:    token,
		ChatData: chatData,
	}

	createRes, err := client.TelegramIntegrationCreate(&tgBody)
	if err != nil {
		return err
	}

	d.Set("integration_id", createRes.Body.ID)

	resID := fmt.Sprintf("%d/%s/%d", clientID, createRes.Body.Type, createRes.Body.ID)
	d.SetId(resID)

	return resourceWallarmTelegramRead(d, m)
}

func resourceWallarmTelegramRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	telegram, err := client.IntegrationRead(clientID, d.Get("integration_id").(int))
	if err != nil {
		return err
	}

	d.Set("integration_id", telegram.ID)
	d.Set("is_active", telegram.Active)
	d.Set("name", telegram.Name)
	d.Set("created_by", telegram.CreatedBy)
	d.Set("type", telegram.Type)
	d.Set("client_id", clientID)

	return nil
}

func resourceWallarmTelegramDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	integrationID := d.Get("integration_id").(int)
	if err := client.IntegrationDelete(integrationID); err != nil {
		return err
	}

	return nil
}
