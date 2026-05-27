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
	for k, v := range flattenAPIDiscoveryConfig(cfg) {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

// Create / Update handlers wire to the same upsert. Full body via expand;
// then read back to refresh Computed fields.
func resourceWallarmAPIDiscoveryCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceWallarmAPIDiscoveryUpdate(ctx, d, m)
}

func resourceWallarmAPIDiscoveryUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceWallarmAPIDiscoveryRead(ctx, d, m)
}

// flattenAPIDiscoveryConfig converts the wallarm-go struct into a map keyed
// by schema attribute name. Nested editable/Computed blocks (TypeList/MaxItems:1)
// are wrapped in []map[string]any so the SDK accepts them via d.Set.
func flattenAPIDiscoveryConfig(cfg *wallarm.APIDiscoveryConfig) map[string]any {
	if cfg == nil {
		return nil
	}
	return map[string]any{
		"client_id":                cfg.ClientID,
		"enabled":                  cfg.Enabled,
		"apply_extended_filter":    cfg.ApplyExtendedFilter,
		"type_detection_threshold": cfg.TypeDetectionThreshold,
		"pii_detection_threshold":  cfg.PIIDetectionThreshold,
		"disabled_apps":            cfg.DisabledApps,
		"protocols": []map[string]any{{
			"rest":    cfg.Protocols.REST,
			"graphql": cfg.Protocols.GraphQL,
			"soap":    cfg.Protocols.SOAP,
			"grpc":    cfg.Protocols.GRPC,
			"mcp":     cfg.Protocols.MCP,
		}},
		"endpoint_stability": []map[string]any{{
			"min_count": cfg.EndpointStability.MinCount,
			"min_time":  cfg.EndpointStability.MinTime,
		}},
		"call_points_storage_limit":      cfg.CallPointsStorageLimit,
		"group_soap":                     cfg.GroupSOAP,
		"allowed_content_types_patterns": cfg.AllowedContentTypesPatterns,
		"sensitive_samples": []map[string]any{{
			"enabled":      cfg.SensitiveSamples.Enabled,
			"min_masked":   cfg.SensitiveSamples.MinMasked,
			"max_masked":   cfg.SensitiveSamples.MaxMasked,
			"mask_symbols": cfg.SensitiveSamples.MaskSymbols,
		}},
		"server_variability": []map[string]any{{
			"enabled":                    cfg.ServerVariability.Enabled,
			"by_date_enabled":            cfg.ServerVariability.ByDateEnabled,
			"by_local_code_enabled":      cfg.ServerVariability.ByLocalCodeEnabled,
			"by_email_enabled":           cfg.ServerVariability.ByEmailEnabled,
			"by_alphanumeric_id_enabled": cfg.ServerVariability.ByAlphanumericIDEnabled,
			"by_custom_paths": []map[string]any{{
				"enabled": cfg.ServerVariability.ByCustomPaths.Enabled,
				"paths":   cfg.ServerVariability.ByCustomPaths.Paths,
			}},
		}},
		"extensions_whitelist": []map[string]any{{
			"enabled":    cfg.ExtensionsWhitelist.Enabled,
			"extensions": cfg.ExtensionsWhitelist.Extensions,
		}},
	}
}

// expandAPIDiscoveryConfig reads the resource state into a wallarm-go struct
// for the POST upsert. Schema Defaults (via *schema.ResourceData) backfill any
// omitted scalars and nested blocks.
func expandAPIDiscoveryConfig(d *schema.ResourceData) *wallarm.APIDiscoveryConfig {
	cfg := &wallarm.APIDiscoveryConfig{
		ClientID:               d.Get("client_id").(int),
		Enabled:                d.Get("enabled").(bool),
		ApplyExtendedFilter:    d.Get("apply_extended_filter").(bool),
		TypeDetectionThreshold: d.Get("type_detection_threshold").(float64),
		PIIDetectionThreshold:  d.Get("pii_detection_threshold").(float64),
	}

	// disabled_apps: TypeList of TypeInt → []int
	if raw, ok := d.GetOk("disabled_apps"); ok {
		list := raw.([]any)
		cfg.DisabledApps = make([]int, 0, len(list))
		for _, v := range list {
			cfg.DisabledApps = append(cfg.DisabledApps, v.(int))
		}
	} else {
		cfg.DisabledApps = []int{}
	}

	// protocols (nested editable): use schema defaults when block omitted.
	cfg.Protocols = wallarm.APIDiscoveryProtocols{
		REST: true, GraphQL: true, SOAP: true, GRPC: true, MCP: true,
	}
	if list, ok := d.Get("protocols").([]any); ok && len(list) > 0 {
		m := list[0].(map[string]any)
		cfg.Protocols.REST = m["rest"].(bool)
		cfg.Protocols.GraphQL = m["graphql"].(bool)
		cfg.Protocols.SOAP = m["soap"].(bool)
		cfg.Protocols.GRPC = m["grpc"].(bool)
		cfg.Protocols.MCP = m["mcp"].(bool)
	}

	// endpoint_stability (nested editable): defaults from schema.
	cfg.EndpointStability = wallarm.APIDiscoveryEndpointStability{
		MinCount: 2, MinTime: 300,
	}
	if list, ok := d.Get("endpoint_stability").([]any); ok && len(list) > 0 {
		m := list[0].(map[string]any)
		cfg.EndpointStability.MinCount = m["min_count"].(int)
		cfg.EndpointStability.MinTime = m["min_time"].(int)
	}

	return cfg
}

func resourceWallarmAPIDiscoveryDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	// Settings are a singleton — cannot be deleted, only modified.
	return nil
}
