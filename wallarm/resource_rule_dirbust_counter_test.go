package wallarm

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccRuleDirbustCounterCreate(t *testing.T) {
	if os.Getenv("WALLARM_EXTRA_PERMISSIONS") == "" {
		t.Skip("Skipping not test as it requires WALLARM_EXTRA_PERMISSIONS set")
	}
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
					//resource.TestMatchResourceAttr(name, "counter", regexp.MustCompile("^d:.+")),
				),
				//ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRuleDirbustCounterCreate(resourceID string) string {
	//return `resource "wallarm_rule_dirbust_counter" "dirbust_counter" {
	//
	//}
	//
	//resource "wallarm_trigger" "dirbust_trigger" {
	// template_id = "forced_browsing_started"
	//
	// filters {
	//   filter_id = "hint_tag"
	//   operator = "eq"
	//   value = [wallarm_rule_dirbust_counter.dirbust_counter.counter]
	// }
	//
	// actions {
	//   action_id = "mark_as_brute"
	// }
	//
	// actions {
	//   action_id = "block_ips"
	//   lock_time = 2592000
	// }
	//
	// threshold = {
	//   period = 30
	//   operator = "gt"
	//   count = 30
	// }
	//}`
	return fmt.Sprintf(`
resource "wallarm_rule_dirbust_counter" "%[1]s" {
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

func testAccCheckWallarmRuleDirbustCounterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

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
			Limit:     1000,
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
