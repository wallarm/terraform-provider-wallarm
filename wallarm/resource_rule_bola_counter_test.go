package wallarm

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccRuleBolaCounterCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_bola_counter." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBolaCounterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBolaCounterCreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "3"),
					resource.TestMatchResourceAttr(name, "counter", regexp.MustCompile("^d:.+")),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRuleBolaCounterCreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_bola_counter" "%[1]s" {
	comment = "This is a comment for a test case"

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

func testAccCheckWallarmRuleBolaCounterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bola_counter" {
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
				Type:     []string{"bola_counter"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Bola counter rule still exists")
		}
	}

	return nil
}
