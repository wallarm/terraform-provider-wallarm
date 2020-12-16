package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccWallarmVulnDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmVulnDefault,
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmVuln("data.wallarm_vuln.vulns"),
				),
			},
		},
	})
}

func TestAccWallarmVulnFilterStatus(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmVulnFilterStatus("open"),
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmVuln("data.wallarm_vuln.vulns"),
				),
			},
			{
				Config: testAccWallarmVulnFilterStatus("falsepositive"),
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmVuln("data.wallarm_vuln.vulns"),
				),
			},
			{
				Config: testAccWallarmVulnFilterStatus("closed"),
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmVuln("data.wallarm_vuln.vulns"),
				),
			},
		},
	})
}

func TestAccWallarmVulnFilterLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmVulnFilterLimit,
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmVuln("data.wallarm_vuln.vulns"),
				),
			},
		},
	})
}

func TestAccWallarmVulnFilterOffset(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmVulnFilterOffset,
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmVuln("data.wallarm_vuln.vulns"),
				),
			},
		},
	})
}

func testAccWallarmVuln(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var (
			limit     int = 1000
			vulnCount int
			err       error
		)

		rs := s.RootModule().Resources[n]
		a := rs.Primary.Attributes

		if rs.Primary.ID == "" {
			return fmt.Errorf("couldn't fetch vulnerabilities from the API")
		}

		if vulnCount, err = strconv.Atoi(a["vuln.#"]); err != nil {
			return err
		}

		if a["filter.0.limit"] != "" {
			if limit, err = strconv.Atoi(a["filter.0.limit"]); err != nil {
				return err
			}
		}

		if vulnCount > limit {
			return fmt.Errorf(`the API returned more vulnerabilites %d than requested %d`, vulnCount, limit)
		}

		return nil
	}
}

const testAccWallarmVulnDefault = `
data "wallarm_vuln" "vulns" {}
`

func testAccWallarmVulnFilterStatus(status string) string {
	return fmt.Sprintf(`
data "wallarm_vuln" "vulns" {
	filter {
		status = "%s"
	}
}`, status)
}

const testAccWallarmVulnFilterLimit = `
data "wallarm_vuln" "vulns" {
	filter {
		limit = 200
	}
}
`

const testAccWallarmVulnFilterOffset = `
data "wallarm_vuln" "vulns" {
	filter {
		offset = 100
	}
}
`

const testAccWallarmVulnFilterFullConfig = `
data "wallarm_vuln" "vulns" {
	filter {
		status = "open"
		limit = 100
		offset = 0
	}
}
`
