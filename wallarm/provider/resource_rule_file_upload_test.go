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
    value = "wenum.wallarm.com"
    point = {
      header = "HOST"
    }
  }
  
  point = [["header_all"]]
  
  size = 1
  size_unit = "mb"

}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleFileUploadSizeLimitDestroy,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmRuleFileUploadSizeLimitDestroy,
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
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_rule_file_upload_size_limit" {
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
			Limit:     APIListLimit,
			Offset:    0,
			OrderBy:   "updated_at",
			OrderDesc: true,
			Filter: &wallarm.HintFilter{
				Clientid: []int{clientID},
				ActionID: []int{actionID},
			},
		}

		rule, err := client.HintRead(hint)
		if err != nil && rule != nil && len(*rule.Body) > 0 {
			return fmt.Errorf("Wallarm Mode Rule still exists")
		}
	}

	return nil
}
