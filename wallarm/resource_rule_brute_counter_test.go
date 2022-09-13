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
				Config: testAccRuleBruteForceCounterCreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestMatchResourceAttr(name, "counter", regexp.MustCompile("^b:.*")),
				),
			},
		},
	})
}

func testAccRuleBruteForceCounterCreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_bruteforce_counter" "%[1]s" {
	action {
		type = "iequal"
		value = "/"
		point = {
			path = 0
		}
	}
}`, resourceID)
}
