package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type maskingTestingRule struct {
	point     string
	matchType []string
	value     string
}

func TestAccRuleMaskingCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_masking." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleMaskingBasicConfig(rnd, "iequal", "masking.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "form_urlencoded"),
					resource.TestCheckResourceAttr(name, "point.1.1", "query"),
				),
			},
		},
	})
}

func TestAccRuleMaskingCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_masking." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleMaskingDestroy,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleMaskingDefaultBranchConfig(rnd, point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckNoResourceAttr(name, "action"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleMaskingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleMaskingFullSettingsConfig(rnd, rule),
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
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testWallarmRuleMaskingBasicConfig(resourceID, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_masking" "%[1]s" {
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

func testWallarmRuleMaskingDefaultBranchConfig(resourceID, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_masking" "%[1]s" {
	point = [%[2]s]
}`, resourceID, point)
}

func testAccRuleMaskingCreateRecreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_masking" "%[1]s" {
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
resource "wallarm_rule_masking" "%[7]s" {

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
		type = "%[2]s"
		point = {
		  action_name = "masking"
		}
	}

	action {
		type = "%[2]s"
		point = {
		  action_name = "masking"
		}
	}

	action {
		type = "%[4]s"
		point = {
		  action_ext = ""
		}
	}
	  
	action {
		type = "%[4]s"
		point = {
		  path = 0
		}
	}

	action {
		type = "%[2]s"
		point = {
		  method = "GET"
		}
	}

	action {
		type = "%[1]s"
		point = {
		  scheme = "https"
		}
	}

	action {
		type = "%[1]s"
		point = {
		  proto = "1.1"
		}
	}

	action {
		type = "%[3]s"
		point = {
		  uri = "/api/token[0-9A-Za-z]+"
		}
	}

	action {
		type = "%[4]s"
		value = "%[5]s"
		point = {
		  header = "HOST"
		}
	}

	point = [%[6]s]
}`, equal, iequal, regex, absent, rule.value, rule.point, resourceID)
}

func testAccCheckWallarmRuleMaskingDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_masking" {
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
				Type:     []string{"sensitive_data"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Sensitive Data Masking rule still exists")
		}
	}

	return nil
}
