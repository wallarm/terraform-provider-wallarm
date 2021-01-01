package wallarm

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type vpatchTestingRule struct {
	attackType string
	point      string
	matchType  []string
	value      string
}

func TestAccRuleVpatchCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	var clientID string
	var ok bool
	if clientID, ok = os.LookupEnv("WALLARM_API_CLIENT_ID"); !ok {
		clientID = "6039"
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchBasicConfig(rnd, clientID, "xss", "iequal", "vpatch.wallarm.com", "HOST", "get_all"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "client_id", clientID),
					resource.TestCheckResourceAttr(name, "attack_type.0", "xss"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "get_all"),
				),
			},
		},
	})
}

func TestAccRuleVpatchCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	attackType := `"crlf", "scanner", "redir", "ldapi"`
	point := `["get", "query"]`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchDefaultBranchConfig(rnd, attackType, point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type.0", "crlf"),
					resource.TestCheckResourceAttr(name, "attack_type.1", "scanner"),
					resource.TestCheckResourceAttr(name, "attack_type.2", "redir"),
					resource.TestCheckResourceAttr(name, "attack_type.3", "ldapi"),
					resource.TestCheckResourceAttr(name, "point.0.0", "get"),
					resource.TestCheckResourceAttr(name, "point.0.1", "query"),
					resource.TestCheckNoResourceAttr(name, "action"),
				),
			},
		},
	})
}

func TestAccRuleVpatchCreate_FullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	attackType := `"any", "sqli", "rce", "crlf", "nosqli", "ptrav", "xxe", "ptrav", "xss", "scanner", "redir", "ldapi"`
	point := `["post"],["json_doc"],["array",0],["hash","password"]`
	matchType := []string{"equal", "iequal", "regex", "absent"}
	value := generateRandomResourceName(10) + ".example.com"

	rule := vpatchTestingRule{
		attackType: attackType,
		point:      point,
		matchType:  matchType,
		value:      value,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchFullSettingsConfig(rnd, rule),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type.#", "12"),
					resource.TestCheckResourceAttr(name, "action.#", "9"),
					resource.TestCheckResourceAttr(name, "point.#", "4"),
					resource.TestCheckResourceAttr(name, "point.0.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.#", "1"),
					resource.TestCheckResourceAttr(name, "point.1.0", "json_doc"),
					resource.TestCheckResourceAttr(name, "point.2.#", "2"),
					resource.TestCheckResourceAttr(name, "point.2.0", "array"),
					resource.TestCheckResourceAttr(name, "point.2.1", "0"),
					resource.TestCheckResourceAttr(name, "point.3.#", "2"),
					resource.TestCheckResourceAttr(name, "point.3.0", "hash"),
					resource.TestCheckResourceAttr(name, "point.3.1", "password"),
				),
			},
		},
	})
}

func testWallarmRuleVpatchBasicConfig(resourceID, clientID, attackType, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" "%[7]s" {
  client_id = %[1]s
  attack_type = ["%[2]s"]
  action {
    type = "%[3]s"
    value = "%[4]s"
    point = {
      header = "%[5]s"
    }
  }
  point = [["%[6]s"]]
}`, clientID, attackType, actionType, actionValue, actionPoint, point, resourceID)
}

func testWallarmRuleVpatchDefaultBranchConfig(resourceID, attackType, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" "%[1]s" {
	attack_type = [%[2]s]
	point = [%[3]s]
}`, resourceID, attackType, point)
}

func testWallarmRuleVpatchFullSettingsConfig(resourceID string, rule vpatchTestingRule) string {
	equal := rule.matchType[0]
	iequal := rule.matchType[1]
	regex := rule.matchType[2]
	absent := rule.matchType[3]
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" "%[8]s" {

	attack_type = [%[1]s]

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
		type = "%[3]s"
		point = {
		  action_name = "masking"
		}
	}

	action {
		type = "%[5]s"
		point = {
		  action_ext = ""
		}
	}
	  
	action {
		type = "%[5]s"
		point = {
		  path = 0
		}
	}

	action {
		type = "%[3]s"
		point = {
		  method = "GET"
		}
	}

	action {
		type = "%[2]s"
		point = {
		  scheme = "https"
		}
	}

	action {
		type = "%[2]s"
		point = {
		  proto = "1.1"
		}
	}

	action {
		type = "%[4]s"
		point = {
		  uri = "/api/token[0-9A-Za-z]+"
		}
	}

	action {
		type = "%[3]s"
		value = "%[6]s"
		point = {
		  header = "HOST"
		}
	}

	point = [%[7]s]
}`, rule.attackType, equal, iequal, regex, absent, rule.value, rule.point, resourceID)
}

func testAccCheckWallarmRuleVpatchDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

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
			Limit:     1000,
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
