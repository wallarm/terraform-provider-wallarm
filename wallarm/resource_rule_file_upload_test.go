package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
		},
	})
}

func testAccCheckWallarmRuleFileUploadSizeLimitDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

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
			Limit:     1000,
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
