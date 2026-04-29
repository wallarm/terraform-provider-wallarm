package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type binaryDataTestingRule struct {
	point     string
	matchType []string
	value     string
}

func TestAccRuleBinaryDataCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_binary_data." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleBinaryDataBasicConfig(rnd, "iequal", "binary_data_basic.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
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

func TestAccRuleBinaryDataCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_binary_data." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBinaryDataCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "headers"),
				),
			},
			{
				Config: testAccRuleBinaryDataCreateRecreate(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "headers"),
				),
			},
		},
	})
}

func TestAccRuleBinaryDataCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_binary_data." + rnd
	point := `["header","HOST"],["pollution"]`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleBinaryDataDefaultBranchConfig(rnd, point),
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

func TestAccRuleBinaryDataCreate_FullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_binary_data." + rnd
	point := `["post"],["xml"],["array_all"],["xml_attr","CDATA"]`
	matchType := []string{"equal", "iequal", "regex", "absent"}
	value := generateRandomResourceName(10) + ".wallarm.com"

	rule := binaryDataTestingRule{
		point:     point,
		matchType: matchType,
		value:     value,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleBinaryDataFullSettingsConfig(rnd, rule),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "9"),
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

func testWallarmRuleBinaryDataBasicConfig(resourceID, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" %[1]q {
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

func testWallarmRuleBinaryDataDefaultBranchConfig(resourceID, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" %[1]q {
	point = [%[2]s]
}`, resourceID, point)
}

func testAccRuleBinaryDataCreateRecreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" %[1]q {
	action {
		point = {
		  method = "POST|GET|PATCH"
		}
    	type = "regex"
	}

  point = [["header", "headers"]]
}`, resourceID)
}

func testWallarmRuleBinaryDataFullSettingsConfig(resourceID string, rule binaryDataTestingRule) string {
	equal := rule.matchType[0]
	iequal := rule.matchType[1]
	regex := rule.matchType[2]
	absent := rule.matchType[3]
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" %[7]q {

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
		  action_name = "foobar"
		}
	}

	action {
		type = %[2]q
		point = {
		  action_name = "foobar"
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
		value = "admin"
		type = %[1]q

		point = {
			query = "user"
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

func TestAccRuleBinaryDataUpdateInPlaceComment(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_binary_data." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleBinaryDataUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleBinaryDataUpdateCommentConfig(rnd, "second comment"),
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

func testAccRuleBinaryDataUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" %[1]q {
  comment = %[2]q
  action {
    type  = "iequal"
    value = "binary_data_comment_update.example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["post"], ["form_urlencoded", "query"]]
}`, resourceID, comment)
}

func testAccCheckWallarmRuleBinaryDataDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_binary_data" {
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
			return fmt.Errorf("wallarm_rule_binary_data %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
