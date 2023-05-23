package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type binaryDataTestingRule struct {
	point     string
	matchType []string
	value     string
}

func TestAccRuleBinaryDataCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_binary_data." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleBinaryDataBasicConfig(rnd, "iequal", "binary.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
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

func TestAccRuleBinaryDataCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_binary_data." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBinaryDataDestroy,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleBinaryDataDefaultBranchConfig(rnd, point),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleBinaryDataDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleBinaryDataFullSettingsConfig(rnd, rule),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "10"),
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

func testWallarmRuleBinaryDataBasicConfig(resourceID, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" "%[1]s" {
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

func testWallarmRuleBinaryDataDefaultBranchConfig(resourceID, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" "%[1]s" {
	point = [%[2]s]
}`, resourceID, point)
}

func testAccRuleBinaryDataCreateRecreate(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_binary_data" "%[1]s" {
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
resource "wallarm_rule_binary_data" "%[7]s" {

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
		  action_name = "foobar"
		}
	}

	action {
		type = "%[2]s"
		point = {
		  action_name = "foobar"
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
		value = "admin"
		type = "%[1]s"

		point = {
			query = "user"
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
		type = "%[3]s"
		value = "%[5]s"
		point = {
		  header = "HOST"
		}
	}

	point = [%[6]s]
}`, equal, iequal, regex, absent, rule.value, rule.point, resourceID)
}

func testAccCheckWallarmRuleBinaryDataDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_binary_data" {
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
				Type:     []string{"binary_data"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Allow Binary Data rule still exists")
		}
	}

	return nil
}
