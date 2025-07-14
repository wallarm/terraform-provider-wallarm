package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccApp(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_application." + rnd
	rndID := generateRandomNumber(100000)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAppConfig(rnd, "tf-test-"+rnd, rndID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "app_id", rndID),
				),
			},
		},
	})
}

func TestAccWallarmApp_CreateThenUpdate(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_application." + rnd
	rndID := generateRandomNumber(100000)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAppConfig(rnd, "tf-test-"+rnd, rndID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "app_id", rndID),
				),
			},
			{
				Config: testWallarmAppConfig(rnd, "tf-test-"+rnd, rndID+"1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "app_id", rndID+"1"),
				),
			},
		},
	})
}

func TestAccWallarmApp_Existing(t *testing.T) {
	rnd := generateRandomResourceName(5)
	name := "wallarm_application." + rnd
	rndID := generateRandomNumber(100000)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testWallarmAppConfigRequiresImportError(rnd, "tf-test-"+rnd, rndID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmAppResourceExists(name),
					resource.TestCheckResourceAttr(name, "name", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(name, "app_id", rndID),
				),
				ExpectError: ResourceExistsError("[0-9]+/tf-test-[a-z]+/[0-9]+", "wallarm_application"),
			},
		},
	})
}

func testWallarmAppConfig(resourceID, hostname, appID string) string {
	return fmt.Sprintf(`
resource "wallarm_application" "%[1]s" {
  name = "%[2]s"
  app_id = %[3]s
}`, resourceID, hostname, appID)
}

func testWallarmAppConfigRequiresImportError(resourceID, hostname, appID string) string {
	return fmt.Sprintf(`
resource "wallarm_application" "%[1]s" {
	name = "%[2]s"
	app_id = %[3]s
}

resource "wallarm_application" "existing" {
  name = "%[2]s"
  app_id = %[3]s
}`, resourceID, hostname, appID)
}

func testAccCheckWallarmAppResourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		// retrieve the configured client from the test setup
		conn := testAccProvider.Meta().(wallarm.API)
		name := rs.Primary.Attributes["name"]
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return err
		}
		appID, err := strconv.Atoi(rs.Primary.Attributes["app_id"])
		if err != nil {
			return err
		}
		appRead := &wallarm.AppRead{
			Limit:  1000,
			Offset: 0,
			Filter: &wallarm.AppReadFilter{
				Clientid: []int{clientID},
			},
		}
		appResp, err := conn.AppRead(appRead)
		if err != nil {
			return err
		}

		for _, app := range appResp.Body {
			if app.ID != nil && *app.ID == appID || app.Name == name {
				return nil
			}
		}
		return fmt.Errorf("Wallarm Application (%s) not found", rs.Primary.ID)
	}
}
