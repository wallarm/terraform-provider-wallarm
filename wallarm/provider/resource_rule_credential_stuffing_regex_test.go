package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccRuleCredentialStuffingRegex_basic(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_regex." + resourceName
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleCredentialStuffingRegexDestroy(),
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
		CheckDestroy:             testAccRuleCredentialStuffingRegexDestroy(),
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

func testAccRuleCredentialStuffingRegexDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api, err := testAccNewAPIClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "wallarm_rule_credential_stuffing_regex" {
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
				return fmt.Errorf("wallarm_rule_credential_stuffing_regex %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
