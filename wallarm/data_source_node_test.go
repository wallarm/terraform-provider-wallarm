package wallarm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccWallarmDataNodeDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmDataNodeDefault,
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmDataNode("data.wallarm_node.waf"),
				),
			},
		},
	})
}

func TestAccWallarmDataNodeFilterType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmDataNodeFilterType,
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmDataNode("data.wallarm_node.waf"),
				),
			},
		},
	})
}

// Only for regular nodes
/*
func TestAccWallarmDataNodeFilterUUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmDataNodeFilterUUID,
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmDataNode("data.wallarm_node.waf"),
				),
			},
		},
	})
}

func TestAccWallarmDataNodeFilterHostname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWallarmDataNodeFilterHostname,
				Check: resource.ComposeTestCheckFunc(
					testAccWallarmDataNode("data.wallarm_node.waf"),
				),
			},
		},
	})
}
*/

func testAccWallarmDataNode(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var (
			nodeCount int
			err       error
		)

		rs := s.RootModule().Resources[n]
		a := rs.Primary.Attributes

		if rs.Primary.ID == "" {
			return fmt.Errorf("couldn't fetch wallarm nodes from the API")
		}

		if nodeCount, err = strconv.Atoi(a["nodes.#"]); err != nil {
			return err
		}

		if nodeCount == 0 {
			return fmt.Errorf(`
			no nodes in the account by applied filter:
			filter {
				type = "%[1]s"
				uuid = "%[2]s"
				hostname = "%[3]s"
			}`, a["filter.0.type"], a["filter.0.uuid"], a["filter.0.hostname"])
		}

		if filterType, ok := a["filter.type"]; ok {
			if filterType != a["nodes.0.type"] {
				return fmt.Errorf("type %[1]s doesn't correspond to the filter: %[2]s", filterType, a["nodes.0.type"])
			}
		}

		return nil
	}
}

const testAccWallarmDataNodeDefault = `
data "wallarm_node" "waf" {}
`

const testAccWallarmDataNodeFilterType = `
data "wallarm_node" "waf" {
	filter {
		type = "cloud_node"
	}
}
`

// This cannot be applied to nodes of type cloud_node
const testAccWallarmDataNodeFilterUUID = `
data "wallarm_node" "waf" {
	filter {
		uuid = "13ef5f5e-01ca-4db1-85ac-8d969d688f25"
	}
}
`

// This cannot be applied to nodes of type cloud_node
const testAccWallarmDataNodeFilterHostname = `
data "wallarm_node" "waf" {
	filter {
		hostname = "ed9763c775f5"
	}
}
`
