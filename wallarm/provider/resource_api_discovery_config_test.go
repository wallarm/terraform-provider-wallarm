package wallarm

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

// fixture mirrors the GET response captured during design probing.
func sampleAPIDiscoveryConfig() *wallarm.APIDiscoveryConfig {
	return &wallarm.APIDiscoveryConfig{
		ClientID: 22510,
		Enabled:  true,
		Protocols: wallarm.APIDiscoveryProtocols{
			REST: true, GraphQL: true, SOAP: false, GRPC: true, MCP: false,
		},
		ApplyExtendedFilter:    true,
		TypeDetectionThreshold: 0.5,
		PIIDetectionThreshold:  0.1,
		CallPointsStorageLimit: 50000,
		SensitiveSamples: wallarm.APIDiscoverySensitiveSamples{
			Enabled: false, MinMasked: 20, MaxMasked: 80, MaskSymbols: false,
		},
		DisabledApps: []int{42, 7},
		EndpointStability: wallarm.APIDiscoveryEndpointStability{
			MinCount: 3, MinTime: 600,
		},
		GroupSOAP: false,
		ServerVariability: wallarm.APIDiscoveryServerVariability{
			Enabled: true, ByDateEnabled: true,
			ByCustomPaths: wallarm.APIDiscoveryServerVariabilityByCustomPaths{
				Enabled: true, Paths: []string{"/api/v1/.*"},
			},
		},
		AllowedContentTypesPatterns: []string{"text/xml", "application/%json"},
		ExtensionsWhitelist: wallarm.APIDiscoveryExtensionsWhitelist{
			Enabled: true, Extensions: []string{"do", "action"},
		},
	}
}

func TestFlattenAPIDiscoveryConfig_FullResponse(t *testing.T) {
	t.Parallel()
	cfg := sampleAPIDiscoveryConfig()
	flat := flattenAPIDiscoveryConfig(cfg)

	// Top-level scalars.
	if flat["client_id"] != 22510 {
		t.Errorf("client_id: got %v", flat["client_id"])
	}
	if flat["enabled"] != true {
		t.Errorf("enabled: got %v", flat["enabled"])
	}
	if flat["apply_extended_filter"] != true {
		t.Errorf("apply_extended_filter: got %v", flat["apply_extended_filter"])
	}
	if flat["type_detection_threshold"] != 0.5 {
		t.Errorf("type_detection_threshold: got %v", flat["type_detection_threshold"])
	}
	if flat["call_points_storage_limit"] != 50000 {
		t.Errorf("call_points_storage_limit: got %v", flat["call_points_storage_limit"])
	}

	// Lists.
	if !reflect.DeepEqual(flat["disabled_apps"], []int{42, 7}) {
		t.Errorf("disabled_apps: got %v", flat["disabled_apps"])
	}
	if !reflect.DeepEqual(flat["allowed_content_types_patterns"], []string{"text/xml", "application/%json"}) {
		t.Errorf("allowed_content_types_patterns: got %v", flat["allowed_content_types_patterns"])
	}

	// Nested editable block: protocols.
	protocols, ok := flat["protocols"].([]map[string]any)
	if !ok || len(protocols) != 1 {
		t.Fatalf("protocols: expected single-element []map[string]any, got %T (%v)", flat["protocols"], flat["protocols"])
	}
	if protocols[0]["rest"] != true || protocols[0]["soap"] != false || protocols[0]["mcp"] != false {
		t.Errorf("protocols flag values wrong: %v", protocols[0])
	}

	// Nested editable block: endpoint_stability.
	stab, ok := flat["endpoint_stability"].([]map[string]any)
	if !ok || len(stab) != 1 {
		t.Fatalf("endpoint_stability: expected single-element list")
	}
	if stab[0]["min_count"] != 3 || stab[0]["min_time"] != 600 {
		t.Errorf("endpoint_stability values wrong: %v", stab[0])
	}

	// Nested Computed: server_variability with nested by_custom_paths.
	sv, ok := flat["server_variability"].([]map[string]any)
	if !ok || len(sv) != 1 {
		t.Fatalf("server_variability: expected single-element list")
	}
	if sv[0]["enabled"] != true || sv[0]["by_date_enabled"] != true {
		t.Errorf("server_variability flags wrong: %v", sv[0])
	}
	bcp, ok := sv[0]["by_custom_paths"].([]map[string]any)
	if !ok || len(bcp) != 1 {
		t.Fatalf("by_custom_paths: expected single-element list")
	}
	if bcp[0]["enabled"] != true {
		t.Errorf("by_custom_paths.enabled wrong: %v", bcp[0])
	}
	if !reflect.DeepEqual(bcp[0]["paths"], []string{"/api/v1/.*"}) {
		t.Errorf("by_custom_paths.paths wrong: %v", bcp[0]["paths"])
	}

	// Nested Computed: extensions_whitelist.
	ew, ok := flat["extensions_whitelist"].([]map[string]any)
	if !ok || len(ew) != 1 {
		t.Fatalf("extensions_whitelist: expected single-element list")
	}
	if ew[0]["enabled"] != true {
		t.Errorf("extensions_whitelist.enabled wrong: %v", ew[0])
	}
	if !reflect.DeepEqual(ew[0]["extensions"], []string{"do", "action"}) {
		t.Errorf("extensions_whitelist.extensions wrong: %v", ew[0]["extensions"])
	}
}

func TestExpandAPIDiscoveryConfig_RoundTrip(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"client_id":                22510,
		"enabled":                  true,
		"apply_extended_filter":    false,
		"type_detection_threshold": 0.8,
		"pii_detection_threshold":  0.2,
		"disabled_apps":            []any{42, 7},
		"protocols": []any{map[string]any{
			"rest": true, "graphql": false, "soap": false, "grpc": true, "mcp": false,
		}},
		"endpoint_stability": []any{map[string]any{
			"min_count": 5, "min_time": 450,
		}},
	}

	d := newAPIDiscoveryConfigResourceData(t, raw)
	d.SetId("22510")

	cfg := expandAPIDiscoveryConfig(d)
	if cfg == nil {
		t.Fatal("expandAPIDiscoveryConfig returned nil")
	}

	if cfg.ClientID != 22510 {
		t.Errorf("ClientID: got %d", cfg.ClientID)
	}
	if !cfg.Enabled {
		t.Errorf("Enabled: got %v", cfg.Enabled)
	}
	if cfg.ApplyExtendedFilter {
		t.Errorf("ApplyExtendedFilter: expected false, got true")
	}
	if cfg.TypeDetectionThreshold != 0.8 {
		t.Errorf("TypeDetectionThreshold: got %v", cfg.TypeDetectionThreshold)
	}
	if cfg.PIIDetectionThreshold != 0.2 {
		t.Errorf("PIIDetectionThreshold: got %v", cfg.PIIDetectionThreshold)
	}
	if !reflect.DeepEqual(cfg.DisabledApps, []int{42, 7}) {
		t.Errorf("DisabledApps: got %v", cfg.DisabledApps)
	}
	if !cfg.Protocols.REST || cfg.Protocols.GraphQL || cfg.Protocols.SOAP || !cfg.Protocols.GRPC || cfg.Protocols.MCP {
		t.Errorf("Protocols flags wrong: %+v", cfg.Protocols)
	}
	if cfg.EndpointStability.MinCount != 5 || cfg.EndpointStability.MinTime != 450 {
		t.Errorf("EndpointStability wrong: %+v", cfg.EndpointStability)
	}
}

func TestParseClientIDFromCompositeID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		id     string
		want   int
		wantOK bool
	}{
		{name: "composite id", id: "22510/apid_config", want: 22510, wantOK: true},
		{name: "bare numeric id (pre-rename state)", id: "22510", want: 22510, wantOK: true},
		{name: "empty id", id: "", want: 0, wantOK: false},
		{name: "non-numeric prefix", id: "notanumber/apid_config", want: 0, wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseClientIDFromCompositeID(tc.id)
			if ok != tc.wantOK {
				t.Errorf("ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("id: got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestExpandAPIDiscoveryConfig_DefaultsApplied(t *testing.T) {
	t.Parallel()

	// Only the bare minimum — let schema Defaults fill the rest.
	raw := map[string]any{
		"client_id": 22510,
	}
	d := newAPIDiscoveryConfigResourceData(t, raw)
	d.SetId("22510")

	cfg := expandAPIDiscoveryConfig(d)
	if cfg == nil {
		t.Fatal("expandAPIDiscoveryConfig returned nil")
	}

	// Defaults from schema.
	if !cfg.Enabled {
		t.Errorf("Enabled default should be true, got %v", cfg.Enabled)
	}
	if !cfg.ApplyExtendedFilter {
		t.Errorf("ApplyExtendedFilter default should be true, got %v", cfg.ApplyExtendedFilter)
	}
	if cfg.TypeDetectionThreshold != 0.5 {
		t.Errorf("TypeDetectionThreshold default should be 0.5, got %v", cfg.TypeDetectionThreshold)
	}
	if cfg.PIIDetectionThreshold != 0.1 {
		t.Errorf("PIIDetectionThreshold default should be 0.1, got %v", cfg.PIIDetectionThreshold)
	}

	// Nested defaults (protocols block omitted → all 5 defaults true).
	if !cfg.Protocols.REST || !cfg.Protocols.GraphQL || !cfg.Protocols.SOAP || !cfg.Protocols.GRPC || !cfg.Protocols.MCP {
		t.Errorf("Protocols defaults should all be true, got %+v", cfg.Protocols)
	}
	// Nested defaults (endpoint_stability omitted → defaults: 2, 300).
	if cfg.EndpointStability.MinCount != 2 || cfg.EndpointStability.MinTime != 300 {
		t.Errorf("EndpointStability defaults should be {2, 300}, got %+v", cfg.EndpointStability)
	}
}

// newAPIDiscoveryConfigResourceData builds a *schema.ResourceData with the resource's
// schema and the given raw HCL state. Used by the unit tests above.
func newAPIDiscoveryConfigResourceData(t *testing.T, raw map[string]any) *schema.ResourceData {
	t.Helper()
	r := resourceWallarmAPIDiscoveryConfig()
	return schema.TestResourceDataRaw(t, r.Schema, raw)
}

// --- Acceptance tests (TF_ACC-gated) ---------------------------------------

// Singleton — tests must run sequentially (no ParallelTest) since they all
// mutate the same per-tenant config record.

func testAccAPIDiscoveryConfigHCL(enabled bool, typeThreshold, piiThreshold float64, disabledApps string) string {
	return fmt.Sprintf(`
resource "wallarm_api_discovery_config" "test" {
  enabled                  = %t
  apply_extended_filter    = true
  type_detection_threshold = %g
  pii_detection_threshold  = %g
  disabled_apps            = %s

  protocols {
    rest    = true
    graphql = true
    soap    = true
    grpc    = true
    mcp     = true
  }

  endpoint_stability {
    min_count = 2
    min_time  = 300
  }
}
`, enabled, typeThreshold, piiThreshold, disabledApps)
}

func TestAccAPIDiscoveryConfig_BasicLifecycle(t *testing.T) {
	const resourceName = "wallarm_api_discovery_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		// CheckDestroy is intentionally nil: the singleton record always exists.
		// Delete is a noop; the API config persists after the test.
		Steps: []resource.TestStep{
			// Step 1: apply with one set of values + empty disabled_apps.
			{
				Config: testAccAPIDiscoveryConfigHCL(true, 0.5, 0.1, "[]"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "apply_extended_filter", "true"),
					resource.TestCheckResourceAttr(resourceName, "type_detection_threshold", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "pii_detection_threshold", "0.1"),
					resource.TestCheckResourceAttr(resourceName, "protocols.0.rest", "true"),
					resource.TestCheckResourceAttr(resourceName, "protocols.0.mcp", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_stability.0.min_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_stability.0.min_time", "300"),
					resource.TestCheckResourceAttr(resourceName, "disabled_apps.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "call_points_storage_limit"),
				),
			},
			// Step 2: mutate enabled + thresholds + populate disabled_apps; in-place update.
			{
				Config: testAccAPIDiscoveryConfigHCL(false, 0.8, 0.2, "[42, 7]"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type_detection_threshold", "0.8"),
					resource.TestCheckResourceAttr(resourceName, "pii_detection_threshold", "0.2"),
					resource.TestCheckResourceAttr(resourceName, "disabled_apps.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "disabled_apps.0", "42"),
					resource.TestCheckResourceAttr(resourceName, "disabled_apps.1", "7"),
				),
			},
			// Step 3: import-verify round-trip.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIDiscoveryConfig_ThresholdOutOfRange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wallarm_api_discovery_config" "bad" {
  type_detection_threshold = 1.5
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected type_detection_threshold to be in the range`),
			},
		},
	})
}
