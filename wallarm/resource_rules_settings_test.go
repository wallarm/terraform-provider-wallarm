package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmRulesSettings(t *testing.T) {
	rnd := generateRandomResourceName(10)
	resourceName := "wallarm_rules_settings." + rnd

	attrs := map[string]string{
		"min_lom_format":            "53",
		"max_lom_format":            "54",
		"max_lom_size":              "10000",
		"lom_disabled":              "false",
		"lom_compilation_delay":     "1000",
		"rules_snapshot_enabled":    "true",
		"rules_snapshot_max_count":  "55",
		"rules_manipulation_locked": "true",
		"heavy_lom":                 "true",
		"parameters_count_weight":   "1",
		"path_variativity_weight":   "2",
		"pii_weight":                "3",
		"request_content_weight":    "4",
		"open_vulns_weight":         "5",
		"serialized_data_weight":    "6",
		"risk_score_algo":           "maximum",
		"pii_fallback":              "true",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRulesSettingsConfig(rnd, attrs),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "min_lom_format", attrs["min_lom_format"]),
					resource.TestCheckResourceAttr(resourceName, "max_lom_format", attrs["max_lom_format"]),
					resource.TestCheckResourceAttr(resourceName, "max_lom_size", attrs["max_lom_size"]),
					resource.TestCheckResourceAttr(resourceName, "lom_disabled", attrs["lom_disabled"]),
					resource.TestCheckResourceAttr(resourceName, "lom_compilation_delay", attrs["lom_compilation_delay"]),
					resource.TestCheckResourceAttr(resourceName, "rules_snapshot_enabled", attrs["rules_snapshot_enabled"]),
					resource.TestCheckResourceAttr(resourceName, "rules_snapshot_max_count", attrs["rules_snapshot_max_count"]),
					resource.TestCheckResourceAttr(resourceName, "rules_manipulation_locked", attrs["rules_manipulation_locked"]),
					resource.TestCheckResourceAttr(resourceName, "heavy_lom", attrs["heavy_lom"]),
					resource.TestCheckResourceAttr(resourceName, "parameters_count_weight", attrs["parameters_count_weight"]),
					resource.TestCheckResourceAttr(resourceName, "path_variativity_weight", attrs["path_variativity_weight"]),
					resource.TestCheckResourceAttr(resourceName, "pii_weight", attrs["pii_weight"]),
					resource.TestCheckResourceAttr(resourceName, "request_content_weight", attrs["request_content_weight"]),
					resource.TestCheckResourceAttr(resourceName, "open_vulns_weight", attrs["open_vulns_weight"]),
					resource.TestCheckResourceAttr(resourceName, "serialized_data_weight", attrs["serialized_data_weight"]),
					resource.TestCheckResourceAttr(resourceName, "risk_score_algo", attrs["risk_score_algo"]),
					resource.TestCheckResourceAttr(resourceName, "pii_fallback", attrs["pii_fallback"]),
				),
			},
		},
	})
}

func testWallarmRulesSettingsConfig(resourceID string, attrs map[string]string) string {
	return fmt.Sprintf(`
resource "wallarm_rules_settings" "%[1]s" {
  min_lom_format = %[2]s
	max_lom_format = %[3]s
	max_lom_size = %[4]s
	lom_disabled = %[5]s
	lom_compilation_delay = %[6]s
	rules_snapshot_enabled = %[7]s
	rules_snapshot_max_count = %[8]s
	rules_manipulation_locked = %[9]s
	heavy_lom = %[10]s
	parameters_count_weight = %[11]s
	path_variativity_weight = %[12]s
	pii_weight = %[13]s
	request_content_weight = %[14]s
	open_vulns_weight = %[15]s
	serialized_data_weight = %[16]s
	risk_score_algo = "%[17]s"
	pii_fallback = %[18]s
}`, resourceID,
		attrs["min_lom_format"],
		attrs["max_lom_format"],
		attrs["max_lom_size"],
		attrs["lom_disabled"],
		attrs["lom_compilation_delay"],
		attrs["rules_snapshot_enabled"],
		attrs["rules_snapshot_max_count"],
		attrs["rules_manipulation_locked"],
		attrs["heavy_lom"],
		attrs["parameters_count_weight"],
		attrs["path_variativity_weight"],
		attrs["pii_weight"],
		attrs["request_content_weight"],
		attrs["open_vulns_weight"],
		attrs["serialized_data_weight"],
		attrs["risk_score_algo"],
		attrs["pii_fallback"],
	)
}
