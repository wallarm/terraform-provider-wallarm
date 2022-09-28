package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	wallarm "github.com/wallarm/wallarm-go"
)

func TestAccRuleVariativeValuesCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_variative_values." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVariativeValuesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVariativeValuesBasicConfig(rnd, "iequal", "variative-values.wallarm.com", "HOST", `["path","99"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "path"),
					resource.TestCheckResourceAttr(name, "point.0.1", "99"),
				),
			},
		},
	})
}

func TestAccRuleVariativeValuesCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_variative_values." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVariativeValuesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleVariativeValuesCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "point.0.0", "action_ext"),
				),
			},
			{
				Config: testAccRuleVariativeValuesCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "point.0.0", "action_ext"),
				),
			},
		},
	})
}

func testWallarmRuleVariativeValuesBasicConfig(resourceID, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_variative_values" "%[1]s" {
  action {
    type = "%[2]s"
    value = "%[3]s"
    point = {
      header = "%[4]s"
    }
  }
  point = [%[5]s]
}`, resourceID, actionType, actionValue, actionPoint, point)
}

func testAccRuleVariativeValuesCreateRecreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_variative_values" "%[1]s" {
  point = [["action_ext"]]
}`, resourceID)
}

func testAccCheckWallarmRuleVariativeValuesDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_variative_values" {
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
				Type:     []string{"variative_values"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Variative values rule still exists")
		}
	}

	return nil
}
