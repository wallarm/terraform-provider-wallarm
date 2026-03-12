package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIntegrationTeamsRequiredFields(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_teams." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationTeamsRequiredOnly(rnd, "https://xxxxx.webhook.office.com/xxxxxxxxx"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "webhook_url", "https://xxxxx.webhook.office.com/xxxxxxxxx"),
				),
			},
		},
	})
}

func TestAccIntegrationTeamsFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_teams." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationTeamsFullConfig(rnd, "tf-test-"+rnd, "https://xxxxx.webhook.office.com/xxxxxxxxx", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", "https://xxxxx.webhook.office.com/xxxxxxxxx"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "7"),
				),
			},
		},
	})
}

func TestAccIntegrationTeamsIncorrectEvents(t *testing.T) {
	rnd := generateRandomResourceName(5)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmIntegrationTeamsIncorrectEvents(rnd, "https://xxxxx.webhook.office.com/xxxxxxxxx"),
				ExpectError: regexp.MustCompile(`expected .* to be one of \[`),
			},
		},
	})
}

func TestAccIntegrationTeamsCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_teams." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationTeamsFullConfig(rnd, "tf-test-"+rnd, "https://xxxxx.webhook.office.com/xxxxxxxxx", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", "https://xxxxx.webhook.office.com/xxxxxxxxx"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "7"),
				),
			},
			{
				Config: testWallarmIntegrationTeamsFullConfig(rnd, "tf-updated-"+rnd, "https://xxxxx.webhook.office.com/xxxxxxxxx", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "webhook_url", "https://xxxxx.webhook.office.com/xxxxxxxxx"),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "7"),
				),
			},
		},
	})
}

func testWallarmIntegrationTeamsRequiredOnly(resourceID, url string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_teams" "%[1]s" {
	webhook_url = "%[2]s"
}`, resourceID, url)
}

func testWallarmIntegrationTeamsFullConfig(resourceID, name, url, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_teams" "%[1]s" {
	name = "%[2]s"
	webhook_url = "%[3]s"
	active = %[4]s

	event {
		event_type = "system"
		active = true
	}
	event {
		event_type = "rules_and_triggers"
		active = "%[4]s"
	}
	event {
		event_type = "security_issue_critical"
		active = "%[4]s"
	}
	event {
		event_type = "security_issue_high"
		active = "%[4]s"
	}
	event {
		event_type = "security_issue_medium"
		active = "%[4]s"
	}
	event {
		event_type = "security_issue_low"
		active = "%[4]s"
	}
	event {
		event_type = "security_issue_info"
		active = "%[4]s"
	}
}`, resourceID, name, url, active)
}

func testWallarmIntegrationTeamsIncorrectEvents(resourceID, url string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_teams" "%[1]s" {
	webhook_url = "%[2]s"
	active = true

	event {
		event_type = "siem"
		active = true
	}
}`, resourceID, url)
}
