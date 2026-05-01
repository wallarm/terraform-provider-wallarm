package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleRateLimit(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_rate_limit." + resourceName
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleRateLimitDestroy,
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
		value = "rate_limit_basic.example.com"
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleRateLimitDestroy,
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

func testAccRuleRateLimitDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_rate_limit")
}

// TestAccRuleRateLimit_RspStatusRequired guards the v2.3.8 schema
// actualisation: `rsp_status` was Optional in the schema but Required at the
// API level (`should be in 400..599, can't be blank`). The schema is now
// Required so plan-time validation catches the omission cleanly. PlanOnly +
// ExpectError so no API contact is needed.
func TestAccRuleRateLimit_RspStatusRequired(t *testing.T) {
	rnd := generateRandomResourceName(5)
	config := fmt.Sprintf(`
resource "wallarm_rule_rate_limit" %[1]q {
  action {
    type  = "iequal"
    value = "ratelimit-rsp-required.example.com"
    point = { header = "HOST" }
  }
  point = [["get_all"]]
  rate  = 100
  # rsp_status omitted on purpose
}
`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)"rsp_status".*required|argument "rsp_status" is required`),
			},
		},
	})
}

// TestAccRuleRateLimit_RateBurstZero is the regression guard for the v2.3.8
// silent-zero-drop bug. wallarm-go v0.12.1 changed Rate/Burst/Delay to *int
// because the previous int+omitempty shape silently dropped a literal 0
// from the wire payload — the API then rejected with `can't be blank`.
//
// Step 1: Create with `rate = 0`, `burst = 0`, `delay = 0` — must succeed.
// Step 2: re-plan, must show no drift.
func TestAccRuleRateLimit_RateBurstZero(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_rate_limit." + rnd
	config := fmt.Sprintf(`
resource "wallarm_rule_rate_limit" %[1]q {
  action {
    type  = "iequal"
    value = "ratelimit-zero.example.com"
    point = { header = "HOST" }
  }
  point      = [["get_all"]]
  rate       = 0
  burst      = 0
  delay      = 0
  rsp_status = 429
  time_unit  = "rps"
}
`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleRateLimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rate", "0"),
					resource.TestCheckResourceAttr(name, "burst", "0"),
					resource.TestCheckResourceAttr(name, "delay", "0"),
				),
			},
			{
				// Same config — second plan should be clean (no drift).
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
