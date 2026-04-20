package wallarm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWallarmAPISpec() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmAPISpecCreate,
		ReadContext:   resourceWallarmAPISpecRead,
		DeleteContext: resourceWallarmAPISpecDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmAPISpecImport,
		},

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

func resourceWallarmAPISpecCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)

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
		return diag.FromErr(err)
	}

	d.Set("api_spec_id", createRes.Body.ID)
	d.SetId(strconv.Itoa(createRes.Body.ID))

	return resourceWallarmAPISpecRead(ctx, d, m)
}

func resourceWallarmAPISpecRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	apiSpec, err := client.ApiSpecRead(clientID, apiSpecID)
	if err != nil {
		if errors.Is(err, wallarm.ErrNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("title", apiSpec.Title)
	d.Set("description", apiSpec.Description)
	d.Set("file_remote_url", apiSpec.FileRemoteURL)
	d.Set("regular_file_update", apiSpec.RegularFileUpdate)
	d.Set("api_detection", apiSpec.ApiDetection)
	d.Set("client_id", apiSpec.ClientID)
	if err := d.Set("instances", apiSpec.Instances); err != nil {
		return diag.FromErr(fmt.Errorf("error setting instances: %w", err))
	}
	if err := d.Set("domains", apiSpec.Domains); err != nil {
		return diag.FromErr(fmt.Errorf("error setting domains: %w", err))
	}

	return nil
}

func resourceWallarmAPISpecDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	err := client.ApiSpecDelete(clientID, apiSpecID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// resourceWallarmAPISpecImport parses a single-integer import ID into
// api_spec_id. The caller must set client_id in the resource config.
func resourceWallarmAPISpecImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	apiSpecID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid id (%q), expected single integer api_spec_id: %w", d.Id(), err)
	}
	d.Set("api_spec_id", apiSpecID)
	d.SetId(strconv.Itoa(apiSpecID))
	return []*schema.ResourceData{d}, nil
}
