package wallarm

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccWallarmTenant(t *testing.T) {
	if os.Getenv("WALLARM_EXTRA_PERMISSIONS") == "" {
		t.Skip("Skipping not test as it requires WALLARM_EXTRA_PERMISSIONS set")
	}
	rnd := generateRandomResourceName(10)
	resourceName := "wallarm_tenant." + rnd
	name := "tf-test-" + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmTenantDestroy(name),
		Steps: []resource.TestStep{
			{
				Config: testWallarmTenantConfig(rnd, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmTenantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
				),
			},
		},
	})
}

func testWallarmTenantConfig(resourceID, name string) string {
	return fmt.Sprintf(`
resource "wallarm_tenant" "%[1]s" {
  name = "%[2]s"
}`, resourceID, name)
}

func testAccCheckWallarmTenantDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(wallarm.API)

		for _, resource := range s.RootModule().Resources {
			if resource.Type != name {
				continue
			}

			tenantID, err := strconv.Atoi(resource.Primary.Attributes["tenant_id"])
			if err != nil {
				return err
			}

			res, err := client.ClientUpdate(&wallarm.ClientUpdate{
				Filter: &wallarm.ClientFilter{ID: tenantID},
				Fields: &wallarm.ClientFields{Enabled: false},
			})
			if err != nil {
				return err
			}

			if res.Body[0].Name == name {
				return fmt.Errorf("Resource still exists: %s", name)
			}

			return nil
		}

		return nil
	}
}

func testAccCheckWallarmTenantExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(wallarm.API)

		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		tenantID, err := strconv.Atoi(resource.Primary.Attributes["tenant_id"])
		if err != nil {
			return err
		}

		name := resource.Primary.Attributes["name"]

		res, err := client.ClientRead(&wallarm.ClientRead{
			Limit: 1,
			Filter: &wallarm.ClientReadFilter{
				ClientFilter: wallarm.ClientFilter{ID: tenantID},
			},
		})
		if err != nil {
			return err
		}

		if res.Body[0].Name == name {
			return nil
		}

		return fmt.Errorf("WallarmTenant not found: %s", name)
	}
}
