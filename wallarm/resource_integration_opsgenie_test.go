package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

var (
	opsgenieAlertsEndpoint = "https://api.opsgenie.com/v2/alerts"
)

func TestAccIntegrationOpsGenieRequiredFields(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_opsgenie." + rnd
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationOpsGenieRequiredOnly(rnd, opsgenieAlertsEndpoint, rndToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "api_url", opsgenieAlertsEndpoint),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
				),
			},
		},
	})
}

func TestAccIntegrationOpsGenieFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_opsgenie." + rnd
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationOpsGenieFullConfig(rnd, "tf-test-"+rnd, opsgenieAlertsEndpoint, rndToken, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", opsgenieAlertsEndpoint),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
				),
			},
		},
	})
}

func TestAccIntegrationOpsGenieCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_opsgenie." + rnd
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationOpsGenieFullConfig(rnd, "tf-test-"+rnd, opsgenieAlertsEndpoint, rndToken, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", opsgenieAlertsEndpoint),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
				),
			},
			{
				Config: testWallarmIntegrationOpsGenieFullConfig(rnd, "tf-updated-"+rnd, opsgenieAlertsEndpoint, rndToken, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", opsgenieAlertsEndpoint),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
				),
			},
		},
	})
}

func testWallarmIntegrationOpsGenieRequiredOnly(resourceID, url, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_opsgenie" "%[1]s" {
    api_url = "%[2]s"
	api_token = "%[3]s"
}`, resourceID, url, token)
}

func testWallarmIntegrationOpsGenieFullConfig(resourceID, name, url, token, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_opsgenie" "%[1]s" {
	name = "%[2]s"
	api_url = "%[3]s"
	api_token = "%[4]s"
	active = "%[5]s"

	event {
		event_type = "hit"
		active = true
	}
	event {
		event_type = "vuln_high"
		active = "%[5]s"
	}
	event {
		event_type = "vuln_medium"
		active = "%[5]s"
	}
	event {
		event_type = "vuln_low"
		active = "%[5]s"
	}
}`, resourceID, name, url, token, active)
}
