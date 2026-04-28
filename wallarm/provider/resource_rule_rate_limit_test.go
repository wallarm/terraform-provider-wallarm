package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccRuleRateLimit(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_rate_limit." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccRuleRateLimitDestroy(),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleRateLimit(resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "point.0.0", "header"),
					resource.TestCheckResourceAttr(resourceAddress, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(resourceAddress, "delay", "100"),
				),
			},
			{
				ResourceName:            resourceAddress,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func testAccRuleRateLimit(resourceName string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_rate_limit" %[1]q {
	point = [["header", "HOST"]]

	action {
		type = "iequal"
		value = "example.com"
		point = {
			header = "HOST"
		}
	}

  comment = "My TF Rate Limit 5"
  delay = 100
  burst = 20
  rate = 300
  rsp_status = 500
  time_unit = "rps"
}
`, resourceName)
}

func TestAccRuleRateLimitUpdateInPlaceDelay(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_rate_limit." + resourceName
	var firstRuleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccRuleRateLimitDestroy(),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleRateLimitUpdateConfig(resourceName, "rate_limit_update.example.com", 100),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "delay", "100"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceAddress].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleRateLimitUpdateConfig(resourceName, "rate_limit_update.example.com", 200),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "delay", "200"),
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

func testAccRuleRateLimitUpdateConfig(resourceName, host string, delay int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_rate_limit" %[1]q {
	point = [["header", "HOST"]]

	action {
		type = "iequal"
		value = %[2]q
		point = {
			header = "HOST"
		}
	}

  delay = %[3]d
  burst = 20
  rate = 300
  rsp_status = 500
  time_unit = "rps"
}
`, resourceName, host, delay)
}

func testAccRuleRateLimitDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ProviderMeta).Client

		for _, resource := range s.RootModule().Resources {
			if resource.Type != "wallarm_rule_rate_limit" {
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
