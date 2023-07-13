package tabler_icons

import (
	"embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/muesli/gamut"
	"github.com/rs/zerolog"
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
	if f, err := resources.Open("icons/" + name + ".svg"); err != nil {
		return nil, err
	} else if icon, readErr := io.ReadAll(f); readErr != nil {
		return nil, readErr
	} else {
		// tabler icons must be stained before use
		cache[cacheId] = NewStainedSVGResource(name, icon, stainColor[0])

		logger.Debug().Str("name", name).Str("stain", gamut.ToHex(stainColor[0])).Msg("tabler-icon has been loaded + cached")

		return cache[cacheId], nil
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
