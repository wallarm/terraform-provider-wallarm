package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type maskingTestingRule struct {
	point     string
	matchType []string
	value     string
}

func TestAccRuleMaskingCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_masking." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleMaskingBasicConfig(rnd, "iequal", "masking_basic.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
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

func TestAccRuleMaskingCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_masking." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleMaskingCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "headers"),
				),
			},
			{
				Config: testAccRuleMaskingCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "headers"),
				),
			},
		},
	})
}

func TestAccRuleMaskingCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_masking." + rnd
	point := `["header","HOST"],["pollution"]`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleMaskingDefaultBranchConfig(rnd, point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckResourceAttr(name, "action.#", "0"),
				),
			},
		},
	})
}

func TestAccRuleMaskingCreate_FullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_masking." + rnd
	point := `["post"],["xml"],["array_all"],["xml_attr","CDATA"]`
	matchType := []string{"equal", "iequal", "regex", "absent"}
	value := generateRandomResourceName(10) + ".wallarm.com"

	rule := maskingTestingRule{
		point:     point,
		matchType: matchType,
		value:     value,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleMaskingFullSettingsConfig(rnd, rule),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "8"),
					resource.TestCheckResourceAttr(name, "point.#", "4"),
					resource.TestCheckResourceAttr(name, "point.0.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.#", "1"),
					resource.TestCheckResourceAttr(name, "point.1.0", "xml"),
					resource.TestCheckResourceAttr(name, "point.2.#", "1"),
					resource.TestCheckResourceAttr(name, "point.2.0", "array_all"),
					resource.TestCheckResourceAttr(name, "point.3.#", "2"),
					resource.TestCheckResourceAttr(name, "point.3.0", "xml_attr"),
					resource.TestCheckResourceAttr(name, "point.3.1", "CDATA"),
				),
			},
		},
	})
}

func testWallarmRuleMaskingBasicConfig(resourceID, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_masking" %[1]q {
  action {
    type = %[2]q
    value = %[3]q
    point = {
      header = %[4]q
    }
  }
  point = [%[5]s]
}`, resourceID, actionType, actionValue, actionPoint, point)
}

func testWallarmRuleMaskingDefaultBranchConfig(resourceID, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_masking" %[1]q {
	point = [%[2]s]
}`, resourceID, point)
}

func testAccRuleMaskingCreateRecreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_masking" %[1]q {
	action {
		point = {
		  method = "POST|GET|PATCH"
		}
    	type = "regex"
	}

  point = [["header", "headers"]]
}`, resourceID)
}

func testWallarmRuleMaskingFullSettingsConfig(resourceID string, rule maskingTestingRule) string {
	equal := rule.matchType[0]
	iequal := rule.matchType[1]
	regex := rule.matchType[2]
	absent := rule.matchType[3]
	return fmt.Sprintf(`
resource "wallarm_rule_masking" %[7]q {

	action {
		point = {
		  instance = 1
		}
	}

	action {
		point = {
		  instance = 1
		}
	}

	action {
		type = %[2]q
		point = {
		  action_name = "masking"
		}
	}

	action {
		type = %[2]q
		point = {
		  action_name = "masking"
		}
	}

	action {
		type = %[4]q
		point = {
		  action_ext = ""
		}
	}

	action {
		type = %[4]q
		point = {
		  path = 0
		}
	}

	action {
		type = %[2]q
		point = {
		  method = "GET"
		}
	}

	action {
		type = %[1]q
		point = {
		  scheme = "https"
		}
	}

	action {
		type = %[1]q
		point = {
		  proto = "1.1"
		}
	}

	action {
		type = %[3]q
		value = %[5]q
		point = {
		  header = "HOST"
		}
	}

	point = [%[6]s]
}`, equal, iequal, regex, absent, rule.value, rule.point, resourceID)
}

func TestAccRuleMaskingUpdateInPlaceComment(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_masking." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleMaskingUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleMaskingUpdateCommentConfig(rnd, "second comment"),
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

func testAccRuleMaskingUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_masking" %[1]q {
  comment = %[2]q
  action {
    type  = "iequal"
    value = "masking_comment_update.example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"], ["form_urlencoded", "query"]]
}`, resourceID, comment)
}

func testAccCheckWallarmRuleMaskingDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_masking")
}
