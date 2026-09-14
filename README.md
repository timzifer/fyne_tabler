# fyne_tabler

[![CI](https://github.com/timzifer/fyne_tabler/actions/workflows/ci.yml/badge.svg)](https://github.com/timzifer/fyne_tabler/actions/workflows/ci.yml)

[Tabler Icons](https://tabler.io/icons) (v2.30.0) for [Fyne](https://fyne.io) apps, embedded into your
binary and colorable at runtime. Requires Go 1.21+ and Fyne 2.3+.

## Install

```sh
go get github.com/timzifer/fyne_tabler
```

## Usage

```go
import (
	"image/color"

	"fyne.io/fyne/v2/widget"

	"github.com/timzifer/fyne_tabler"
)

// stained with the current theme foreground color
icon := widget.NewIcon(fyne_tabler.Icon(fyne_tabler.IconHome))

// stained with a custom color (alpha is respected)
red := fyne_tabler.Icon(fyne_tabler.IconAlertTriangle, color.NRGBA{R: 0xd0, A: 0xff})

// re-color an existing icon
blue := red.(fyne_tabler.StainableResource).MustStain(color.NRGBA{B: 0xd0, A: 0xff})

// raw / stained SVG bytes, or a rasterized 64×64 PNG
svg := fyne_tabler.StainedSource(fyne_tabler.IconHome, color.Black)
png, _ := fyne_tabler.PNG(fyne_tabler.IconHome, 64, color.Black)

// names from configuration (as listed on https://tabler.io/icons)
if i, ok := fyne_tabler.Lookup("arrow-up"); ok {
	icon.SetResource(fyne_tabler.Icon(i))
}
```

Icons are values of an unexported type that only the generated `Icon*`
variables and `Lookup` produce – arbitrary strings do not compile, so `Icon`
cannot fail. Since the type is unexported, store `fyne.Resource`s or names
(`i.String()`) rather than icon values in your own structs.

The default stain color is taken when the icon is requested; request the icon
again (or call `MustStain`) after a theme change.

Built on [fyne_iconkit](https://github.com/timzifer/fyne_iconkit). After
updating the SVGs, run `go generate ./...` to refresh the `Icon*` variables.

## License

Code: [MIT](LICENSE). Icons: MIT, © Paweł Kuna – see [LICENSE-tabler-icons](LICENSE-tabler-icons).
