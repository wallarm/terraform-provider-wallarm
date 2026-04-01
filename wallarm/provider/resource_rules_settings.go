package wallarm

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmRulesSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceWallarmRulesSettingsRead,
		CreateContext: resourceWallarmRulesSettingsCreate,
		UpdateContext: resourceWallarmRulesSettingsUpdate,
		DeleteContext: resourceWallarmRulesSettingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "The Client ID to perform changes",
				ValidateFunc: validation.IntBetween(1, math.MaxInt32),
			},

			// min_lom_format and max_lom_format are nullable on the API side:
			// nil means "use server default". Since Terraform SDK v2 TypeInt
			// cannot represent nil, we use 0 as the sentinel:
			//   = 0   → clear to server default (sends JSON null)
			//   = N   → set to N (sends JSON integer)
			//   null  → SDK v2 limitation: treated as "keep current value"
			//           (use 0 instead of null to clear)
			"min_lom_format": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Minimum LOM format version. Set to 0 to use server default (nil).",
			},
			"max_lom_format": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum LOM format version. Set to 0 to use server default (nil).",
			},
			"max_lom_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1025),
			},
			"lom_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"lom_compilation_delay": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxInt32),
			},
			"rules_snapshot_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"rules_snapshot_max_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 99),
			},
			"rules_manipulation_locked": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"heavy_lom": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"parameters_count_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"path_variativity_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"pii_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"request_content_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"open_vulns_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"serialized_data_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"risk_score_algo": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"maximum", "average"}, false),
			},
		},
	}
}

func resourceWallarmRulesSettingsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)

	// Parse client_id from the composite ID on import.
	// ID format: "{clientID}/rules_settings"
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id := d.Id(); strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) == 2 {
			var parsed int
			if _, err := fmt.Sscanf(parts[0], "%d", &parsed); err == nil {
				clientID = parsed
			}
		}
	}

	res, err := client.RulesSettingsRead(clientID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("client_id", clientID)

	if res.Body == nil || res.Body.RuleSettingsParams == nil {
		return nil
	}

	s := res.Body.RuleSettingsParams

	// For nullable fields (min/max_lom_format): API nil → 0 in state (sentinel).
	if s.MinLomFormat != nil {
		d.Set("min_lom_format", *s.MinLomFormat)
	} else {
		d.Set("min_lom_format", 0)
	}
	if s.MaxLomFormat != nil {
		d.Set("max_lom_format", *s.MaxLomFormat)
	} else {
		d.Set("max_lom_format", 0)
	}

	// Standard pointer fields: skip d.Set on nil to keep current state.
	if s.MaxLomSize != nil {
		d.Set("max_lom_size", *s.MaxLomSize)
	}
	if s.LomDisabled != nil {
		d.Set("lom_disabled", *s.LomDisabled)
	}
	if s.LomCompilationDelay != nil {
		d.Set("lom_compilation_delay", *s.LomCompilationDelay)
	}
	if s.RulesSnapshotEnabled != nil {
		d.Set("rules_snapshot_enabled", *s.RulesSnapshotEnabled)
	}
	if s.RulesSnapshotMaxCount != nil {
		d.Set("rules_snapshot_max_count", *s.RulesSnapshotMaxCount)
	}
	if s.RulesManipulationLocked != nil {
		d.Set("rules_manipulation_locked", *s.RulesManipulationLocked)
	}
	if s.HeavyLom != nil {
		d.Set("heavy_lom", *s.HeavyLom)
	}
	if s.ParametersCountWeight != nil {
		d.Set("parameters_count_weight", *s.ParametersCountWeight)
	}
	if s.PathVariativityWeight != nil {
		d.Set("path_variativity_weight", *s.PathVariativityWeight)
	}
	if s.PiiWeight != nil {
		d.Set("pii_weight", *s.PiiWeight)
	}
	if s.RequestContentWeight != nil {
		d.Set("request_content_weight", *s.RequestContentWeight)
	}
	if s.OpenVulnsWeight != nil {
		d.Set("open_vulns_weight", *s.OpenVulnsWeight)
	}
	if s.SerializedDataWeight != nil {
		d.Set("serialized_data_weight", *s.SerializedDataWeight)
	}
	if s.RiskScoreAlgo != nil {
		d.Set("risk_score_algo", *s.RiskScoreAlgo)
	}

	return nil
}

func resourceWallarmRulesSettingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := updateRulesSettings(d, m); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d/rules_settings", clientID))
	return resourceWallarmRulesSettingsRead(ctx, d, m)
}

func resourceWallarmRulesSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := updateRulesSettings(d, m); err != nil {
		return diag.FromErr(err)
	}
	return resourceWallarmRulesSettingsRead(ctx, d, m)
}

func resourceWallarmRulesSettingsDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Settings are a singleton — cannot be deleted, only modified.
	return nil
}

// shouldUpdate returns true if the given key should be included in the API
// request. For new resources (Create), it checks whether the user explicitly
// configured the key. For existing resources (Update), it checks whether
// the value changed.
func shouldUpdate(d *schema.ResourceData, key string) bool {
	if d.IsNewResource() {
		return isConfigured(d, key)
	}
	return d.HasChange(key)
}

// setNullableInt handles min_lom_format and max_lom_format — fields where the
// API accepts null (server default). Since their JSON tags lack omitempty,
// a nil pointer serializes as JSON null and a non-nil pointer as the integer.
//
// We must ALWAYS populate these fields in the request to avoid accidentally
// sending null during unrelated updates. The convention is:
//
//	Terraform 0 → nil pointer  → JSON null  → API clears to server default
//	Terraform N → &N pointer   → JSON N     → API sets to N
//	Not changing → &current    → JSON N     → API keeps current value
func setNullableInt(d *schema.ResourceData, key string, target **int) {
	if shouldUpdate(d, key) {
		// User changed (or configured on Create): send the new value.
		val := d.Get(key).(int)
		if val != 0 {
			*target = &val
		}
		// val == 0: leave nil → JSON null → clear to server default
	} else if !d.IsNewResource() {
		// Not changing on this Update: preserve current state value.
		if current := d.Get(key).(int); current != 0 {
			*target = &current
		}
		// current == 0: already null on API, nil → JSON null → no change
	}
}

func updateRulesSettings(d *schema.ResourceData, m interface{}) error {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return err
	}

	params := &wallarm.RuleSettingsParams{}

	// Nullable fields — must always be populated (see setNullableInt doc).
	setNullableInt(d, "min_lom_format", &params.MinLomFormat)
	setNullableInt(d, "max_lom_format", &params.MaxLomFormat)

	// Standard fields — only sent when changed (omitempty omits nil).
	if shouldUpdate(d, "max_lom_size") {
		val := d.Get("max_lom_size").(int)
		params.MaxLomSize = &val
	}
	if shouldUpdate(d, "lom_disabled") {
		val := d.Get("lom_disabled").(bool)
		params.LomDisabled = &val
	}
	if shouldUpdate(d, "lom_compilation_delay") {
		val := d.Get("lom_compilation_delay").(int)
		params.LomCompilationDelay = &val
	}
	if shouldUpdate(d, "rules_snapshot_enabled") {
		val := d.Get("rules_snapshot_enabled").(bool)
		params.RulesSnapshotEnabled = &val
	}
	if shouldUpdate(d, "rules_snapshot_max_count") {
		val := d.Get("rules_snapshot_max_count").(int)
		params.RulesSnapshotMaxCount = &val
	}
	if shouldUpdate(d, "rules_manipulation_locked") {
		val := d.Get("rules_manipulation_locked").(bool)
		params.RulesManipulationLocked = &val
	}
	if shouldUpdate(d, "heavy_lom") {
		val := d.Get("heavy_lom").(bool)
		params.HeavyLom = &val
	}
	if shouldUpdate(d, "parameters_count_weight") {
		val := d.Get("parameters_count_weight").(int)
		params.ParametersCountWeight = &val
	}
	if shouldUpdate(d, "path_variativity_weight") {
		val := d.Get("path_variativity_weight").(int)
		params.PathVariativityWeight = &val
	}
	if shouldUpdate(d, "pii_weight") {
		val := d.Get("pii_weight").(int)
		params.PiiWeight = &val
	}
	if shouldUpdate(d, "request_content_weight") {
		val := d.Get("request_content_weight").(int)
		params.RequestContentWeight = &val
	}
	if shouldUpdate(d, "open_vulns_weight") {
		val := d.Get("open_vulns_weight").(int)
		params.OpenVulnsWeight = &val
	}
	if shouldUpdate(d, "serialized_data_weight") {
		val := d.Get("serialized_data_weight").(int)
		params.SerializedDataWeight = &val
	}
	if shouldUpdate(d, "risk_score_algo") {
		val := d.Get("risk_score_algo").(string)
		params.RiskScoreAlgo = &val
	}

	_, err = client.RulesSettingsUpdate(params, clientID)
	if err != nil {
		return err
	}

	return nil
}

// isConfigured checks whether the user explicitly set a given key in their
// Terraform configuration. Unlike d.GetOk, this correctly detects zero values
// (false, 0, "") as explicitly configured. Uses d.GetRawConfig() which is
// the SDK v2 replacement for the deprecated GetOkExists.
func isConfigured(d *schema.ResourceData, key string) bool {
	raw := d.GetRawConfig()
	if raw.IsNull() || !raw.IsKnown() {
		return false
	}
	attrs := raw.AsValueMap()
	v, ok := attrs[key]
	return ok && !v.IsNull()
}
