package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmGlobalMode_FiltrationSafeBlocking(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_global_mode." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalMode_FiltrationConfig(rnd, "safe_blocking"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "block"),
				),
			},
		},
	})
}

func TestAccWallarmGlobalMode_ScannerOff(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_global_mode." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalMode_ScannerConfig(rnd, "off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "scanner_mode", "off"),
				),
			},
		},
	})
}

func TestAccWallarmGlobalMode_RecheckerOn(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_global_mode." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalMode_RecheckerConfig(rnd, "on"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "rechecker_mode", "on"),
				),
			},
		},
	})
}

func TestAccWallarmGlobalMode_FiltrationBlock_ScannerOff_RechkeckerOn(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_global_mode." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalMode_FullConfig(rnd, "block", "off", "on"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "block"),
					resource.TestCheckResourceAttr(name, "scanner_mode", "off"),
					resource.TestCheckResourceAttr(name, "rechecker_mode", "on"),
				),
			},
		},
	})
}

func TestAccWallarmGlobalMode_FiltrationDefault_ScannerOn_RecheckerOff(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_global_mode." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalMode_FullConfig(rnd, "default", "on", "off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "default"),
					resource.TestCheckResourceAttr(name, "scanner_mode", "on"),
					resource.TestCheckResourceAttr(name, "rechecker_mode", "off"),
				),
			},
		},
	})
}

func testWallarmGlobalMode_FiltrationConfig(resourceID, filtrationMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  filtration_mode = "%[2]s"
}`, resourceID, filtrationMode)
}

func testWallarmGlobalMode_ScannerConfig(resourceID, scannerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  scanner_mode = "%[2]s"
}`, resourceID, scannerMode)
}

func testWallarmGlobalMode_RecheckerConfig(resourceID, recheckerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  rechecker_mode = "%[2]s"
}`, resourceID, recheckerMode)
}

func testWallarmGlobalMode_FullConfig(resourceID, filtrationMode, scannerMode, recheckerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  filtration_mode = "%[2]s"
  scanner_mode = "%[3]s"
  rechecker_mode = "%[4]s"
}`, resourceID, filtrationMode, scannerMode, recheckerMode)
}
