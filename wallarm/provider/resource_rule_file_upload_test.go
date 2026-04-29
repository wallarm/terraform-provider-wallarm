package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	return testAccCheckHintDestroyed(s, "wallarm_rule_file_upload_size_limit")
}
