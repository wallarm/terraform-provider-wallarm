package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccRuleAttackRecheckerRewriteCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_attack_rechecker_rewrite." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAttackRecheckerRewriteCreate(rnd, "my.awesome-application.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rules.0", "my.awesome-application.com"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
				),
			},
		},
	})
}
func TestAccRuleAttackRecheckerRewriteCreateWithAction(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_attack_rechecker_rewrite." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAttackRecheckerRewriteCreateWithAction(rnd, "my.awesome-application.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rules.0", "my.awesome-application.com"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleAttackRecheckerRewriteCreate(resourceID, enabled string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_attack_rechecker_rewrite" "%[1]s" {
	rules = ["%[2]s"]
	point = [["header", "HOST"]]
}`, resourceID, enabled)
}

func testAccRuleAttackRecheckerRewriteCreateWithAction(resourceID, enabled string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_attack_rechecker_rewrite" "%[1]s" {
	rules = ["%[2]s"]

	action {
		type = "iequal"
		value = "rewrite.example.com"
		point = {
		  header = "HOST"
		}
	}

	point = [["header", "HOST"]]
}`, resourceID, enabled)
}
