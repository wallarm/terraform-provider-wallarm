package wallarm

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIntegrationTelegramRequiredFields(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_telegram." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationTelegramRequiredOnly(rnd, "WallarmIntegrationTest", "+y86q0LOQ4QG3hK9QgVDfw=="),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "telegram_username", "WallarmIntegrationTest"),
				),
			},
		},
	})
}

func TestAccIntegrationTelegramFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_telegram." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationTelegramFullConfig(rnd, "tf-test-"+rnd, "WallarmIntegrationTest", "+y86q0LOQ4QG3hK9QgVDfw==", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "telegram_username", "WallarmIntegrationTest"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "10"),
				),
			},
		},
	})
}

func TestAccIntegrationTelegramIncorrectEvents(t *testing.T) {
	rnd := generateRandomResourceName(5)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmIntegrationTelegramIncorrectEvents(rnd, "WallarmIntegrationTest", "+y86q0LOQ4QG3hK9QgVDfw=="),
				ExpectError: regexp.MustCompile(`expected .* to be one of \[`),
			},
		},
	})
}

func TestAccIntegrationTelegramCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_telegram." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationTelegramFullConfig(rnd, "tf-test-"+rnd, "WallarmIntegrationTest", "+y86q0LOQ4QG3hK9QgVDfw==", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "10"),
				),
			},
			{
				Config: testWallarmIntegrationTelegramFullConfig(rnd, "tf-updated-"+rnd, "WallarmIntegrationTest", "+y86q0LOQ4QG3hK9QgVDfw==", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "10"),
				),
			},
		},
	})
}

func testWallarmIntegrationTelegramRequiredOnly(resourceID, telegramUsername, chatData string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_telegram" "%[1]s" {
	telegram_username = "%[2]s"
	chat_data         = "%[3]s"
}`, resourceID, telegramUsername, chatData)
}

func testWallarmIntegrationTelegramFullConfig(resourceID, name, telegramUsername, chatData, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_telegram" "%[1]s" {
	name              = "%[2]s"
	telegram_username = "%[3]s"
	chat_data         = "%[4]s"
	active            = %[5]s

	event {
		event_type = "system"
		active = true
	}
	event {
		event_type = "rules_and_triggers"
		active = "%[5]s"
	}
	event {
		event_type = "security_issue_critical"
		active = "%[5]s"
	}
	event {
		event_type = "security_issue_high"
		active = "%[5]s"
	}
	event {
		event_type = "security_issue_medium"
		active = "%[5]s"
	}
	event {
		event_type = "security_issue_low"
		active = "%[5]s"
	}
	event {
		event_type = "security_issue_info"
		active = "%[5]s"
	}
	event {
		event_type = "report_daily"
		active = "%[5]s"
	}
	event {
		event_type = "report_weekly"
		active = "%[5]s"
	}
	event {
		event_type = "report_monthly"
		active = "%[5]s"
	}
}`, resourceID, name, telegramUsername, chatData, active)
}

func testWallarmIntegrationTelegramIncorrectEvents(resourceID, telegramUsername, chatData string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_telegram" "%[1]s" {
	telegram_username = "%[2]s"
	chat_data         = "%[3]s"
	active            = true

	event {
		event_type = "siem"
		active = true
	}
}`, resourceID, telegramUsername, chatData)
}
