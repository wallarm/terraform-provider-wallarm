package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleDisableAttackTypeCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_attack_type." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableAttackTypeDestroy,
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
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleDisableAttackTypeCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_attack_type." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableAttackTypeDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableAttackTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleDisableAttackTypeDefaultBranchConfig(rnd, "ssi", point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "ssi"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckResourceAttr(name, "action.#", "0"),
				),
			},
		},
	})
}

func testWallarmRuleDisableAttackTypeBasicConfig(resourceID, attackType, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_attack_type" %[1]q {
  action {
    type = %[2]q
    value = %[3]q
    point = {
      header = %[4]q
    }
  }
  point = [%[5]s]
  attack_type = %[6]q
}`, resourceID, actionType, actionValue, actionPoint, point, attackType)
}

func testWallarmRuleDisableAttackTypeDefaultBranchConfig(resourceID, attackType, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_attack_type" %[1]q {
  point = [%[2]s]
  attack_type = %[3]q
}`, resourceID, point, attackType)
}

func testAccRuleDisableAttackTypeCreateRecreate(resourceID, attackType string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_attack_type" %[1]q {
  point = [["header", "X-FOOBAR"]]
  attack_type = %[2]q
}`, resourceID, attackType)
}

func TestAccRuleDisableAttackTypeUpdateInPlaceComment(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_disable_attack_type." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleDisableAttackTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDisableAttackTypeUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleDisableAttackTypeUpdateCommentConfig(rnd, "second comment"),
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

func testAccRuleDisableAttackTypeUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_disable_attack_type" %[1]q {
  comment     = %[2]q
  attack_type = "sqli"
  action {
    type  = "iequal"
    value = "disable_attack_type_comment_update.example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"], ["form_urlencoded", "query"]]
}`, resourceID, comment)
}

func testAccCheckWallarmRuleDisableAttackTypeDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_disable_attack_type" {
			continue
		}

		ruleID, err := strconv.Atoi(rs.Primary.Attributes["rule_id"])
		if err != nil {
			return fmt.Errorf("invalid rule_id for %s: %w", rs.Primary.ID, err)
		}
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return fmt.Errorf("invalid client_id for %s: %w", rs.Primary.ID, err)
		}

		// OrderBy is required by the API — HintRead returns 400 without it.
		resp, err := api.HintRead(&wallarm.HintRead{
			Limit:   1,
			OrderBy: "updated_at",
			Filter:  &wallarm.HintFilter{Clientid: []int{clientID}, ID: []int{ruleID}},
		})
		if err != nil {
			return fmt.Errorf("checking hint %d still exists: %w", ruleID, err)
		}
		if resp.Body != nil && len(*resp.Body) > 0 {
			return fmt.Errorf("wallarm_rule_disable_attack_type %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
