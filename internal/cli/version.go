package cli

import "runtime/debug"

// Set at build time by GoReleaser.
var buildVersion string

func version() string {
	if buildVersion != "" {
		return buildVersion
	}

	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" &&
		info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}
