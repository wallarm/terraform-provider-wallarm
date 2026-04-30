package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleEnumExact(t *testing.T) {
	const config = `
resource "wallarm_rule_enum" "wallarm_rule_enum_exact" {
  mode = "block"

  action {
    type = "iequal"
    value = "enum_exact.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = 5
    period = 30
  }

  enumerated_parameters {
    mode = "exact"
    points {
      point     = ["header", "REFERER"]
      sensitive = false
    }
    points {
      point     = ["get", "id"]
      sensitive = true
    }
  }
}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_exact", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_exact", "action.#", "1"),
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_exact", "enumerated_parameters.0.mode", "exact"),
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_exact", "enumerated_parameters.0.points.0.sensitive", "false"),
				),
			},
			{
				ResourceName:            "wallarm_rule_enum.wallarm_rule_enum_exact",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleEnumRegexp(t *testing.T) {
	const config = `
resource "wallarm_rule_enum" "wallarm_rule_enum_regexp" {
  mode = "block"

  action {
    type = "iequal"
    value = "enum_regexp.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = 5
    period = 30
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = false
    plain_parameters      = false
  }
}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_regexp", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_regexp", "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleEnumWithAdvancedConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_enum" "wallarm_rule_enum_advanced_conditions" {
  mode = "block"

  action {
    type = "iequal"
    value = "enum_advanced.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = 5
    period = 30
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = false
    plain_parameters      = false
  }

  advanced_conditions {
    field    = "status_code"
    value    = ["200"]
    operator = "eq"
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_advanced_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_advanced_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleEnumWithArbitraryConditions(t *testing.T) {
	const config = `
resource "wallarm_rule_enum" "wallarm_rule_enum_arbitrary_conditions" {
  mode = "block"

  action {
    type = "iequal"
    value = "enum_arbitrary.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000

  }

  threshold {
    count = 5
    period = 30
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = false
    plain_parameters      = false
  }

  arbitrary_conditions {
    point = [["header", "X-LOGIN"]]
    value    = ["value"]
    operator = "ne"
  }

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_arbitrary_conditions", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_enum.wallarm_rule_enum_arbitrary_conditions", "action.#", "1"),
				),
			},
		},
	})
}

func testAccRuleEnumUpdateConfig(additional bool) string {
	return fmt.Sprintf(`
resource "wallarm_rule_enum" "update_additional" {
  mode = "block"

  action {
    type = "iequal"
    value = "enum_update.example.com"
    point = {
      header = "HOST"
    }
  }

  reaction {
    block_by_session = 3000
    block_by_ip = 4000
  }

  threshold {
    count = 5
    period = 30
  }

  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo", "bar"]
    value_regexps         = ["baz"]
    additional_parameters = %[1]t
    plain_parameters      = false
  }
}
`, additional)
}

func TestAccRuleEnumUpdateInPlaceAdditionalParameters(t *testing.T) {
	resourceName := "wallarm_rule_enum.update_additional"
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleEnumUpdateConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enumerated_parameters.0.additional_parameters", "false"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceName].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleEnumUpdateConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enumerated_parameters.0.additional_parameters", "true"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[resourceName].Primary.Attributes["rule_id"]
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

func testAccCheckWallarmRuleEnumDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_enum")
}

// TestAccRuleEnum_DefaultsToActiveOnCreate guards the v2.3.8 fix for
// resourcerule.Create defaulting `active` to false instead of true. Without
// the fix, mitigation controls (brute, bola, enum, forced_browsing,
// graphql_detection, file_upload_size_limit, rate_limit_enum) shipped to the
// API as inactive when the user omitted `active` from HCL — and a subsequent
// `active = false` produced no plan diff because state already matched.
//
// Step 1: create without `active`, assert state has true.
// Step 2: add `active = false`, assert plan applies and state flips.
func TestAccRuleEnum_DefaultsToActiveOnCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_enum." + rnd

	withActive := func(activeLine string) string {
		return fmt.Sprintf(`
resource "wallarm_rule_enum" %[1]q {
  mode = "block"
  %[2]s
  action {
    type  = "iequal"
    value = "enum_active_default.example.com"
    point = { header = "HOST" }
  }
  reaction { block_by_ip = 600 }
  threshold {
    count  = 5
    period = 30
  }
  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo"]
    value_regexps         = ["bar"]
  }
}`, rnd, activeLine)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: withActive(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "active", "true"),
				),
			},
			{
				Config: withActive(`active = false`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "active", "false"),
				),
			},
		},
	})
}

// TestAccRuleEnum_ExactRejectsAdditionalParameters verifies that
// EnumeratedParamsCustomizeDiff fails plan when the user populates
// `additional_parameters` while `mode = "exact"`. Without this validator the
// mapper silently dropped the field on PUT, causing a perpetual diff loop.
// PlanOnly + ExpectError so no API contact is required.
func TestAccRuleEnum_ExactRejectsAdditionalParameters(t *testing.T) {
	rnd := generateRandomResourceName(5)
	config := fmt.Sprintf(`
resource "wallarm_rule_enum" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "enum_exact_reject.example.com"
    point = { header = "HOST" }
  }
  reaction { block_by_ip = 600 }
  threshold {
    count  = 5
    period = 30
  }
  enumerated_parameters {
    mode                  = "exact"
    additional_parameters = true
    points {
      point     = ["header", "REFERER"]
      sensitive = true
    }
  }
}`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`additional_parameters.*not allowed when mode = "exact"`),
			},
		},
	})
}

// TestAccRuleEnum_ExactRejectsAdditionalParametersFalse verifies the strict
// denial added in v2.3.8: writing `additional_parameters = false` (any value)
// in `mode = "exact"` errors at plan, regardless of the literal value. The
// SDK's `Default: false` makes `d.Get` return false even when the user
// omitted the field, so the validator uses `d.GetRawConfig()` to detect
// HCL presence and reject any explicit boolean.
func TestAccRuleEnum_ExactRejectsAdditionalParametersFalse(t *testing.T) {
	rnd := generateRandomResourceName(5)
	config := fmt.Sprintf(`
resource "wallarm_rule_enum" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "enum_exact_reject_false.example.com"
    point = { header = "HOST" }
  }
  reaction { block_by_ip = 600 }
  threshold {
    count  = 5
    period = 30
  }
  enumerated_parameters {
    mode                  = "exact"
    additional_parameters = false
    points {
      point     = ["header", "REFERER"]
      sensitive = true
    }
  }
}`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`additional_parameters.*not allowed when mode = "exact"`),
			},
		},
	})
}

// TestAccRuleEnum_RegexpRequiresValueRegexps reproduces the v2.3.8 drift-loop
// bug: in `mode = "regexp"`, omitting `value_regexps` previously caused the
// mapper to substitute `[""]` (silently). The API echoed it back as `[null]`
// in state, producing `~ value_regexps = [- null,]` on every subsequent
// plan. The fix: drop the substitution and require both lists at plan time.
// User must write `value_regexps = [""]` explicitly to opt out of value
// matching.
func TestAccRuleEnum_RegexpRequiresValueRegexps(t *testing.T) {
	rnd := generateRandomResourceName(5)
	config := fmt.Sprintf(`
resource "wallarm_rule_enum" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "enum_regexp_required.example.com"
    point = { header = "HOST" }
  }
  reaction { block_by_ip = 600 }
  threshold {
    count  = 5
    period = 30
  }
  enumerated_parameters {
    mode         = "regexp"
    name_regexps = ["foo", "bar"]
  }
}`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`value_regexps.*required when mode = "regexp"`),
			},
		},
	})
}

// TestAccRuleEnum_UpdateAdditionalParametersToFalse regression-tests the
// Optional+Computed zero-value bug for booleans: flipping
// `additional_parameters: true → false` previously silently preserved true.
func TestAccRuleEnum_UpdateAdditionalParametersToFalse(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_enum." + rnd
	withParams := func(additional bool) string {
		return fmt.Sprintf(`
resource "wallarm_rule_enum" %[1]q {
  mode = "block"
  action {
    type  = "iequal"
    value = "enum_bool_update.example.com"
    point = { header = "HOST" }
  }
  reaction {
    block_by_session = 3000
    block_by_ip      = 4000
  }
  threshold {
    count  = 5
    period = 30
  }
  enumerated_parameters {
    mode                  = "regexp"
    name_regexps          = ["foo"]
    value_regexps         = ["bar"]
    additional_parameters = %[2]t
    plain_parameters      = false
  }
}`, rnd, additional)
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleEnumDestroy,
		Steps: []resource.TestStep{
			{
				Config: withParams(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "enumerated_parameters.0.additional_parameters", "true"),
				),
			},
			{
				Config: withParams(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "enumerated_parameters.0.additional_parameters", "false"),
				),
			},
		},
	})
}
