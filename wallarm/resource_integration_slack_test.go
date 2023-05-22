package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIntegrationSlackRequiredFields(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_slack." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSlackRequiredOnly(rnd, "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "webhook_url", "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"),
				),
			},
		},
	})
}

func TestAccIntegrationSlackFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_slack." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSlackFullConfig(rnd, "tf-test-"+rnd, "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "5"),
				),
			},
		},
	})
}

func TestAccIntegrationSlackIncorrectEvents(t *testing.T) {
	rnd := generateRandomResourceName(5)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmIntegrationSlackIncorrectEvents(rnd, "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"),
				ExpectError: regexp.MustCompile(`event: attribute supports 5 item maximum, config has [0-9]+ declared`),
			},
		},
	})
}
func TestAccIntegrationSlackCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_slack." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSlackFullConfig(rnd, "tf-test-"+rnd, "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "5"),
				),
			},
			{
				Config: testWallarmIntegrationSlackFullConfig(rnd, "tf-updated-"+rnd, "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "5"),
				),
			},
		},
	})
}

func testWallarmIntegrationSlackRequiredOnly(resourceID, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_slack" "%[1]s" {
	webhook_url = "%[2]s"
}`, resourceID, token)
}

func testWallarmIntegrationSlackFullConfig(resourceID, name, token, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_slack" "%[1]s" {
	name = "%[2]s"
	webhook_url = "%[3]s"
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

func testWallarmIntegrationSlackIncorrectEvents(resourceID, url string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_slack" "%[1]s" {
	webhook_url = "%[2]s"
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
		event_type = "vuln_high"
		active = true
	}
	event {
		event_type = "vuln_medium"
		active = true
	}
	event {
		event_type = "vuln_low"
		active = true
	}
}`, resourceID, url)
}
