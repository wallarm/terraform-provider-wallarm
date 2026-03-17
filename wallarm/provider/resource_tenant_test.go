package wallarm

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"prevent_destroy"},
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
		client := testAccProvider.Meta().(*ProviderMeta).Client

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "wallarm_tenant" {
				continue
			}

			tenantClientID, err := strconv.Atoi(rs.Primary.ID)
			if err != nil {
				return err
			}

			tenant, err := readTenantByID(client, tenantClientID)
			if err != nil {
				return err
			}

			if tenant != nil && tenant.Enabled && tenant.Name == name {
				return fmt.Errorf("Resource still exists: %s", name)
			}
		}

		return nil
	}
}

func testAccCheckWallarmTenantExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ProviderMeta).Client

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		tenantClientID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		name := rs.Primary.Attributes["name"]

		tenant, err := readTenantByID(client, tenantClientID)
		if err != nil {
			return err
		}

		if tenant != nil && tenant.Name == name {
			return nil
		}

		return fmt.Errorf("WallarmTenant not found: %s", name)
	}
}
