package wallarm

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"
)

func resourceWallarmGraylist() *schema.Resource {
	return resourceWallarmIPList(wallarm.GraylistType)
}
