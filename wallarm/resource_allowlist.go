package wallarm

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/wallarm/wallarm-go"
)

func resourceWallarmAllowlist() *schema.Resource {
	return resourceWallarmIPList(wallarm.AllowlistType)
}
