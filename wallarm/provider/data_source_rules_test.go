package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceRules(t *testing.T) {
	rnd := generateRandomResourceName(5)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRulesConfig(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceRules("data.wallarm_rules.all"),
				),
			},
		},
	})
}

func TestAccDataSourceRulesFilterType(t *testing.T) {
	rnd := generateRandomResourceName(5)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRulesFilterTypeConfig(rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceRulesType("data.wallarm_rules.filtered", "wallarm_mode"),
				),
			},
		},
	})
}

func testAccCheckDataSourceRules(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		ruleCount, err := strconv.Atoi(rs.Primary.Attributes["rules.#"])
		if err != nil {
			return err
		}

		if ruleCount == 0 {
			return fmt.Errorf("no rules returned by data source")
		}

		return nil
	}
}

func testAccCheckDataSourceRulesType(n, expectedType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ruleCount, err := strconv.Atoi(rs.Primary.Attributes["rules.#"])
		if err != nil {
			return err
		}

		for i := 0; i < ruleCount; i++ {
			ruleType := rs.Primary.Attributes[fmt.Sprintf("rules.%d.type", i)]
			if ruleType != expectedType {
				return fmt.Errorf("rule %d has type %q, expected %q", i, ruleType, expectedType)
			}
		}

		return nil
	}
}

func testAccDataSourceRulesConfig(rnd string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_mode" "%[1]s" {
  mode = "monitoring"
  action {
    type  = "iequal"
    point = {
      header = "HOST"
    }
    value = "tf-test-%[1]s.example.com"
  }
}

data "wallarm_rules" "all" {
  depends_on = [wallarm_rule_mode.%[1]s]
}`, rnd)
}

func testAccDataSourceRulesFilterTypeConfig(rnd string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_mode" "%[1]s" {
  mode = "monitoring"
  action {
    type  = "iequal"
    point = {
      header = "HOST"
    }
    value = "tf-test-%[1]s.example.com"
  }
}

data "wallarm_rules" "filtered" {
  type       = ["wallarm_mode"]
  depends_on = [wallarm_rule_mode.%[1]s]
}`, rnd)
}
