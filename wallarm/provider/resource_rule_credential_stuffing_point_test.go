package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccRuleCredentialStuffingPoint_basic(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_point." + resourceName
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccRuleCredentialStuffingPointDestroy(),
		Providers:    testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleCredentialStuffingPointBasic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "point.0.0", "header"),
					resource.TestCheckResourceAttr(resourceAddress, "point.0.1", "HOST"),
					resource.TestCheckResourceAttr(resourceAddress, "login_point.0.0", "header"),
					resource.TestCheckResourceAttr(resourceAddress, "login_point.0.1", "SESSION-ID"),
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

func TestAccRuleCredentialStuffingPointUpdateInPlaceComment(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_point." + resourceName
	var firstRuleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccRuleCredentialStuffingPointDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleCredentialStuffingPointUpdateCommentConfig(resourceName, "first comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "comment", "first comment"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceAddress].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleCredentialStuffingPointUpdateCommentConfig(resourceName, "second comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "comment", "second comment"),
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

func testAccRuleCredentialStuffingPointUpdateCommentConfig(resourceName, comment string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_credential_stuffing_point" %[1]q {
	point = [["header", "HOST"]]
	login_point = [["header", "SESSION-ID"]]
	cred_stuff_type = "custom"
	comment = %[2]q

	action {
		type = "iequal"
		value = "credstuff_point_comment_update.example.com"
		point = {
			header = "HOST"
		}
	}
}
`, resourceName, comment)
}

func testAccRuleCredentialStuffingPointBasic(resourceName string) string {
	return fmt.Sprintf(`
resource "wallarm_rule_credential_stuffing_point" %[1]q {
	point = [["header", "HOST"]]
	login_point = [["header", "SESSION-ID"]]
	cred_stuff_type = "custom"

	action {
		type = "iequal"
		value = "example.com"
		point = {
			header = "HOST"
		}
	}
}
`, resourceName)
}

func testAccRuleCredentialStuffingPointDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ProviderMeta).Client

		for _, resource := range s.RootModule().Resources {
			if resource.Type != "wallarm_rule_credential_stuffing_point" {
				continue
			}

			clientID, err := strconv.Atoi(resource.Primary.Attributes["client_id"])
			if err != nil {
				return err
			}
			ruleID, err := strconv.Atoi(resource.Primary.Attributes["rule_id"])
			if err != nil {
				return err
			}

			resp, err := client.HintRead(&wallarm.HintRead{
				Limit:   1,
				OrderBy: "updated_at",
				Filter: &wallarm.HintFilter{
					Clientid: []int{clientID},
					ID:       []int{ruleID},
				},
			})
			if err != nil {
				return err
			}

			if resp != nil && resp.Body != nil && len(*resp.Body) != 0 {
				return fmt.Errorf("Resource still exists: %s", resource.Primary.ID)
			}
		}

		return nil
	}
}
