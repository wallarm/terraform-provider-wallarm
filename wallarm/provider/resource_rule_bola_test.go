package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleBolaExact(t *testing.T) {
	const config = `
resource "wallarm_rule_bola" "wallarm_rule_bola_exact" {
  mode = "block"

  action {
    type = "iequal"
    value = "bola_exact.example.com"
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
		CheckDestroy:             testAccCheckWallarmRuleBolaDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_exact", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_exact", "action.#", "1"),
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_exact", "enumerated_parameters.0.mode", "exact"),
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_exact", "enumerated_parameters.0.points.0.sensitive", "false"),
				),
			},
			{
				ResourceName:            "wallarm_rule_bola.wallarm_rule_bola_exact",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleBolaRegexp(t *testing.T) {
	const config = `
resource "wallarm_rule_bola" "wallarm_rule_bola_regexp" {
  mode = "block"

  action {
    type = "iequal"
    value = "bola_regexp.example.com"
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
		CheckDestroy:             testAccCheckWallarmRuleBolaDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_regexp", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_regexp", "action.#", "1"),
				),
			},
			{
				ResourceName:            "wallarm_rule_bola.wallarm_rule_bola_regexp",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleBolaWithAdvancedConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_bola" "wallarm_rule_bola_advanced_conditions" {
  mode = "block"

  action {
    type = "iequal"
    value = "bola_advanced.example.com"
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
    value    = ["400"]
    operator = "eq"
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBolaDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_advanced_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_advanced_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleBolaWithArbitraryConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_bola" "wallarm_rule_bola_arbitrary_conditions" {
  mode = "block"

  action {
    type = "iequal"
    value = "bola_arbitrary.example.com"
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
		CheckDestroy:             testAccCheckWallarmRuleBolaDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_arbitrary_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_bola.wallarm_rule_bola_arbitrary_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleBolaUpdateConfig(count int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_bola" "update_count" {
  mode = "block"

  action {
    type = "iequal"
    value = "bola_update.example.com"
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
`, count)
}

func TestAccRuleBolaUpdateInPlaceThresholdCount(t *testing.T) {
	resourceName := "wallarm_rule_bola.update_count"
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBolaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBolaUpdateConfig(5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold.0.count", "5"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceName].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleBolaUpdateConfig(10),
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

func testAccCheckWallarmRuleBolaDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_bola")
}
