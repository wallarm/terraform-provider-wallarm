package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccRuleFileUploadSizeLimit(t *testing.T) {
	const config = `
resource "wallarm_rule_file_upload_size_limit" "wallarm_rule_file_upload_size_limit_1" {
  mode = "block"

  action {
    type = "iequal"
    value = "file_upload_basic.example.com"
    point = {
      header = "HOST"
    }
  }

  point = [["header_all"]]

  size = 1
  size_unit = "mb"

}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleFileUploadSizeLimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("wallarm_rule_file_upload_size_limit.wallarm_rule_file_upload_size_limit_1", "mode", "block"),
					resource.TestCheckResourceAttr("wallarm_rule_file_upload_size_limit.wallarm_rule_file_upload_size_limit_1", "action.#", "1"),
				),
			},
			{
				ResourceName:            "wallarm_rule_file_upload_size_limit.wallarm_rule_file_upload_size_limit_1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_type"},
			},
		},
	})
}

func TestAccRuleFileUploadSizeLimitUpdateInPlaceSize(t *testing.T) {
	resourceAddress := "wallarm_rule_file_upload_size_limit.update_size"
	var firstRuleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmRuleFileUploadSizeLimitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleFileUploadSizeLimitUpdateConfig("file_upload_update.example.com", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "size", "1"),
					func(s *terraform.State) error {
						firstRuleID = s.RootModule().Resources[resourceAddress].Primary.Attributes["rule_id"]
						return nil
					},
				),
			},
			{
				Config: testAccRuleFileUploadSizeLimitUpdateConfig("file_upload_update.example.com", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "size", "2"),
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

func testAccRuleFileUploadSizeLimitUpdateConfig(host string, size int) string {
	return fmt.Sprintf(`
resource "wallarm_rule_file_upload_size_limit" "update_size" {
  mode = "block"

  action {
    type  = "iequal"
    value = %[1]q
    point = { header = "HOST" }
  }

  point     = [["header_all"]]
  size      = %[2]d
  size_unit = "mb"
}
`, host, size)
}

func testAccCheckWallarmRuleFileUploadSizeLimitDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_file_upload_size_limit" {
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
			return fmt.Errorf("wallarm_rule_file_upload_size_limit %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
