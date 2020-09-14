package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccRuleRegexCreateUserAgent(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_regex." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleRegexCreateUserAgent(rnd, "^(Mozilla(~(.*d833810e8a84cd2432e95893c36d8bff.*)))$"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "regex", "^(Mozilla(~(.*d833810e8a84cd2432e95893c36d8bff.*)))$"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "USER-AGENT"),
				),
			},
		},
	})
}
func TestAccRuleRegexCreateOpenDir(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_regex." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleRegexCreateOpenDir(rnd, "/[.]git"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "regex", "/[.]git"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header_all"),
				),
			},
			{
				Config: testWallarmRuleRegexCreateOpenDir(rnd, "/[.]env"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "regex", "/[.]env"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header_all"),
				),
			},
		},
	})
}

func TestAccRuleRegexCreateNotANumber(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_regex." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleRegexCreateNotANumber(rnd, "\\\\D"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "regex", "\\D"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "path"),
					resource.TestCheckResourceAttr(name, "point.0.1", "3"),
				),
			},
		},
	})
}

func testWallarmRuleRegexCreateUserAgent(resourceID, regex string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_regex" "%[1]s" {
	regex = "%[2]s"
	experimental = true
	attack_type =  "scanner"
	
	action {
		type = "iequal"
		value = "%[1]s.wallarm-demo.com"
		point = {
			header = "HOST"
		}
	}
	point = [["header", "USER-AGENT"]]
}`, resourceID, regex)
}

func testWallarmRuleRegexCreateOpenDir(resourceID, regex string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_regex" "%[1]s" {
	regex = "%[2]s"
	experimental = false
	attack_type =  "vpatch"
	
	action {
		type = "iequal"
		value = "%[1]s.wallarm-demo.com"
		point = {
			header = "HOST"
		}
	}
	point = [["header_all"]]
}`, resourceID, regex)
}

func testWallarmRuleRegexCreateNotANumber(resourceID, regex string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_regex" "%[1]s" {
	regex = "%[2]s"
	experimental = false
	attack_type =  "scanner"
	
	action {
		type = "iequal"
		value = "%[1]s.wallarm-demo.com"
		point = {
			header = "HOST"
		}
	}
	point = [["path", 3]]
}`, resourceID, regex)
}

func testAccCheckWallarmRuleRegexDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_regex" {
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
				Type:     []string{"regex"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Regular Expression rule still exists")
		}
	}

	return nil
}
