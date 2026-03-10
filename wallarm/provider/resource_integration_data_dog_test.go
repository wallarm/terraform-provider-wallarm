package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIntegrationDataDogRequiredFields(t *testing.T) {
	name := "wallarm_integration_data_dog.test"
	rndToken := generateRandomUUID() // 32+ hex chars

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationDataDogRequiredOnly(rndToken, "US1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "region", "US1"),
				),
			},
		},
	})
}

func TestAccIntegrationDataDogFullSettings(t *testing.T) {
	name := "wallarm_integration_data_dog.test"
	rnd := generateRandomResourceName(10)
	rndToken := generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationDataDogFullConfig("tf-test-"+rnd, rndToken, "US1", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "region", "US1"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
		},
	})
}

func TestAccIntegrationDataDogCreateThenUpdate(t *testing.T) {
	name := "wallarm_integration_data_dog.test"
	rnd := generateRandomResourceName(10)
	rndToken := generateRandomUUID() + generateRandomUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationDataDogFullConfig("tf-test-"+rnd, rndToken, "US1", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "region", "US1"),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
			{
				Config: testWallarmIntegrationDataDogFullConfig("tf-updated-"+rnd, rndToken, "US1", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "region", "US1"),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
		},
	})
}

func testWallarmIntegrationDataDogRequiredOnly(token, region string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_data_dog" "test" {
	token = "%[1]s"
	region = "%[2]s"
}`, token, region)
}

func testWallarmIntegrationDataDogFullConfig(name, token, region, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_data_dog" "test" {
	name = "%[1]s"
	token = "%[2]s"
	region = "%[3]s"
	active = %[4]s

	event {
		event_type = "siem"
		active = true
		with_headers = true
	}
	event {
		event_type = "rules_and_triggers"
		active = %[4]s
	}
	event {
		event_type = "number_of_requests_per_hour"
		active = %[4]s
	}
	event {
		event_type = "security_issue_critical"
		active = %[4]s
	}
	event {
		event_type = "security_issue_high"
		active = %[4]s
	}
	event {
		event_type = "security_issue_medium"
		active = %[4]s
	}
	event {
		event_type = "security_issue_low"
		active = %[4]s
	}
	event {
		event_type = "security_issue_info"
		active = %[4]s
	}
	event {
		event_type = "system"
		active = true
	}

}`, name, token, region, active)
}
