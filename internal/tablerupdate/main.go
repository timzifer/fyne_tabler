// Command tablerupdate imports a release of Tabler Icons from the npm registry
// (package @tabler/icons) into this package, using fyne_iconkit/iconupdate.
//
// Outline icons keep their name, filled icons get a "-filled" suffix
// (icons/outline/home.svg → home.svg, icons/filled/home.svg → home-filled.svg),
// matching the flat naming of Tabler Icons 2.x. Renames published in Tabler's
// aliases.json become deprecated aliases.
//
// It runs as part of go generate, before iconconst regenerates the Icon*
// variables:
//
//	//go:generate go run ./internal/tablerupdate
package main

import "github.com/timzifer/fyne_iconkit/iconupdate"

func main() {
	iconupdate.Main(iconupdate.Source{
		Name:    "Tabler Icons",
		URL:     "https://tabler.io/icons",
		Package: "@tabler/icons",
		Dirs: map[string]string{
			"package/icons/outline/": "",
			"package/icons/filled/":  filledSuffix,
		},
		License: "package/LICENSE",
		CompareURL: func(from, to string) string {
			return "https://github.com/tabler/tabler-icons/compare/v" + from + "...v" + to
		},
		Aliases: fetchAliases,

		Command:      "go run ./internal/tablerupdate",
		GoPackage:    "fyne_tabler",
		VersionConst: "TablerVersion",
		StateFile:    "tabler.json",
		Out:          "tabler_gen.go",
		LicenseFile:  "LICENSE-tabler-icons",
	})
}
