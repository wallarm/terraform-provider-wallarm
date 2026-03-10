package wallarm

import (
	"fmt"
	"math"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmRulesSettings() *schema.Resource {
	return &schema.Resource{
		Read:   resourceWallarmRulesSettingsRead,
		Create: resourceWallarmRulesSettingsCreate,
		Update: resourceWallarmRulesSettingsUpdate,
		Delete: resourceWallarmRulesSettingsDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "The Client ID to perform changes",
				ValidateFunc: validation.IntBetween(1, math.MaxInt32),
			},
			"min_lom_format": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, math.MaxInt32),
			},
			"max_lom_format": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, math.MaxInt32),
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
				Computed: false,
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

func resourceWallarmRulesSettingsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	res, err := client.RulesSettingsRead(clientID)
	if err != nil {
		return err
	}

	d.Set("min_lom_format", res.Body.MinLomFormat)

	d.Set("max_lom_format", res.Body.MaxLomFormat)

	d.Set("max_lom_size", res.Body.MaxLomSize)

	d.Set("lom_disabled", res.Body.LomDisabled)

	d.Set("lom_compilation_delay", res.Body.LomCompilationDelay)

	d.Set("rules_snapshot_enabled", res.Body.RulesSnapshotEnabled)

	d.Set("rules_snapshot_max_count", res.Body.RulesSnapshotMaxCount)

	d.Set("rules_manipulation_locked", res.Body.RulesManipulationLocked)

	d.Set("heavy_lom", res.Body.HeavyLom)

	d.Set("parameters_count_weight", res.Body.ParametersCountWeight)

	d.Set("path_variativity_weight", res.Body.PathVariativityWeight)

	d.Set("pii_weight", res.Body.PiiWeight)

	d.Set("request_content_weight", res.Body.RequestContentWeight)

	d.Set("open_vulns_weight", res.Body.OpenVulnsWeight)

	d.Set("serialized_data_weight", res.Body.SerializedDataWeight)

	d.Set("risk_score_algo", res.Body.RiskScoreAlgo)

	return nil
}

func resourceWallarmRulesSettingsCreate(d *schema.ResourceData, m interface{}) error {
	clientID := retrieveClientID(d)

	err := updateRulesSettings(d, m)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%d/rules_settings", clientID)
	d.SetId(id)
	return resourceWallarmRulesSettingsRead(d, m)
}

func resourceWallarmRulesSettingsUpdate(d *schema.ResourceData, m interface{}) error {
	return updateRulesSettings(d, m)
}

func resourceWallarmRulesSettingsDelete(_ *schema.ResourceData, _ interface{}) error {
	return nil
}

// nolint:gocyclo
func updateRulesSettings(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)

	params := &wallarm.RuleSettingsParams{}

	// Use d.HasChange for the Update path so that setting a bool to false
	// or an int to 0 actually gets sent to the API. d.GetOk treats zero
	// values as "not set" and silently skips them.
	//
	// For the Create path (d.IsNewResource()), use isConfigured which checks
	// d.GetRawConfig() — the SDK v2 replacement for the deprecated GetOkExists.
	// This correctly detects zero values (false, 0, "") as explicitly configured.

	if d.IsNewResource() {
		// Create: send every field the user configured
		if isConfigured(d, "min_lom_format") {
			val := d.Get("min_lom_format").(int)
			params.MinLomFormat = &val
		}
		if isConfigured(d, "max_lom_format") {
			val := d.Get("max_lom_format").(int)
			params.MaxLomFormat = &val
		}
		if isConfigured(d, "max_lom_size") {
			val := d.Get("max_lom_size").(int)
			params.MaxLomSize = &val
		}
		if isConfigured(d, "lom_disabled") {
			val := d.Get("lom_disabled").(bool)
			params.LomDisabled = &val
		}
		if isConfigured(d, "lom_compilation_delay") {
			val := d.Get("lom_compilation_delay").(int)
			params.LomCompilationDelay = &val
		}
		if isConfigured(d, "rules_snapshot_enabled") {
			val := d.Get("rules_snapshot_enabled").(bool)
			params.RulesSnapshotEnabled = &val
		}
		if isConfigured(d, "rules_snapshot_max_count") {
			val := d.Get("rules_snapshot_max_count").(int)
			params.RulesSnapshotMaxCount = &val
		}
		if isConfigured(d, "rules_manipulation_locked") {
			val := d.Get("rules_manipulation_locked").(bool)
			params.RulesManipulationLocked = &val
		}
		if isConfigured(d, "heavy_lom") {
			val := d.Get("heavy_lom").(bool)
			params.HeavyLom = &val
		}
		if isConfigured(d, "parameters_count_weight") {
			val := d.Get("parameters_count_weight").(int)
			params.ParametersCountWeight = &val
		}
		if isConfigured(d, "path_variativity_weight") {
			val := d.Get("path_variativity_weight").(int)
			params.PathVariativityWeight = &val
		}
		if isConfigured(d, "pii_weight") {
			val := d.Get("pii_weight").(int)
			params.PiiWeight = &val
		}
		if isConfigured(d, "request_content_weight") {
			val := d.Get("request_content_weight").(int)
			params.RequestContentWeight = &val
		}
		if isConfigured(d, "open_vulns_weight") {
			val := d.Get("open_vulns_weight").(int)
			params.OpenVulnsWeight = &val
		}
		if isConfigured(d, "serialized_data_weight") {
			val := d.Get("serialized_data_weight").(int)
			params.SerializedDataWeight = &val
		}
		if isConfigured(d, "risk_score_algo") {
			val := d.Get("risk_score_algo").(string)
			params.RiskScoreAlgo = &val
		}
	} else {
		// Update: send only fields that changed, using d.Get to capture zero values
		if d.HasChange("min_lom_format") {
			val := d.Get("min_lom_format").(int)
			params.MinLomFormat = &val
		}
		if d.HasChange("max_lom_format") {
			val := d.Get("max_lom_format").(int)
			params.MaxLomFormat = &val
		}
		if d.HasChange("max_lom_size") {
			val := d.Get("max_lom_size").(int)
			params.MaxLomSize = &val
		}
		if d.HasChange("lom_disabled") {
			val := d.Get("lom_disabled").(bool)
			params.LomDisabled = &val
		}
		if d.HasChange("lom_compilation_delay") {
			val := d.Get("lom_compilation_delay").(int)
			params.LomCompilationDelay = &val
		}
		if d.HasChange("rules_snapshot_enabled") {
			val := d.Get("rules_snapshot_enabled").(bool)
			params.RulesSnapshotEnabled = &val
		}
		if d.HasChange("rules_snapshot_max_count") {
			val := d.Get("rules_snapshot_max_count").(int)
			params.RulesSnapshotMaxCount = &val
		}
		if d.HasChange("rules_manipulation_locked") {
			val := d.Get("rules_manipulation_locked").(bool)
			params.RulesManipulationLocked = &val
		}
		if d.HasChange("heavy_lom") {
			val := d.Get("heavy_lom").(bool)
			params.HeavyLom = &val
		}
		if d.HasChange("parameters_count_weight") {
			val := d.Get("parameters_count_weight").(int)
			params.ParametersCountWeight = &val
		}
		if d.HasChange("path_variativity_weight") {
			val := d.Get("path_variativity_weight").(int)
			params.PathVariativityWeight = &val
		}
		if d.HasChange("pii_weight") {
			val := d.Get("pii_weight").(int)
			params.PiiWeight = &val
		}
		if d.HasChange("request_content_weight") {
			val := d.Get("request_content_weight").(int)
			params.RequestContentWeight = &val
		}
		if d.HasChange("open_vulns_weight") {
			val := d.Get("open_vulns_weight").(int)
			params.OpenVulnsWeight = &val
		}
		if d.HasChange("serialized_data_weight") {
			val := d.Get("serialized_data_weight").(int)
			params.SerializedDataWeight = &val
		}
		if d.HasChange("risk_score_algo") {
			val := d.Get("risk_score_algo").(string)
			params.RiskScoreAlgo = &val
		}
	}

	_, err := client.RulesSettingsUpdate(params, clientID)
	if err != nil {
		return err
	}

	return resourceWallarmRulesSettingsRead(d, m)
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
