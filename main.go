package main

import (
	"github.com/416e64726579/terraform-provider-wallarm/wallarm"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: wallarm.Provider})
}
