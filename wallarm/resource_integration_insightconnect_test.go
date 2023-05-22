package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

var (
	insightConnectURL = "https://example.com/insight/connect"
)

func TestAccIntegrationInsightConnectRequiredFields(t *testing.T) {
	name := "wallarm_integration_insightconnect.test"
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationInsightConnectRequiredOnly(insightConnectURL, rndToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "api_url", insightConnectURL),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
				),
			},
		},
	})
}

func TestAccIntegrationInsightConnectFullSettings(t *testing.T) {
	name := "wallarm_integration_insightconnect.test"
	rnd := generateRandomResourceName(10)
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationInsightConnectFullConfig("tf-test-"+rnd, insightConnectURL, rndToken, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", insightConnectURL),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
		},
	})
}

func TestAccIntegrationInsightConnectCreateThenUpdate(t *testing.T) {
	name := "wallarm_integration_insightconnect.test"
	rnd := generateRandomResourceName(10)
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationInsightConnectFullConfig("tf-test-"+rnd, insightConnectURL, rndToken, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", insightConnectURL),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
			{
				Config: testWallarmIntegrationInsightConnectFullConfig("tf-updated-"+rnd, insightConnectURL, rndToken, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", insightConnectURL),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
		},
	})
}

func testWallarmIntegrationInsightConnectRequiredOnly(url, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_insightconnect" "test" {
	api_url = "%[1]s"
	api_token = "%[2]s"
}`, url, token)
}

func testWallarmIntegrationInsightConnectFullConfig(name, url, token, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_insightconnect" "test" {
	name = "%[1]s"
	api_url = "%[2]s"
	api_token = "%[3]s"
	active = %[4]s

	event {
		event_type = "hit"
		active = true
	}
	event {
		event_type = "vuln_high"
		active = %[4]s
	}
	event {
		event_type = "vuln_medium"
		active = %[4]s
	}
	event {
		event_type = "vuln_low"
		active = %[4]s
	}
	event {
		event_type = "system"
		active = true
	}
	event {
		event_type = "scope"
		active = %[4]s
	}

}`, name, url, token, active)
}
