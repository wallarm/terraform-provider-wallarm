package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleGraphqlDetection(t *testing.T) {
	const config = `
resource "wallarm_rule_graphql_detection" "wallarm_rule_graphql_detection_1" {
  mode = "block"

  action {
    type = "iequal"
    value = "graphql_basic.example.com"
    point = {
      header = "HOST"
    }
  }

  max_depth = 10
  max_value_size_kb = 10
  max_doc_size_kb = 100
  max_aliases = 5
  max_doc_per_batch = 10
  introspection = true
  debug_enabled = true

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleGraphqlDetectionDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleGraphqlDetectionDestroy,
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
	return testAccCheckHintDestroyed(s, "wallarm_rule_graphql_detection")
}

// Create with only `mode` should land all API defaults in state and re-plan
// clean. Int fields use Optional+Computed so the API echoes back into state.
// Bool fields use Optional+Default(true) so the SDK fills d.Get("introspection")
// with the schema default; the wire payload sends true and the API echoes true
// back. Schema default and API default match by design — keeps removal
// symmetric (see TestAccRuleGraphqlDetection_RemovingBoolFromHCLRestoresDefault).
func TestAccRuleGraphqlDetection_MinimalCreatePreservesAPIDefaults(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_graphql_detection." + rnd
	config := fmt.Sprintf(`
resource "wallarm_rule_graphql_detection" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "graphql-defaults.example.com"
    point = { header = "HOST" }
  }
}
`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleGraphqlDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// Int API defaults preserved via Optional+Computed.
					resource.TestCheckResourceAttr(name, "max_depth", "10"),
					resource.TestCheckResourceAttr(name, "max_value_size_kb", "10"),
					resource.TestCheckResourceAttr(name, "max_doc_size_kb", "100"),
					resource.TestCheckResourceAttr(name, "max_doc_per_batch", "10"),
					// Bools omitted from HCL → wire skip → API defaults (true) win.
					resource.TestCheckResourceAttr(name, "introspection", "true"),
					resource.TestCheckResourceAttr(name, "debug_enabled", "true"),
				),
			},
			{
				// Re-plan with the same minimal config — Computed must
				// preserve state, no drift.
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccRuleGraphqlDetection_UpdateAfterMinimalCreate is the regression
// guard for the v2.3.8 silent-int-zeroing-on-update bug: previously, after
// Create with only `mode`, adding `introspection = false` to HCL produced
// a plan that wanted to send zero for max_depth/max_value_size_kb/etc. (the
// SDK treated state-only values as drift), which the API rejected with
// `should be in 1..N`. With Computed: true, the unchanged ints are
// preserved by SDK plan logic.
func TestAccRuleGraphqlDetection_UpdateAfterMinimalCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_graphql_detection." + rnd
	configMinimal := fmt.Sprintf(`
resource "wallarm_rule_graphql_detection" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "graphql-update.example.com"
    point = { header = "HOST" }
  }
}
`, rnd)
	configIntrospectionFalse := fmt.Sprintf(`
resource "wallarm_rule_graphql_detection" %[1]q {
  mode          = "block"
  introspection = false
  action {
    type  = "iequal"
    value = "graphql-update.example.com"
    point = { header = "HOST" }
  }
}
`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleGraphqlDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: configMinimal,
				// Primary contract: int-zeroing on Update (regression guard).
				// Bool assertion is a sanity check that omitted-from-HCL preserves API default.
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "max_depth", "10"),
					resource.TestCheckResourceAttr(name, "introspection", "true"),
				),
			},
			{
				// Update with introspection explicitly set — the int fields
				// stay unchanged in state and aren't zeroed (Computed
				// preserves), so the API keeps its values; Update succeeds.
				// Pre-fix this would have errored with "should be in 1..N".
				Config: configIntrospectionFalse,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "introspection", "false"),
					resource.TestCheckResourceAttr(name, "max_depth", "10"),
					resource.TestCheckResourceAttr(name, "max_value_size_kb", "10"),
				),
			},
		},
	})
}

// TestAccRuleGraphqlDetection_RemovingBoolFromHCLRestoresDefault guards the
// v2.3.9 manual-test finding: with Optional+Computed, setting `introspection
// = false` then removing the line left state stuck at false. With
// Optional+Default(true), removing the line plans `false → true` and Update
// restores the default — symmetric with adding `= false` planning `true → false`.
func TestAccRuleGraphqlDetection_RemovingBoolFromHCLRestoresDefault(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_graphql_detection." + rnd
	cfg := func(extra string) string {
		return fmt.Sprintf(`
resource "wallarm_rule_graphql_detection" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "graphql-restore-default.example.com"
    point = { header = "HOST" }
  }
%[2]s
}
`, rnd, extra)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleGraphqlDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg("  introspection = false\n  debug_enabled = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "introspection", "false"),
					resource.TestCheckResourceAttr(name, "debug_enabled", "false"),
				),
			},
			{
				// Remove both lines from HCL → schema default (true) wins,
				// plan shows diff, Update restores the API default.
				Config: cfg(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "introspection", "true"),
					resource.TestCheckResourceAttr(name, "debug_enabled", "true"),
				),
			},
		},
	})
}

// TestAccRuleGraphqlDetection_MaxAliasesMutableInPlace guards the v2.3.9 schema
// flip on `max_aliases` (Optional+Computed+ForceNew → Optional+Default(5)).
// Probe-confirmed mutable on PUT; wallarm-go v0.12.3 added the field to
// HintUpdateV3Params and resourcerule.WithMaxAliases wires it. The test
// flips the value and asserts rule_id stays stable across the Update.
func TestAccRuleGraphqlDetection_MaxAliasesMutableInPlace(t *testing.T) {
	rnd := generateRandomResourceName(5)
	addr := "wallarm_rule_graphql_detection." + rnd
	var firstRuleID string
	cfg := func(maxAliases int) string {
		return fmt.Sprintf(`
resource "wallarm_rule_graphql_detection" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "graphql-max-aliases.example.com"
    point = { header = "HOST" }
  }
  max_aliases = %[2]d
}
`, rnd, maxAliases)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleGraphqlDetectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg(3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "max_aliases", "3"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[addr].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: cfg(7),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "max_aliases", "7"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[addr].Primary.Attributes["rule_id"]
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
