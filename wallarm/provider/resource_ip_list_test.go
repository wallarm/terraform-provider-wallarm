package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// These tests cover wallarm_denylist (which delegates to the shared
// resourceWallarmIPList implementation) with non-IP rule types:
// country, datacenter, and proxy_type.  The ip_range / subnet path
// is already tested in resource_denylist_test.go and resource_allowlist_test.go.

func TestAccWallarmDenylistCountry(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_denylist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmDenylistCountry(rnd, "tf-test-"+rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "country.#", "2"),
				),
			},
		},
	})
}

func TestAccWallarmDenylistDatacenter(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_denylist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmDenylistDatacenter(rnd, "tf-test-"+rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "datacenter.#", "1"),
				),
			},
		},
	})
}

func TestAccWallarmDenylistProxyType(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_denylist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmDenylistProxyType(rnd, "tf-test-"+rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "proxy_type.#", "1"),
				),
			},
		},
	})
}

func TestAccWallarmDenylistForever(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_denylist." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmDenylistForever(rnd, "tf-test-"+rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "reason", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "time_format", "Forever"),
				),
			},
		},
	})
}

func TestAccWallarmDenylistConflictingFields(t *testing.T) {
	rnd := generateRandomResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmDenylistConflict(rnd),
				ExpectError: regexp.MustCompile(`conflicts with`),
			},
		},
	})
}

func testWallarmDenylistCountry(resourceID, reason string) string {
	return fmt.Sprintf(`
resource "wallarm_denylist" "%[1]s" {
  country     = ["US", "DE"]
  reason      = "%[2]s"
  time_format = "Minutes"
  time        = 60
}`, resourceID, reason)
}

func testWallarmDenylistDatacenter(resourceID, reason string) string {
	return fmt.Sprintf(`
resource "wallarm_denylist" "%[1]s" {
  datacenter  = ["aws"]
  reason      = "%[2]s"
  time_format = "Minutes"
  time        = 60
}`, resourceID, reason)
}

func testWallarmDenylistProxyType(resourceID, reason string) string {
	return fmt.Sprintf(`
resource "wallarm_denylist" "%[1]s" {
  proxy_type  = ["TOR"]
  reason      = "%[2]s"
  time_format = "Minutes"
  time        = 60
}`, resourceID, reason)
}

func testWallarmDenylistForever(resourceID, reason string) string {
	return fmt.Sprintf(`
resource "wallarm_denylist" "%[1]s" {
  ip_range    = ["10.0.0.1"]
  reason      = "%[2]s"
  time_format = "Forever"
}`, resourceID, reason)
}

func testWallarmDenylistConflict(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_denylist" "%[1]s" {
  ip_range    = ["10.0.0.1"]
  country     = ["US"]
  reason      = "conflict-test"
  time_format = "Minutes"
  time        = 60
}`, resourceID)
}
