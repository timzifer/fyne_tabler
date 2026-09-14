package tabler_icons

import (
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/muesli/gamut"
	"image/color"
)

type (
	StainableResource interface {
		fyne.Resource

		StainColor() color.Color

		Stain(color color.Color) (fyne.Resource, error)
		MustStain(color color.Color) fyne.Resource
	}

	stainedSvgResource struct {
		svg        []byte
		name       string
		stainColor color.Color
		cacheName  string
	}
)

func (s stainedSvgResource) StainColor() color.Color {
	return s.stainColor
}

func (s stainedSvgResource) MustStain(color color.Color) fyne.Resource {

	if stained, err := s.Stain(color); err != nil {
		panic(err)
	} else {
		return stained
	}
}

func (s stainedSvgResource) Name() string {
	return s.cacheName
}

func (s stainedSvgResource) Content() []byte {
	return s.svg
}

func (s stainedSvgResource) Stain(color color.Color) (fyne.Resource, error) {
	return Icon(s.name, color)
}

func NewStainedSVGResource(name string, source []byte, stainColor color.Color) StainableResource {
	return &stainedSvgResource{
		name:       name,
		cacheName:  fmt.Sprintf("%s-%s", name, gamut.ToHex(stainColor)),
		stainColor: stainColor,
		svg:        bytes.Replace(source, []byte("currentColor"), []byte(gamut.ToHex(stainColor)), -1),
	}
}
