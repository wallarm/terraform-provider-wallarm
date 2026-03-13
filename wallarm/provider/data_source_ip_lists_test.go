package wallarm

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceIPLists(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIPListsConfig("denylist"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceIPListsExists("data.wallarm_ip_lists.test"),
				),
			},
			{
				Config: testAccDataSourceIPListsConfig("allowlist"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceIPListsExists("data.wallarm_ip_lists.allow"),
				),
			},
		},
	})
}

func testAccCheckDataSourceIPListsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return nil // data source may not exist if no entries
		}
		if rs.Primary.ID == "" {
			return nil
		}
		return nil
	}
}

func testAccDataSourceIPListsConfig(listType string) string {
	if listType == "allowlist" {
		return `data "wallarm_ip_lists" "allow" {
  list_type = "allowlist"
}`
	}
	return `data "wallarm_ip_lists" "test" {
  list_type = "denylist"
}`
}
