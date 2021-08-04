package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/wallarm/terraform-provider-wallarm/wallarm"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: wallarm.Provider})
}
