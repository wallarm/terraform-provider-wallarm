package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccRuleBruteForceCounterCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_bruteforce_counter." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBruteForceCounterCreate(rnd, "b:root"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "counter", "b:root"),
				),
			},
		},
	})
}

func TestAccRuleBruteForceCounterIncorrectName(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_bruteforce_counter." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBruteForceCounterIncorrectName(rnd, "root"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "counter", "root"),
				),
				ExpectError: regexp.MustCompile(`config is invalid: invalid value for counter \(name of the counter always starts with "b:"\)`),
			},
		},
	})
}

func testAccRuleBruteForceCounterCreate(resourceID, counter string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_bruteforce_counter" "%[1]s" {
	counter = "%[2]s"
	action {
		type = "iequal"
		value = "/"
		point = {
			path = 0
		}
	}
}`, resourceID, counter)
}

func testAccRuleBruteForceCounterIncorrectName(resourceID, counter string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_bruteforce_counter" "%[1]s" {
	counter = "%[2]s"
	action {
		type = "iequal"
		value = "/"
		point = {
			path = 0
		}
	}
}`, resourceID, counter)
}
