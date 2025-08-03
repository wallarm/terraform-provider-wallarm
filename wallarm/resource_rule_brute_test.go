package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TODO add brute exact too
func TestAccRuleBruteRegexp(t *testing.T) {
	const config = `
resource "wallarm_rule_brute" "wallarm_rule_brute_regexp" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "wbrute.wallarm.com"
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

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = false
    plain_parameters      = false
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBruteDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_regexp", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_regexp", "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleBruteWithAdvancedConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_brute" "wallarm_rule_brute_advanced_conditions" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "wbrute.wallarm.com"
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

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = false
    plain_parameters      = false
  }

  advanced_conditions {
    field    = "proxy_type"
    value    = ["ABC"]
    operator = "eq"
  }

  advanced_conditions {
    field    = "datacenter"
    value    = ["abc"]
    operator = "ne"
  }

}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBruteDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_advanced_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_advanced_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleBruteWithArbitraryConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_brute" "wallarm_rule_brute_arbitrary_conditions" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "wbrute.wallarm.com"
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

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = false
    plain_parameters      = false
  }

  arbitrary_conditions {
    point = [["header", "X-LOGIN"]]
    value    = ["datacenter_value"]
    operator = "ne"
  }

}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBruteDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_arbitrary_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_arbitrary_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func testAccCheckWallarmRuleBruteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_brute" {
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
