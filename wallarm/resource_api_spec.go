package wallarm

import (
	"fmt"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmApiSpec() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmApiSpecCreate,
		Read:   resourceWallarmApiSpecRead,
		Delete: resourceWallarmApiSpecDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The Client ID to perform changes",
			},
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The title of the API specification",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the API specification",
			},
			"file_remote_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The remote URL of the API specification file",
			},
			"regular_file_update": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if the file should be updated regularly",
			},
			"api_detection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if API detection is enabled",
			},
			"api_spec_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWallarmApiSpecCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	apiSpecBody := wallarm.ApiSpecCreate{
		Title:             d.Get("title").(string),
		Description:       d.Get("description").(string),
		FileRemoteURL:     d.Get("file_remote_url").(string),
		RegularFileUpdate: d.Get("regular_file_update").(bool),
		ApiDetection:      d.Get("api_detection").(bool),
		ClientID:          d.Get("client_id").(int),
	}

	createRes, err := client.ApiSpecCreate(&apiSpecBody)
	if err != nil {
		return err
	}

	d.Set("api_spec_id", createRes.Body.ID)
	d.SetId(fmt.Sprintf("%d", createRes.Body.ID))

	return resourceWallarmApiSpecRead(d, m)
}

func resourceWallarmApiSpecRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Id()

	apiSpec, err := client.ApiSpecRead(clientID, apiSpecID)
	if err != nil {
		if err.Error() == "Not found" {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("title", apiSpec.Title)
	d.Set("description", apiSpec.Description)
	d.Set("file_remote_url", apiSpec.FileRemoteURL)
	d.Set("regular_file_update", apiSpec.RegularFileUpdate)
	d.Set("api_detection", apiSpec.ApiDetection)
	d.Set("client_id", apiSpec.ClientID)

	return nil
}

func resourceWallarmApiSpecDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Id()

	err := client.ApiSpecDelete(clientID, apiSpecID)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
