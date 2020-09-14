package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmNode(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_node." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmNodeConfig(rnd, "tf-test-"+rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "hostname", "tf-test-"+rnd),
				),
			},
		},
	})
}

func testWallarmNodeConfig(resourceID, hostname string) string {
	return fmt.Sprintf(`
resource "wallarm_node" "%[1]s" {
  hostname = "%[2]s"
}`, resourceID, hostname)
}
