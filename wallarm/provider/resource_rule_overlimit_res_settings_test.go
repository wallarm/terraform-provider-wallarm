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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleOverlimitResSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleOverlimitResSettingsConfig(resourceName, "overlimit_basic.example.com", "monitoring", 1000),
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
	first := testAccRuleOverlimitResSettingsConfig("first", "overlimit_exists.example.com", "monitoring", 1000)
	dup := first + testAccRuleOverlimitResSettingsConfig("duplicate", "overlimit_exists.example.com", "blocking", 5000)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleOverlimitResSettingsDestroy,
		Steps: []resource.TestStep{
			{Config: first},
			{
				Config:      dup,
				ExpectError: ResourceExistsError(`[0-9]+/[0-9]+/[0-9]+`, "wallarm_rule_overlimit_res_settings"),
			},
		},
	})
}

func TestAccOverlimitResSettingsUpdateInPlaceOverlimitTime(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_overlimit_res_settings." + resourceName
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleOverlimitResSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleOverlimitResSettingsConfig(resourceName, "overlimit_update.example.com", "monitoring", 1000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "overlimit_time", "1000"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceAddress].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleOverlimitResSettingsConfig(resourceName, "overlimit_update.example.com", "monitoring", 2000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "overlimit_time", "2000"),
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

func testAccRuleOverlimitResSettingsDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_overlimit_res_settings" {
			continue
		}

		ruleID, err := strconv.Atoi(rs.Primary.Attributes["rule_id"])
		if err != nil {
			return fmt.Errorf("invalid rule_id for %s: %w", rs.Primary.ID, err)
		}
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return fmt.Errorf("invalid client_id for %s: %w", rs.Primary.ID, err)
		}

		// OrderBy is required by the API — HintRead returns 400 without it.
		resp, err := api.HintRead(&wallarm.HintRead{
			Limit:   1,
			OrderBy: "updated_at",
			Filter:  &wallarm.HintFilter{Clientid: []int{clientID}, ID: []int{ruleID}},
		})
		if err != nil {
			return fmt.Errorf("checking hint %d still exists: %w", ruleID, err)
		}
		if resp.Body != nil && len(*resp.Body) > 0 {
			return fmt.Errorf("wallarm_rule_overlimit_res_settings %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
