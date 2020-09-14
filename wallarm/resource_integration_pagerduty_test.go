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
				Config: testWallarmIntegrationPagerDutyFullConfig(rnd, "tf-test-"+rnd, rndKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "integration_key", rndKey),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
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

func testWallarmIntegrationPagerDutyRequiredOnly(resourceID, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_pagerduty" "%[1]s" {
	integration_key = "%[2]s"
}`, resourceID, token)
}

func testWallarmIntegrationPagerDutyFullConfig(resourceID, name, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_pagerduty" "%[1]s" {
	name = "%[2]s"
	integration_key = "%[3]s"
	active = true
	
	event {
		event_type = "system"
		active = true
	}
	event {
		event_type = "scope"
		active = true
	}
	event {
		event_type = "hit"
		active = true
	}
	event {
		event_type = "vuln"
		active = true
	}
}`, resourceID, name, token)
}
