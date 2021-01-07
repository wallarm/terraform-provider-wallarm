package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

func testAccCheckWallarmRuleIgnoreRegexDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_ignore_regex" {
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
				Type:     []string{"disable_regex"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Disable Regular Expression rule still exists")
		}
	}

	return nil
}
