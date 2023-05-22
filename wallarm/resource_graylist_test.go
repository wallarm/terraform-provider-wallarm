package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmGraylistMinutes(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_graylist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGraylistMinutes(rnd, "tf-test-"+rnd, "Minutes", "60"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "60"),
				),
			},
		},
	})
}

func TestAccWallarmGraylistRFC3339(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_graylist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGraylistRFC3339(rnd, "tf-test-"+rnd, "RFC3339", "2026-01-02T15:04:05+07:00"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "2026-01-02T15:04:05+07:00"),
				),
			},
		},
	})
}

func TestAccWallarmGraylistBigSubnet(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_graylist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmGraylistBigSubnet(rnd, "tf-test-"+rnd, "60"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "60"),
				),
			},
		},
	})
}

func testWallarmGraylistMinutes(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_graylist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = %[4]s # Minutes
}`, resourceID, reason, timeFormat, time)
}

func testWallarmGraylistRFC3339(resourceID, reason, timeFormat, time string) string {
	return fmt.Sprintf(`
resource "wallarm_graylist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time_format = "%[3]s"
	time = "%[4]s" # Date
}`, resourceID, reason, timeFormat, time)
}

func testWallarmGraylistBigSubnet(resourceID, reason, time string) string {
	return fmt.Sprintf(`
resource "wallarm_graylist" "%[1]s" {
	ip_range = ["3.3.3.3/23"]
	application = [1]
	reason = "%[2]s"
	time_format = "Minutes"
	time = %[3]s # Minutes
}`, resourceID, reason, time)
}

func TestAccWallarmGraylistIncorrectSubnet(t *testing.T) {
	rnd := generateRandomResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmGraylistIncorrectSubnet(rnd, "tf-test-"+rnd, "60"),
				ExpectError: regexp.MustCompile(`subnet must be >= /20, got [0-9]+`),
			},
		},
	})
}

func testWallarmGraylistIncorrectSubnet(resourceID, reason, time string) string {
	return fmt.Sprintf(`
resource "wallarm_graylist" "%[1]s" {
	ip_range = ["4.4.4.4/18"]
	application = [1]
	reason = "%[2]s"
	time_format = "Minutes"
	time = %[3]s # Minutes
}`, resourceID, reason, time)
}
