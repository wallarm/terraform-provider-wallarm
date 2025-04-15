package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	wallarm "github.com/wallarm/wallarm-go"
)

func TestAccRuleCredentialStuffingPoint_basic(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_point." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccRuleCredentialStuffingPointDestroy(),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleCredentialStuffingPoint_basic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "point.0.0", "header"),
					resource.TestCheckResourceAttr(resourceAddress, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(resourceAddress, "login_point.0.0", "header"),
					resource.TestCheckResourceAttr(resourceAddress, "login_point.0.1", "SESSION-ID"),
					resource.TestCheckResourceAttr(resourceAddress, "cred_stuff_type", "custom"),
					resource.TestCheckResourceAttr(resourceAddress, "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleCredentialStuffingPoint_basic(resourceName string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_credential_stuffing_point" %[1]q {
	point = [["header", "HOST"]]
	login_point = [["header", "SESSION-ID"]]
	cred_stuff_type = "custom"

	action {
		type = "iequal"
		value = "example.com"
		point = {
			header = "HOST"
		}
	}
}
`, resourceName)
}

func testAccRuleCredentialStuffingPointDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(wallarm.API)

		for _, resource := range s.RootModule().Resources {
			if resource.Type != "wallarm_rule_credential_stuffing_point" {
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
