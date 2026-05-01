package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleBruteExact(t *testing.T) {
	const config = `
resource "wallarm_rule_brute" "wallarm_rule_brute_exact" {
  mode = "block"

  action {
    type = "iequal"
    value = "brute_exact.example.com"
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
    mode = "exact"
    points {
      point     = ["header", "REFERER"]
      sensitive = false
    }
    points {
      point     = ["get", "id"]
      sensitive = true
    }
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
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_exact", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_exact", "action.#", "1"),
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_exact", "enumerated_parameters.0.mode", "exact"),
					resource.TestCheckResourceAttr("wallarm_rule_brute.wallarm_rule_brute_exact", "enumerated_parameters.0.points.0.sensitive", "false"),
				),
			},
			{
				ResourceName:            "wallarm_rule_brute.wallarm_rule_brute_exact",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

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
	return testAccCheckHintDestroyed(s, "wallarm_rule_brute")
}

// TestAccRuleBrute_RegexpRejectsPoints verifies plan-time rejection of
// `points` populated alongside `mode = "regexp"`. Mapper would silently
// drop the field on PUT — perpetual diff. PlanOnly + ExpectError, no API contact.
func TestAccRuleBrute_RegexpRejectsPoints(t *testing.T) {
	rnd := generateRandomResourceName(5)
	config := fmt.Sprintf(`
resource "wallarm_rule_brute" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "brute_regexp_reject.example.com"
    point = { header = "HOST" }
  }
  reaction { block_by_ip = 600 }
  threshold {
    count  = 5
    period = 30
  }
  enumerated_parameters {
    mode          = "regexp"
    name_regexps  = ["foo"]
    value_regexps = ["bar"]
    points {
      point     = ["header", "REFERER"]
      sensitive = true
    }
  }
}`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("`points` not allowed when mode = \"regexp\""),
			},
		},
	})
}
