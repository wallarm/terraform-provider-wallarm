package wallarm

import (
	"fmt"
	"math"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

	if v, ok := d.GetOk("min_lom_format"); ok {
		value := v.(int)
		params.MinLomFormat = &value
	}

	if v, ok := d.GetOk("max_lom_format"); ok {
		if val, ok := v.(int); ok {
			params.MaxLomFormat = &val
		}
	}

	if v, ok := d.GetOk("max_lom_size"); ok {
		if val, ok := v.(int); ok {
			params.MaxLomSize = &val
		}
	}

	if v, ok := d.GetOk("lom_disabled"); ok {
		if val, ok := v.(bool); ok {
			params.LomDisabled = &val
		}
	}

	if v, ok := d.GetOk("lom_compilation_delay"); ok {
		if val, ok := v.(int); ok {
			params.LomCompilationDelay = &val
		}
	}

	if v, ok := d.GetOk("rules_snapshot_enabled"); ok {
		if val, ok := v.(bool); ok {
			params.RulesSnapshotEnabled = &val
		}
	}

	if v, ok := d.GetOk("rules_snapshot_max_count"); ok {
		if val, ok := v.(int); ok {
			params.RulesSnapshotMaxCount = &val
		}
	}

	if v, ok := d.GetOk("rules_manipulation_locked"); ok {
		if val, ok := v.(bool); ok {
			params.RulesManipulationLocked = &val
		}
	}

	if v, ok := d.GetOk("heavy_lom"); ok {
		if val, ok := v.(bool); ok {
			params.HeavyLom = &val
		}
	}

	if v, ok := d.GetOk("parameters_count_weight"); ok {
		if val, ok := v.(int); ok {
			params.ParametersCountWeight = &val
		}
	}

	if v, ok := d.GetOk("path_variativity_weight"); ok {
		if val, ok := v.(int); ok {
			params.PathVariativityWeight = &val
		}
	}

	if v, ok := d.GetOk("pii_weight"); ok {
		if val, ok := v.(int); ok {
			params.PiiWeight = &val
		}
	}

	if v, ok := d.GetOk("request_content_weight"); ok {
		if val, ok := v.(int); ok {
			params.RequestContentWeight = &val
		}
	}

	if v, ok := d.GetOk("open_vulns_weight"); ok {
		if val, ok := v.(int); ok {
			params.OpenVulnsWeight = &val
		}
	}

	if v, ok := d.GetOk("serialized_data_weight"); ok {
		if val, ok := v.(int); ok {
			params.SerializedDataWeight = &val
		}
	}

	if v, ok := d.GetOk("risk_score_algo"); ok {
		if val, ok := v.(string); ok {
			params.RiskScoreAlgo = &val
		}
	}

	_, err := client.RulesSettingsUpdate(params, clientID)
	if err != nil {
		return err
	}

	return resourceWallarmRulesSettingsRead(d, m)
}
