package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccRuleVariativeKeysCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_variative_keys." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVariativeKeysDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVariativeKeysBasicConfig(rnd, "iequal", "variative-keys.wallarm.com", "HOST", `["post"],["form_urlencoded_all"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "form_urlencoded_all"),
				),
			},
		},
	})
}

func TestAccRuleVariativeKeysCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_variative_keys." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVariativeKeysDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleVariativeKeysCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "json_doc"),
					resource.TestCheckResourceAttr(name, "point.2.0", "hash_all"),
				),
			},
			{
				Config: testAccRuleVariativeKeysCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "json_doc"),
					resource.TestCheckResourceAttr(name, "point.2.0", "hash_all"),
				),
			},
		},
	})
}

func testWallarmRuleVariativeKeysBasicConfig(resourceID, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_variative_keys" "%[1]s" {
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

func testAccRuleVariativeKeysCreateRecreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_variative_keys" "%[1]s" {
  point = [["post"],["json_doc"],["hash_all"]]
}`, resourceID)
}

func testAccCheckWallarmRuleVariativeKeysDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_variative_keys" {
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
				Type:     []string{"variative_keys"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Variative Keys rule still exists")
		}
	}

	return nil
}
