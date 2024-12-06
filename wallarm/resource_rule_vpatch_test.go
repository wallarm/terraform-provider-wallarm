package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccRuleVpatchCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchBasicConfig(rnd, "xss", "iequal", "vpatch.wallarm.com", "HOST", "get_all"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "xss"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "get_all"),
				),
			},
		},
	})
}

func TestAccRuleVpatchCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	point := `["get", "query"]`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchDefaultBranchConfig(rnd, `"crlf"`, point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "crlf"),
					resource.TestCheckResourceAttr(name, "point.0.0", "get"),
					resource.TestCheckResourceAttr(name, "point.0.1", "query"),
					resource.TestCheckNoResourceAttr(name, "action"),
				),
			},
		},
	})
}

func testWallarmRuleVpatchBasicConfig(resourceID, attackType, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" "%[1]s" {
  attack_type = "%[2]s"
  action {
    type = "%[3]s"
    value = "%[4]s"
    point = {
      header = "%[5]s"
    }
  }
  point = [["%[6]s"]]
}`, resourceID, attackType, actionType, actionValue, actionPoint, point)
}

func testWallarmRuleVpatchDefaultBranchConfig(resourceID, attackType, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" "%[1]s" {
	attack_type = %[2]s
	point = [%[3]s]
}`, resourceID, attackType, point)
}

func testAccCheckWallarmRuleVpatchDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_vpatch" {
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
				Type:     []string{"vpatch"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Virtual Patch rule still exists")
		}
	}

	return nil
}
