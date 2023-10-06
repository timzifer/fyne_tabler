package tabler_icons

import (
	"bytes"
	"embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/muesli/gamut"
	"github.com/rs/zerolog"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"image/color"
	"io"
	"sync"
)

var (
	mutex  = sync.Mutex{}
	logger = zerolog.Nop()
	cache  = map[string]fyne.Resource{}

	//go:embed "icons"
	resources embed.FS
)

func RegisterLogger(l zerolog.Logger) {
	logger = l
}

func cacheIdentifier(name string, stainColor color.Color) string {
	return fmt.Sprintf("%s(%s)", name, gamut.ToHex(stainColor))
}

func Source(name string) ([]byte, error) {

	if f, err := resources.Open("icons/" + name + ".svg"); err != nil {
		return nil, err
	} else if icon, readErr := io.ReadAll(f); readErr != nil {
		return nil, readErr
	} else {
		return icon, nil
	}
}

func StainedSource(name string, stainColor color.Color) ([]byte, error) {
	if source, err := Source(name); err != nil {
		return nil, err
	} else {
		return bytes.Replace(source, []byte("currentColor"), []byte(gamut.ToHex(stainColor)), -1), nil
	}
}

func Icon(name string, stainColor ...color.Color) (fyne.Resource, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if len(stainColor) == 0 {
		stainColor = []color.Color{theme.ForegroundColor()}
	}

	cacheId := cacheIdentifier(name, stainColor[0])
	if i, exists := cache[cacheId]; exists {
		return i, nil
	}
	if icon, err := Source(name); err != nil {
		return nil, err
	} else {
		// tabler icons must be stained before use
		cache[cacheId] = NewStainedSVGResource(name, icon, stainColor[0])

		logger.Debug().Str("name", name).Str("stain", gamut.ToHex(stainColor[0])).Msg("tabler-icon has been loaded + cached")

		return cache[cacheId], nil
	}

}

func PNG(name string, stainColor ...color.Color) ([]byte, error) {
	var (
		source []byte
		err    error
	)
	if len(stainColor) > 0 {
		source, err = StainedSource(name, stainColor[0])
	} else {
		source, err = Source(name)
	}
	if err != nil {
		return nil, err
	}

	if svg, parseErr := canvas.ParseSVG(bytes.NewBuffer(source)); parseErr != nil {
		return nil, parseErr
	} else {
		var pngBuffer bytes.Buffer
		if rasterErr := renderers.PNG()(&pngBuffer, svg); rasterErr != nil {
			return nil, rasterErr
		} else {
			return pngBuffer.Bytes(), nil
		}
	}
}

func MustIcon(name string, stainColor ...color.Color) fyne.Resource {
	if icon, err := Icon(name, stainColor...); err != nil {
		logger.Warn().Err(err).Str("name", name).Msg("could not find tabler-icon")

		return theme.ErrorIcon()
	} else {
		return icon
	}
}
