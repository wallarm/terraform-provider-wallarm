package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccRuleCredentialStuffingMode_basic(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_mode." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccRuleCredentialStuffingModeDestroy(),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleCredentialStuffingMode_basic(resourceName, "default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "mode", "default"),
					resource.TestCheckResourceAttr(resourceAddress, "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleCredentialStuffingMode_basic(resourceName string, mode string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_credential_stuffing_mode" %[1]q {
	mode = %[2]q
	action {
		type = "iequal"
		value = "example.com"
		point = {
			header = "HOST"
		}
	}
}
`, resourceName, mode)
}

func testAccRuleCredentialStuffingModeDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(wallarm.API)

		for _, resource := range s.RootModule().Resources {
			if resource.Type != "wallarm_rule_credential_stuffing_mode" {
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
