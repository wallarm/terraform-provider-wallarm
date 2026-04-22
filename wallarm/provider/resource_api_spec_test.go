package wallarm

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccPreCheckAPISpec(t *testing.T) {
	testAccPreCheck(t)
	v := os.Getenv("WALLARM_API_CLIENT_ID")
	if v == "" {
		t.Skip("WALLARM_API_CLIENT_ID must be set for wallarm_api_spec acceptance tests")
	}
	if _, err := strconv.Atoi(v); err != nil {
		t.Fatalf("WALLARM_API_CLIENT_ID must be an integer, got %q", v)
	}
}

func TestAccWallarmAPISpec_basic(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckAPISpec(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWallarmAPISpecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecBasic(rnd),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(resourceName, "description", "Created by Terraform acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "file_remote_url", "https://raw.githubusercontent.com/concentrator/petstore/refs/heads/main/spec.yaml"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domains.0", "petstore-tf-test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "api_spec_id"),
				),
			},
		},
	})
}

func testAccWallarmAPISpecBasic(resourceID string) string {
	return fmt.Sprintf(`
resource "wallarm_api_spec" "%[1]s" {
  client_id           = %[2]s
  title               = "tf-test-%[1]s"
  description         = "Created by Terraform acceptance test"
  file_remote_url     = "https://raw.githubusercontent.com/concentrator/petstore/refs/heads/main/spec.yaml"
  regular_file_update = false
  api_detection       = false
  domains             = ["petstore-tf-test.example.com"]
  instances           = [1]
}`, resourceID, os.Getenv("WALLARM_API_CLIENT_ID"))
}

func testAccCheckWallarmAPISpecDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_api_spec" {
			continue
		}

		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return err
		}
		apiSpecID, err := strconv.Atoi(rs.Primary.Attributes["api_spec_id"])
		if err != nil {
			return err
		}

		_, err = client.APISpecReadByID(clientID, apiSpecID)
		if err == nil {
			return fmt.Errorf("API Spec %d still exists", apiSpecID)
		}
	}

	return nil
}
