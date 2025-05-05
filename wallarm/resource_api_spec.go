package wallarm

import (
	"strconv"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceWallarmAPISpec() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmAPISpecCreate,
		Read:   resourceWallarmAPISpecRead,
		Delete: resourceWallarmAPISpecDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The Client ID to perform changes",
				ForceNew:    true,
			},
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The title of the API specification",
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the API specification",
				ForceNew:    true,
			},
			"file_remote_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The remote URL of the API specification file",
				ForceNew:    true,
			},
			"regular_file_update": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if the file should be updated regularly",
				ForceNew:    true,
			},
			"api_detection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if API detection is enabled",
				ForceNew:    true,
			},
			"domains": {
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				Description: "List of domains",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"instances": {
				Type:        schema.TypeList,
				ForceNew:    true,
				Required:    true,
				Description: "List of instance IDs",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"api_spec_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceWallarmAPISpecCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)

	apiSpecBody := wallarm.ApiSpecCreate{
		Title:             d.Get("title").(string),
		Description:       d.Get("description").(string),
		FileRemoteURL:     d.Get("file_remote_url").(string),
		RegularFileUpdate: d.Get("regular_file_update").(bool),
		ApiDetection:      d.Get("api_detection").(bool),
		ClientID:          d.Get("client_id").(int),
		Instances:         d.Get("instances").([]interface{}),
		Domains:           d.Get("domains").([]interface{}),
	}

	createRes, err := client.ApiSpecCreate(&apiSpecBody)
	if err != nil {
		return err
	}

	d.Set("api_spec_id", createRes.Body.ID)
	d.SetId(strconv.Itoa(createRes.Body.ID))

	return resourceWallarmAPISpecRead(d, m)
}

func resourceWallarmAPISpecRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

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
	d.Set("instances", apiSpec.Instances)
	d.Set("domains", apiSpec.Domains)

	return nil
}

func resourceWallarmAPISpecDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	err := client.ApiSpecDelete(clientID, apiSpecID)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
