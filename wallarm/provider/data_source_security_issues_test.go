package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceSecurityIssues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityIssuesConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceSecurityIssuesExists("data.wallarm_security_issues.all"),
				),
			},
		},
	})
}

func testAccCheckDataSourceSecurityIssuesExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		issueCount, err := strconv.Atoi(rs.Primary.Attributes["issues.#"])
		if err != nil {
			return err
		}

		// It's valid for there to be zero issues, so just check the attribute exists
		_ = issueCount

		return nil
	}
}

func testAccDataSourceSecurityIssuesConfig() string {
	return `data "wallarm_security_issues" "all" {}`
}
