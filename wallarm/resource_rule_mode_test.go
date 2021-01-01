package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type wmodeTestingRule struct {
	mode      string
	matchType []string
	value     string
}

func TestAccRuleWmodeCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleWmodeBasicConfig(rnd, "monitoring", "iequal", "wmode.wallarm.com", "HOST"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "monitoring"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleWmodeCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleWmodeDefaultBranchConfig(rnd, "default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(name, "action"),
				),
				ExpectError: ResourceExistsError("[0-9]+/[0-9]+/[0-9]+/wallarm_mode", "wallarm_rule_mode"),
			},
		},
	})
}

func TestAccRuleWmodeCreate_FullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_mode." + rnd
	matchType := []string{"equal", "iequal", "regex", "absent"}
	value := "https://docs.wallarm.com/admin-en/installation-nginx-en/"

	rule := wmodeTestingRule{
		mode:      "off",
		matchType: matchType,
		value:     value,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleWmodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleWmodeFullSettingsConfig(rnd, rule),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "off"),
					resource.TestCheckResourceAttr(name, "action.#", "10"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testWallarmRuleWmodeBasicConfig(resourceID, mode, actionType, actionValue, actionPoint string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_mode" "%[1]s" {
  mode = "%[2]s"
  action {
    type = "%[3]s"
    value = "%[4]s"
    point = {
      header = "%[5]s"
    }
  }
}`, resourceID, mode, actionType, actionValue, actionPoint)
}

func testWallarmRuleWmodeDefaultBranchConfig(resourceID, mode string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_mode" "%[1]s" {
	mode = "%[2]s"
}`, resourceID, mode)
}

func testWallarmRuleWmodeFullSettingsConfig(resourceID string, rule wmodeTestingRule) string {
	equal := rule.matchType[0]
	iequal := rule.matchType[1]
	regex := rule.matchType[2]
	absent := rule.matchType[3]
	return fmt.Sprintf(`
resource "wallarm_rule_mode" "%[7]s" {

	mode = "%[1]s"

	action {
		point = {
		  instance = 9
		}
	}

	# Intentionally create a duplicate which is supposed to be removed by Set
	action {
		point = {
		  instance = 9
		}
	}

	action {
		type = "%[3]s"
		point = {
		  action_name = "wmode"
		}
	}

	action {
		type = "%[5]s"
		point = {
		  action_ext = ""
		}
	}

	action {
		value = "api"
		type = "%[2]s"
		point = {
		  path = 1
		}
	}

	action {
		value = "login"
		type = "%[3]s"
		point = {
		  path = 3
		}
	}

	action {
		type = "%[3]s"
		point = {
		  method = "PUT"
		}
	}

	action {
		type = "%[2]s"
		point = {
		  scheme = "http"
		}
	}

	action {
		type = "%[2]s"
		point = {
		  proto = "1.0"
		}
	}

	action {
		type = "%[4]s"
		point = {
		  uri = "/console/username[0-9A-Za-z]+"
		}
	}

	action {
		type = "%[3]s"
		value = "%[6]s"
		point = {
		  header = "referer"
		}
	}

}`, rule.mode, equal, iequal, regex, absent, rule.value, resourceID)
}

func testAccCheckWallarmRuleWmodeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_mode" {
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
				Type:     []string{"wallarm_mode"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Wallarm Mode Rule still exists")
		}
	}

	return nil
}
