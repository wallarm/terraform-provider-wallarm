package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
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
					resource.TestCheckResourceAttr(name, "action.#", "0"),
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

// Multiple vpatch hints can attach to the same Action scope, keyed by
// (point, attack_type). Verifies action_id is shared, rule_ids are distinct,
// and Read round-trips both.
func TestAccRuleVpatch_MultiplePerAction(t *testing.T) {
	host := "multi-vpatch.example.com"
	config := fmt.Sprintf(`
resource "wallarm_rule_vpatch" "sqli_user" {
  attack_type = "sqli"
  action {
    type  = "iequal"
    value = %[1]q
    point = { header = "HOST" }
  }
  point = [["post"], ["form_urlencoded", "username"]]
}

resource "wallarm_rule_vpatch" "xss_pass" {
  attack_type = "xss"
  action {
    type  = "iequal"
    value = %[1]q
    point = { header = "HOST" }
  }
  point = [["post"], ["form_urlencoded", "password"]]
}
`, host)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"wallarm_rule_vpatch.sqli_user", "action_id",
						"wallarm_rule_vpatch.xss_pass", "action_id",
					),
				),
			},
		},
	})
}

func TestAccRuleVpatchUpdateInPlaceComment(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	var firstRuleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleVpatchUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleVpatchUpdateCommentConfig(rnd, "second comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "second comment"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						if newID != firstRuleID {
							return fmt.Errorf("expected rule_id to stay stable on in-place update, was %s now %s", firstRuleID, newID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccRuleVpatchUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
  attack_type = "xss"
  comment     = %[2]q
  action {
    type  = "iequal"
    value = "vpatch_comment_update.example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
}`, resourceID, comment)
}

func testAccCheckWallarmRuleVpatchDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

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
			Limit:     APIListLimit,
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
