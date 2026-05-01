package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRuleCredentialStuffingPoint_basic(t *testing.T) {
	resourceName := generateRandomResourceName(5)
	resourceAddress := "wallarm_rule_credential_stuffing_point." + resourceName
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleCredentialStuffingPointDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleCredentialStuffingPointDestroy,
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
		value = "credstuff_point_basic.example.com"
		point = {
			header = "HOST"
		}
	}
}
`, resourceName)
}

func testAccRuleCredentialStuffingPointDestroy(s *terraform.State) error {
	return testAccCheckHintDestroyed(s, "wallarm_rule_credential_stuffing_point")
}
