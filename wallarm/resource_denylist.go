package wallarm

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	wallarm "github.com/wallarm/wallarm-go"
)

func resourceWallarmDenylist() *schema.Resource {
	return resourceWallarmIPList(wallarm.DenylistType)
}
