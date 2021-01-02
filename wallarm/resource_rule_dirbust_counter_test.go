package wallarm

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	wallarm "github.com/416e64726579/wallarm-go"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
				Config: testAccRuleDirbustCounterCreate(rnd, "d:login"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "3"),
					resource.TestCheckResourceAttr(name, "counter", "d:login"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRuleDirbustCounterIncorrectName(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_dirbust_counter." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDirbustCounterIncorrectName(rnd, "aspx"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "counter", "aspx"),
				),
				ExpectError: regexp.MustCompile(`config is invalid: invalid value for counter \(name of the counter always starts with "d:"\)`),
			},
		},
	})
}

func testAccRuleDirbustCounterCreate(resourceID, counter string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_dirbust_counter" "%[1]s" {
	counter = "%[2]s"

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
}`, resourceID, counter)
}

func testAccRuleDirbustCounterIncorrectName(resourceID, counter string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_dirbust_counter" "%[1]s" {
	counter = "%[2]s"
	action {
    	type = "iequal"
    	point = {
      		action_ext = "aspx"
    	}
  	}
}`, resourceID, counter)
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
