package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmScannerCreate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_scanner." + rnd
	scannerElements := `"1.1.1.1", "example.com", "2.2.2.2/31"`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmScannerCreate(rnd, scannerElements),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "element.0", "1.1.1.1"),
					resource.TestCheckResourceAttr(name, "element.1", "example.com"),
					resource.TestCheckResourceAttr(name, "element.2", "2.2.2.2/31"),
					resource.TestCheckResourceAttr(name, "disabled", "true"),
				),
			},
		},
	})
}

func testWallarmScannerCreate(resourceID, element string) string {
	return fmt.Sprintf(`
resource "wallarm_scanner" "%[1]s" {
	element = ["1.1.1.1", "example.com", "2.2.2.2/31"]
	disabled = true
}`, resourceID, element)
}
