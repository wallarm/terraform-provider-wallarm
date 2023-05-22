package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

var (
	webhookURL = "https://example.com/webhook"
)

func TestAccIntegrationWebhookRequiredFields(t *testing.T) {
	name := "wallarm_integration_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationWebhookRequiredOnly(webhookURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "webhook_url", webhookURL),
				),
			},
		},
	})
}

func TestAccIntegrationWebhookFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_webhook." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationWebhookFullConfig(rnd, "tf-test-"+rnd, webhookURL, "POST", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg==", "application/json", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", webhookURL),
					resource.TestCheckResourceAttr(name, "http_method", "POST"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "headers.Authorization", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg=="),
					resource.TestCheckResourceAttr(name, "headers.Content-Type", "application/json"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
		},
	})
}

func TestAccIntegrationWebhookCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_webhook." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationWebhookFullConfig(rnd, "tf-test-"+rnd, webhookURL, "POST", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg==", "application/json", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", webhookURL),
					resource.TestCheckResourceAttr(name, "http_method", "POST"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "headers.Authorization", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg=="),
					resource.TestCheckResourceAttr(name, "headers.Content-Type", "application/json"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
				),
			},
			{
				Config: testWallarmIntegrationWebhookFullConfig(rnd, "tf-updated-"+rnd, webhookURL, "POST", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg==", "application/json", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", webhookURL),
					resource.TestCheckResourceAttr(name, "http_method", "POST"),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "headers.Authorization", "Basic SGkgYXR0ZW50aXZlIFdhbGxhcm0gdXNlcg=="),
					resource.TestCheckResourceAttr(name, "headers.Content-Type", "application/json"),
					resource.TestCheckResourceAttr(name, "event.#", "6"),
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

func testWallarmIntegrationWebhookFullConfig(resourceID, name, url, httpMethod, auth, cntype, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_webhook" "%[1]s" {
	name = "%[2]s"
	webhook_url = "%[3]s"
	http_method = "%[4]s"
	active = %[7]s

	event {
		event_type = "hit"
		active = true
	}
	event {
		event_type = "vuln_high"
		active = "%[7]s"
	}
	event {
		event_type = "vuln_medium"
		active = "%[7]s"
	}
	event {
		event_type = "vuln_low"
		active = "%[7]s"
	}
	event {
		event_type = "system"
		active = true
	}
	event {
		event_type = "scope"
		active = %[7]s
	}

	headers = {
		Authorization = "%[5]s"
		Content-Type = "%[6]s"
	}
}`, resourceID, name, url, httpMethod, auth, cntype, active)
}
