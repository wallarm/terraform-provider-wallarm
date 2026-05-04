package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type wmodeTestingRule struct {
	mode      string
	matchType []string
	value     string
}

func TestAccRuleWmodeCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleWmodeBasicConfig(rnd, "monitoring", "iequal", "wmode_basic.example.com", "HOST"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "monitoring"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleWmodeCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleWmodeDefaultBranchConfig(rnd, "default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "0"),
				),
				ExpectError: ResourceExistsError("[0-9]+/[0-9]+/[0-9]+/\\w+", "wallarm_rule_mode"),
			},
		},
	})
}

func TestAccRuleWmodeCreate_FullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	matchType := []string{"equal", "iequal", "regex", "absent"}
	value := "https://docs.wallarm.com/admin-en/installation-nginx-en/"

	rule := wmodeTestingRule{
		mode:      "off",
		matchType: matchType,
		value:     value,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleWmodeFullSettingsConfig(rnd, rule),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "off"),
					resource.TestCheckResourceAttr(name, "action.#", "9"),
				),
			},
		},
	})
}

func testWallarmRuleWmodeBasicConfig(resourceID, mode, actionType, actionValue, actionPoint string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_mode" %[1]q {
  mode = %[2]q
  action {
    type = %[3]q
    value = %[4]q
    point = {
      header = %[5]q
    }
  }
}`, resourceID, mode, actionType, actionValue, actionPoint)
}

func testWallarmRuleWmodeDefaultBranchConfig(resourceID, mode string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_mode" %[1]q {
	mode = %[2]q
}`, resourceID, mode)
}

func testWallarmRuleWmodeFullSettingsConfig(resourceID string, rule wmodeTestingRule) string {
	equal := rule.matchType[0]
	iequal := rule.matchType[1]
	regex := rule.matchType[2]
	absent := rule.matchType[3]
	return fmt.Sprintf(`
resource "wallarm_rule_mode" %[7]q {

	mode = %[1]q

	action {
		point = {
		  instance = 9
		}
	}

	# Intentionally create a duplicate which is supposed to be removed by Set
	action {
		point = {
		  instance = 9
		}
	}

	action {
		type = %[3]q
		point = {
		  action_name = "wmode"
		}
	}

	action {
		type = %[5]q
		point = {
		  action_ext = ""
		}
	}

	action {
		value = "api"
		type = %[2]q
		point = {
		  path = 1
		}
	}

	action {
		value = "login"
		type = %[3]q
		point = {
		  path = 3
		}
	}

	action {
		type = %[3]q
		point = {
		  method = "PUT"
		}
	}

	action {
		type = %[2]q
		point = {
		  scheme = "http"
		}
	}

	action {
		type = %[2]q
		point = {
		  proto = "1.0"
		}
	}

	action {
		type = %[3]q
		value = %[6]q
		point = {
		  header = "referer"
		}
	}

}`, rule.mode, equal, iequal, regex, absent, rule.value, resourceID)
}

func TestAccRuleModeUpdateInPlaceMode(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleWmodeBasicConfig(rnd, "monitoring", "iequal", "wmode_update.example.com", "HOST"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "monitoring"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testWallarmRuleWmodeBasicConfig(rnd, "block", "iequal", "wmode_update.example.com", "HOST"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "block"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						if newID != firstRuleID {
							return fmt.Errorf("expected rule_id to stay stable on in-place update, was %s now %s", firstRuleID, newID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckWallarmRuleWmodeDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_mode")
}

// TestAccRuleMode_RemovingActiveFromHCLRestoresDefault guards the v2.3.9
// active-asymmetry fix on the shared commonResourceRuleFields.active schema:
// previously Optional+Computed left state stuck at the user's last value when
// the line was removed; now Optional+Default(true) plans
// `false → true` symmetrically. Uses wallarm_rule_mode as a representative
// rule but the schema is shared, so this guards every rule resource.
func TestAccRuleMode_RemovingActiveFromHCLRestoresDefault(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	cfg := func(extra string) string {
		return fmt.Sprintf(`
resource "wallarm_rule_mode" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "active-restore.example.com"
    point = { header = "HOST" }
  }
%[2]s
}
`, rnd, extra)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg("  active = false"),
				Check:  resource.TestCheckResourceAttr(name, "active", "false"),
			},
			{
				// Remove the line → schema default (true) wins, plan shows
				// diff `false → true`, Update restores the API default.
				Config: cfg(""),
				Check:  resource.TestCheckResourceAttr(name, "active", "true"),
			},
		},
	})
}
