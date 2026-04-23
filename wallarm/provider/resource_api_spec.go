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

func expandAPISpecAuthHeaders(raw []interface{}) []wallarm.APISpecAuthHeader {
	out := make([]wallarm.APISpecAuthHeader, 0, len(raw))
	for _, r := range raw {
		m := r.(map[string]interface{})
		out = append(out, wallarm.APISpecAuthHeader{
			Key:   m["key"].(string),
			Value: m["value"].(string),
		})
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

// setStateFromAPISpecBody writes every API-returned field onto the Terraform
// resource state. Used by Read after APISpecReadByID; Create and Update delegate
// to Read because POST/PUT responses omit some fields (file_remote_url,
// auth_headers, file, policy) that only GET returns.
func setStateFromAPISpecBody(d *schema.ResourceData, spec wallarm.APISpecBody) error {
	d.Set("client_id", spec.ClientID)
	d.Set("title", spec.Title)
	d.Set("description", spec.Description)
	d.Set("file_remote_url", spec.FileRemoteURL)
	d.Set("regular_file_update", spec.RegularFileUpdate)
	d.Set("api_detection", spec.APIDetection)
	if err := d.Set("domains", spec.Domains); err != nil {
		return fmt.Errorf("error setting domains: %w", err)
	}
	if err := d.Set("instances", spec.Instances); err != nil {
		return fmt.Errorf("error setting instances: %w", err)
	}
	if err := d.Set("auth_headers", flattenAPISpecAuthHeaders(spec.AuthHeaders)); err != nil {
		return fmt.Errorf("error setting auth_headers: %w", err)
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
		return fmt.Errorf("error setting file: %w", err)
	}
	return nil
}

func resourceWallarmAPISpec() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmAPISpecCreate,
		ReadContext:   resourceWallarmAPISpecRead,
		UpdateContext: resourceWallarmAPISpecUpdate,
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
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Authorization headers sent by the spec fetcher when downloading the file_remote_url.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key":   {Type: schema.TypeString, Required: true, Description: "Header name."},
						"value": {Type: schema.TypeString, Required: true, Sensitive: true, Description: "Header value (sensitive)."},
					},
				},
			},
			"api_spec_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Server-assigned ID of the API specification.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Upload and processing status reported by the API (e.g., 'ready', 'pending').",
			},
			"spec_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The 'info.version' value declared inside the uploaded OpenAPI document.",
			},
			"openapi_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The OpenAPI/Swagger version declared by the uploaded document (e.g., '3.0.3').",
			},
			"endpoints_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of endpoints parsed from the specification.",
			},
			"shadow_endpoints_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of shadow endpoints (live but not in the spec).",
			},
			"orphan_endpoints_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of orphan endpoints (in the spec but not seen live).",
			},
			"zombie_endpoints_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of zombie endpoints (previously live, no longer seen).",
			},
			"format": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Internal format code of the stored specification file.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Monotonically increasing revision of this specification.",
			},
			"node_sync_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Revision most recently synchronized to filtering nodes.",
			},
			"last_synced_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of the last successful spec sync to nodes.",
			},
			"last_compared_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of the last live-traffic comparison against the spec.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of the last modification to this resource.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp of this resource.",
			},
			"file_changed_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the underlying spec file was last changed.",
			},
			"file": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Metadata of the stored specification file.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name":       {Type: schema.TypeString, Computed: true, Description: "Stored filename of the specification."},
						"signed_url": {Type: schema.TypeString, Computed: true, Sensitive: true, Description: "Time-limited signed URL for downloading the raw file."},
						"checksum":   {Type: schema.TypeString, Computed: true, Description: "Checksum of the stored file, used to detect content changes."},
						"mime_type":  {Type: schema.TypeString, Computed: true, Description: "MIME type of the stored file."},
						"version":    {Type: schema.TypeInt, Computed: true, Description: "Monotonically increasing revision of the stored file."},
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
		AuthHeaders:       expandAPISpecAuthHeaders(d.Get("auth_headers").([]interface{})),
	}

	createRes, err := client.APISpecCreate(&apiSpecBody)
	if err != nil {
		return diag.FromErr(err)
	}
	if createRes.Body == nil {
		return diag.Errorf("APISpecCreate: empty response body")
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

	if err := setStateFromAPISpecBody(d, spec); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceWallarmAPISpecUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID := d.Get("client_id").(int)
	apiSpecID := d.Get("api_spec_id").(int)

	body := &wallarm.APISpecUpdate{}
	if d.HasChange("title") {
		v := d.Get("title").(string)
		body.Title = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		body.Description = &v
	}
	if d.HasChange("file_remote_url") {
		v := d.Get("file_remote_url").(string)
		body.FileRemoteURL = &v
	}
	if d.HasChange("regular_file_update") {
		v := d.Get("regular_file_update").(bool)
		body.RegularFileUpdate = &v
	}
	if d.HasChange("api_detection") {
		v := d.Get("api_detection").(bool)
		body.APIDetection = &v
	}
	if d.HasChange("domains") {
		body.Domains = d.Get("domains").([]interface{})
	}
	if d.HasChange("instances") {
		body.Instances = d.Get("instances").([]interface{})
	}
	if d.HasChange("auth_headers") {
		body.AuthHeaders = expandAPISpecAuthHeaders(d.Get("auth_headers").([]interface{}))
	}

	if _, err := client.APISpecUpdate(clientID, apiSpecID, body); err != nil {
		return diag.FromErr(err)
	}
	return resourceWallarmAPISpecRead(ctx, d, m)
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
	parts := strings.Split(d.Id(), "/")
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
