// Package fyne_tabler provides Tabler Icons (https://tabler.io/icons) for Fyne apps.
//
// Icons are addressed by the generated Icon* variables (or Lookup for names
// from configuration). They are monochrome and stained before use: without
// an explicit color the current theme foreground color is applied.
package fyne_tabler

import (
	"embed"
	"image/color"

	"fyne.io/fyne/v2"
	"github.com/timzifer/fyne_iconkit"
)

//go:generate go run github.com/timzifer/fyne_iconkit/cmd/iconconst -dir icons -type icon -prefix Icon -out icons_gen.go

// icon identifies one embedded icon. It is unexported on purpose: values only
// come from the generated Icon* variables or Lookup, so every icon exists.
type icon struct{ name string }

// String returns the icon name, e.g. "arrow-up". Lookup accepts it again.
func (i icon) String() string { return i.name }

// StainableResource is the resource type returned by Icon.
type StainableResource = fyne_iconkit.StainableResource

var (
	//go:embed icons
	icons embed.FS

	set = fyne_iconkit.NewStainableSet(icons, "icons")
)

// Lookup returns the icon with the given name (as listed on https://tabler.io/icons),
// e.g. for names read from configuration.
func Lookup(name string) (icon, bool) {
	if !set.Has(name) {
		return icon{}, false
	}
	return icon{name}, true
}

// All returns every icon of this package, sorted by name.
func All() []icon {
	names := set.Names()
	all := make([]icon, len(names))
	for n, name := range names {
		all[n] = icon{name}
	}
	return all
}

// Icon returns i as StainableResource, stained with stainColor or the theme
// foreground color. Resources are cached per icon and color.
func Icon(i icon, stainColor ...color.Color) fyne.Resource {
	return set.MustIcon(i.name, stainColor...)
}

// Source returns the unmodified SVG content of i.
func Source(i icon) []byte {
	src, _ := set.Source(i.name)
	return src
}

// StainedSource returns the SVG content of i stained with c.
func StainedSource(i icon, c color.Color) []byte {
	src, _ := set.StainedSource(i.name, c)
	return src
}

// PNG rasterizes i to a size×size PNG, stained with stainColor or black.
func PNG(i icon, size int, stainColor ...color.Color) ([]byte, error) {
	return set.PNG(i.name, size, stainColor...)
}
