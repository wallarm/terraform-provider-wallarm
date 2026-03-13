package wallarm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceHits(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceHitsConfig("nonexistent-request-id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceHitsExists("data.wallarm_hits.test"),
					resource.TestCheckResourceAttr("data.wallarm_hits.test", "hits.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceHitsAttackMode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceHitsAttackModeConfig("nonexistent-request-id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceHitsExists("data.wallarm_hits.test_attack"),
					resource.TestCheckResourceAttr("data.wallarm_hits.test_attack", "hits.#", "0"),
					resource.TestCheckResourceAttr("data.wallarm_hits.test_attack", "mode", "attack"),
				),
			},
		},
	})
}

func testAccCheckDataSourceHitsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		return nil
	}
}

func testAccDataSourceHitsConfig(requestID string) string {
	return fmt.Sprintf(`data "wallarm_hits" "test" {
  request_id = "%s"
}`, requestID)
}

func testAccDataSourceHitsAttackModeConfig(requestID string) string {
	return fmt.Sprintf(`data "wallarm_hits" "test_attack" {
  request_id = "%s"
  mode       = "attack"
}`, requestID)
}
