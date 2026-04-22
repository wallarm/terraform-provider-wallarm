package wallarm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenAPISpecAuthHeaders(hdrs []wallarm.APISpecAuthHeader) []interface{} {
	out := make([]interface{}, 0, len(hdrs))
	for _, h := range hdrs {
		out = append(out, map[string]interface{}{"key": h.Key, "value": h.Value})
	}
	return out
}

func flattenAPISpecFile(f *wallarm.APISpecFile) []interface{} {
	if f == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"name":       f.Name,
		"signed_url": f.SignedURL,
		"checksum":   f.Checksum,
		"mime_type":  f.MimeType,
		"version":    f.Version,
	}}
}

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
			"domains": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of domains",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"instances": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of instance IDs",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"auth_headers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key":   {Type: schema.TypeString, Required: true},
						"value": {Type: schema.TypeString, Required: true, Sensitive: true},
					},
				},
			},
			"api_spec_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status":                 {Type: schema.TypeString, Computed: true},
			"spec_version":           {Type: schema.TypeString, Computed: true},
			"openapi_version":        {Type: schema.TypeString, Computed: true},
			"endpoints_count":        {Type: schema.TypeInt, Computed: true},
			"shadow_endpoints_count": {Type: schema.TypeInt, Computed: true},
			"orphan_endpoints_count": {Type: schema.TypeInt, Computed: true},
			"zombie_endpoints_count": {Type: schema.TypeInt, Computed: true},
			"format":                 {Type: schema.TypeInt, Computed: true},
			"version":                {Type: schema.TypeInt, Computed: true},
			"node_sync_version":      {Type: schema.TypeInt, Computed: true},
			"last_synced_at":         {Type: schema.TypeString, Computed: true},
			"last_compared_at":       {Type: schema.TypeString, Computed: true},
			"updated_at":             {Type: schema.TypeString, Computed: true},
			"created_at":             {Type: schema.TypeString, Computed: true},
			"file_changed_at":        {Type: schema.TypeString, Computed: true},
			"file": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name":       {Type: schema.TypeString, Computed: true},
						"signed_url": {Type: schema.TypeString, Computed: true, Sensitive: true},
						"checksum":   {Type: schema.TypeString, Computed: true},
						"mime_type":  {Type: schema.TypeString, Computed: true},
						"version":    {Type: schema.TypeInt, Computed: true},
					},
				},
			},
		},
	}
}

func resourceWallarmAPISpecCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)

	apiSpecBody := wallarm.APISpecCreate{
		Title:             d.Get("title").(string),
		Description:       d.Get("description").(string),
		FileRemoteURL:     d.Get("file_remote_url").(string),
		RegularFileUpdate: d.Get("regular_file_update").(bool),
		APIDetection:      d.Get("api_detection").(bool),
		ClientID:          d.Get("client_id").(int),
		Instances:         d.Get("instances").([]interface{}),
		Domains:           d.Get("domains").([]interface{}),
	}

	createRes, err := client.APISpecCreate(&apiSpecBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("api_spec_id", createRes.Body.ID)
	d.SetId(fmt.Sprintf("%d/%d", d.Get("client_id").(int), createRes.Body.ID))

	return resourceWallarmAPISpecRead(ctx, d, m)
}

func resourceWallarmAPISpecRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	spec, err := client.APISpecReadByID(clientID, apiSpecID)
	if err != nil {
		if errors.Is(err, wallarm.ErrNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("client_id", spec.ClientID)
	d.Set("title", spec.Title)
	d.Set("description", spec.Description)
	d.Set("file_remote_url", spec.FileRemoteURL)
	d.Set("regular_file_update", spec.RegularFileUpdate)
	d.Set("api_detection", spec.APIDetection)
	if err := d.Set("domains", spec.Domains); err != nil {
		return diag.FromErr(fmt.Errorf("error setting domains: %w", err))
	}
	if err := d.Set("instances", spec.Instances); err != nil {
		return diag.FromErr(fmt.Errorf("error setting instances: %w", err))
	}
	if err := d.Set("auth_headers", flattenAPISpecAuthHeaders(spec.AuthHeaders)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting auth_headers: %w", err))
	}

	d.Set("status", spec.Status)
	d.Set("spec_version", spec.SpecVersion)
	d.Set("openapi_version", spec.OpenAPIVersion)
	d.Set("endpoints_count", spec.EndpointsCount)
	d.Set("shadow_endpoints_count", spec.ShadowEndpointsCount)
	d.Set("orphan_endpoints_count", spec.OrphanEndpointsCount)
	d.Set("zombie_endpoints_count", spec.ZombieEndpointsCount)
	d.Set("format", spec.Format)
	d.Set("version", spec.Version)
	d.Set("node_sync_version", spec.NodeSyncVersion)
	d.Set("last_synced_at", spec.LastSyncedAt)
	d.Set("last_compared_at", spec.LastComparedAt)
	d.Set("updated_at", spec.UpdatedAt)
	d.Set("created_at", spec.CreatedAt)
	d.Set("file_changed_at", spec.FileChangedAt)
	if err := d.Set("file", flattenAPISpecFile(spec.File)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting file: %w", err))
	}

	return nil
}

func resourceWallarmAPISpecDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	err := client.APISpecDelete(clientID, apiSpecID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// resourceWallarmAPISpecImport parses a 2-part import ID "{client_id}/{api_spec_id}".
func resourceWallarmAPISpecImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 3)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{client_id}/{api_spec_id}\"", d.Id())
	}
	clientID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid client_id: %w", err)
	}
	apiSpecID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid api_spec_id: %w", err)
	}
	d.Set("client_id", clientID)
	d.Set("api_spec_id", apiSpecID)
	d.SetId(fmt.Sprintf("%d/%d", clientID, apiSpecID))
	return []*schema.ResourceData{d}, nil
}
