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
				Config: testWallarmGlobalModeFiltrationConfig(rnd, "safe_blocking"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "safe_blocking"),
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
				Config: testWallarmGlobalModeScannerConfig(rnd, "off"),
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
				Config: testWallarmGlobalModeRecheckerConfig(rnd, "on"),
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
				Config: testWallarmGlobalModeFullConfig(rnd, "block", "off", "on"),
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
				Config: testWallarmGlobalModeFullConfig(rnd, "default", "on", "off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "default"),
					resource.TestCheckResourceAttr(name, "scanner_mode", "on"),
					resource.TestCheckResourceAttr(name, "rechecker_mode", "off"),
				),
			},
		},
	})
}

func testWallarmGlobalModeFiltrationConfig(resourceID, filtrationMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  filtration_mode = "%[2]s"
}`, resourceID, filtrationMode)
}

func testWallarmGlobalModeScannerConfig(resourceID, scannerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  scanner_mode = "%[2]s"
}`, resourceID, scannerMode)
}

func testWallarmGlobalModeRecheckerConfig(resourceID, recheckerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  rechecker_mode = "%[2]s"
}`, resourceID, recheckerMode)
}

func testWallarmGlobalModeFullConfig(resourceID, filtrationMode, scannerMode, recheckerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  filtration_mode = "%[2]s"
  scanner_mode = "%[3]s"
  rechecker_mode = "%[4]s"
}`, resourceID, filtrationMode, scannerMode, recheckerMode)
}
