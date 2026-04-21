package wallarm

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	wallarm "github.com/wallarm/wallarm-go"
)

// Shared HCL — minimal "enabled" rule with a single header match.
func testAccRuleAPIAbuseModeConfigBasic(name, host, mode string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_api_abuse_mode" %[1]q {
  mode  = %[3]q
  title = "acc test %[1]s"
  action {
    type  = "iequal"
    value = %[2]q
    point = { header = "HOST" }
  }
}
`, name, host, mode)
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
	cached := testAccProvider.Meta().(*ProviderMeta).Client.(*CachedClient)
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

		// Ground-truth: bypass cache, go straight to wallarm-go API.
		// OrderBy is REQUIRED — API returns 400 without it.
		rawReq := &wallarm.HintRead{
			Limit: 1, Offset: 0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter:    &wallarm.HintFilter{Clientid: []int{clientID}, ID: []int{ruleID}},
		}
		rawResp, rawErr := cached.API.HintRead(rawReq)
		rawCount := 0
		if rawErr == nil && rawResp != nil && rawResp.Body != nil {
			rawCount = len(*rawResp.Body)
		}
		log.Printf("[TRACE] CheckDestroy: RAW api.HintRead(client=%d, id=%d) → err=%v count=%d",
			clientID, ruleID, rawErr, rawCount)

		// Cached path (same call the real provider code uses)
		cachedResp, cachedErr := cached.HintRead(rawReq)
		cachedCount := 0
		if cachedErr == nil && cachedResp != nil && cachedResp.Body != nil {
			cachedCount = len(*cachedResp.Body)
		}
		log.Printf("[TRACE] CheckDestroy: CACHED client.HintRead(client=%d, id=%d) → err=%v count=%d",
			clientID, ruleID, cachedErr, cachedCount)

		if rawCount != cachedCount {
			log.Printf("[TRACE] CheckDestroy: DISAGREEMENT raw=%d vs cached=%d for id=%d",
				rawCount, cachedCount, ruleID)
		}

		if cachedErr == nil && cachedCount > 0 {
			return fmt.Errorf("wallarm_rule_api_abuse_mode %s still exists (raw_count=%d, cached_count=%d)",
				rs.Primary.ID, rawCount, cachedCount)
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
				Config: testAccRuleAPIAbuseModeConfigBasic("basic", "basic.example.com", "enabled"),
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
				Config: testAccRuleAPIAbuseModeConfigBasic("force_new", "force_new.example.com", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmRuleAPIAbuseModeExists("wallarm_rule_api_abuse_mode.force_new"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources["wallarm_rule_api_abuse_mode.force_new"].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleAPIAbuseModeConfigBasic("force_new", "force_new.example.com", "disabled"),
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
				Config:      testAccRuleAPIAbuseModeConfigBasic("bad", "bad.example.com", "monitoring"),
				ExpectError: regexp.MustCompile(`expected mode to be one of \["enabled" "disabled"\]`),
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
				Config: testAccRuleAPIAbuseModeConfigBasic("import_me", "import_me.example.com", "disabled"),
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
	configFirst := testAccRuleAPIAbuseModeConfigBasic("first", "exists.example.com", "enabled")
	configDup := configFirst + testAccRuleAPIAbuseModeConfigBasic("duplicate", "exists.example.com", "disabled")

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
