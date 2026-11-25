# fyne_tabler

[![CI](https://github.com/timzifer/fyne_tabler/actions/workflows/ci.yml/badge.svg)](https://github.com/timzifer/fyne_tabler/actions/workflows/ci.yml)

[Tabler Icons](https://tabler.io/icons) (v3.53.0) for [Fyne](https://fyne.io) apps, embedded into your
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

`Icon` serves the strokes as filled outlines, so Fyne can recolor icons in
disabled buttons and menu items, important buttons, selected tabs and
`theme.NewThemedResource`/`NewColoredResource`/`NewDisabledResource`, and
strokes scale with the icon size. `Source` and `StainedSource` return the
original stroked SVG.

Filled icons are named `<name>-filled`. `fyne_tabler.TablerVersion` holds the
embedded Tabler Icons release.

Built on [fyne_iconkit](https://github.com/timzifer/fyne_iconkit).

## Renamed icons

When Tabler renames an icon, the old `Icon*` variable stays available as a
deprecated alias of the new one (and `Lookup` accepts the old name). Like
Tabler's own framework packages, aliases are kept until the next major
release of Tabler Icons (aliases from Tabler 3.x are removed with Tabler 4).
See [CHANGELOG.md](CHANGELOG.md) for renamed, removed, added and changed
icons per release.

## Versioning

Releases are tagged `v1.<minor>.<patch>`; the embedded Tabler Icons release
is named in the release title, in [CHANGELOG.md](CHANGELOG.md) and in
`fyne_tabler.TablerVersion`. Every Tabler update merged into `main` is
released automatically as the next minor version, with its changelog section
as release notes ([release workflow](.github/workflows/release.yml)).

> [!IMPORTANT]
> Unlike strict semantic versioning, **minor releases may remove `Icon*`
> variables**: deprecated aliases once Tabler Icons has a new major release,
> and icons that Tabler drops without a successor. Every removal is announced
> in CHANGELOG.md, and deprecated aliases are marked `// Deprecated:`
> (reported by `staticcheck` and gopls) for the whole Tabler major. Replace
> deprecated identifiers when you see them, or pin the version in your
> `go.mod` and read the changelog before upgrading.

## Updating the icons

A [daily workflow](.github/workflows/update-icons.yml) opens a pull request
for every new Tabler Icons release, with the release notes as description.

Locally, `go generate ./...` does the same: it downloads the latest release
from npm (`@tabler/icons`), replaces the SVGs, license and the version above,
updates the aliases and `CHANGELOG.md`, and regenerates the `Icon*` variables.

```sh
go generate ./...                                    # latest release
go run ./internal/tablerupdate -version 3.52.0 && \
  go generate -run iconconst ./...                   # specific release
go run ./internal/tablerupdate -offline && \
  go generate -run iconconst ./...                   # regenerate without downloading
```

Renames are taken from Tabler's `aliases.json`, from names (`cart-x` →
`shopping-cart-x`) and from identical shapes. [`tabler.json`](tabler.json)
records the imported releases and aliases; there you can

- add an alias by hand (`"source": "manual"`, `"since"` = current version),
  then run with `-offline`,
- add a hint to `renames` (`"old": "new"`) that is used when a future release
  removes `old`.

## License

Code: [MIT](LICENSE). Icons: MIT, © Paweł Kuna – see [LICENSE-tabler-icons](LICENSE-tabler-icons).
