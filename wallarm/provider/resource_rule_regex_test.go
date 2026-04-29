package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleRegexCreateUserAgent(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_regex." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleRegexDestroy,
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
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}
func TestAccRuleRegexCreateOpenDir(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_regex." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleRegexDestroy,
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleRegexCreateNotANumber(rnd, "\\D"),
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
resource "wallarm_rule_regex" %[1]q {
	regex = %[2]q
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
resource "wallarm_rule_regex" %[1]q {
	regex = %[2]q
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
resource "wallarm_rule_regex" %[1]q {
	regex = %[2]q
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
	return testAccCheckHintDestroyed(s, "wallarm_rule_regex")
}
