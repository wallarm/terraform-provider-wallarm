package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmAllowlistMinutes(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_allowlist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAllowlistMinutes(rnd, "tf-test-"+rnd, "Minutes", "60"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "60"),
				),
			},
		},
	})
}

func TestAccWallarmAllowlistHours(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_allowlist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAllowlistHours(rnd, "tf-test-"+rnd, "Hours", "5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "5"),
				),
			},
		},
	})
}

func TestAccWallarmAllowlistDays(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_allowlist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAllowlistDays(rnd, "tf-test-"+rnd, "Days", "7"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "7"),
				),
			},
		},
	})
}

func TestAccWallarmAllowlistWeeks(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_allowlist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAllowlistWeeks(rnd, "tf-test-"+rnd, "Weeks", "4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "4"),
				),
			},
		},
	})
}

func TestAccWallarmAllowlistMonths(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_allowlist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAllowlistMonths(rnd, "tf-test-"+rnd, "Months", "12"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "12"),
				),
			},
		},
	})
}

func TestAccWallarmAllowlistRFC3339(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_allowlist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAllowlistRFC3339(rnd, "tf-test-"+rnd, "RFC3339", "2026-01-02T15:04:05+07:00"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "2026-01-02T15:04:05+07:00"),
				),
			},
		},
	})
}

func TestAccWallarmAllowlistBigSubnet(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_allowlist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAllowlistBigSubnet(rnd, "tf-test-"+rnd, "60"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "60"),
				),
			},
		},
	})
}

func testWallarmAllowlistMinutes(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = %[4]s # Minutes
}`, resourceID, reason, timeFormat, time)
}

func testWallarmAllowlistHours(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = %[4]s # Hours
}`, resourceID, reason, timeFormat, time)
}

func testWallarmAllowlistDays(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = %[4]s # Days
}`, resourceID, reason, timeFormat, time)
}

func testWallarmAllowlistWeeks(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = %[4]s # Weeks
}`, resourceID, reason, timeFormat, time)
}

func testWallarmAllowlistMonths(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = %[4]s # Months
}`, resourceID, reason, timeFormat, time)
}

func testWallarmAllowlistRFC3339(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = "%[4]s" # Date
}`, resourceID, reason, timeFormat, time)
}

func testWallarmAllowlistBigSubnet(resourceID, reason, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["3.3.3.3/23"]
	application = [1]
	reason = "%[2]s"
	time_format = "Minutes"
	time = %[3]s # Minutes
}`, resourceID, reason, time)
}

func TestAccWallarmAllowlistIncorrectSubnet(t *testing.T) {
	rnd := generateRandomResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmAllowlistIncorrectSubnet(rnd, "tf-test-"+rnd, "60"),
				ExpectError: regexp.MustCompile(`subnet must be >= /20, got [0-9]+`),
			},
		},
	})
}

func testWallarmAllowlistIncorrectSubnet(resourceID, reason, time string) string {
	return fmt.Sprintf(`
resource "wallarm_allowlist" "%[1]s" {
	ip_range = ["4.4.4.4/18"]
	application = [1]
	reason = "%[2]s"
	time_format = "Minutes"
	time = %[3]s # Minutes
}`, resourceID, reason, time)
}
