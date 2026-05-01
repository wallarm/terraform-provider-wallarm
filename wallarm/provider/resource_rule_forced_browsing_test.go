package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO add enum exact too
func TestAccRuleForcedBrowsingRegexp(t *testing.T) {
	const config = `
resource "wallarm_rule_forced_browsing" "wallarm_rule_forced_browsing_regexp" {
  mode = "block"

  action {
    type = "iequal"
    value = "forced_browsing_regexp.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = 5
    period = 30
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleForcedBrowsingDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_rule_forced_browsing_regexp", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_rule_forced_browsing_regexp", "action.#", "1"),
				),
			},
			{
				ResourceName:            "wallarm_rule_forced_browsing.wallarm_rule_forced_browsing_regexp",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleForcedBrowsingWithAdvancedConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_forced_browsing" "wallarm_forced_browsing_advanced_conditions" {
  mode = "block"

  action {
    type = "iequal"
    value = "forced_browsing_advanced.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = 5
    period = 30
  }

  advanced_conditions {
    field    = "status_code"
    value    = ["200"]
    operator = "eq"
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleForcedBrowsingDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_forced_browsing_advanced_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_forced_browsing_advanced_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleForcedBrowsingWithArbitraryConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_forced_browsing" "wallarm_forced_browsing_arbitrary_conditions" {
  mode = "block"

  action {
    type = "iequal"
    value = "forced_browsing_arbitrary.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000

  }

  threshold {
    count = 5
    period = 30
  }

  arbitrary_conditions {
    point = [["header", "X-LOGIN"]]
    value    = ["value"]
    operator = "ne"
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleForcedBrowsingDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_forced_browsing_arbitrary_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_forced_browsing_arbitrary_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleForcedBrowsingUpdateConfig(period int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_forced_browsing" "update_period" {
  mode = "block"

  action {
    type = "iequal"
    value = "forced_browsing_update.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = 5
    period = %[1]d
  }
}
`, period)
}

func TestAccRuleForcedBrowsingUpdateInPlaceThresholdPeriod(t *testing.T) {
	resourceName := "wallarm_rule_forced_browsing.update_period"
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleForcedBrowsingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleForcedBrowsingUpdateConfig(30),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold.0.period", "30"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceName].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleForcedBrowsingUpdateConfig(60),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold.0.period", "60"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[resourceName].Primary.Attributes["rule_id"]
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

func testAccCheckWallarmRuleForcedBrowsingDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_forced_browsing")
}
