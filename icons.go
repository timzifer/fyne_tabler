// Package fyne_tabler provides Tabler Icons (https://tabler.io/icons) for Fyne apps.
//
// Icons are monochrome and stained before use: without an explicit color the
// current theme foreground color is applied. The generated Icon* constants
// list all available names.
package fyne_tabler

import (
	"embed"
	"image/color"

	"fyne.io/fyne/v2"
	"github.com/timzifer/fyne_iconkit"
)

//go:generate go run github.com/timzifer/fyne_iconkit/cmd/iconconst -dir icons -prefix Icon -out icons_gen.go

// StainableResource is the resource type returned by Icon and MustIcon.
type StainableResource = fyne_iconkit.StainableResource

var (
	//go:embed icons
	icons embed.FS

	set = fyne_iconkit.NewStainableSet(icons, "icons")
)

// Icon returns the named icon as StainableResource, stained with stainColor
// or the theme foreground color. Unknown names yield an error wrapping
// fs.ErrNotExist.
func Icon(name string, stainColor ...color.Color) (fyne.Resource, error) {
	return set.Icon(name, stainColor...)
}

// MustIcon is like Icon but logs the error via fyne.LogError and returns
// theme.ErrorIcon() for unknown names.
func MustIcon(name string, stainColor ...color.Color) fyne.Resource {
	return set.MustIcon(name, stainColor...)
}

// Source returns the unmodified SVG content of the named icon.
func Source(name string) ([]byte, error) {
	return set.Source(name)
}

// StainedSource returns the SVG content of the named icon stained with c.
func StainedSource(name string, c color.Color) ([]byte, error) {
	return set.StainedSource(name, c)
}

// PNG rasterizes the named icon to a size×size PNG, stained with stainColor
// or black.
func PNG(name string, size int, stainColor ...color.Color) ([]byte, error) {
	return set.PNG(name, size, stainColor...)
}

// Names returns the names of all available icons.
func Names() []string {
	return set.Names()
}
