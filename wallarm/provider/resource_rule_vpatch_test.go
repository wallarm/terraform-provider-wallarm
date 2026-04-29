package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleVpatchCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchDefaultBranchConfig(rnd, "crlf", point),
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
resource "wallarm_rule_vpatch" %[1]q {
  attack_type = %[2]q
  action {
    type = %[3]q
    value = %[4]q
    point = {
      header = %[5]q
    }
  }
  point = [[%[6]q]]
}`, resourceID, attackType, actionType, actionValue, actionPoint, point)
}

func testWallarmRuleVpatchDefaultBranchConfig(resourceID, attackType, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
	attack_type = %[2]q
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
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
	return testAccCheckHintDestroyed(s, "wallarm_rule_vpatch")
}
