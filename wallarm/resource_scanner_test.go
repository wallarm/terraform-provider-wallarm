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
				Config: testWallarmScannerCreate(rnd, scannerElements, "true"),
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

func TestAccWallarmScannerCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_scanner." + rnd
	scannerElements := `"1.1.1.1", "example.com", "2.2.2.2/31"`
	scannerElements2 := `"5.5.5.5", "example2.com", "4.4.4.4/30"`
	scannerElements3 := `"6.6.6.6", "example3.com", "8.8.8.8/32"`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmScannerCreate(rnd, scannerElements, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "element.0", "1.1.1.1"),
					resource.TestCheckResourceAttr(name, "element.1", "example.com"),
					resource.TestCheckResourceAttr(name, "element.2", "2.2.2.2/31"),
					resource.TestCheckResourceAttr(name, "disabled", "true"),
				),
			},
			{
				Config: testWallarmScannerCreate(rnd, scannerElements2, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "element.0", "5.5.5.5"),
					resource.TestCheckResourceAttr(name, "element.1", "example2.com"),
					resource.TestCheckResourceAttr(name, "element.2", "4.4.4.4/30"),
					resource.TestCheckResourceAttr(name, "disabled", "true"),
				),
			},
			{
				Config: testWallarmScannerCreate(rnd, scannerElements3, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "element.0", "6.6.6.6"),
					resource.TestCheckResourceAttr(name, "element.1", "example3.com"),
					resource.TestCheckResourceAttr(name, "element.2", "8.8.8.8/32"),
					resource.TestCheckResourceAttr(name, "disabled", "false"),
				),
			},
		},
	})
}

func testWallarmScannerCreate(resourceID, element, disabled string) string {
	return fmt.Sprintf(`
resource "wallarm_scanner" "%[1]s" {
	element = [%[2]s]
	disabled = %[3]s
}`, resourceID, element, disabled)
}
