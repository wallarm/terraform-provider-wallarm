package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	apiURL = "https://api.wallarm.com"
)

var (
	splunkURL = "https://httpbin.org:443"
)

func TestAccIntegrationSplunkRequiredFields(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_splunk." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSplunkRequiredOnly(rnd, splunkURL, "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "api_url", splunkURL),
					resource.TestCheckResourceAttr(name, "api_token", "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"),
				),
			},
		},
	})
}

func TestAccIntegrationSplunkFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_splunk." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSplunkFullConfig(rnd, "tf-test-"+rnd, splunkURL, "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", splunkURL),
					resource.TestCheckResourceAttr(name, "api_token", "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
		},
	})
}

func TestAccIntegrationSplunkCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_splunk." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSplunkFullConfig(rnd, "tf-test-"+rnd, splunkURL, "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", splunkURL),
					resource.TestCheckResourceAttr(name, "api_token", "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
			{
				Config: testWallarmIntegrationSplunkFullConfig(rnd, "tf-updated-"+rnd, apiURL, "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", apiURL),
					resource.TestCheckResourceAttr(name, "api_token", "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
		},
	})
}

func testWallarmIntegrationSplunkRequiredOnly(resourceID, url, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_splunk" "%[1]s" {
	api_url = "%[2]s"
    api_token = "%[3]s"
}`, resourceID, url, token)
}

func testWallarmIntegrationSplunkFullConfig(resourceID, name, url, token, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_splunk" "%[1]s" {
	name = "%[2]s"
	api_url = "%[3]s"
    api_token = "%[4]s"
	active = %[5]s

	event {
		event_type = "siem"
		active = true
		with_headers = true
	}
	event {
		event_type = "rules_and_triggers"
		active = %[5]s
	}
	event {
		event_type = "number_of_requests_per_hour"
		active = %[5]s
	}
	event {
		event_type = "security_issue_critical"
		active = %[5]s
	}
	event {
		event_type = "security_issue_high"
		active = %[5]s
	}
	event {
		event_type = "security_issue_medium"
		active = %[5]s
	}
	event {
		event_type = "security_issue_low"
		active = %[5]s
	}
	event {
		event_type = "security_issue_info"
		active = %[5]s
	}
	event {
		event_type = "system"
		active = true
	}
}`, resourceID, name, url, token, active)
}
