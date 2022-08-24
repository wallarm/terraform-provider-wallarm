package wallarm

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	wallarm "github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccRuleParserStateCreate_Basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_parser_state." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleParserStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleParserStateBasicConfig(rnd, "base64", "enabled", "iequal", "parsers.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "parser", "base64"),
					resource.TestCheckResourceAttr(name, "state", "enabled"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
					resource.TestCheckResourceAttr(name, "point.0.0", "post"),
					resource.TestCheckResourceAttr(name, "point.1.0", "form_urlencoded"),
					resource.TestCheckResourceAttr(name, "point.1.1", "query"),
				),
			},
		},
	})
}

func TestAccRuleParserStateCreate_IncorrectState(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmRuleParserStateBasicConfig(rnd, "base64", "incorrect", "iequal", "parsers.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				ExpectError: regexp.MustCompile(`expected state to be one of \[enabled disabled\], got incorrect`),
			},
		},
	})
}

func TestAccRuleParserStateCreate_IncorrectParser(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmRuleParserStateBasicConfig(rnd, "incorrect", "enabled", "iequal", "parsers.wallarm.com", "HOST", `["post"],["form_urlencoded","query"]`),
				ExpectError: regexp.MustCompile(`config is invalid: expected parser to be one of \[base64 cookie form_urlencoded gzip grpc json_doc multipart percent protobuf htmljs viewstate xml\], got incorrect`),
			},
		},
	})
}

func TestAccRuleParserStateCreateRecreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_parser_state." + rnd
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleParserStateDestroy,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleParserStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmRuleParserStateDefaultBranchConfig(rnd, "gzip", "disabled", point),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "parser", "gzip"),
					resource.TestCheckResourceAttr(name, "state", "disabled"),
					resource.TestCheckResourceAttr(name, "point.0.0", "header"),
					resource.TestCheckResourceAttr(name, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(name, "point.1.0", "pollution"),
					resource.TestCheckNoResourceAttr(name, "action"),
				),
			},
		},
	})
}

func testWallarmRuleParserStateBasicConfig(resourceID, parser, state, actionType, actionValue, actionPoint, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_parser_state" "%[1]s" {
  action {
    type = "%[2]s"
    value = "%[3]s"
    point = {
      header = "%[4]s"
    }
  }
  point = [%[5]s]
  parser = "%[6]s"
  state = "%[7]s"
}`, resourceID, actionType, actionValue, actionPoint, point, parser, state)
}

func testWallarmRuleParserStateDefaultBranchConfig(resourceID, parser, state, point string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_parser_state" "%[1]s" {
  point = [%[2]s]
  parser = "%[3]s"
  state = "%[4]s"
}`, resourceID, point, parser, state)
}

func testAccRuleParserStateCreateRecreate(resourceID, parser, state string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_parser_state" "%[1]s" {
  point = [["uri"]]
  parser = "%[2]s"
  state = "%[3]s"
}`, resourceID, parser, state)
}

func testAccCheckWallarmRuleParserStateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_parser_state" {
			continue
		}

		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return err
		}
		actionID, err := strconv.Atoi(rs.Primary.Attributes["action_id"])
		if err != nil {
			return err
		}

		hint := &wallarm.HintRead{
			Limit:     1000,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ActionID: []int{actionID},
				Type:     []string{"parser_state"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Disable/Enable Parsers rule still exists")
		}
	}

	return nil
}
