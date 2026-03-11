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
				Config: testWallarmIntegrationTelegramRequiredOnly(rnd, "testbot", "ytMxjwmqzIit067MD0vpSw=="),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "telegram_username", "testbot"),
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
				Config: testWallarmIntegrationTelegramFullConfig(rnd, "tf-test-"+rnd, "testbot", "ytMxjwmqzIit067MD0vpSw==", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "telegram_username", "testbot"),
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
				Config:      testWallarmIntegrationTelegramIncorrectEvents(rnd, "testbot", "ytMxjwmqzIit067MD0vpSw=="),
				ExpectError: regexp.MustCompile(`event: attribute supports 10 items maximum, config has [0-9]+ declared`),
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
				Config: testWallarmIntegrationTelegramFullConfig(rnd, "tf-test-"+rnd, "testbot", "ytMxjwmqzIit067MD0vpSw==", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "10"),
				),
			},
			{
				Config: testWallarmIntegrationTelegramFullConfig(rnd, "tf-updated-"+rnd, "testbot", "ytMxjwmqzIit067MD0vpSw==", "false"),
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
		event_type = "system"
		active = true
	}
	event {
		event_type = "rules_and_triggers"
		active = true
	}
	event {
		event_type = "security_issue_critical"
		active = true
	}
	event {
		event_type = "security_issue_high"
		active = true
	}
	event {
		event_type = "security_issue_medium"
		active = true
	}
	event {
		event_type = "security_issue_low"
		active = true
	}
	event {
		event_type = "security_issue_info"
		active = true
	}
	event {
		event_type = "report_daily"
		active = true
	}
	event {
		event_type = "report_weekly"
		active = true
	}
	event {
		event_type = "report_monthly"
		active = true
	}
	event {
		event_type = "system"
		active = false
	}
}`, resourceID, telegramUsername, chatData)
}
