package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TODO add enum exact too
func TestAccRuleForcedBrowsingRegexp(t *testing.T) {
	const config = `
resource "wallarm_rule_forced_browsing" "wallarm_rule_forced_browsing_regexp" {
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
		CheckDestroy: testAccCheckWallarmRuleForcedBrowsingDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_rule_forced_browsing_regexp", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_forced_browsing.wallarm_rule_forced_browsing_regexp", "action.#", "1"),
				),
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
		CheckDestroy: testAccCheckWallarmRuleForcedBrowsingDestroy,
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
		CheckDestroy: testAccCheckWallarmRuleForcedBrowsingDestroy,
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

func testAccCheckWallarmRuleForcedBrowsingDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_forced_browsing" {
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
			Limit:     1000,
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
