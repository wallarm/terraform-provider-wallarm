package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccOverlimitResSettings(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_overlimit_res_settings." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccRuleOverlimitResSettingsDestroy(),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleOverlimitResSettings(resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "overlimit_time", "1000"),
					resource.TestCheckResourceAttr(resourceAddress, "mode", "monitoring"),
				),
			},
		},
	})
}

func testAccRuleOverlimitResSettings(resourceName string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_overlimit_res_settings" %[1]q {

	mode = "monitoring"
	overlimit_time = 1000
	action {
		type = "iequal"
		value = "example.com"
		point = {
			header = "HOST"
		}
	}

  comment = "My TF Overlimit Res Setting"
}
`, resourceName)
}

func testAccRuleOverlimitResSettingsDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(wallarm.API)

		for _, resource := range s.RootModule().Resources {
			if resource.Type != "wallarm_rule_overlimit_res_settings" {
				continue
			}

			clientID, err := strconv.Atoi(resource.Primary.Attributes["client_id"])
			if err != nil {
				return err
			}
			ruleID, err := strconv.Atoi(resource.Primary.Attributes["rule_id"])
			if err != nil {
				return err
			}

			resp, err := client.HintRead(&wallarm.HintRead{
				Limit:   1,
				OrderBy: "updated_at",
				Filter: &wallarm.HintFilter{
					Clientid: []int{clientID},
					ID:       []int{ruleID},
				},
			})
			if err != nil {
				return err
			}

			if resp != nil && resp.Body != nil && len(*resp.Body) != 0 {
				return fmt.Errorf("Resource still exists: %s", resource.Primary.ID)
			}
		}

		return nil
	}
}
