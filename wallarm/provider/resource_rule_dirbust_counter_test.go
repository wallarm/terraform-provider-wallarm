package wallarm

import (
	"fmt"
	// "os"
	"regexp"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleDirbustCounterCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_dirbust_counter." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleDirbustCounterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDirbustCounterCreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "3"),
					resource.TestMatchResourceAttr(name, "counter", regexp.MustCompile("^d:.+")),
				),
				ExpectNonEmptyPlan: false,
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

func testAccRuleDirbustCounterCreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_dirbust_counter" "%[1]s" {
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

func testAccCheckWallarmRuleDirbustCounterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dirbust_counter" {
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
			Limit:     APIListLimit,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ActionID: []int{actionID},
				Type:     []string{"dirbust_counter"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Dirbust counter rule still exists")
		}
	}

	return nil
}
