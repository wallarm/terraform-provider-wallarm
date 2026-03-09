package version

import (
	"runtime/debug"
)

var (
	// ProviderVersion is set during the release process to the release version of the binary
	ProviderVersion = "2.0.0"
)

func SDK(ver int) string {
	var sdkVersion string
	var suffix string
	if ver == 2 {
		suffix = "/v2"
	} else {
		suffix = ""
	}
	var depPath string = "github.com/hashicorp/terraform-plugin-sdk" + suffix

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == depPath {
				sdkVersion = dep.Version
				break
			}
		}
	}
	return sdkVersion
}

func Provider() string {
	var providerVersion string

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		providerVersion = buildInfo.Main.Version
	}

	return providerVersion
}
