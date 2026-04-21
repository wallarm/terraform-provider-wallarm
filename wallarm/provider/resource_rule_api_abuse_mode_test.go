package wallarm

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	wallarm "github.com/wallarm/wallarm-go"
)

// Shared HCL — minimal "enabled" rule with a single header match.
func testAccRuleAPIAbuseModeConfigBasic(name, mode string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_api_abuse_mode" %q {
  mode  = %q
  title = "acc test %[1]s"
  action {
    type  = "iequal"
    value = "example.com"
    point = { header = "HOST" }
  }
}
`, name, mode)
}

// Pinterest-style scope from the design spec.
func testAccRuleAPIAbuseModeConfigPinterest() string {
	return `
resource "wallarm_rule_api_abuse_mode" "pinterest" {
  mode    = "disabled"
  title   = "Allow Pinterest"
  comment = "Allow Pinterest through protections"

  action {
    type  = "regex"
    value = ".*(Pinterest|Pinterestbot)/(0.2|1.0);?\\s[(]?[+]https?://www[.]pinterest[.]com/bot[.]html[)].*"
    point = { header = "USER-AGENT" }
  }
  action {
    type  = "equal"
    value = "api"
    point = { path = "0" }
  }
  action {
    type  = "regex"
    value = "v\\d"
    point = { path = "1" }
  }
  action {
    type = "absent"
    point = { action_ext = "" }
  }
}
`
}

func testAccCheckWallarmRuleAPIAbuseModeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for %s", resourceName)
		}
		return nil
	}
}

func testAccCheckWallarmRuleAPIAbuseModeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_api_abuse_mode" {
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
		resp, err := client.HintRead(&wallarm.HintRead{
			Limit: 1, Offset: 0,
			Filter: &wallarm.HintFilter{Clientid: []int{clientID}, ID: []int{ruleID}},
		})
		if err == nil && resp != nil && resp.Body != nil && len(*resp.Body) > 0 {
			return fmt.Errorf("wallarm_rule_api_abuse_mode %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func TestAccRuleAPIAbuseModeCreate_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleAPIAbuseModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAPIAbuseModeConfigBasic("basic", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmRuleAPIAbuseModeExists("wallarm_rule_api_abuse_mode.basic"),
					resource.TestCheckResourceAttr("wallarm_rule_api_abuse_mode.basic", "mode", "enabled"),
					resource.TestCheckResourceAttr("wallarm_rule_api_abuse_mode.basic", "rule_type", "api_abuse_mode"),
					resource.TestCheckResourceAttr("wallarm_rule_api_abuse_mode.basic", "action.#", "1"),
				),
			},
			{
				ResourceName:            "wallarm_rule_api_abuse_mode.basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleAPIAbuseModeCreate_Disabled(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleAPIAbuseModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAPIAbuseModeConfigPinterest(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmRuleAPIAbuseModeExists("wallarm_rule_api_abuse_mode.pinterest"),
					resource.TestCheckResourceAttr("wallarm_rule_api_abuse_mode.pinterest", "mode", "disabled"),
					resource.TestCheckResourceAttr("wallarm_rule_api_abuse_mode.pinterest", "title", "Allow Pinterest"),
					resource.TestCheckResourceAttr("wallarm_rule_api_abuse_mode.pinterest", "action.#", "4"),
				),
			},
		},
	})
}

func TestAccRuleAPIAbuseModeForceNewOnMode(t *testing.T) {
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleAPIAbuseModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAPIAbuseModeConfigBasic("force_new", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmRuleAPIAbuseModeExists("wallarm_rule_api_abuse_mode.force_new"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources["wallarm_rule_api_abuse_mode.force_new"].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleAPIAbuseModeConfigBasic("force_new", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmRuleAPIAbuseModeExists("wallarm_rule_api_abuse_mode.force_new"),
					resource.TestCheckResourceAttr("wallarm_rule_api_abuse_mode.force_new", "mode", "disabled"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources["wallarm_rule_api_abuse_mode.force_new"].Primary.Attributes["rule_id"]
						if newID == firstRuleID {
							return fmt.Errorf("expected rule_id to change on ForceNew, still %s", newID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccRuleAPIAbuseModeInvalidMode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAPIAbuseModeConfigBasic("bad", "monitoring"),
				ExpectError: regexp.MustCompile(`expected mode to be one of \[enabled disabled\]`),
			},
		},
	})
}

func TestAccRuleAPIAbuseModeImport(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleAPIAbuseModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAPIAbuseModeConfigBasic("import_me", "disabled"),
				Check:  testAccCheckWallarmRuleAPIAbuseModeExists("wallarm_rule_api_abuse_mode.import_me"),
			},
			{
				ResourceName:            "wallarm_rule_api_abuse_mode.import_me",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleAPIAbuseModeExistsError(t *testing.T) {
	// Same action scope, different "resource" label → existsAction must block the second create.
	configFirst := testAccRuleAPIAbuseModeConfigBasic("first", "enabled")
	configDup := configFirst + testAccRuleAPIAbuseModeConfigBasic("duplicate", "disabled")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleAPIAbuseModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: configFirst,
				Check:  testAccCheckWallarmRuleAPIAbuseModeExists("wallarm_rule_api_abuse_mode.first"),
			},
			{
				Config:      configDup,
				ExpectError: ResourceExistsError(`[0-9]+/[0-9]+/[0-9]+`, "wallarm_rule_api_abuse_mode"),
			},
		},
	})
}
