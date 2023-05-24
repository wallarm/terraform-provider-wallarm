package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccIntegrationEmailRequiredFields(t *testing.T) {
	name := "wallarm_integration_email.test"
	rnd := generateRandomResourceName(8)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationEmailRequiredOnly(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "emails.0", fmt.Sprintf("%s@wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "emails.1", fmt.Sprintf("%s-rnd@wallarm.com", rnd)),
				),
			},
		},
	})
}

func TestAccIntegrationEmailFullSettings(t *testing.T) {
	name := "wallarm_integration_email.test"
	rnd := generateRandomResourceName(8)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationEmailFullConfig("tf-test-"+rnd, rnd, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "emails.0", fmt.Sprintf("%s@wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "emails.1", fmt.Sprintf("%s-rnd@wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "8"),
				),
			},
		},
	})
}

func TestAccIntegrationEmailCreateThenUpdate(t *testing.T) {
	name := "wallarm_integration_email.test"
	rnd := generateRandomResourceName(8)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationEmailFullUpdatedConfig("tf-test-"+rnd, rnd, "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "emails.0", fmt.Sprintf("%s@wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "emails.1", fmt.Sprintf("%s-rnd@wallarm.com", rnd)),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "8"),
				),
			},
			{
				Config: testWallarmIntegrationEmailFullUpdatedConfig("tf-test-"+rnd, rnd+"updated", "false", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "emails.0", fmt.Sprintf("%s@wallarm.com", rnd+"updated")),
					resource.TestCheckResourceAttr(name, "emails.1", fmt.Sprintf("%s-rnd@wallarm.com", rnd+"updated")),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "8"),
				),
			},
			{
				Config: testWallarmIntegrationEmailFullUpdatedConfig("tf-test-"+rnd, rnd+"updated", "true", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "emails.0", fmt.Sprintf("%s@wallarm.com", rnd+"updated")),
					resource.TestCheckResourceAttr(name, "emails.1", fmt.Sprintf("%s-rnd@wallarm.com", rnd+"updated")),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "8"),
				),
			},
			{
				Config: testWallarmIntegrationEmailFullUpdatedConfig("tf-test-new"+rnd, rnd+"new", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-new"+rnd),
					resource.TestCheckResourceAttr(name, "emails.0", fmt.Sprintf("%s@wallarm.com", rnd+"new")),
					resource.TestCheckResourceAttr(name, "emails.1", fmt.Sprintf("%s-rnd@wallarm.com", rnd+"new")),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "8"),
				),
			},
		},
	})
}

func testWallarmIntegrationEmailRequiredOnly(email string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_email" "test" {
	emails = ["%[1]s@wallarm.com", "%[1]s-rnd@wallarm.com"]
}`, email)
}

func testWallarmIntegrationEmailFullConfig(name, email, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_email" "test" {
	name = "%[1]s"
	active = %[3]s
	emails = ["%[2]s@wallarm.com", "%[2]s-rnd@wallarm.com"]

	event {
		event_type = "report_daily"
		active = true
	}
	event {
		event_type = "report_weekly"
		active = %[3]s
	}
	event {
		event_type = "report_monthly"
		active = true
	}
	event {
		event_type = "vuln_high"
		active = %[3]s
	}
	event {
		event_type = "vuln_medium"
		active = %[3]s
	}
	event {
		event_type = "vuln_low"
		active = %[3]s
	}
	event {
		event_type = "system"
		active = true
	}
	event {
		event_type = "scope"
		active = %[3]s
	}

}`, name, email, active)
}

func testWallarmIntegrationEmailFullUpdatedConfig(name, email, globalActive, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_email" "test" {
	name = "%[1]s"
	active = %[3]s
	emails = ["%[2]s@wallarm.com", "%[2]s-rnd@wallarm.com"]

	event {
		event_type = "report_daily"
		active = %[4]s
	}
	event {
		event_type = "report_weekly"
		active = %[4]s
	}
	event {
		event_type = "report_monthly"
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
	event {
		event_type = "system"
		active = %[4]s
	}
	event {
		event_type = "scope"
		active = %[4]s
	}

}`, name, email, globalActive, active)
}
