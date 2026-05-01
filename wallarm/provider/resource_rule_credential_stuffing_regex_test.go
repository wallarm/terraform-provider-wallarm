package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleCredentialStuffingRegex_basic(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_regex." + resourceName
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleCredentialStuffingRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleCredentialStuffingRegexBasic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "regex", "abc"),
					resource.TestCheckResourceAttr(resourceAddress, "login_regex", "def"),
					resource.TestCheckResourceAttr(resourceAddress, "case_sensitive", "true"),
					resource.TestCheckResourceAttr(resourceAddress, "cred_stuff_type", "custom"),
					resource.TestCheckResourceAttr(resourceAddress, "action.#", "1"),
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

func TestAccRuleCredentialStuffingRegexUpdateInPlaceCaseSensitive(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_regex." + resourceName
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleCredentialStuffingRegexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleCredentialStuffingRegexUpdateConfig(resourceName, "credstuff_update.example.com", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "case_sensitive", "false"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceAddress].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleCredentialStuffingRegexUpdateConfig(resourceName, "credstuff_update.example.com", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "case_sensitive", "true"),
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

func testAccRuleCredentialStuffingRegexUpdateConfig(resourceName, host string, caseSensitive bool) string {
	return fmt.Sprintf(`
resource "wallarm_rule_credential_stuffing_regex" %[1]q {
	regex           = "abc"
	login_regex     = "def"
	case_sensitive  = %[3]t
	cred_stuff_type = "custom"

	action {
		type  = "iequal"
		value = %[2]q
		point = {
			header = "HOST"
		}
	}
}
`, resourceName, host, caseSensitive)
}

func testAccRuleCredentialStuffingRegexBasic(resourceName string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_credential_stuffing_regex" %[1]q {
	regex = "abc"
	login_regex = "def"
	case_sensitive = true
	cred_stuff_type = "custom"

	action {
		type = "iequal"
		value = "credstuff_regex_basic.example.com"
		point = {
			header = "HOST"
		}
	}
}
`, resourceName)
}

func testAccRuleCredentialStuffingRegexDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_credential_stuffing_regex")
}
