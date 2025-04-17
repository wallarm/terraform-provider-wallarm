package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccRuleIgnoreRegexCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_ignore_regex." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleIgnoreRegexCreate(rnd, "/[.]example[.]com[.]php$"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-AUTH"),
				),
			},
		},
	})
}

func testAccRuleIgnoreRegexCreate(resourceID, regex string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_regex" "%[1]s" {
	regex = "%[2]s"
	experimental = true
	attack_type =  "scanner"
	
	action {
		type = "iequal"
		value = "%[1]s.wallarm.com"
		point = {
			header = "HOST"
		}
	}
	point = [["header_all"]]
}

resource "wallarm_rule_ignore_regex" "%[1]s" {
	regex_id =  wallarm_rule_regex.%[1]s.regex_id
	action {
		type = "iequal"
		value = "%[1]s.wallarm.com"
		point = {
			header = "HOST"
		}
	}
	point = [["header", "X-AUTH"]]
}`, resourceID, regex)
}
