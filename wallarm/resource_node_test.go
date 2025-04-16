package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/wallarm/wallarm-go"
)

func TestAccWallarmNode(t *testing.T) {
	rnd := generateRandomResourceName(10)
	name := "wallarm_node." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmDestroy(name),
		Steps: []resource.TestStep{
			{
				Config: testWallarmNodeConfig(rnd, "tf-test-"+rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWallarmNodeExists(name),
					resource.TestCheckResourceAttr(name, "hostname", "tf-test-"+rnd),
				),
			},
		},
	})
}

func testWallarmNodeConfig(resourceID, hostname string) string {
	return fmt.Sprintf(`
resource "wallarm_node" "%[1]s" {
  hostname = "%[2]s"
}`, resourceID, hostname)
}

func testAccCheckWallarmDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(wallarm.API)

		for _, resource := range s.RootModule().Resources {
			if resource.Type != name {
				continue
			}

			clientID, err := strconv.Atoi(resource.Primary.Attributes["client_id"])
			if err != nil {
				return err
			}

			hostname := resource.Primary.Attributes["hostname"]

			// TODO: Add filter by hostname in wallarm-go and use here and resource
			nodes, err := client.NodeRead(clientID, "all")
			if err != nil {
				return err
			}

			for _, node := range nodes.Body {
				if node.Hostname == hostname {
					return fmt.Errorf("Resource still exists: %s", name)
				}
			}

			return nil
		}

		return nil
	}
}

func testAccCheckWallarmNodeExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(wallarm.API)

		resource, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		clientID, err := strconv.Atoi(resource.Primary.Attributes["client_id"])
		if err != nil {
			return err
		}

		hostname := resource.Primary.Attributes["hostname"]

		// TODO: Add filter by hostname in wallarm-go and use here and resource
		nodes, err := client.NodeRead(clientID, "all")
		if err != nil {
			return err
		}

		for _, node := range nodes.Body {
			if node.Hostname == hostname {
				return nil
			}
		}

		return fmt.Errorf("WallarmNode not found: %s", name)
	}
}
