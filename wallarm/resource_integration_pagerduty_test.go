package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIntegrationPagerDutyRequiredFields(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_pagerduty." + rnd
	rndKey := generateRandomResourceName(32)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationPagerDutyRequiredOnly(rnd, rndKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "integration_key", rndKey),
				),
			},
		},
	})
}

func TestAccIntegrationPagerDutyFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_pagerduty." + rnd
	rndKey := generateRandomResourceName(32)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationPagerDutyFullConfig(rnd, "tf-test-"+rnd, rndKey, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "integration_key", rndKey),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
		},
	})
}

func TestAccIntegrationPagerDutyIncorrectKeyLength(t *testing.T) {
	rnd := generateRandomResourceName(5)
	rndKey := generateRandomResourceName(28)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmIntegrationPagerDutyRequiredOnly(rnd, rndKey),
				ExpectError: regexp.MustCompile(`length of "[a-z_]+" must be equal to 32, got: [0-9]+`),
			},
		},
	})
}

func TestAccIntegrationPagerDutyCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_pagerduty." + rnd
	rndKey := generateRandomResourceName(32)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationPagerDutyFullConfig(rnd, "tf-test-"+rnd, rndKey, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "integration_key", rndKey),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
			{
				Config: testWallarmIntegrationPagerDutyFullConfig(rnd, "tf-updated-"+rnd, rndKey, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "integration_key", rndKey),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
		},
	})
}

func testWallarmIntegrationPagerDutyRequiredOnly(resourceID, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_pagerduty" "%[1]s" {
	integration_key = "%[2]s"
}`, resourceID, token)
}

func testWallarmIntegrationPagerDutyFullConfig(resourceID, name, token, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_pagerduty" "%[1]s" {
	name = "%[2]s"
	integration_key = "%[3]s"
	active = %[4]s

	event {
		event_type = "system"
		active = %[4]s
	}
	event {
		event_type = "scope"
		active = %[4]s
	}
	event {
		event_type = "hit"
		active = %[4]s
	}
	event {
		event_type = "vuln_high"
		active = "%[4]s"
	}
	event {
		event_type = "vuln_medium"
		active = "%[4]s"
	}
	event {
		event_type = "vuln_low"
		active = "%[4]s"
	}
}`, resourceID, name, token, active)
}
