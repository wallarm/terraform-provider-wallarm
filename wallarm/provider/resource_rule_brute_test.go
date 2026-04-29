package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/wallarm/wallarm-go"
)

// TODO add brute exact too
func TestAccRuleBruteRegexp(t *testing.T) {
	const config = `
resource "wallarm_rule_brute" "wallarm_rule_brute_regexp" {
  mode = "block"

  action {
    type = "iequal"
    value = "wbrute_regexp.example.com"
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBruteDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_regexp", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_regexp", "action.#", "1"),
				),
			},
			{
				ResourceName:            "wallarm_rule_brute.wallarm_rule_brute_regexp",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
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
    value = "wbrute_advanced.example.com"
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
    field    = "status_code"
    value    = ["200"]
    operator = "eq"
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBruteDestroy,
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
    value = "wbrute_arbitrary.example.com"
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
    value    = ["value"]
    operator = "ne"
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBruteDestroy,
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

func testAccRuleBruteUpdateConfig(thresholdCount int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_brute" "update_threshold" {
  mode = "block"

  action {
    type = "iequal"
    value = "wbrute_update.example.com"
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

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = false
    plain_parameters      = false
  }
}
`, thresholdCount)
}

func TestAccRuleBruteUpdateInPlaceThresholdCount(t *testing.T) {
	resourceName := "wallarm_rule_brute.update_threshold"
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBruteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBruteUpdateConfig(5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold.0.count", "5"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceName].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleBruteUpdateConfig(10),
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

func testAccCheckWallarmRuleBruteDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_brute" {
			continue
		}
		ruleID, err := strconv.Atoi(rs.Primary.Attributes["rule_id"])
		if err != nil {
			return fmt.Errorf("invalid rule_id for %s: %w", rs.Primary.ID, err)
		}
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return fmt.Errorf("invalid client_id for %s: %w", rs.Primary.ID, err)
		}

		// OrderBy is required by the API — HintRead returns 400 without it.
		resp, err := api.HintRead(&wallarm.HintRead{
			Limit:   1,
			OrderBy: "updated_at",
			Filter:  &wallarm.HintFilter{Clientid: []int{clientID}, ID: []int{ruleID}},
		})
		if err != nil {
			return fmt.Errorf("checking hint %d still exists: %w", ruleID, err)
		}
		if resp.Body != nil && len(*resp.Body) > 0 {
			return fmt.Errorf("wallarm_rule_brute %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
