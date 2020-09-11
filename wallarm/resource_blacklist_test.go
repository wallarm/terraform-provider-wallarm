package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccWallarmBlacklistMinutes(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_blacklist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmBlacklistMinutes(rnd, "tf-test-"+rnd, "60"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "60"),
				),
			},
		},
	})
}

func testWallarmBlacklistMinutes(resourceID, reason, time string) string {
	return fmt.Sprintf(`
resource "wallarm_blacklist" "%[1]s" {
	ip_range = ["1.1.1.1/30", "2.2.2.2", "3.3.3.3/32"]
	application = [1, 2]
	reason = "%[2]s"
	time = %[3]s # Minutes
}`, resourceID, reason, time)
}

func TestAccWallarmBlacklistBigSubnet(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_blacklist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmBlacklistBigSubnet(rnd, "tf-test-"+rnd, "60"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time", "60"),
				),
			},
		},
	})
}

func testWallarmBlacklistBigSubnet(resourceID, reason, time string) string {
	return fmt.Sprintf(`
resource "wallarm_blacklist" "%[1]s" {
	ip_range = ["3.3.3.3/23"]
	application = [1]
	reason = "%[2]s"
	time = %[3]s # Minutes
}`, resourceID, reason, time)
}
