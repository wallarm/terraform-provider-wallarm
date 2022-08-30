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

func TestAccRuleDisableAttackTypeCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_attack_type." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleDisableAttackTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleDisableAttackTypeBasicConfig(rnd, "sqli", "iequal", "attack-types.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "sqli"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "form_urlencoded"),
					resource.TestCheckResourceAttr(name, "point.1.1", "query"),
				),
			},
		},
	})
}

func TestAccRuleDisableAttackTypeCreate_IncorrectAttackType(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmRuleDisableAttackTypeBasicConfig(rnd, "incorrect", "iequal", "attack-types.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				ExpectError: regexp.MustCompile(`config is invalid: expected attack_type to be one of \[sqli xss rce ptrav crlf nosqli xxe ldapi scanner ssti ssi mail_injection vpatch\], got incorrect`),
			},
		},
	})
}

func TestAccRuleDisableAttackTypeCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_attack_type." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleDisableAttackTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDisableAttackTypeCreateRecreate(rnd, "xss"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "xss"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
			{
				Config: testAccRuleDisableAttackTypeCreateRecreate(rnd, "xss"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "xss"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "X-FOOBAR"),
				),
			},
		},
	})
}

func TestAccRuleDisableAttackTypeCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_attack_type." + rnd
	point := `["header","HOST"],["pollution"]`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleDisableAttackTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleDisableAttackTypeDefaultBranchConfig(rnd, "ssi", point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "ssi"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckNoResourceAttr(name, "action"),
				),
			},
		},
	})
}

func testWallarmRuleDisableAttackTypeBasicConfig(resourceID, attackType, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_attack_type" "%[1]s" {
  action {
    type = "%[2]s"
    value = "%[3]s"
    point = {
      header = "%[4]s"
    }
  }
  point = [%[5]s]
  attack_type = "%[6]s"
}`, resourceID, actionType, actionValue, actionPoint, point, attackType)
}

func testWallarmRuleDisableAttackTypeDefaultBranchConfig(resourceID, attackType, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_attack_type" "%[1]s" {
  point = [%[2]s]
  attack_type = "%[3]s"
}`, resourceID, point, attackType)
}

func testAccRuleDisableAttackTypeCreateRecreate(resourceID, attackType string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_attack_type" "%[1]s" {
  point = [["header", "X-FOOBAR"]]
  attack_type = "%[2]s"
}`, resourceID, attackType)
}

func testAccCheckWallarmRuleDisableAttackTypeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_disable_attack_type" {
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
				Type:     []string{"disable_attack_type"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Ignore Certain Attack Type rule still exists")
		}
	}

	return nil
}
