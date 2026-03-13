package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccRuleSetResponseHeader_basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_set_response_header." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleSetResponseHeaderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetResponseHeaderBasic(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "append"),
					resource.TestCheckResourceAttr(name, "name", "X-Test-Header"),
					resource.TestCheckResourceAttr(name, "values.#", "1"),
					resource.TestCheckResourceAttr(name, "action.#", "1"),
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

func TestAccRuleSetResponseHeader_replace(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_set_response_header." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleSetResponseHeaderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetResponseHeaderReplace(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "replace"),
					resource.TestCheckResourceAttr(name, "name", "X-Frame-Options"),
					resource.TestCheckResourceAttr(name, "values.#", "1"),
				),
			},
		},
	})
}

func TestAccRuleSetResponseHeader_multipleValues(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_set_response_header." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleSetResponseHeaderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetResponseHeaderMultipleValues(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "append"),
					resource.TestCheckResourceAttr(name, "name", "X-Multi-Header"),
					resource.TestCheckResourceAttr(name, "values.#", "2"),
				),
			},
		},
	})
}

func testAccRuleSetResponseHeaderBasic(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_set_response_header" "%[1]s" {
  mode   = "append"
  name   = "X-Test-Header"
  values = ["test-value"]

  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
}`, resourceID)
}

func testAccRuleSetResponseHeaderReplace(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_set_response_header" "%[1]s" {
  mode   = "replace"
  name   = "X-Frame-Options"
  values = ["DENY"]

  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
}`, resourceID)
}

func testAccRuleSetResponseHeaderMultipleValues(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_set_response_header" "%[1]s" {
  mode   = "append"
  name   = "X-Multi-Header"
  values = ["value-one", "value-two"]

  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }
}`, resourceID)
}

func testAccCheckWallarmRuleSetResponseHeaderDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_set_response_header" {
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
			Limit:     DefaultAPIListLimit,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ActionID: []int{actionID},
				Type:     []string{"set_response_header"},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && len(*rule.Body) != 0 {
			return fmt.Errorf("Set Response Header rule still exists")
		}
	}

	return nil
}
