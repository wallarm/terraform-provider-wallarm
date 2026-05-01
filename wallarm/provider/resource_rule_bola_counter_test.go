package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRuleBolaCounterCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_bola_counter." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBolaCounterCreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "3"),
					resource.TestMatchResourceAttr(name, "counter", regexp.MustCompile("^i:.+")),
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

func testAccRuleBolaCounterCreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_bola_counter" %[1]q {
	action {
		type = "absent"
    	point = {
			path = 0
    	}
	}

	action {
		type = "iequal"
    	point = {
			action_name = "login"
    	}
  	}
	action {
		type = "equal"
    	point = {
			action_ext = "aspx"
    	}
  	}
}`, resourceID)
}
