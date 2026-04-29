package wallarm

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleParserStateCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_parser_state." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleParserStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleParserStateBasicConfig(rnd, "base64", "enabled", "iequal", "parser_state_basic.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "parser", "base64"),
					resource.TestCheckResourceAttr(name, "state", "enabled"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "form_urlencoded"),
					resource.TestCheckResourceAttr(name, "point.1.1", "query"),
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

func TestAccRuleParserStateCreate_IncorrectState(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmRuleParserStateBasicConfig(rnd, "base64", "incorrect", "iequal", "parser_state_invalid.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				ExpectError: regexp.MustCompile(`expected state to be one of \["enabled" "disabled"\], got incorrect`),
			},
		},
	})
}

func TestAccRuleParserStateCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_parser_state." + rnd
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleParserStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleParserStateCreateRecreate(rnd, "htmljs", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "parser", "htmljs"),
					resource.TestCheckResourceAttr(name, "state", "enabled"),
					resource.TestCheckResourceAttr(name, "point.0.0", "uri"),
				),
			},
			{
				Config: testAccRuleParserStateCreateRecreate(rnd, "htmljs", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "parser", "htmljs"),
					resource.TestCheckResourceAttr(name, "state", "enabled"),
					resource.TestCheckResourceAttr(name, "point.0.0", "uri"),
				),
			},
		},
	})
}

func TestAccRuleParserStateCreate_DefaultBranch(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_parser_state." + rnd
	point := `["header","HOST"],["pollution"]`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleParserStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleParserStateDefaultBranchConfig(rnd, "gzip", "disabled", point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "parser", "gzip"),
					resource.TestCheckResourceAttr(name, "state", "disabled"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckResourceAttr(name, "action.#", "0"),
				),
			},
		},
	})
}

func TestAccRuleParserStateUpdateInPlaceState(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_parser_state." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleParserStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleParserStateUpdateConfig(rnd, "parser_state_update.example.com", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "state", "enabled"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleParserStateUpdateConfig(rnd, "parser_state_update.example.com", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "state", "disabled"),
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

func testAccRuleParserStateUpdateConfig(resourceID, host, state string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_parser_state" %[1]q {
  action {
    type  = "iequal"
    value = %[2]q
    point = {
      header = "HOST"
    }
  }
  point  = [["post"],["form_urlencoded","query"]]
  parser = "base64"
  state  = %[3]q
}`, resourceID, host, state)
}

func testWallarmRuleParserStateBasicConfig(resourceID, parser, state, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_parser_state" %[1]q {
  action {
    type = %[2]q
    value = %[3]q
    point = {
      header = %[4]q
    }
  }
  point = [%[5]s]
  parser = %[6]q
  state = %[7]q
}`, resourceID, actionType, actionValue, actionPoint, point, parser, state)
}

func testWallarmRuleParserStateDefaultBranchConfig(resourceID, parser, state, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_parser_state" %[1]q {
  point = [%[2]s]
  parser = %[3]q
  state = %[4]q
}`, resourceID, point, parser, state)
}

func testAccRuleParserStateCreateRecreate(resourceID, parser, state string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_parser_state" %[1]q {
  point = [["uri"]]
  parser = %[2]q
  state = %[3]q
}`, resourceID, parser, state)
}

func testAccCheckWallarmRuleParserStateDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_parser_state" {
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
			return fmt.Errorf("wallarm_rule_parser_state %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
