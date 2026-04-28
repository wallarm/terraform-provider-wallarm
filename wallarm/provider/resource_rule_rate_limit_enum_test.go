package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO add enum exact too
func TestAccRuleRateLimitEnumRegexp(t *testing.T) {
	const config = `
resource "wallarm_rule_rate_limit_enum" "wallarm_rule_rate_limit_enum_regexp" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "wenum.wallarm.com"
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRateLimitEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_rate_limit_enum.wallarm_rule_rate_limit_enum_regexp", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_rate_limit_enum.wallarm_rule_rate_limit_enum_regexp", "action.#", "1"),
				),
			},
			{
				ResourceName:            "wallarm_rule_rate_limit_enum.wallarm_rule_rate_limit_enum_regexp",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleRateLimitEnumWithAdvancedConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_rate_limit_enum" "wallarm_rule_rate_limit_enum_advanced_conditions" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "wenum.wallarm.com"
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRateLimitEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_rate_limit_enum.wallarm_rule_rate_limit_enum_advanced_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_rate_limit_enum.wallarm_rule_rate_limit_enum_advanced_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleRateLimitEnumWithArbitraryConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_rate_limit_enum" "wallarm_rule_rate_limit_enum_arbitrary_conditions" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "wenum.wallarm.com"
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRateLimitEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_rate_limit_enum.wallarm_rule_rate_limit_enum_arbitrary_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_rate_limit_enum.wallarm_rule_rate_limit_enum_arbitrary_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleRateLimitEnumUpdateConfig(count int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_rate_limit_enum" "update_count" {
  mode = "block"

  action {
    type = "iequal"
    value = "wrlenum_update.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = %[1]d
    period = 30
  }
}
`, count)
}

func TestAccRuleRateLimitEnumUpdateInPlaceThresholdCount(t *testing.T) {
	resourceName := "wallarm_rule_rate_limit_enum.update_count"
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleRateLimitEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleRateLimitEnumUpdateConfig(5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold.0.count", "5"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceName].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleRateLimitEnumUpdateConfig(10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold.0.count", "10"),
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

func testAccCheckWallarmRuleRateLimitEnumDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_rate_limit_enum" {
			continue
		}

		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return err
		}
		actionID, err := strconv.Atoi(rs.Primary.Attributes["action_id"])
		if err != nil {
			return err
		}

		hint := &wallarm.HintRead{
			Limit:     APIListLimit,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ActionID: []int{actionID},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && rule != nil && len(*rule.Body) > 0 {
			return fmt.Errorf("Wallarm Mode Rule still exists")
		}
	}

	return nil
}
