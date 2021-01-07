package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	wallarm "github.com/416e64726579/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccWallarmTriggerOnlyRequiredWithError(t *testing.T) {
	rnd := generateRandomResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testWallarmTriggerOnlyRequiredWithError(rnd, "attacks_exceeded", "send_notification", "wallarm_integration_email.test.integration_id"),
				ExpectError: ArgumentMustBePresented("threshold", "attacks_exceeded"),
			},
			{
				Config:      testWallarmTriggerOnlyRequiredWithError(rnd, "hits_exceeded", "send_notification", "wallarm_integration_email.test.integration_id"),
				ExpectError: ArgumentMustBePresented("threshold", "hits_exceeded"),
			},
			{
				Config:      testWallarmTriggerOnlyRequiredWithError(rnd, "incidents_exceeded", "send_notification", "wallarm_integration_email.test.integration_id"),
				ExpectError: ArgumentMustBePresented("threshold", "incidents_exceeded"),
			},
			{
				Config:      testWallarmTriggerOnlyRequiredWithError(rnd, "vector_attack", "send_notification", "wallarm_integration_email.test.integration_id"),
				ExpectError: ArgumentMustBePresented("threshold", "vector_attack"),
			},
			{
				Config:      testWallarmTriggerOnlyRequiredWithError(rnd, "bruteforce_started", "send_notification", "wallarm_integration_email.test.integration_id"),
				ExpectError: ArgumentMustBePresented("threshold", "bruteforce_started"),
			},
		},
	})
}

func TestAccWallarmTriggerAttacksWithThreshold(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_trigger." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmTriggerAttacksWithThreshold(rnd, "attacks_exceeded", "send_notification", "wallarm_integration_email.test.integration_id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "template_id", "attacks_exceeded"),
					resource.TestCheckResourceAttr(name, "actions.#", "1"),
				),
			},
		},
	})
}

func TestAccWallarmTriggerAttacksWithFilters(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_trigger." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmTriggerAttacksWithFilters(rnd, "attacks_exceeded", "send_notification", "wallarm_integration_email.test.integration_id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "template_id", "attacks_exceeded"),
					resource.TestCheckResourceAttr(name, "actions.#", "1"),
					resource.TestCheckResourceAttr(name, "filters.#", "6"),
				),
			},
		},
	})
}

func TestAccWallarmTriggerAttacksWithResponse5xx(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_trigger." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmTriggerAttacksWithResponse5xx(rnd, "attacks_exceeded", "send_notification"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "template_id", "attacks_exceeded"),
					resource.TestCheckResourceAttr(name, "actions.#", "1"),
					resource.TestCheckResourceAttr(name, "filters.#", "6"),
				),
			},
		},
	})
}

func TestAccWallarmTriggerBruteforce(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_trigger." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWallarmTriggerBruteforce(rnd, "bruteforce_started", "mark_as_brute", "block_ips"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "template_id", "bruteforce_started"),
					resource.TestCheckResourceAttr(name, "actions.#", "2"),
					resource.TestCheckResourceAttr(name, "filters.#", "2"),
					resource.TestCheckResourceAttr(name, "filters.0.filter_id", "url"),
					resource.TestCheckResourceAttr(name, "filters.0.value.0", "example.com:443/brute"),
					resource.TestCheckResourceAttr(name, "filters.1.filter_id", "ip_address"),
					resource.TestCheckResourceAttr(name, "filters.1.value.0", "1.1.1.1"),
					resource.TestCheckResourceAttr(name, "threshold.%", "3"),
					resource.TestCheckResourceAttr(name, "threshold.count", "30"),
					resource.TestCheckResourceAttr(name, "threshold.period", "30"),
				),
			},
		},
	})
}

func testWallarmTriggerOnlyRequiredWithError(resourceID, templateID, actionID, integrationID string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_email" "test" {
	name = "New Terraform Integration"
	emails = ["%[1]s@wallarm.com"]
	
	event {
		event_type = "vuln"
		active = true
	}
}
	
resource "wallarm_trigger" "%[1]s" {
	template_id = "%[2]s"
	actions {
		action_id = "%[3]s"
		integration_id = [%[4]s]
	}
}`, resourceID, templateID, actionID, integrationID)
}

func testWallarmTriggerAttacksWithThreshold(resourceID, templateID, actionID, integrationID string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_email" "test" {
	name = "New Terraform Integration"
	emails = ["%[1]s@wallarm.com"]
	
	event {
		event_type = "vuln"
		active = true
	}
}
	
resource "wallarm_trigger" "%[1]s" {
	template_id = "%[2]s"
	actions {
		action_id = "%[3]s"
		integration_id = [%[4]s]
	}

	threshold = {
		period = 86400
		operator = "gt"
		count = 10000
	}
}`, resourceID, templateID, actionID, integrationID)
}

func testWallarmTriggerAttacksWithFilters(resourceID, templateID, actionID, integrationID string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_email" "test" {
	name = "New Terraform Integration"
	emails = ["%[1]s@wallarm.com"]
	
	event {
		event_type = "vuln"
		active = true
	}
}
	
resource "wallarm_trigger" "%[1]s" {
	template_id = "%[2]s"
	actions {
		action_id = "%[3]s"
		integration_id = [%[4]s]
	}

	filters {
		filter_id = "ip_address"
		operator = "eq"
		value = ["1.1.1.1"]
	}
	
	filters {
		filter_id = "pool"
		operator = "eq"
		value = [-1]
	}

	filters {
		filter_id = "attack_type"
		operator = "eq"
		value = ["sqli"]
	}

	filters {
		filter_id = "domain"
		operator = "eq"
		value = ["example.com"]
	}

	filters {
		filter_id = "target"
		operator = "eq"
		value = ["server", "database"]
	}

	filters {
		filter_id = "response_status"
		operator = "eq"
		value = [504, 503]
	}

	threshold = {
		period = 86400
		operator = "gt"
		count = 10000
	}
}`, resourceID, templateID, actionID, integrationID)
}

func testWallarmTriggerAttacksWithResponse5xx(resourceID, templateID, actionID string) string {
	return fmt.Sprintf(`

resource "wallarm_application" "%[1]s" {
		name = "tf-testacc-app"
		app_id = 42
}

resource "wallarm_integration_email" "%[1]s" {
	name = "New Terraform Integration"
	emails = ["%[1]s@wallarm.com"]
	
	event {
		event_type = "vuln"
		active = true
	}
}
	
resource "wallarm_trigger" "%[1]s" {
	template_id = "%[2]s"
	actions {
		action_id = "%[3]s"
		integration_id = [wallarm_integration_email.%[1]s.integration_id]
	}

	filters {
		filter_id = "ip_address"
		operator = "eq"
		value = ["1.1.1.1", "2.2.2.2"]
	}
	
	filters {
		filter_id = "pool"
		operator = "eq"
		value = [wallarm_application.%[1]s.app_id]
	}

	filters {
		filter_id = "attack_type"
		operator = "eq"
		value = ["xss"]
	}

	filters {
		filter_id = "domain"
		operator = "eq"
		value = ["example.com"]
	}

	filters {
		filter_id = "target"
		operator = "eq"
		value = ["client"]
	}

	filters {
		filter_id = "response_status"
		operator = "eq"
		value = ["5xx", "4xx"]
	}

	threshold = {
		period = 1
		operator = "gt"
		count = 1
	}
}`, resourceID, templateID, actionID)
}

func testWallarmTriggerBruteforce(resourceID, templateID, actionID, actionIDExtra string) string {
	return fmt.Sprintf(`
resource "wallarm_trigger" "%[1]s" {
	template_id = "%[2]s"

	filters {
		filter_id = "url"
		operator = "eq"
		value = ["example.com:443/brute"]
	}
	
	filters {
		filter_id = "ip_address"
		operator = "eq"
		value = ["1.1.1.1"]
	}
	
	actions {
		action_id = "%[3]s"
	}
	  
	actions {
		action_id = "%[4]s"
		lock_time = 2592000
	}

	threshold = {
		period = 30
		operator = "gt"
		count = 30
	}
}`, resourceID, templateID, actionID, actionIDExtra)
}

func testAccCheckWallarmTriggerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(wallarm.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_trigger" {
			continue
		}

		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return err
		}

		triggerID, err := strconv.Atoi(rs.Primary.Attributes["trigger_id"])
		if err != nil {
			return err
		}

		triggers, err := client.TriggerRead(clientID)
		if err != nil {
			return nil
		}

		for _, t := range triggers.Triggers {
			if t.ID == triggerID {
				return fmt.Errorf("Wallarm Trigger %d for client %d still exists", triggerID, clientID)
			}
		}
	}
	return nil
}
