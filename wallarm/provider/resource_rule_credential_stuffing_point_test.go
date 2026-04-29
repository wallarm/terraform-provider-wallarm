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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccRuleCredentialStuffingPointDestroy(),
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
		CheckDestroy:             testAccRuleCredentialStuffingPointDestroy(),
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

func testAccRuleCredentialStuffingPointDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api, err := testAccNewAPIClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "wallarm_rule_credential_stuffing_point" {
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
				return fmt.Errorf("wallarm_rule_credential_stuffing_point %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
