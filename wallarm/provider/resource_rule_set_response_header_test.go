package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleSetResponseHeader_basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_set_response_header." + rnd

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleSetResponseHeaderDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleSetResponseHeaderDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleSetResponseHeaderDestroy,
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

func TestAccRuleSetResponseHeaderUpdateInPlaceMode(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_rule_set_response_header." + rnd
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleSetResponseHeaderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleSetResponseHeaderUpdateConfig(rnd, "set_response_header_update.example.com", "append"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "append"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[name].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleSetResponseHeaderUpdateConfig(rnd, "set_response_header_update.example.com", "replace"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "mode", "replace"),
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

func testAccRuleSetResponseHeaderUpdateConfig(resourceID, host, mode string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_set_response_header" %[1]q {
  mode   = %[3]q
  name   = "X-Test-Header"
  values = ["test-value"]

  action {
    type  = "iequal"
    value = %[2]q
    point = {
      header = "HOST"
    }
  }
}`, resourceID, host, mode)
}

func testAccRuleSetResponseHeaderBasic(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_set_response_header" %[1]q {
  mode   = "append"
  name   = "X-Test-Header"
  values = ["test-value"]

  action {
    type = "iequal"
    value = "set_response_header_basic.example.com"
    point = {
      header = "HOST"
    }
  }
}`, resourceID)
}

func testAccRuleSetResponseHeaderReplace(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_set_response_header" %[1]q {
  mode   = "replace"
  name   = "X-Frame-Options"
  values = ["DENY"]

  action {
    type = "iequal"
    value = "set_response_header_replace.example.com"
    point = {
      header = "HOST"
    }
  }
}`, resourceID)
}

func testAccRuleSetResponseHeaderMultipleValues(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_set_response_header" %[1]q {
  mode   = "append"
  name   = "X-Multi-Header"
  values = ["value-one", "value-two"]

  action {
    type = "iequal"
    value = "set_response_header_multi.example.com"
    point = {
      header = "HOST"
    }
  }
}`, resourceID)
}

func testAccCheckWallarmRuleSetResponseHeaderDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_set_response_header")
}
