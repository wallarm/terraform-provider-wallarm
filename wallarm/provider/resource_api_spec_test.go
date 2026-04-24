package wallarm

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	wallarm "github.com/wallarm/wallarm-go"
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
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecBasic(rnd, clientID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", "tf-test-"+rnd),
					resource.TestCheckResourceAttr(resourceName, "description", "Created by Terraform acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "file_remote_url", "https://raw.githubusercontent.com/concentrator/petstore/refs/heads/main/spec.yaml"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domains.0", "petstore-tf-test-"+rnd+".example.com"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "api_spec_id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints_count"),
					resource.TestCheckResourceAttrSet(resourceName, "openapi_version"),
					resource.TestCheckResourceAttrSet(resourceName, "spec_version"),
					resource.TestCheckResourceAttrSet(resourceName, "last_synced_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "file.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "file.0.checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "file.0.version"),
				),
			},
		},
	})
}

func testAccWallarmAPISpecBasic(resourceID, clientID string) string {
	return fmt.Sprintf(`
resource "wallarm_api_spec" %[1]q {
  client_id           = %[2]s
  title               = "tf-test-%[1]s"
  description         = "Created by Terraform acceptance test"
  file_remote_url     = "https://raw.githubusercontent.com/concentrator/petstore/refs/heads/main/spec.yaml"
  regular_file_update = false
  api_detection       = false
  domains             = ["petstore-tf-test-%[1]s.example.com"]
  instances           = [1]
}`, resourceID, clientID)
}

func TestAccWallarmAPISpec_Update(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec." + rnd
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	var firstID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecBasic(rnd, clientID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", "tf-test-"+rnd),
					func(s *terraform.State) error {
						firstID = s.RootModule().Resources[resourceName].Primary.Attributes["api_spec_id"]
						return nil
					},
				),
			},
			{
				Config: testAccWallarmAPISpecUpdated(rnd, clientID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", "tf-test-"+rnd+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					func(s *terraform.State) error {
						newID := s.RootModule().Resources[resourceName].Primary.Attributes["api_spec_id"]
						if newID != firstID {
							return fmt.Errorf("expected api_spec_id to remain %s across Update, got %s (destroy/recreate instead of in-place Update)", firstID, newID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccWallarmAPISpecUpdated(resourceID, clientID string) string {
	return fmt.Sprintf(`
resource "wallarm_api_spec" %[1]q {
  client_id           = %[2]s
  title               = "tf-test-%[1]s-updated"
  description         = "Updated description"
  file_remote_url     = "https://raw.githubusercontent.com/concentrator/petstore/refs/heads/main/spec.yaml"
  regular_file_update = false
  api_detection       = false
  domains             = ["petstore-tf-test-%[1]s.example.com"]
  instances           = [1]
}`, resourceID, clientID)
}

func TestAccWallarmAPISpec_Import(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec." + rnd
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecDestroy,
		Steps: []resource.TestStep{
			{Config: testAccWallarmAPISpecBasic(rnd, clientID)},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"file.0.signed_url"},
			},
		},
	})
}

func TestAccWallarmAPISpec_AuthHeaders(t *testing.T) {
	rnd := generateRandomResourceName(5)
	resourceName := "wallarm_api_spec." + rnd
	clientID := os.Getenv("WALLARM_API_CLIENT_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckAPISpec(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWallarmAPISpecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmAPISpecWithAuthHeader(rnd, clientID, "X-Token", "first-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auth_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_headers.0.key", "X-Token"),
					resource.TestCheckResourceAttr(resourceName, "auth_headers.0.value", "first-value"),
				),
			},
			{
				Config: testAccWallarmAPISpecWithAuthHeader(rnd, clientID, "X-Token", "second-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auth_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_headers.0.key", "X-Token"),
					resource.TestCheckResourceAttr(resourceName, "auth_headers.0.value", "second-value"),
				),
			},
			{
				Config: testAccWallarmAPISpecBasic(rnd, clientID),
				Check:  resource.TestCheckResourceAttr(resourceName, "auth_headers.#", "0"),
			},
		},
	})
}

func testAccWallarmAPISpecWithAuthHeader(resourceID, clientID, key, value string) string {
	return fmt.Sprintf(`
resource "wallarm_api_spec" %[1]q {
  client_id           = %[2]s
  title               = "tf-test-%[1]s"
  description         = "Auth headers test"
  file_remote_url     = "https://raw.githubusercontent.com/concentrator/petstore/refs/heads/main/spec.yaml"
  regular_file_update = false
  api_detection       = false
  domains             = ["petstore-tf-test-%[1]s.example.com"]
  instances           = [1]
  auth_headers {
    key   = %[3]q
    value = %[4]q
  }
}`, resourceID, clientID, key, value)
}

func testAccCheckWallarmAPISpecDestroy(s *terraform.State) error {
	api, err := testAccNewAPIClient()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "wallarm_api_spec" {
			continue
		}
		clientID, err := strconv.Atoi(rs.Primary.Attributes["client_id"])
		if err != nil {
			return fmt.Errorf("invalid client_id for %s: %w", rs.Primary.ID, err)
		}
		apiSpecID, err := strconv.Atoi(rs.Primary.Attributes["api_spec_id"])
		if err != nil {
			return fmt.Errorf("invalid api_spec_id for %s: %w", rs.Primary.ID, err)
		}
		_, err = api.APISpecReadByID(clientID, apiSpecID)
		if err == nil {
			return fmt.Errorf("API Spec %d still exists", apiSpecID)
		}
		if !errors.Is(err, wallarm.ErrNotFound) {
			return fmt.Errorf("checking api_spec %d: %w", apiSpecID, err)
		}
	}
	return nil
}
