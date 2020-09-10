package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIntegrationWebhook_RequiredFields(t *testing.T) {
	name := "wallarm_integration_webhook.test"
	rnd := generateRandomResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationWebhookRequiredOnly(fmt.Sprintf("https://%s.wallarm.com", rnd)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "webhook_url", fmt.Sprintf("https://%s.wallarm.com", rnd)),
				),
			},
		},
	})
}

func TestAccIntegrationWebhook_FullSettings(t *testing.T) {
	name := "wallarm_integration_webhook.test"
	rnd := generateRandomResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationWebhookFullConfig("tf-test-"+rnd, fmt.Sprintf("https://%s.wallarm.com", rnd), "POST", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg==", "application/json"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", fmt.Sprintf("https://%s.wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "http_method", "POST"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "headers.Authorization", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg=="),
					resource.TestCheckResourceAttr(name, "headers.Content-Type", "application/json"),
					resource.TestCheckResourceAttr(name, "event.#", "4"),
				),
			},
		},
	})
}

func testWallarmIntegrationWebhookRequiredOnly(url string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_webhook" "test" {
	webhook_url = "%[1]s"
}`, url)
}

func testWallarmIntegrationWebhookFullConfig(name, url, httpMethod, auth, cntype string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_webhook" "test" {
	name = "%[1]s"
	webhook_url = "%[2]s"
	http_method = "%[3]s"
	active = true
	
	event {
		event_type = "hit"
		active = true
	}
	event {
		event_type = "vuln"
		active = true
	}
	event {
		event_type = "system"
		active = true
	}
	event {
		event_type = "scope"
		active = true
	}

	headers = {
		Authorization = "%[4]s"
		Content-Type = "%[5]s"
	}
}`, name, url, httpMethod, auth, cntype)
}
