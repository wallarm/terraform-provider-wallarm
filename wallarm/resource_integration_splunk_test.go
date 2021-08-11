package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

var (
	splunkURL = "https://example.com/splunk"
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
					resource.TestCheckResourceAttr(name, "event.#", "4"),
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
					resource.TestCheckResourceAttr(name, "event.#", "4"),
				),
			},
			{
				Config: testWallarmIntegrationSplunkFullConfig(rnd, "tf-updated-"+rnd, apiURL, "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", apiURL),
					resource.TestCheckResourceAttr(name, "api_token", "b1e2d6dc-e4b5-400d-9dae-270c39c5daa2"),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
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
		event_type = "system"
		active = true
	}
	event {
		event_type = "scope"
		active = %[5]s
	}
	event {
		event_type = "vuln"
		active = true
	}
	event {
		event_type = "hit"
		active = %[5]s
	}
}`, resourceID, name, url, token, active)
}
