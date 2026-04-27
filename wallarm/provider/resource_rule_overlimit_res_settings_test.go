package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
				Config: testAccRuleOverlimitResSettingsConfig(resourceName, "example.com", "monitoring", 1000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "overlimit_time", "1000"),
					resource.TestCheckResourceAttr(resourceAddress, "mode", "monitoring"),
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

// overlimit_res_settings has no per-rule discriminator (no point, no
// enumerated_parameters), so two same-scope rules collide and existingHintForAction
// must block the second create.
func TestAccOverlimitResSettings_ExistsError(t *testing.T) {
	first := testAccRuleOverlimitResSettingsConfig("first", "exists.example.com", "monitoring", 1000)
	dup := first + testAccRuleOverlimitResSettingsConfig("duplicate", "exists.example.com", "blocking", 5000)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccRuleOverlimitResSettingsDestroy(),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{Config: first},
			{
				Config:      dup,
				ExpectError: ResourceExistsError(`[0-9]+/[0-9]+/[0-9]+`, "wallarm_rule_overlimit_res_settings"),
			},
		},
	})
}

func testAccRuleOverlimitResSettingsConfig(label, host, mode string, overlimitTime int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_overlimit_res_settings" %[1]q {
  mode           = %[3]q
  overlimit_time = %[4]d
  action {
    type  = "iequal"
    value = %[2]q
    point = { header = "HOST" }
  }
}
`, label, host, mode, overlimitTime)
}

func testAccRuleOverlimitResSettingsDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ProviderMeta).Client

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
