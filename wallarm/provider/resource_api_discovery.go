package wallarm

import (
	"context"
	"math"
	"strconv"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceWallarmAPIDiscovery — singleton config per client_id, mirroring the
// console's Settings → API Discovery page.
func resourceWallarmAPIDiscovery() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceWallarmAPIDiscoveryRead,
		CreateContext: resourceWallarmAPIDiscoveryCreate,
		UpdateContext: resourceWallarmAPIDiscoveryUpdate,
		DeleteContext: resourceWallarmAPIDiscoveryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "Tenant client ID. Uses the provider's default when omitted.",
				ValidateFunc: validation.IntBetween(1, math.MaxInt32),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Master toggle for API Discovery.",
			},
			"apply_extended_filter": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Filter endpoints by response content type.",
			},
			"type_detection_threshold": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      0.5,
				Description:  "Fraction of requests used to determine parameter types (0.0–1.0).",
				ValidateFunc: validation.FloatBetween(0.0, 1.0),
			},
			"pii_detection_threshold": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      0.1,
				Description:  "Fraction of requests used to detect sensitive data (0.0–1.0).",
				ValidateFunc: validation.FloatBetween(0.0, 1.0),
			},
			"disabled_apps": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Pool IDs excluded from API Discovery.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"protocols": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Which protocols API Discovery should analyse.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rest":    {Type: schema.TypeBool, Optional: true, Default: true},
						"graphql": {Type: schema.TypeBool, Optional: true, Default: true},
						"soap":    {Type: schema.TypeBool, Optional: true, Default: true},
						"grpc":    {Type: schema.TypeBool, Optional: true, Default: true},
						"mcp":     {Type: schema.TypeBool, Optional: true, Default: true},
					},
				},
			},
			"endpoint_stability": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Thresholds for promoting an endpoint to the discovered inventory.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							Description:  "Minimum number of requests.",
							ValidateFunc: validation.IntBetween(1, 100),
						},
						"min_time": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      300,
							Description:  "Minimum time window in seconds.",
							ValidateFunc: validation.IntBetween(1, 900),
						},
					},
				},
			},

			// --- Computed / read-only ------------------------------------

			"call_points_storage_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Storage limit for call points (read-only).",
			},
			"group_soap": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether SOAP endpoints are grouped (read-only).",
			},
			"allowed_content_types_patterns": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Content-type patterns considered for discovery (read-only).",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"sensitive_samples": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Sensitive-sample masking config (read-only).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled":      {Type: schema.TypeBool, Computed: true},
						"min_masked":   {Type: schema.TypeInt, Computed: true},
						"max_masked":   {Type: schema.TypeInt, Computed: true},
						"mask_symbols": {Type: schema.TypeBool, Computed: true},
					},
				},
			},
			"server_variability": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Server-variability heuristics (read-only).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled":                    {Type: schema.TypeBool, Computed: true},
						"by_date_enabled":            {Type: schema.TypeBool, Computed: true},
						"by_local_code_enabled":      {Type: schema.TypeBool, Computed: true},
						"by_email_enabled":           {Type: schema.TypeBool, Computed: true},
						"by_alphanumeric_id_enabled": {Type: schema.TypeBool, Computed: true},
						"by_custom_paths": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {Type: schema.TypeBool, Computed: true},
									"paths":   {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
								},
							},
						},
					},
				},
			},
			"extensions_whitelist": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "File extensions filtered during discovery (read-only).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled":    {Type: schema.TypeBool, Computed: true},
						"extensions": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
					},
				},
			},
		},
	}
}

func resourceWallarmAPIDiscoveryRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	// Import sets d.Id() = "<client_id>"; honour it.
	if id := d.Id(); id != "" {
		if parsed, perr := strconv.Atoi(id); perr == nil {
			clientID = parsed
		}
	}

	cfg, err := client.APIDiscoveryConfigRead(clientID)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg == nil {
		d.SetId("")
		return nil
	}

	d.SetId(strconv.Itoa(clientID))
	d.Set("client_id", clientID)
	d.Set("enabled", cfg.Enabled)
	d.Set("apply_extended_filter", cfg.ApplyExtendedFilter)
	d.Set("type_detection_threshold", cfg.TypeDetectionThreshold)
	d.Set("pii_detection_threshold", cfg.PIIDetectionThreshold)
	d.Set("disabled_apps", cfg.DisabledApps)

	d.Set("protocols", []map[string]any{{
		"rest":    cfg.Protocols.REST,
		"graphql": cfg.Protocols.GraphQL,
		"soap":    cfg.Protocols.SOAP,
		"grpc":    cfg.Protocols.GRPC,
		"mcp":     cfg.Protocols.MCP,
	}})

	d.Set("endpoint_stability", []map[string]any{{
		"min_count": cfg.EndpointStability.MinCount,
		"min_time":  cfg.EndpointStability.MinTime,
	}})

	// Computed-only attributes
	d.Set("call_points_storage_limit", cfg.CallPointsStorageLimit)
	d.Set("group_soap", cfg.GroupSOAP)
	d.Set("allowed_content_types_patterns", cfg.AllowedContentTypesPatterns)

	d.Set("sensitive_samples", []map[string]any{{
		"enabled":      cfg.SensitiveSamples.Enabled,
		"min_masked":   cfg.SensitiveSamples.MinMasked,
		"max_masked":   cfg.SensitiveSamples.MaxMasked,
		"mask_symbols": cfg.SensitiveSamples.MaskSymbols,
	}})

	d.Set("server_variability", []map[string]any{{
		"enabled":                    cfg.ServerVariability.Enabled,
		"by_date_enabled":            cfg.ServerVariability.ByDateEnabled,
		"by_local_code_enabled":      cfg.ServerVariability.ByLocalCodeEnabled,
		"by_email_enabled":           cfg.ServerVariability.ByEmailEnabled,
		"by_alphanumeric_id_enabled": cfg.ServerVariability.ByAlphanumericIDEnabled,
		"by_custom_paths": []map[string]any{{
			"enabled": cfg.ServerVariability.ByCustomPaths.Enabled,
			"paths":   cfg.ServerVariability.ByCustomPaths.Paths,
		}},
	}})

	d.Set("extensions_whitelist", []map[string]any{{
		"enabled":    cfg.ExtensionsWhitelist.Enabled,
		"extensions": cfg.ExtensionsWhitelist.Extensions,
	}})

	return nil
}

// Create / Update / Delete handlers — bodies land in Task 5 once the expand
// helper exists in Task 4. Until then, they're no-ops that delegate to Read so
// the resource is registrable and Read tests can run end-to-end.
func resourceWallarmAPIDiscoveryCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	_ = wallarm.APIDiscoveryConfig{} // referenced from Task 5
	return resourceWallarmAPIDiscoveryRead(ctx, d, m)
}

func resourceWallarmAPIDiscoveryUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceWallarmAPIDiscoveryRead(ctx, d, m)
}

func resourceWallarmAPIDiscoveryDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	// Settings are a singleton — cannot be deleted, only modified.
	return nil
}
