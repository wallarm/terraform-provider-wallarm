package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccRuleDisableStampCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_stamp." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleDisableStampDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleDisableStampBasicConfig(rnd, 1234, "iequal", "stamp.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "1234"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "form_urlencoded"),
					resource.TestCheckResourceAttr(name, "point.1.1", "query"),
				),
			},
		},
	})
}

func TestAccRuleDisableStampCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_stamp." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleDisableStampDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDisableStampCreateRecreate(rnd, 5678),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "5678"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
			{
				Config: testAccRuleDisableStampCreateRecreate(rnd, 5678),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "5678"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
		},
	})
}

func TestAccRuleDisableStampCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_stamp." + rnd
	point := `["header","HOST"],["pollution"]`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleDisableStampDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleDisableStampDefaultBranchConfig(rnd, 9012, point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "stamp", "9012"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckNoResourceAttr(name, "action"),
				),
			},
		},
	})
}

func testWallarmRuleDisableStampBasicConfig(resourceID string, stamp int, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_stamp" "%[1]s" {
  action {
    type = "%[2]s"
    value = "%[3]s"
    point = {
      header = "%[4]s"
    }
  }
  point = [%[5]s]
  stamp = %[6]d
}`, resourceID, actionType, actionValue, actionPoint, point, stamp)
}

func testWallarmRuleDisableStampDefaultBranchConfig(resourceID string, stamp int, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_stamp" "%[1]s" {
  point = [%[2]s]
  stamp = %[3]d
}`, resourceID, point, stamp)
}

func testAccRuleDisableStampCreateRecreate(resourceID string, stamp int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_stamp" "%[1]s" {
  point = [["header", "X-FOOBAR"]]
  stamp = %[2]d
}`, resourceID, stamp)
}

func testAccCheckWallarmRuleDisableStampDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_disable_stamp" {
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
				Type:     []string{"disable_stamp"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Disable Stamp rule still exists")
		}
	}

	return nil
}
