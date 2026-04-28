package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccRuleGraphqlDetection(t *testing.T) {
	const config = `
resource "wallarm_rule_graphql_detection" "wallarm_rule_graphql_detection_1" {
  mode = "block"
  
  action {
    type = "iequal"
    value = "wenum.wallarm.com"
    point = {
      header = "HOST"
    }
  }
  
  max_depth = 10
  max_value_size_kb = 10
  max_doc_size_kb = 100
  max_alias_size_kb = 5
  max_doc_per_batch = 10
  introspection = true
  debug_enabled = true

}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleGraphqlDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_graphql_detection.wallarm_rule_graphql_detection_1", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_graphql_detection.wallarm_rule_graphql_detection_1", "action.#", "1"),
				),
			},
			{
				ResourceName:            "wallarm_rule_graphql_detection.wallarm_rule_graphql_detection_1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleGraphqlDetectionUpdateInPlaceDebugEnabled(t *testing.T) {
	resourceAddress := "wallarm_rule_graphql_detection.update_debug"
	var firstRuleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleGraphqlDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGraphqlDetectionUpdateConfig("graphql_update.example.com", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "debug_enabled", "false"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceAddress].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleGraphqlDetectionUpdateConfig("graphql_update.example.com", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "debug_enabled", "true"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[resourceAddress].Primary.Attributes["rule_id"]
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

func testAccRuleGraphqlDetectionUpdateConfig(host string, debugEnabled bool) string {
	return fmt.Sprintf(`
resource "wallarm_rule_graphql_detection" "update_debug" {
  mode = "block"

  action {
    type  = "iequal"
    value = %[1]q
    point = { header = "HOST" }
  }

  max_depth         = 10
  max_value_size_kb = 10
  max_doc_size_kb   = 100
  max_doc_per_batch = 10
  introspection     = true
  debug_enabled     = %[2]t
}
`, host, debugEnabled)
}

func testAccCheckWallarmRuleGraphqlDetectionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_graphql_detection" {
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
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && rule != nil && len(*rule.Body) > 0 {
			return fmt.Errorf("Wallarm Mode Rule still exists")
		}
	}

	return nil
}
