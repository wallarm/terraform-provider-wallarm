package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIntegrationSumologicRequiredFields(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_sumologic." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSumologicRequiredOnly(rnd, "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA=="),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "sumologic_url", "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA=="),
				),
			},
		},
	})
}

func TestAccIntegrationSumologicFullSettings(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_sumologic." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSumologicFullConfig(rnd, "tf-test-"+rnd, "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA==", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "sumologic_url", "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA=="),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
		},
	})
}

func TestAccIntegrationSumologicCreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_integration_sumologic." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmIntegrationSumologicFullConfig(rnd, "tf-test-"+rnd, "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA==", "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "sumologic_url", "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA=="),
					resource.TestCheckResourceAttr(name, "active", "true"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
			{
				Config: testWallarmIntegrationSumologicFullConfig(rnd, "tf-updated-"+rnd, "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA==", "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-updated-"+rnd),
					resource.TestCheckResourceAttr(name, "sumologic_url", "https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/ZaVnC4dhaV123gN3o--AIj3q8y9GrwxSrAgcOJMvltRVnEIAIyR001VBlDsTYGBpieGxBxyJZA1eFIZcuyJ_ivkjPZ6Ynl8x3kLBJi4arZ479cD8ePJsqA=="),
					resource.TestCheckResourceAttr(name, "active", "false"),
					resource.TestCheckResourceAttr(name, "event.#", "9"),
				),
			},
		},
	})
}

func testWallarmIntegrationSumologicRequiredOnly(resourceID, token string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_sumologic" "%[1]s" {
	sumologic_url = "%[2]s"
}`, resourceID, token)
}

func testWallarmIntegrationSumologicFullConfig(resourceID, name, token, active string) string {
	return fmt.Sprintf(`
resource "wallarm_integration_sumologic" "%[1]s" {
	name = "%[2]s"
	sumologic_url = "%[3]s"
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
}`, resourceID, name, token, active)
}
