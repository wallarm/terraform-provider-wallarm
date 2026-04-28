package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleIgnoreRegexUpdateInPlaceComment(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_ignore_regex." + rnd
	var firstRuleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleIgnoreRegexUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleIgnoreRegexUpdateCommentConfig(rnd, "second comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "second comment"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[name].Primary.Attributes["rule_id"]
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

func testAccRuleIgnoreRegexUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_regex" %[1]q {
	regex = "/[.]example[.]com[.]php$"
	experimental = true
	attack_type =  "scanner"

	action {
		type = "iequal"
		value = "ignore_regex_comment_update.example.com"
		point = {
			header = "HOST"
		}
	}
	point = [["header_all"]]
}

resource "wallarm_rule_ignore_regex" %[1]q {
	regex_id = wallarm_rule_regex.%[1]s.regex_id
	comment  = %[2]q
	action {
		type = "iequal"
		value = "ignore_regex_comment_update.example.com"
		point = {
			header = "HOST"
		}
	}
	point = [["header", "X-AUTH"]]
}`, resourceID, comment)
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
