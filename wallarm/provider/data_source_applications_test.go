package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceApplications(t *testing.T) {
	rnd := generateRandomResourceName(5)
	rndID := generateRandomNumber(100000)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceApplicationsConfig(rnd, "tf-test-"+rnd, rndID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceApplications("data.wallarm_applications.all"),
				),
			},
		},
	})
}

func testAccCheckDataSourceApplications(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		appCount, err := strconv.Atoi(rs.Primary.Attributes["applications.#"])
		if err != nil {
			return err
		}

		if appCount == 0 {
			return fmt.Errorf("no applications returned by data source")
		}

		return nil
	}
}

func testAccDataSourceApplicationsConfig(resourceID, name, appID string) string {
	return fmt.Sprintf(`
resource "wallarm_application" "%[1]s" {
  name   = "%[2]s"
  app_id = %[3]s
}

data "wallarm_applications" "all" {
  depends_on = [wallarm_application.%[1]s]
}`, resourceID, name, appID)
}
