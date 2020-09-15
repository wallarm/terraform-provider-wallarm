package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIntegrationInsightConnectRequiredFields(t *testing.T) {
	name := "wallarm_integration_insightconnect.test"
	rnd := generateRandomResourceName(10)
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationInsightConnectRequiredOnly(fmt.Sprintf("https://%s.wallarm.com", rnd), rndToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "api_url", fmt.Sprintf("https://%s.wallarm.com", rnd)),
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
				Config: testWallarmIntegrationInsightConnectFullConfig("tf-test-"+rnd, fmt.Sprintf("https://%s.wallarm.com", rnd), rndToken, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", fmt.Sprintf("https://%s.wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
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
				Config: testWallarmIntegrationInsightConnectFullConfig("tf-test-"+rnd, fmt.Sprintf("https://%s.wallarm.com", rnd), rndToken, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", fmt.Sprintf("https://%s.wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
				),
			},
			{
				Config: testWallarmIntegrationInsightConnectFullConfig("tf-updated-"+rnd, fmt.Sprintf("https://%s.wallarm.com", rnd), rndToken, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "api_url", fmt.Sprintf("https://%s.wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "api_token", rndToken),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
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
		event_type = "vuln"
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
