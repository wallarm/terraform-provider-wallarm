package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleVpatchCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchBasicConfig(rnd, "xss", "iequal", "vpatch.wallarm.com", "HOST", "get_all"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "xss"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "get_all"),
				),
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleVpatchCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	point := `["get", "query"]`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleVpatchDefaultBranchConfig(rnd, "crlf", point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "attack_type", "crlf"),
					resource.TestCheckResourceAttr(name, "point.0.0", "get"),
					resource.TestCheckResourceAttr(name, "point.0.1", "query"),
					resource.TestCheckResourceAttr(name, "action.#", "0"),
				),
			},
		},
	})
}

func testWallarmRuleVpatchBasicConfig(resourceID, attackType, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
  attack_type = %[2]q
  action {
    type = %[3]q
    value = %[4]q
    point = {
      header = %[5]q
    }
  }
  point = [[%[6]q]]
}`, resourceID, attackType, actionType, actionValue, actionPoint, point)
}

func testWallarmRuleVpatchDefaultBranchConfig(resourceID, attackType, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
	attack_type = %[2]q
	point = [%[3]s]
}`, resourceID, attackType, point)
}

// Multiple vpatch hints can attach to the same Action scope, keyed by
// (point, attack_type). Verifies action_id is shared, rule_ids are distinct,
// and Read round-trips both.
func TestAccRuleVpatch_MultiplePerAction(t *testing.T) {
	host := "multi-vpatch.example.com"
	config := fmt.Sprintf(`
resource "wallarm_rule_vpatch" "sqli_user" {
  attack_type = "sqli"
  action {
    type  = "iequal"
    value = %[1]q
    point = { header = "HOST" }
  }
  point = [["post"], ["form_urlencoded", "username"]]
}

resource "wallarm_rule_vpatch" "xss_pass" {
  attack_type = "xss"
  action {
    type  = "iequal"
    value = %[1]q
    point = { header = "HOST" }
  }
  point = [["post"], ["form_urlencoded", "password"]]
}
`, host)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"wallarm_rule_vpatch.sqli_user", "action_id",
						"wallarm_rule_vpatch.xss_pass", "action_id",
					),
				),
			},
		},
	})
}

func TestAccRuleVpatchUpdateInPlaceComment(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleVpatchUpdateCommentConfig(rnd, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleVpatchUpdateCommentConfig(rnd, "second comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "comment", "second comment"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[name].Primary.Attributes["rule_id"]
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

func testAccRuleVpatchUpdateCommentConfig(resourceID, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
  attack_type = "xss"
  comment     = %[2]q
  action {
    type  = "iequal"
    value = "vpatch_comment_update.example.com"
    point = {
      header = "HOST"
    }
  }
  point = [["get_all"]]
}`, resourceID, comment)
}

func testAccCheckWallarmRuleVpatchDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_vpatch")
}

// TestAccRuleVpatchActionScope_UriOnly exercises ActionScopeCustomizeDiff happy
// path: a single `action {}` block with a `uri`-only point map. Verifies that
// the customizer accepts uri as a standalone scope (PointValuePoint, value goes
// in the point map; sibling `value` field must be empty).
func TestAccRuleVpatchActionScope_UriOnly(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	config := fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
  attack_type = "xss"
  action {
    type  = "iequal"
    point = { uri = "/api/v1/uri-scope-test/%[1]s" }
  }
  point = [["get_all"]]
}
`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "attack_type", "xss"),
				),
			},
		},
	})
}

// TestAccRuleVpatchActionScope_PathDecomposed exercises the decomposed-path
// scope shape: separate `action {}` blocks for `path`/`action_name`/`action_ext`
// (uriConflictPoints) without `uri`. Customizer must accept this — they only
// conflict when paired with `uri`.
func TestAccRuleVpatchActionScope_PathDecomposed(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_vpatch." + rnd
	config := fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
  attack_type = "xss"
  action {
    type  = "equal"
    value = "0"
    point = { path = "0" }
  }
  action {
    type  = "iequal"
    point = { action_name = "users" }
  }
  point = [["get_all"]]
}
`, rnd)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleVpatchDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "action.#", "2"),
				),
			},
		},
	})
}

// vpatchActionScopeErrorConfig produces an HCL with two action blocks. The
// `extraBlock` is the full second `action {}` body — caller controls whether
// it includes `value = ...` (required for header/query, must be empty for
// action_name/action_ext, free for path) so the test exercises the intended
// validator branch.
func vpatchActionScopeErrorConfig(rnd, extraBlock string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_vpatch" %[1]q {
  attack_type = "xss"
  action {
    type  = "iequal"
    point = { uri = "/api/error-test/%[1]s" }
  }
  action {
    %[2]s
  }
  point = [["get_all"]]
}
`, rnd, extraBlock)
}

// validation error tests use PlanOnly + ExpectError so plan-time
// CustomizeDiff failures are caught without an apply (no API contact).

func TestAccRuleVpatchActionScope_UriPathConflict(t *testing.T) {
	rnd := generateRandomResourceName(5)
	// path is not a PointValuePoint — value field is unconstrained.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: vpatchActionScopeErrorConfig(rnd, `type = "equal"
    value = "0"
    point = { path = "0" }`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`action condition "uri" conflicts with`),
			},
		},
	})
}

func TestAccRuleVpatchActionScope_UriActionNameConflict(t *testing.T) {
	rnd := generateRandomResourceName(5)
	// action_name is a PointValuePoint — value MUST be empty (or omitted).
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: vpatchActionScopeErrorConfig(rnd, `type = "iequal"
    point = { action_name = "users" }`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`action condition "uri" conflicts with`),
			},
		},
	})
}

func TestAccRuleVpatchActionScope_UriActionExtConflict(t *testing.T) {
	rnd := generateRandomResourceName(5)
	// action_ext is a PointValuePoint — value MUST be empty (or omitted).
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: vpatchActionScopeErrorConfig(rnd, `type = "iequal"
    point = { action_ext = "json" }`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`action condition "uri" conflicts with`),
			},
		},
	})
}

func TestAccRuleVpatchActionScope_UriQueryConflict(t *testing.T) {
	rnd := generateRandomResourceName(5)
	// query is required to have a non-empty value (it's the matched content).
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: vpatchActionScopeErrorConfig(rnd, `type = "iequal"
    value = "extra"
    point = { query = "search" }`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`action condition "uri" conflicts with`),
			},
		},
	})
}

func TestAccRuleVpatchActionScope_InvalidPointKey(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: vpatchActionScopeErrorConfig(rnd, `type = "iequal"
    value = "x"
    point = { totally_made_up = "x" }`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`unknown action point key`),
			},
		},
	})
}

func TestAccRuleVpatchActionScope_MultipleKeysInPoint(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: vpatchActionScopeErrorConfig(rnd, `type = "iequal"
    value = "HOST"
    point = { header = "HOST", query = "q" }`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`action block "point" must contain exactly one key`),
			},
		},
	})
}
