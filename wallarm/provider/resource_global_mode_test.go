package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalModeScannerConfig(rnd),
				Check:  resource.ComposeTestCheckFunc(),
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
				Config: testWallarmGlobalModeFullConfig(rnd, "block", "on"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "block"),
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
				Config: testWallarmGlobalModeFullConfig(rnd, "default", "off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "default"),
					resource.TestCheckResourceAttr(name, "rechecker_mode", "off"),
				),
			},
		},
	})
}

func TestAccWallarmGlobalMode_OverlimitResSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_global_mode." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalModeOverlimitConfig(rnd, 1000, "blocking"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "overlimit_time", "1000"),
					resource.TestCheckResourceAttr(name, "overlimit_mode", "blocking"),
				),
			},
			{
				Config: testWallarmGlobalModeOverlimitConfig(rnd, 500, "monitoring"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "overlimit_time", "500"),
					resource.TestCheckResourceAttr(name, "overlimit_mode", "monitoring"),
				),
			},
		},
	})
}

func TestAccWallarmGlobalMode_FullWithOverlimit(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_global_mode." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGlobalModeAllConfig(rnd, "block", "on", 1000, "blocking"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "filtration_mode", "block"),
					resource.TestCheckResourceAttr(name, "rechecker_mode", "on"),
					resource.TestCheckResourceAttr(name, "overlimit_time", "1000"),
					resource.TestCheckResourceAttr(name, "overlimit_mode", "blocking"),
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

func testWallarmGlobalModeScannerConfig(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
}`, resourceID)
}

func testWallarmGlobalModeRecheckerConfig(resourceID, recheckerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  rechecker_mode = "%[2]s"
}`, resourceID, recheckerMode)
}

func testWallarmGlobalModeFullConfig(resourceID, filtrationMode, recheckerMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  filtration_mode = "%[2]s"
  rechecker_mode = "%[3]s"
}`, resourceID, filtrationMode, recheckerMode)
}

func testWallarmGlobalModeOverlimitConfig(resourceID string, overlimitTime int, overlimitMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  overlimit_time = %[2]d
  overlimit_mode = "%[3]s"
}`, resourceID, overlimitTime, overlimitMode)
}

func testWallarmGlobalModeAllConfig(resourceID, filtrationMode, recheckerMode string, overlimitTime int, overlimitMode string) string {
	return fmt.Sprintf(`
resource "wallarm_global_mode" "%[1]s" {
  filtration_mode = "%[2]s"
  rechecker_mode  = "%[3]s"
  overlimit_time  = %[4]d
  overlimit_mode  = "%[5]s"
}`, resourceID, filtrationMode, recheckerMode, overlimitTime, overlimitMode)
}
